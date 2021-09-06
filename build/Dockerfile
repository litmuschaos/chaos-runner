# Multi-stage docker build
# Build stage
FROM golang:alpine AS builder

LABEL maintainer="LitmusChaos"

ARG TARGETPLATFORM=linux/amd64

ADD . /chaos-runner
WORKDIR /chaos-runner

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN CGO_ENABLED=0 go build -o /output/chaos-runner -v ./bin

# Image source: https://github.com/litmuschaos/test-tools/blob/master/custom/hardend-alpine/control-plane/Dockerfile
# The base image is non-root (have litmus user) with default litmus directory.
FROM litmuschaos/hardend-alpine

LABEL maintainer="LitmusChaos"

ENV RUNNER=/usr/local/bin/chaos-runner 

# copy CA certificates
COPY --from=builder /output/chaos-runner ${RUNNER}

ENTRYPOINT ["/usr/local/bin/chaos-runner"]
