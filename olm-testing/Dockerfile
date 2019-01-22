FROM quay.io/openshift/origin-operator-registry:latest

COPY manifests manifests
RUN initializer

CMD ["registry-server"]