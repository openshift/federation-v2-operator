# This Dockerfile represents a multistage build. The stages, respectively:
#
# 1. build federation binaries
# 2. copy binaries in, add OLM manifests, labels, etc

# build stage 1: build federation binaries
FROM openshift/origin-release:golang-1.11 as builder
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

# HACK: DOCKER_BUILD is set here to workaround the use of the docker command in the federation-v2 Makefile
# HACK: GIT_VERSION is set explicitly due to an issue with how the .git directory is copied in
RUN DOCKER_BUILD="/bin/sh -c " GIT_VERSION="0.0.10" make hyperfed

# build stage 2: copy in binaries, add OLM manifest, labels, etc.
FROM openshift/origin-base

# copy in binaries
WORKDIR /root/
COPY --from=builder /go/src/github.com/kubernetes-sigs/federation-v2/bin/hyperfed-linux /root/hyperfed
RUN ln -s hyperfed controller-manager && ln -s hyperfed kubefed2

# user directive - this image does not require root
USER 1001

# TODO: copy in OLM manifests
# ADD manifests/ /manifests

ENTRYPOINT ["/root/controller-manager"]
CMD ["--install-crds=false"]

# apply labels to final image
LABEL io.k8s.display-name="OpenShift Federation v2" \
      io.k8s.description="This is a component that allows management of Kubernetes/OpenShift resources across multiple clusters" \
      maintainer="AOS Multicluster Team <aos-multicluster@redhat.com>"