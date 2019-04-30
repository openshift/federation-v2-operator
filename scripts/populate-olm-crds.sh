#!/bin/bash

# This script populates custom resource definitions for the OLM configuration
# for federation in the 'manifests' directory from the vendored federation-v2
# helm chart.
#
# This script should be run from the root directory of the
# federation-v2-operator repository.
#
# There are different classes of CRDs this script handles:
#
# 1. CRDs for the cluster-registry
# 2. CRDs 'core' to federation itself - FederatedCluster, etc.
# 3. CRDs for the 'built-in' federation APIs that come with federation.

FEDERATION_CHART_DIR=vendor/github.com/kubernetes-sigs/federation-v2/charts/federation-v2/
PACKAGE=${PACKAGE:-"federation"}
VERSION=${VERSION:-"0.0.9"}
FLAVOR=${FLAVOR:-"upstream-manifests"}
MANIFESTS_DIR=${FLAVOR}/${PACKAGE}/${VERSION}

echo "Populating OLM manifests for package ${PACKAGE} version ${VERSION}"

# handle cluster-registry
cp ${FEDERATION_CHART_DIR}/charts/clusterregistry/templates/crds.yaml ${MANIFESTS_DIR}/cluster-registry.crd.yaml

# handle core federation CRDs; they are in the vendored federation-v2 repo in a
# single file and must be split into individual files per resource in order for
# the operator registry to handle them correctly.
all_core_crds=${MANIFESTS_DIR}/federation-core-all.crd.yaml
cp ${FEDERATION_CHART_DIR}/charts/controllermanager/templates/crds.yaml ${all_core_crds}
csplit ${all_core_crds} --prefix=${MANIFESTS_DIR}/${PACKAGE}-core-split -- /---/ {*}
rm -f ${all_core_crds}
core_crds=$(find ${MANIFESTS_DIR} -name ${PACKAGE}-core-split*)
for f in $core_crds
do
  kind=$(grep kind: $f | head -n 2 | tail -n 1 | cut -b 11-)
  mv -f $f ${MANIFESTS_DIR}/${PACKAGE}-core-${kind}.crd.yaml
done

# For now, don't handle federation API CRDs.
#
# There is work required to be able to deploy the FederatedTypeConfigs required
# for federation to use these APIs, so for now we will deploy federation in a
# state where it doesn't yet know about any APIs and users will have to run
# 'kubefed federate enable' themselves.
exit

# handle federated API CRDs
declare -a filenames=("clusterroles.rbac.authorization.k8s.io"
                      "configmaps"
                      "deployments.apps"
                      "ingresses.extensions"
                      "jobs.batch"
                      "namespaces"
                      "replicasets.apps"
                      "secrets"
                      "serviceaccounts"
                      "services"
)

for i in "${filenames[@]}"
do
  # The manifest files in the vendored federation repo include the
  # FederatedTypeConfigs that drive federation for the federated APIs for k8s.
  # However, OLM doesn't currently support distributing arbitrary resources as
  # part of an OperatorBundle, so we strip them out here.
  src=${FEDERATION_CHART_DIR}/templates/${i}.yaml
  # find the line number for the _second_ '---' resource delimiter in the source file
  line=$(grep -n -- --- $src | head -n 2 | tail -n 1 | cut -d: -f1)
  # populate the manifest directory with _only_ the CRD definition (ie,
  # everything after the second '---')
  awk -v line=$line 'NR > line' $src > ${MANIFESTS_DIR}/federation-api-${i}.crd.yaml
done
