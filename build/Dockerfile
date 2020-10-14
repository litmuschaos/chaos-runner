# BUILD STAGE
FROM golang:1.14 AS builder

LABEL maintainer="LitmusChaos"

ENV RUNNER=/usr/local/bin/chaos-runner

ADD . ${RUNNER}

WORKDIR ${RUNNER}

RUN go build -o build/_output/bin/chaos-runner ./bin

## DEPLOY STAGE
FROM registry.access.redhat.com/ubi7-dev-preview/ubi-minimal:7.6

LABEL maintainer="LitmusChaos"

ENV RUNNER=/usr/local/bin/chaos-runner  \
    USER_UID=1001 \
    USER_NAME=chaos-runner

COPY --from=builder ${RUNNER}/build/_output/bin/chaos-runner ${RUNNER}/binary

ENTRYPOINT ["/usr/local/bin/chaos-runner/binary"]

USER ${USER_UID}
