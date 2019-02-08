#!/bin/bash

# This script builds and pushes an operator registry image and pushes to
# quay.io/$registry/federation-operator-registry
#
# This script should be run from the root directory of the
# federation-v2-operator repository.

registry=${REGISTRY:-$1}
if [[ -z "${registry}" ]]; then
  echo "registry must be set by running \`push-operator-registry <registry>\` or by setting \$REGISTRY"
  exit 1
fi

tag="quay.io/$registry/federation-operator-registry"
echo "Building operator registry with tag $tag"
docker build . -f olm-testing/Dockerfile -t $tag
echo "Pushing $tag"
docker push $tag
