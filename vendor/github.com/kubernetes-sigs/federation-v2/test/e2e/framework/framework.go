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

package framework

import (
	"context"
	"fmt"

	"github.com/kubernetes-sigs/federation-v2/pkg/apis/core/typeconfig"
	fedv1a1 "github.com/kubernetes-sigs/federation-v2/pkg/apis/core/v1alpha1"
	fedclientset "github.com/kubernetes-sigs/federation-v2/pkg/client/clientset/versioned"
	"github.com/kubernetes-sigs/federation-v2/pkg/controller/util"
	"github.com/kubernetes-sigs/federation-v2/test/common"
	"github.com/kubernetes-sigs/federation-v2/test/e2e/framework/managed"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeclientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	crclientset "k8s.io/cluster-registry/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FederationFrameworkImpl interface {
	BeforeEach()
	AfterEach()

	ControllerConfig() *util.ControllerConfig

	Logger() common.TestLogger

	KubeConfig() *restclient.Config

	KubeClient(userAgent string) kubeclientset.Interface
	FedClient(userAgent string) fedclientset.Interface
	CrClient(userAgent string) crclientset.Interface

	ClusterConfigs(userAgent string) map[string]common.TestClusterConfig
	ClusterDynamicClients(apiResource *metav1.APIResource, userAgent string) map[string]common.TestCluster
	ClusterKubeClients(userAgent string) map[string]kubeclientset.Interface
	ClusterNames(userAgent string) []string

	FederationSystemNamespace() string

	// Name of the namespace for the current test to target
	TestNamespaceName() string

	// Internal method that accepts the namespace placement api
	// resource to support namespaced controllers.
	setUpSyncControllerFixture(typeConfig typeconfig.Interface, namespacePlacement *metav1.APIResource) managed.TestFixture
}

// FederationFramework provides an interface to a test federation so
// that the implementation can vary without affecting tests.
type FederationFramework interface {
	FederationFrameworkImpl

	// Registering a fixture ensures it will be torn down after the
	// current test has executed.
	RegisterFixture(fixture managed.TestFixture)

	// Setup a sync controller if necessary and return the fixture.
	// This is implemented commonly to support running a sync
	// controller for tests that require it.
	SetUpSyncControllerFixture(typeConfig typeconfig.Interface) managed.TestFixture

	// Ensure propagation of the test namespace to member clusters
	EnsureTestNamespacePropagation()

	// Ensure a placement resource for the test namespace exists so
	// that the namespace will be propagated to either all or no
	// member clusters.
	EnsureTestNamespacePlacement(allClusters bool) *unstructured.Unstructured
}

// A framework needs to be instantiated before tests are executed to
// ensure registration of BeforeEach/AfterEach.  However, flags that
// dictate which framework should be used won't have been parsed yet.
// The workaround is using a wrapper that performs late-binding on the
// framework flavor.
type frameworkWrapper struct {
	impl                FederationFrameworkImpl
	baseName            string
	namespaceTypeConfig typeconfig.Interface

	// Fixtures to cleanup after each test
	fixtures []managed.TestFixture
}

func NewFederationFramework(baseName string) FederationFramework {
	f := &frameworkWrapper{
		baseName: baseName,
		fixtures: []managed.TestFixture{},
	}
	AfterEach(f.AfterEach)
	BeforeEach(f.BeforeEach)
	return f
}

func (f *frameworkWrapper) framework() FederationFrameworkImpl {
	if f.impl == nil {
		if TestContext.TestManagedFederation {
			f.impl = NewManagedFramework(f.baseName)
		} else {
			f.impl = NewUnmanagedFramework(f.baseName)
		}
	}
	return f.impl
}

func (f *frameworkWrapper) BeforeEach() {
	f.framework().BeforeEach()
}

func (f *frameworkWrapper) AfterEach() {
	f.framework().AfterEach()

	// TODO(font): Delete the namespace finalizer manually rather than
	// relying on the federated namespace sync controller. This would
	// remove the dependency of namespace removal on fixture teardown,
	// which allows the teardown to be moved outside the defer, but before
	// the DumpEventsInNamespace that may contain assertions that could
	// exit the function.
	logger := f.framework().Logger()
	for len(f.fixtures) > 0 {
		fixture := f.fixtures[0]
		fixture.TearDown(logger)
		f.fixtures = f.fixtures[1:]
	}
}

func (f *frameworkWrapper) ControllerConfig() *util.ControllerConfig {
	return f.framework().ControllerConfig()
}

func (f *frameworkWrapper) Logger() common.TestLogger {
	return f.framework().Logger()
}

func (f *frameworkWrapper) KubeConfig() *restclient.Config {
	return f.framework().KubeConfig()
}

func (f *frameworkWrapper) KubeClient(userAgent string) kubeclientset.Interface {
	return f.framework().KubeClient(userAgent)
}

func (f *frameworkWrapper) FedClient(userAgent string) fedclientset.Interface {
	return f.framework().FedClient(userAgent)
}

func (f *frameworkWrapper) CrClient(userAgent string) crclientset.Interface {
	return f.framework().CrClient(userAgent)
}

func (f *frameworkWrapper) ClusterConfigs(userAgent string) map[string]common.TestClusterConfig {
	return f.framework().ClusterConfigs(userAgent)
}

func (f *frameworkWrapper) ClusterNames(userAgent string) []string {
	return f.framework().ClusterNames(userAgent)
}

