package main_test

// The imports in this test stub exist to get dep to populate the dependencies
// we need for the federation-v2 tests.

import (
	"testing"

	_ "github.com/kubernetes-sigs/kubebuilder/pkg/test"
	_ "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
	_ "github.com/pborman/uuid"
	_ "github.com/stretchr/testify/assert"
	_ "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestStub(t *testing.T) {

}
