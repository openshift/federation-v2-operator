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

PACKAGE=${PACKAGE:-"federation"}
VERSION=${VERSION:-"0.0.6"}
AREA=${AREA:-"community-operators"}

dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")
community_operators_dir=$(realpath "${dir}/../../operator-framework/community-operators")
target_dir="${community_operators_dir}/${AREA}/${PACKAGE}"
rm -rf $target_dir
mkdir -p $target_dir

cp -r $dir/manifests/${PACKAGE}/${VERSION}/* $target_dir
cp $dir/manifests/${PACKAGE}/${PACKAGE}.package.yaml $target_dir
