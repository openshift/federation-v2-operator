#!/bin/bash

# This script automates the work of installing federation using the
# CatalogSource in the `olm-testing` directory.
#
# This script can be run from any directory.

dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")

registry=${REGISTRY:-$1}
if [[ -z "${registry}" ]]; then
  echo "registry must be set by running \`install-using-catalog-source.sh <registry>\` or by setting \$REGISTRY"
  exit 1
fi

sed -e "s,quay.io/openshift/federation-operator-registry,quay.io/$registry/federation-operator-registry," $dir/olm-testing/catalog-source.yaml | oc apply -f -
oc apply -f $dir/olm-testing/operator-group.yaml
oc apply -f $dir/olm-testing/subscription.yaml
