/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schedulingmanager

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/kubernetes-sigs/federation-v2/pkg/apis/core/typeconfig"
	corev1a1 "github.com/kubernetes-sigs/federation-v2/pkg/apis/core/v1alpha1"
	fedclientset "github.com/kubernetes-sigs/federation-v2/pkg/client/clientset/versioned"
	corev1alpha1client "github.com/kubernetes-sigs/federation-v2/pkg/client/clientset/versioned/typed/core/v1alpha1"
	"github.com/kubernetes-sigs/federation-v2/pkg/controller/schedulingpreference"
	"github.com/kubernetes-sigs/federation-v2/pkg/controller/util"
	"github.com/kubernetes-sigs/federation-v2/pkg/schedulingtypes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type SchedulerController struct {
	// Store for the FederatedTypeConfig objects
	store cache.Store
	// Informer for the FederatedTypeConfig objects
	controller cache.Controller

	worker util.ReconcileWorker

	scheduler map[string]schedulingtypes.Scheduler

	config *util.ControllerConfig

	runningPlugins sets.String
	// mapping qualifiedname to template kind for managing plugins in scheduler
	templateKindMap map[string]string
}

func StartSchedulerController(config *util.ControllerConfig, stopChan <-chan struct{}) {

	userAgent := "SchedulerController"
	kubeConfig := config.KubeConfig
	restclient.AddUserAgent(kubeConfig, userAgent)
	client := fedclientset.NewForConfigOrDie(kubeConfig).CoreV1alpha1()

	controller := newController(config, client)

	glog.Infof("Starting scheduler controller")
	controller.Run(stopChan)
}

func newController(config *util.ControllerConfig, client corev1alpha1client.CoreV1alpha1Interface) *SchedulerController {
	c := &SchedulerController{
		config:          config,
		scheduler:       make(map[string]schedulingtypes.Scheduler),
		runningPlugins:  sets.String{},
		templateKindMap: make(map[string]string),
	}

	fedNamespace := config.FederationNamespace
	c.worker = util.NewReconcileWorker(c.reconcile, util.WorkerTiming{})

	c.store, c.controller = cache.NewInformer(
		&cache.ListWatch{
			// Only watch the federation namespace to ensure
			// restrictive authz can be applied to a namespaced
			// control plane.
			ListFunc: func(options metav1.ListOptions) (pkgruntime.Object, error) {
				return client.FederatedTypeConfigs(fedNamespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return client.FederatedTypeConfigs(fedNamespace).Watch(options)
			},
		},
		&corev1a1.FederatedTypeConfig{},
		util.NoResyncPeriod,
		util.NewTriggerOnAllChanges(c.worker.EnqueueObject),
	)

	return c
}

// Run runs the Controller.
func (c *SchedulerController) Run(stopChan <-chan struct{}) {
	go c.controller.Run(stopChan)

	// wait for the caches to synchronize before starting the worker
	if !cache.WaitForCacheSync(stopChan, c.controller.HasSynced) {
		runtime.HandleError(errors.New("Timed out waiting for cache to sync"))
		return
	}

	c.worker.Run(stopChan)
}

func (c *SchedulerController) reconcile(qualifiedName util.QualifiedName) util.ReconciliationStatus {
	key := qualifiedName.String()

	glog.V(3).Infof("Running reconcile FederatedTypeConfig for %q", key)

	schedulingType := schedulingtypes.GetSchedulingType(qualifiedName.Name)
	if schedulingType == nil {
		// No scheduler supported for this resource
		return util.StatusAllOK
	}

	cachedObj, exist, err := c.store.GetByKey(key)
	if err != nil {
		runtime.HandleError(errors.Wrapf(err, "Failed to query FederatedTypeConfig store for %q", key))
		return util.StatusError
	}

	if !exist {
		c.stopScheduler(schedulingType.Kind, qualifiedName)
		return util.StatusAllOK
	}

	typeConfig := cachedObj.(*corev1a1.FederatedTypeConfig)
	if typeConfig.Spec.PropagationEnabled == false || typeConfig.DeletionTimestamp != nil {
		c.stopScheduler(schedulingType.Kind, qualifiedName)
		return util.StatusAllOK
	}

	if c.runningPlugins.Has(qualifiedName.Name) {
		// Scheduler and plugin are already running
		return util.StatusAllOK
	}

	// set name and group for the type config target
	corev1a1.SetFederatedTypeConfigDefaults(typeConfig)

	// TODO(marun) Replace with validation webhook
	err = typeconfig.CheckTypeConfigName(typeConfig)
	if err != nil {
		runtime.HandleError(err)
		return util.StatusError
	}

	// Scheduling preference controller is started on demand
	schedulingKind := schedulingType.Kind
	scheduler, ok := c.scheduler[schedulingKind]
	if !ok {
		var err error

		scheduler, err = schedulingpreference.StartSchedulingPreferenceController(c.config, *schedulingType)
		if err != nil {
			runtime.HandleError(errors.Wrapf(err, "Error starting schedulingpreference controller for %q", schedulingKind))
			return util.StatusError
		}
		c.scheduler[schedulingKind] = scheduler
	}

	templateKind := typeConfig.GetTemplate().Kind
	glog.Infof("Starting plugin kind %s for scheduling type %s", templateKind, schedulingKind)

	err = scheduler.StartPlugin(typeConfig)
	if err != nil {
		runtime.HandleError(errors.Wrapf(err, "Error starting plugin for %q", templateKind))
		return util.StatusError
	}
	c.runningPlugins.Insert(qualifiedName.Name)
	c.templateKindMap[qualifiedName.Name] = templateKind

	return util.StatusAllOK
}

func (c *SchedulerController) stopScheduler(schedulingKind string, qualifiedName util.QualifiedName) {
	scheduler, ok := c.scheduler[schedulingKind]
	if !ok {
		return
	}

	if !c.runningPlugins.Has(qualifiedName.Name) {
		return
	}

	glog.Infof("Stopping plugin for %q with kind %q", qualifiedName.Name, c.templateKindMap[qualifiedName.Name])

	scheduler.StopPlugin(c.templateKindMap[qualifiedName.Name])
	c.runningPlugins.Delete(qualifiedName.Name)
	delete(c.templateKindMap, qualifiedName.Name)

	// if all resources registered to same scheduler are deleted, the scheduler should be stopped
	resources := schedulingtypes.GetSchedulingKinds(qualifiedName.Name)
	result := c.runningPlugins.Intersection(resources)
	if result.Len() == 0 {
		glog.Infof("Stopping scheduler schedulingpreference controller for %q", schedulingKind)
		scheduler.Stop()

		delete(c.scheduler, schedulingKind)
	}
}
