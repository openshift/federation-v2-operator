## `federation-v2-operator`

This repository contains source code and resource manifests to build
[Federation-v2](https://github.com/kubernetes-sigs/federation-v2) images
using [ci-operator](https://github.com/openshift/ci-operator) as part of the
[OpenShift](https://openshift.com) build process.

This repository also contains exploratory work to deploy Federation-v2 as an
[operator](https://coreos.com/operators) using
[OLM](https://github.com/operator-framework/operator-lifecycle-manager).

### Relationship between this repository and OperatorHub

This repository is the canonical store for manifests that appear on
[operatorhub.io](https://operatorhub.io) and the in-cluster OperatorHub in
OpenShift 4. In turn, the different OperatorHub locations are all populated from
the
[community-operators](https://github.com/operator-framework/community-operators)
repository. Here's an example of how a change makes its way into this repository
and then into OperatorHub:

1. Changes are made via a pull request to this repository and merged
2. Someone uses the `scripts/populate-community-operators.sh` script to update
   their local fork of the `community-operators` repo with the latest manifests
   in the `community-operators` and `upstream-community-operators` directories
3. The same person makes a pull request to the `community-operators` repository
   with the latest version of the manifests
4. Once the pull request to `community-operators` is merged:
    1. The OperatorHub team manually updates [operatorhub.io](https://operatorhub.io)
    2. OpenShift 4.x clusters pick up the change automatically within 1 hour

### Layout of this repository

- The `stub.go` and `stub_test.go` files are intentionally-non-compileable stub
  files that contain imports to drive `dep` to put the right things into the
  vendor directory
- The `vendor/` directory holds a pinned version of federation-v2 and its
  dependencies
- The `Dockerfile` in the root directory is used to build enterprise images and
  performs binary builds only
- The `upstream-manifests/` directory contains manifests to configure OLM to
  deploy federation-v2 _using the [upstream images](https://quay.io/repository/kubernetes-multicluster/federation-v2?tab=tags);
  this is the directory that populates manifests in
  [OperatorHub](https://operatorhub.io)
- The `olm-testing/` directory contains a `Dockerfile` for building an operator
  registry that hosts the OLM manifests
- The `scripts/` directory holds scripts to populate `manifests` and
  `upstream-manifests`, the
  [community-operators repo](https://github.com/operator-framework/community-operators), push
  operator registries, etc.

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
- The `kubectl` binary in your `PATH`

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
