#!/bin/bash

# This script populates the community-operators repository with OLM manifests
# from this repository. This is a temporary measure until automation exists that
# will consume the manifests directly from an annotated container image.
#
# This script can be run from anywhere; it expects the following preconditions:
#
# - This repo and the community-operators repo are checked out locally into the
#   correct locations within a GOPATH:
#   - src/github.com/openshift/federation-v2-operator
#   - src/github.com/operator-framework/community-operators

dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")
community_operators_dir=$(realpath "${dir}/../../operator-framework/community-operators")
target_dir="${community_operators_dir}/community-operators/federation"
mkdir -p $target_dir

cp -r $dir/manifests/federation $target_dir