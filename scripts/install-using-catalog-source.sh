#!/bin/bash

# This script automates the work of installing federation using the
# CatalogSource in the `olm-testing` directory.
#
# This script can be run from any directory and takes two arguments:
#
# - The quay.io account name to use for the federation-operator-registry image
# - The type of federation to subscribe to. There are two valid types:
#
# - namespaced
# - cluster-scoped
#
# If the second argument is omitted, the default will be 'namespaced'.

dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")
kubectl=${KUBECTL:-"kubectl"}

registry=${REGISTRY:-$1}
if [[ -z "${registry}" ]]; then
  echo "registry must be set by running \`install-using-catalog-source.sh <registry>\` or by setting \$REGISTRY"
  exit 1
fi

subscription_type=${2:-"namespaced"}
declare -a subscription_types=("namespaced" "cluster-scoped")
if ! [[ "${subscription_types[@]}" =~ (^|[[:space:]])"$subscription_type"($|[[:space:]]) ]] ; then
  echo "${subscription_type} is an invalid type to subscribe to; use ${subscription_types[@]}"
  exit 1
fi

sed -e "s,quay.io/openshift/federation-operator-registry,quay.io/$registry/federation-operator-registry," $dir/olm-testing/catalog-source.yaml | oc apply -f -
$kubectl apply -f $dir/olm-testing/${subscription_type}-operator-group.yaml
$kubectl apply -f $dir/olm-testing/${subscription_type}-subscription.yaml
