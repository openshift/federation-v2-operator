## `federation-v2-operator`

This repository contains source code and resource manifests to build
[Federation-v2](https://github.com/kubernetes-sigs/federation-v2) images
using [ci-operator](https://github.com/openshift/ci-operator) as part of the
[OpenShift](https://openshift.com) build process.

Federation v2 is deployed as an [operator](https://coreos.com/operators) using
[OLM](https://github.com/operator-framework/operator-lifecycle-manager).

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
- [ci-operator configuration](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/federation-v2-operator/openshift-federation-v2-operator-master.yaml)

### Developing

This project has tooling allowing you to develop against your own image
repositories in [quay.io](quay.io) without having to make local changes. Quick
development guide follows.

#### Prerequisites

You must have:

- An OpenShift 4.0 cluster and cluster-admin rights for that cluster
  - The `federation-test` namespace must exist in your cluster
- Your own quay.io account and the following image repositories under that account:
  - `origin-federation-controller`
  - `federation-operator-registry`
- The `oc` binary in your `PATH`

#### Build the container image

Build and push the container image to your `origin-federation-controller` image
repository using this command:

For this step, use image tag `quay.io/<your quay account>/origin-federation-controller:v4.0.0`.

```
$ docker build . -t quay.io/<your quay account>/origin-federation-controller:v4.0.0
$ docker push quay.io/<your quay account>/origin-federation-controller:v4.0.0
```

#### Create an operator registry

Use the `scripts/push-operator-registry.sh` script to push an image containing
an operator registry. This script takes your quay.io account name as an argument:

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
into the `federation-test` namespace using OLM. This script takes your quay
account name as an argument, and optionally the type of deployment to subscribe
to:

```
$ scripts/install-using-catalog-source.sh pmorie <namespaced|cluster-scoped>
catalogsource.operators.coreos.com/federation created
operatorgroup.operators.coreos.com/namespaced-federation created
subscription.operators.coreos.com/namespaced-federation-sub created
```

This script:

- Configures a `CatalogSource` for OLM that references the operator registry you built
- Creates an `OperatorGroup` 
- Creates a `Subscription` to the operator that drives OLM to install it in your namespace

Note, if you run this script without an argument, you will get federation
deployed in a namespaced scope. To deploy the cluster-scoped version, you should run:

```
$ ./scripts/install-using-catalog-source.sh pmorie cluster-scoped
catalogsource.operators.coreos.com/federation created
operatorgroup.operators.coreos.com/cluster-scoped-federation created
subscription.operators.coreos.com/cluster-scoped-federation-sub created
```
