# Multi-stage docker build
# Build stage
FROM golang:1.16 AS builder

LABEL maintainer="LitmusChaos"

ARG TARGETPLATFORM

ADD . /chaos-runner
WORKDIR /chaos-runner

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

RUN go env

RUN CGO_ENABLED=0 go build -o /output/chaos-runner -v ./bin

FROM golang:alpine as cert
RUN apk --no-cache add ca-certificates

FROM scratch

LABEL maintainer="LitmusChaos"

ENV RUNNER=/usr/local/bin/chaos-runner \
    USER_UID=1001 \
    USER_NAME=chaos-runner

# copy CA certificates
COPY --from=cert /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /output/chaos-runner ${RUNNER}

ENTRYPOINT ["/usr/local/bin/chaos-runner"]

USER ${USER_UID}
