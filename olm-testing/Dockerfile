FROM alpine as munger

ARG new_image_name
COPY manifests manifests
RUN sed "s,quay.io/openshift/origin-federation-controller,$new_image_name," -i manifests/federation/0.0.6/federation.v0.0.6.clusterserviceversion.yaml
RUN sed "s,quay.io/openshift/origin-federation-controller,$new_image_name," -i manifests/cluster-federation/0.0.6/cluster-federation.v0.0.6.clusterserviceversion.yaml

FROM quay.io/openshift/origin-operator-registry:latest

COPY --from=munger manifests manifests
RUN initializer

CMD ["registry-server"]
