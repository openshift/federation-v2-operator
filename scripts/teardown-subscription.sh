#!/bin/bash

oc delete -n federation-test subscriptions --all
oc delete -n federation-test installplans --all
oc delete -n federation-test clusterserviceversions --all
