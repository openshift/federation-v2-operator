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

new_image_name="quay.io/$registry/federation-operator-registry:v4.0.0"
if [[ ! -z "$OVERRIDE_IMAGE" ]]; then
  new_image_name="$registry"
fi

sed -e "s,quay.io/openshift/federation-operator-registry:v4.0.0,$new_image_name," \
  $dir/olm-testing/catalog-source.yaml | oc apply -f -
oc apply -f $dir/olm-testing/${subscription_type}-operator-group.yaml
oc apply -f $dir/olm-testing/${subscription_type}-subscription.yaml
