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
