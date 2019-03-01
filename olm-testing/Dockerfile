FROM alpine as munger

ARG new_image_name
COPY manifests manifests
RUN sed "s,quay.io/openshift/origin-federation-controller,$new_image_name," -i manifests/federation/0.0.5/federation.v0.0.5.clusterserviceversion.yaml

FROM quay.io/openshift/origin-operator-registry:latest

COPY --from=munger manifests manifests
RUN initializer

CMD ["registry-server"]
