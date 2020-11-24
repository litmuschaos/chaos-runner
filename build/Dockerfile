FROM registry.access.redhat.com/ubi8/ubi-minimal:8.3

ARG TARGETARCH

LABEL maintainer="LitmusChaos"

ENV RUNNER=/usr/local/bin/chaos-runner \
    USER_UID=1001 \
    USER_NAME=chaos-runner

COPY build/_output/bin/chaos-runner-${TARGETARCH} ${RUNNER}

ENTRYPOINT ["/usr/local/bin/chaos-runner"]

USER ${USER_UID}
