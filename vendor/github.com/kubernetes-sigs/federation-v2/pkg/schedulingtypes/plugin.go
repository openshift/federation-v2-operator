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

package schedulingtypes

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/federation-v2/pkg/apis/core/typeconfig"
	"github.com/kubernetes-sigs/federation-v2/pkg/controller/util"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

const (
	replicasPath = "spec.replicas"
)

type Plugin struct {
	targetInformer util.FederatedInformer

	templateStore cache.Store
	// Informer for the templates of the federated type
	templateController cache.Controller

	// Store for the override directives of the federated type
	overrideStore cache.Store
	// Informer controller for override directives of the federated type
	overrideController cache.Controller
	// Dynamic client for override type
	overrideClient util.ResourceClient

	// Store for the placements of the federated type
	placementStore cache.Store
	// Informer controller for placements of the federated type
	placementController cache.Controller
	// Dynamic client for placement type
	placementClient util.ResourceClient

	typeConfig typeconfig.Interface

	stopChannel chan struct{}
}

func NewPlugin(controllerConfig *util.ControllerConfig, eventHandlers SchedulerEventHandlers, typeConfig typeconfig.Interface) (*Plugin, error) {
	targetAPIResource := typeConfig.GetTarget()
	fedClient, kubeClient, crClient := controllerConfig.AllClients(fmt.Sprintf("%s-replica-scheduler", strings.ToLower(targetAPIResource.Kind)))
	p := &Plugin{
		targetInformer: util.NewFederatedInformer(
			fedClient,
			kubeClient,
			crClient,
			controllerConfig.FederationNamespaces,
			&targetAPIResource,
			eventHandlers.ClusterEventHandler,
			eventHandlers.ClusterLifecycleHandlers,
		),
		typeConfig:  typeConfig,
		stopChannel: make(chan struct{}),
	}

	targetNamespace := controllerConfig.TargetNamespace
	federationEventHandler := eventHandlers.FederationEventHandler

	templateAPIResource := typeConfig.GetTemplate()
	templateClient, err := util.NewResourceClient(controllerConfig.KubeConfig, &templateAPIResource)
	if err != nil {
		return nil, err
	}
	p.templateStore, p.templateController = util.NewResourceInformer(templateClient, targetNamespace, federationEventHandler)

	placementAPIResource := typeConfig.GetPlacement()
	p.placementClient, err = util.NewResourceClient(controllerConfig.KubeConfig, &placementAPIResource)
	if err != nil {
		return nil, err
	}
	p.placementStore, p.placementController = util.NewResourceInformer(p.placementClient, targetNamespace, federationEventHandler)

	overrideAPIResource := typeConfig.GetOverride()
	p.overrideClient, err = util.NewResourceClient(controllerConfig.KubeConfig, &overrideAPIResource)
	if err != nil {
		return nil, err
	}
	p.overrideStore, p.overrideController = util.NewResourceInformer(p.overrideClient, targetNamespace, federationEventHandler)

	return p, nil
}

func (p *Plugin) Start() {
	p.targetInformer.Start()

	go p.templateController.Run(p.stopChannel)
	go p.overrideController.Run(p.stopChannel)
	go p.placementController.Run(p.stopChannel)
}

func (p *Plugin) Stop() {
	p.targetInformer.Stop()
	close(p.stopChannel)
}

func (p *Plugin) HasSynced() bool {
	if !p.targetInformer.ClustersSynced() {
		glog.V(2).Infof("Cluster list not synced")
		return false
	}

	if !p.templateController.HasSynced() {
		return false
	}
	if !p.placementController.HasSynced() {
		return false
	}
	if !p.overrideController.HasSynced() {
		return false
	}

	clusters, err := p.targetInformer.GetReadyClusters()
	if err != nil {
		runtime.HandleError(errors.Wrap(err, "Failed to get ready clusters"))
		return false
	}

	if !p.targetInformer.GetTargetStore().ClustersSynced(clusters) {
		return false
	}

	return true
}

func (p *Plugin) TemplateExists(key string) bool {
	_, exist, err := p.templateStore.GetByKey(key)
	if err != nil {
		glog.Errorf("Failed to query store while reconciling RSP controller for key %q: %v", key, err)
		wrappedErr := errors.Wrapf(err, "Failed to query store while reconciling RSP controller for key %q", key)
		runtime.HandleError(wrappedErr)
		return false
	}
	return exist
}

