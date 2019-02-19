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

Quick development guide follows.

#### Prerequisites

You must have:

- An OpenShift 4.0 cluster and cluster-admin rights for that cluster
  - The `federation-testing` namespace must exist in your cluster
- Your own quay.io account
- Under your quay.io account, these image repositories:
  - `origin-federation-controller`
  - `federation-operator-registry`
- The `oc` binary in your `PATH`

#### Local changes for your development flow

You must first make a couple of changes locally:

- You must alter the image refered to in the `manifests/federation/0.0.4` file
  to refer to the image you build and push to an image registry you control
- OLM is configured via a `CatalogSource` that uses the operator registry image.
  This lives in the `olm-testing/catalog-source.yaml` file; to test/develop your
  own instance of the operator, you will need to substitute the coordinates of
  your operator registry image.

#### Build the container image

Build and push the container image to your `origin-federation-controller` image
repository using this command:

```
$ docker build . -t <your image tag>
$ docker push <your image tag>
```

For this step, use image tag `quay.io/<your quay account>/origin-federation-controller:v4.0.0`.

#### Create an operator registry

Use the `scripts/push-operator-registry.sh` script to push an image containing
an operator registry:

```
$ ./scripts/push-operator-registry.sh pmorie
Building operator registry with tag quay.io/pmorie/federation-operator-registry
Sending build context to Docker daemon  53.34MB
Step 1/4 : FROM quay.io/openshift/origin-operator-registry:latest
...
```

Note, this script accepts a parameter for the name of the repository to push to;
use this to inject your quay account name.

#### Install federation using OLM

Run the `scripts/install-using-catalog-source.sh` script to install federation
into the `federation-testing` namespace using OLM. This script:

- Configures a `CatalogSource` for OLM that references the operator registry you built
- Creates an `OperatorGroup` 
- Creates a `Subscription` to the operator that drives OLM to install it in your namespace
