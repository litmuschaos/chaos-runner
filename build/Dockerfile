FROM registry.access.redhat.com/ubi7-dev-preview/ubi-minimal:7.6

LABEL maintainer="LitmusChaos"

ENV RUNNER=/usr/local/bin/chaos-runner \
    USER_UID=1001 \
    USER_NAME=chaos-runner

COPY build/_output/bin/chaos-runner ${RUNNER}

ENTRYPOINT ["/usr/local/bin/chaos-runner"]

USER ${USER_UID}