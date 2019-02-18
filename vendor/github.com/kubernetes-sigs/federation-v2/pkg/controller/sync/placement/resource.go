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

package placement

import (
	fedv1a1 "github.com/kubernetes-sigs/federation-v2/pkg/apis/core/v1alpha1"
	"github.com/kubernetes-sigs/federation-v2/pkg/controller/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
)

type ResourcePlacementPlugin struct {
	// Store for the placement directives of the federated type
	store cache.Store
	// Informer controller for placement directives of the federated type
	controller cache.Controller

	kind string
}

func NewResourcePlacementPlugin(client util.ResourceClient, targetNamespace string, triggerFunc func(pkgruntime.Object)) PlacementPlugin {
	return newResourcePlacementPluginWithOk(client, targetNamespace, triggerFunc)
}

func newResourcePlacementPluginWithOk(client util.ResourceClient, targetNamespace string, triggerFunc func(pkgruntime.Object)) *ResourcePlacementPlugin {
	store, controller := util.NewResourceInformer(client, targetNamespace, triggerFunc)
	return &ResourcePlacementPlugin{
		store:      store,
		controller: controller,
		kind:       client.Kind(),
	}
}

func (p *ResourcePlacementPlugin) Run(stopCh <-chan struct{}) {
	p.controller.Run(stopCh)
}

func (p *ResourcePlacementPlugin) HasSynced() bool {
	return p.controller.HasSynced()
}

func (p *ResourcePlacementPlugin) ComputePlacement(qualifiedName util.QualifiedName, clusters []*fedv1a1.FederatedCluster) (selectedClusters, unselectedClusters []string, err error) {
	selectedClusters, unselectedClusters, _, err = p.computePlacementWithOk(qualifiedName, clusters)
	return selectedClusters, unselectedClusters, err
}

func (p *ResourcePlacementPlugin) GetPlacement(key string) (*unstructured.Unstructured, error) {
	return util.ObjFromCache(p.store, p.kind, key)
}

func (p *ResourcePlacementPlugin) computePlacementWithOk(qualifiedName util.QualifiedName, clusters []*fedv1a1.FederatedCluster) (selectedClusters, unselectedClusters []string, ok bool, err error) {
	unstructuredObj, err := p.GetPlacement(qualifiedName.String())
	if err != nil {
		return nil, nil, false, err
	}
	clusterNames := getClusterNames(clusters)
	if unstructuredObj == nil {
		return []string{}, clusterNames, false, nil
	}

	selectedNames, err := selectedClusterNames(unstructuredObj, clusters)
	if err != nil {
		return nil, nil, false, err
	}

	clusterSet := sets.NewString(clusterNames...)
	selectedSet := sets.NewString(selectedNames...)
	return clusterSet.Intersection(selectedSet).List(), clusterSet.Difference(selectedSet).List(), true, nil
}

func getClusterNames(clusters []*fedv1a1.FederatedCluster) []string {
	clusterNames := []string{}
	for _, cluster := range clusters {
		clusterNames = append(clusterNames, cluster.Name)
	}
	return clusterNames
}

func selectedClusterNames(rawPlacement *unstructured.Unstructured, clusters []*fedv1a1.FederatedCluster) ([]string, error) {
	directive, err := util.GetPlacementDirective(rawPlacement)
	if err != nil {
		return nil, err
	}

	if directive.ClusterNames != nil {
		return directive.ClusterNames, nil
	}

	selectedNames := []string{}
	for _, cluster := range clusters {
		if directive.ClusterSelector.Matches(labels.Set(cluster.Labels)) {
			selectedNames = append(selectedNames, cluster.Name)
		}
	}
	return selectedNames, nil
}
