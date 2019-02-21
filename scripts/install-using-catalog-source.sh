#!/bin/bash

# This script automates the work of installing federation using the
# CatalogSource in the `olm-testing` directory.
#
# This script can be run from any directory.

dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")

oc create -f $dir/olm-testing/catalog-source.yaml
oc create -f $dir/olm-testing/operator-group.yaml
oc create -f $dir/olm-testing/subscription.yaml
