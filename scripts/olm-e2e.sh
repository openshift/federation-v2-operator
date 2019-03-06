#!/bin/bash -e

# olm-e2e.sh orchestrates deployent of the OLM manifests and images from this
# repo, with images references transformed to those built locally or in a CI
# pipeline.
#
# Key points of workflow:
#
# - Make an operator registry
# - Make a CatalogSource pointing to the operator registry
# - Make a Subscription to the desired package
# - Wait for CSV to report status succeeded

# TODO: RIPPED from the upstream kube test shell library. Everyone will need
# this. What do we do?
readonly reset=$(tput sgr0)
readonly  bold=$(tput bold)
readonly black=$(tput setaf 0)
readonly   red=$(tput setaf 1)
readonly green=$(tput setaf 2)

# TODO: change to the right thing once this is working
kubectl=${KUBECTL:-"oc"}

test::object_assert() {
  local tries=$1
  local object=$2
  local request=$3
  local expected=$4
  local args=${5:-}

  for j in $(seq 1 ${tries}); do
    res=$(eval $kubectl get ${args} ${object} -o jsonpath=\"${request}\")
    if [[ "${res}" =~ ^$expected$ ]]; then
        echo -n "${green}"
        echo "Successful get ${object} ${request}: ${res}"
        echo -n "${reset}"
        return 0
    fi
    echo "Waiting for Get ${object} ${request} ${args}: expected: ${expected}, got: ${res}"
    sleep $((${j}-1))
  done

  echo "${bold}${red}"
  echo "FAIL!"
  echo "Get ${object} ${request}"
  echo "  Expected: ${expected}"
  echo "  Got:      ${res}"
  echo "${reset}${red}"
  caller
  echo "${reset}"
  return 1
}

# Env param checking. Required:
#
# - KUBECONFIG
#
# if IN_CI is set:
# - IMAGE_FORMAT
# otherwise:
# - QUAY_ACCOUNT
if [ -z "$KUBECONFIG" ]; then
  echo "KUBECONFIG must be set"
  exit 1
fi

# TODO: right way to do 'i-am-in-ci'
IN_CI="${IN_CI:-""}"
if [[ ! -z "$IN_CI" ]]; then
  # validate arguments for CI flow
  if [ -z "$IMAGE_FORMAT" ]; then
      echo "IMAGE_FORMAT must be set"
      exit 1
  fi
else
  # validate arguments for development flow
  if [ -z "$QUAY_ACCOUNT" ]; then
    echo "QUAY_ACCOUNT must be set"
    exit 1
  fi
fi

subscription_type="${SUBSCRIPTION_TYPE:-"namespaced"}"

set -eu
set -x

# change to root directory of federation-v2-operator repo
dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")
cd $dir

# todo: parameter or random ns?
# note: this is currently hard-coded in assets under olm-testing dir
oc new-project federation-test

version=${FEDERATION_VERSION:-"0.0.6"}

# Deploy operator registry and subscriptions
if [[ ! -z "${IN_CI}" ]]; then
  # IMAGE_FORMAT is a string of the form:
  #
  # registry.svc.ci.openshift.org/ci-op-<input-hash>/stable:${component}
  #
  # The 'component' variable defined here is used to interpolate into IMAGE_FORMAT
  # in the following eval statement.
  echo "IMAGE_FORMAT=$IMAGE_FORMAT"
  component="federation-controller"
  pipeline_federation_image=${IMAGE_FORMAT/\$\{component\}/$component}
  operator_registry_imagestream_tag="operator-registry:latest"
  sed "s,quay.io/openshift/origin-federation-controller,$pipeline_federation_image," -i manifests/federation/${version}/federation.v${version}.clusterserviceversion.yaml
  sed "s,quay.io/openshift/origin-federation-controller,$pipeline_federation_image," -i manifests/cluster-federation/${version}/cluster-federation.v${version}.clusterserviceversion.yaml

  # build and push operator registry image, with the OLM CSV munged to reference
  # the pipeline image for federation
  oc new-build --name=operator-registry --dockerfile="$(<olm-testing/Dockerfile)"
  oc start-build operator-registry \
    --from-dir=. \
    --wait

  # install catalog source pointing to operator
  # TODO OVERRIDE_IMAGE
  OVERRIDE_IMAGE=1 scripts/install-using-catalog-source.sh $operator_registry_imagestream_tag $subscription_type
else
  scripts/push-operator-registry.sh $QUAY_ACCOUNT
  scripts/install-using-catalog-source.sh $QUAY_ACCOUNT $subscription_type
fi

test::object_assert 30 subscriptions.operators.coreos.com/${subscription_type}-federation-sub "{.status.state}" AtLastKnown
test::object_assert 30 clusterserviceversions.operators.coreos.com/federation.v${version} "{.status.phase}" Succeeded

# TODO: do we ever need to do this?
#
# if [[ ! -z "${SETUP_FEDERATION_REPO}" ]]; then
#   cd vendor/github.com/kubernetes-sigs/federation-v2
# fi

# It's party time, let's run some e2es

# # construct e2e cmd
# managed_e2e_test_cmd="go test -v ./test/e2e -args -ginkgo.v -single-call-timeout=1m -ginkgo.trace -ginkgo.randomizeAllSpecs"
# unmanaged_e2e_test_cmd="${managed_e2e_test_cmd} -kubeconfig=${KUBECONFIG}"
# e2e_namespaced_arg="-limited-scope=true"
# final_e2e_test_cmd="${unmanaged_e2e_test_cmd} -federation-namespace=foo -registry-namespace=foo ${e2e_namespaced_arg}"

# ${final_e2e_test_cmd}
