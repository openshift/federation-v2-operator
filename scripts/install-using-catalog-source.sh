#!/bin/bash

# This script automates the work of installing federation using 

oc create -f olm-testing/catalog-source.yaml
oc create -f olm-testing/operator-group.yaml
oc create -f olm-testing/subscription.yaml
