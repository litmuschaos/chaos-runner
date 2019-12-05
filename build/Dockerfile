FROM ubuntu:18.04

LABEL maintainer="LitmusChaos"

<<<<<<< HEAD
COPY ./exec /usr/local/bin/exec

ENTRYPOINT ["exec"]
=======
ENV EXECUTOR=/usr/local/bin/chaos-executor

COPY build/_output/bin/chaos-executor ${EXECUTOR}

ENTRYPOINT ["/usr/local/bin/chaos-executor"]
>>>>>>> f4d3ecaad25fc7d819dac4b280e97320b7803b28
