# This Dockerfile represents a multistage build. The stages, respectively:
#
# 1. build federation binaries
# 2. copy binaries in, add OLM manifests, labels, etc

# build stage 1: build federation binaries
FROM openshift/origin-release:golang-1.10 as builder
RUN yum update -y
RUN yum install -y make git

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /go/src/github.com/kubernetes-sigs/federation-v2

# copy in git info to support recording version in binaries
COPY .git/HEAD .git/HEAD
COPY .git/refs/heads/. .git/refs/heads
RUN mkdir -p .git/objects

COPY vendor/github.com/kubernetes-sigs/federation-v2/Makefile Makefile
COPY vendor/github.com/kubernetes-sigs/federation-v2/pkg pkg
COPY vendor/github.com/kubernetes-sigs/federation-v2/cmd cmd
COPY vendor/github.com/kubernetes-sigs/federation-v2/test test
COPY vendor vendor
RUN rm -rf vendor/github.com/kubernetes-sigs/federation-v2
RUN ls -l cmd/hyperfed
# HACK: DOCKER_CMD is set here to workaround the use of the docker command in the federation-v2 Makefile
RUN DOCKER_BUILD="echo && " make hyperfed controller kubefed2

# build stage 2: copy in binaries, add OLM manifest, labels, etc.
FROM openshift/origin-base

# TODO: copy in OLM manifests
# ADD manifests/ /manifests

# copy in binaries
COPY --from=builder /go/src/github.com/kubernetes-sigs/federation-v2/bin/controller-manager

# user directive - this image does not require root
USER 1001

# apply labels to final image
LABEL io.k8s.display-name="OpenShift Federation v2" \
      io.k8s.description="This is a component that allows management of Kubernetes/OpenShift resources across multiple clusters" \
      maintainer="AOS Multicluster Team <aos-multicluster@redhat.com>"