# Multi-stage docker build
# Build stage
FROM golang:alpine AS builder

LABEL maintainer="LitmusChaos"

ARG TARGETPLATFORM

ADD . /chaos-runner
WORKDIR /chaos-runner

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN CGO_ENABLED=0 go build -buildvcs=false -o /output/chaos-runner -v ./bin

# Packaging stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.4

LABEL maintainer="LitmusChaos"

ENV RUNNER=/usr/local/bin/chaos-runner 

COPY --from=builder /output/chaos-runner ${RUNNER}
RUN chown 65534:0 ${RUNNER} && chmod 755 ${RUNNER}

USER 65534

ENTRYPOINT ["/usr/local/bin/chaos-runner"]
