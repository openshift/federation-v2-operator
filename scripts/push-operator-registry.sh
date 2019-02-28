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

new_image_name=quay.io/$registry/origin-federation-controller
tag="quay.io/$registry/federation-operator-registry:v4.0.0"
echo "Building operator registry with tag $tag"
docker build . -f olm-testing/Dockerfile -t $tag --build-arg new_image_name=$new_image_name
echo "Pushing $tag"
docker push $tag
