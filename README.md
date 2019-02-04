## `federation-v2-operator`

This repository contains source code and resource manifests to build
[Federation-v2](https://github.com/kubernetes-sigs/federation-v2) images
using [ci-operator](https://github.com/openshift/ci-operator) as part of the
[OpenShift](https://openshift.com) build process.

Federation v2 is deployed as an [operator](https://coreos.com/operators) using
[OLM](https://github.com/operator-framework/operator-lifecycle-management).

### Layout of this repository

- The `stub.go` and `stub_test.go` files are intentionally-non-compileable stub
  files that contain imports to drive `dep` to put the right things into the
  vendor directory
- The `vendor/` directory holds a pinned version of federation-v2 and its
  dependencies
- The `Dockerfile` in the root directory is used to build enterprise images and
  performs binary builds only
- The `Dockerfile.federation` file in the root directory is used to build an
  image that contains the vendored source for `federation-v2` and is used to run
  vet checks, unit, and e2e tests

### Continuous Integration

- [Prow Status](https://deck-ci.svc.ci.openshift.org/?repo=openshift%2Ffederation-v2-operator)