func (f *frameworkWrapper) ClusterDynamicClients(apiResource *metav1.APIResource, userAgent string) map[string]common.TestCluster {
	return f.framework().ClusterDynamicClients(apiResource, userAgent)
}

func (f *frameworkWrapper) ClusterKubeClients(userAgent string) map[string]kubeclientset.Interface {
	return f.framework().ClusterKubeClients(userAgent)
}

func (f *frameworkWrapper) FederationSystemNamespace() string {
	return f.framework().FederationSystemNamespace()
}

func (f *frameworkWrapper) TestNamespaceName() string {
	return f.framework().TestNamespaceName()
}

// setUpSyncControllerFixture is not intended to be called on the
// wrapper.  It's only implemented by the concrete frameworks to avoid
// having callers pass in the namespacePlacement arg.
func (f *frameworkWrapper) setUpSyncControllerFixture(typeConfig typeconfig.Interface, namespacePlacement *metav1.APIResource) managed.TestFixture {
	return nil
}

func (f *frameworkWrapper) SetUpSyncControllerFixture(typeConfig typeconfig.Interface) managed.TestFixture {
	namespaceTypeConfig := f.namespaceTypeConfigOrDie()
	placement := namespaceTypeConfig.GetPlacement()
	return f.framework().setUpSyncControllerFixture(typeConfig, &placement)
}

func (f *frameworkWrapper) RegisterFixture(fixture managed.TestFixture) {
	if fixture != nil {
		f.fixtures = append(f.fixtures, fixture)
	}
}

func (f *frameworkWrapper) EnsureTestNamespacePropagation() {
	// When targeting a single namespace, the test namespace will
	// already exist in member clusters.
	if TestContext.LimitedScope {
		return
	}

	// Ensure placement selecting all clusters exists for the test
	// namespace.
	f.EnsureTestNamespacePlacement(true)

	// Start the namespace sync controller to propagate the namespace
	namespaceTypeConfig := f.namespaceTypeConfigOrDie()
	fixture := f.framework().setUpSyncControllerFixture(namespaceTypeConfig, nil)
	f.RegisterFixture(fixture)
}

func (f *frameworkWrapper) EnsureTestNamespacePlacement(allClusters bool) *unstructured.Unstructured {
	tl := f.Logger()

	dynclient, err := client.New(f.KubeConfig(), client.Options{})
	if err != nil {
		tl.Fatalf("Error initializing dynamic client: %v", err)
	}

	apiResource := f.namespaceTypeConfigOrDie().GetPlacement()
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   apiResource.Group,
		Kind:    apiResource.Kind,
		Version: apiResource.Version,
	})

	namespace := f.TestNamespaceName()

	// Return an existing namespace placement if it already exists.
	key := client.ObjectKey{Namespace: namespace, Name: namespace}
	err = dynclient.Get(context.Background(), key, obj)
	if err == nil {
		return obj
	}
	if err != nil && !errors.IsNotFound(err) {
		tl.Fatalf("Error retrieving %s %q: %v", apiResource.Kind, err)
	}

	// Othewise create it.
	obj.SetName(namespace)
	obj.SetNamespace(namespace)
	if allClusters {
		// An empty cluster selector field selects all clusters
		obj.Object[util.SpecField] = map[string]interface{}{
			util.ClusterSelectorField: map[string]interface{}{},
		}
	}

	err = dynclient.Create(context.Background(), obj)
	if err != nil {
		tl.Fatalf("Error creating %s %q: %v", apiResource.Kind, key, err)
	}
	tl.Logf("Created new %s %q", apiResource.Kind, key)

	return obj
}

func (f *frameworkWrapper) namespaceTypeConfigOrDie() typeconfig.Interface {
	if f.namespaceTypeConfig == nil {
		tl := f.Logger()
		dynClient, err := client.New(f.KubeConfig(), client.Options{})
		if err != nil {
			tl.Fatalf("Error initializing dynamic client: %v", err)
		}
		key := client.ObjectKey{
			Namespace: f.FederationSystemNamespace(),
			Name:      util.NamespaceName,
		}
		typeConfig := &fedv1a1.FederatedTypeConfig{}
		err = dynClient.Get(context.Background(), key, typeConfig)
		if err != nil {
			tl.Fatalf("Error retrieving federatedtypeconfig for %q: %v", key.Name, err)
		}
		f.namespaceTypeConfig = typeConfig
	}
	return f.namespaceTypeConfig
}

func createTestNamespace(client kubeclientset.Interface, baseName string) string {
	By("Creating a namespace to execute the test in")
	namespaceName, err := createNamespace(client, baseName)
	Expect(err).NotTo(HaveOccurred())
	By(fmt.Sprintf("Created test namespace %s", namespaceName))
	return namespaceName
}

func createNamespace(client kubeclientset.Interface, baseName string) (string, error) {
	namespaceObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("e2e-tests-%v-", baseName),
		},
	}
	// Be robust about making the namespace creation call.
	// TODO(marun) should all api calls be made 'robustly'?
	var namespaceName string
	if err := wait.PollImmediate(PollInterval, TestContext.SingleCallTimeout, func() (bool, error) {
		namespace, err := client.Core().Namespaces().Create(namespaceObj)
		if err != nil {
			Logf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		namespaceName = namespace.Name
		return true, nil
	}); err != nil {
		return "", err
	}
	return namespaceName, nil
}