func (p *Plugin) ReconcilePlacement(qualifiedName util.QualifiedName, newClusterNames []string) error {
	placement, err := p.placementClient.Resources(qualifiedName.Namespace).Get(qualifiedName.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		newPlacement := newUnstructured(p.typeConfig.GetPlacement(), qualifiedName)
		setPlacementSpec(newPlacement, newClusterNames)
		_, err := p.placementClient.Resources(qualifiedName.Namespace).Create(newPlacement, metav1.CreateOptions{})
		return err
	}

	clusterNames, err := util.GetClusterNames(placement)
	if err != nil {
		return err
	}
	if PlacementUpdateNeeded(clusterNames, newClusterNames) {
		setPlacementSpec(placement, newClusterNames)
		_, err := p.placementClient.Resources(qualifiedName.Namespace).Update(placement, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) ReconcileOverride(qualifiedName util.QualifiedName, result map[string]int64) error {
	override, err := p.overrideClient.Resources(qualifiedName.Namespace).Get(qualifiedName.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		newOverride := newUnstructured(p.typeConfig.GetOverride(), qualifiedName)
		err := setOverrides(newOverride, nil, result)
		if err != nil {
			return err
		}
		_, err = p.overrideClient.Resources(qualifiedName.Namespace).Create(newOverride, metav1.CreateOptions{})
		return err
	}

	overridesMap, err := util.GetOverrides(override)
	if err != nil {
		return errors.Wrapf(err, "Error reading cluster overrides for %s %q", p.typeConfig.GetOverride().Kind, qualifiedName)
	}

	if OverrideUpdateNeeded(overridesMap, result) {
		err := setOverrides(override, overridesMap, result)
		if err != nil {
			return err
		}
		_, err = p.overrideClient.Resources(qualifiedName.Namespace).Update(override, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func newUnstructured(apiResource metav1.APIResource, qualifiedName util.QualifiedName) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetKind(apiResource.Kind)
	gv := schema.GroupVersion{Group: apiResource.Group, Version: apiResource.Version}
	obj.SetAPIVersion(gv.String())
	obj.SetName(qualifiedName.Name)
	obj.SetNamespace(qualifiedName.Namespace)
	return obj
}

func setPlacementSpec(obj *unstructured.Unstructured, clusterNames []string) {
	obj.Object[util.SpecField] = map[string]interface{}{
		util.ClusterNamesField: clusterNames,
	}
}

// These assume that there would be no duplicate clusternames
func PlacementUpdateNeeded(names, newNames []string) bool {
	sort.Strings(names)
	sort.Strings(newNames)
	return !reflect.DeepEqual(names, newNames)
}

func setOverrides(obj *unstructured.Unstructured, overridesMap util.OverridesMap, replicasMap map[string]int64) error {
	if overridesMap == nil {
		overridesMap = make(util.OverridesMap)
	}
	updateOverridesMap(overridesMap, replicasMap)
	return util.SetOverrides(obj, overridesMap)
}

func updateOverridesMap(overridesMap util.OverridesMap, replicasMap map[string]int64) {
	// Remove replicas override for clusters that are not scheduled
	for clusterName, clusterOverridesMap := range overridesMap {
		if _, ok := replicasMap[clusterName]; !ok {
			delete(clusterOverridesMap, replicasPath)
		}
	}
	// Add/update replicas override for clusters that are scheduled
	for clusterName, replicas := range replicasMap {
		clusterOverridesMap, ok := overridesMap[clusterName]
		if !ok {
			clusterOverridesMap = make(util.ClusterOverridesMap)
			overridesMap[clusterName] = clusterOverridesMap
		}
		clusterOverridesMap[replicasPath] = replicas
	}
}

func OverrideUpdateNeeded(overridesMap util.OverridesMap, result map[string]int64) bool {
	resultLen := len(result)
	checkLen := 0
	for clusterName, clusterOverridesMap := range overridesMap {
		for path, rawValue := range clusterOverridesMap {
			if path != replicasPath {
				continue
			}
			// The type of the value will be float64 due to how json
			// marshalling works for interfaces.
			floatValue, ok := rawValue.(float64)
			if !ok {
				return true
			}
			value := int64(floatValue)
			replicas, ok := result[clusterName]
			if !ok || value != int64(replicas) {
				return true
			}
			checkLen += 1
		}
	}

	return checkLen != resultLen
}
