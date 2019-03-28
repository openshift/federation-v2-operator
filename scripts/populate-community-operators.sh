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

# PACKAGE is the specific OLM package being deployed - federation or cluster-federation
PACKAGE=${PACKAGE:-"federation"}
# VERSION is the version of PACKAGE you want to populate in a community-operators fork
VERSION=${VERSION:-"0.0.7"}
# SOURCE is which source manifest you want to move - upstream-manifests or manifests
SOURCE=${SOURCE:-"upstream-manifests"}
# AREA is the place you want the manifests to go within community-operators,
# community-operators or upstream-community-operators
AREA=${AREA:-"community-operators"}

dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")
community_operators_dir=$(realpath "${dir}/../../operator-framework/community-operators")
target_dir="${community_operators_dir}/${AREA}/${PACKAGE}"
rm -rf $target_dir
mkdir -p $target_dir

cp -r $dir/${SOURCE}/${PACKAGE}/${VERSION}/* $target_dir
cp $dir/${SOURCE}/${PACKAGE}/${PACKAGE}.package.yaml $target_dir
