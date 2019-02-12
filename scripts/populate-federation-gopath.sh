#!/bin/bash

federation_gp=github.com/kubernetes-sigs/federation-v2
federation_src="/go/src/${federation_gp}"
vendor_src="vendor/${federation_gp}"
mkdir -p $federation_src
cp $vendor_src/Makefile $federation_src/Makefile
cp -r $vendor_src/pkg $federation_src/pkg
cp -r $vendor_src/cmd $federation_src/cmd
cp -r $vendor_src/test $federation_src/test
cp -r $vendor_src/hack $federation_src/hack
cp -r $vendor_src/scripts $federation_src/scripts
cp -r $vendor_src/config $federation_src/config
cp -r vendor $federation_src/vendor
