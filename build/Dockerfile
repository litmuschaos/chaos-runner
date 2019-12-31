FROM ubuntu:18.04
# FROM registry.access.redhat.com/ubi7-dev-preview/ubi-minimal:7.6

LABEL maintainer="LitmusChaos"

ENV EXECUTOR=/usr/local/bin/chaos-executor

COPY build/_output/bin/chaos-executor ${EXECUTOR}

ENTRYPOINT ["/usr/local/bin/chaos-executor"]
