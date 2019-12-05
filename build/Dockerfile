FROM ubuntu:18.04

LABEL maintainer="LitmusChaos"

ENV EXECUTOR=/usr/local/bin/chaos-executor

COPY build/_output/bin/chaos-executor ${EXECUTOR}

ENTRYPOINT ["/usr/local/bin/chaos-executor"]