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
- The `manifests/` directory contains manifests to configure OLM to deploy
  federation-v2
- The `olm-testing/` directory contains a `Dockerfile` for building an operator
  registry that hosts the OLM manifests
- The `scripts/` directory holds scripts to populate `manifests`

### Continuous Integration

- [Prow Status](https://deck-ci.svc.ci.openshift.org/?repo=openshift%2Ffederation-v2-operator)

### Developing

Use the `scripts/push-operator-registry.sh` script to push an image containing an operator registry:

```
$ ./scripts/push-operator-registry.sh pmorie
Building operator registry with tag quay.io/pmorie/federation-operator-registry
Sending build context to Docker daemon  53.34MB
Step 1/4 : FROM quay.io/openshift/origin-operator-registry:latest
...
```

OLM is configured via a `CatalogSource` that uses the operator registry image:

```
$ kubectl create -f olm-testing/catalog-source.yaml
```
