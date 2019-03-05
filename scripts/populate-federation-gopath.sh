#!/bin/bash -eu

# This script populates the federation-v2 directory within a gopath directory
# structure with the vendored federation-v2 source from this repo. It can be run
# from any directory. The gopath directory structure is assumed to be rooted at
# '/go' but can be overriden by setting the GOPATH_DIR environment variable.

dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")
gopath_dir="${GOPATH_DIR:-"/go"}"

federation_gp=github.com/kubernetes-sigs/federation-v2
federation_src="${gopath_dir}/src/${federation_gp}"
vendor_src="${dir}/vendor/${federation_gp}"

echo "Populating federation-v2 gopath at ${federation_src}"
mkdir -p $federation_src
cp $vendor_src/Makefile $federation_src/Makefile
cp -r $vendor_src/pkg $federation_src/pkg
cp -r $vendor_src/cmd $federation_src/cmd
cp -r $vendor_src/test $federation_src/test
cp -r $vendor_src/hack $federation_src/hack
cp -r $vendor_src/scripts $federation_src/scripts
cp -r $vendor_src/config $federation_src/config
cp -r vendor $federation_src/vendor
