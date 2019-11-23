FROM ubuntu:18.04

LABEL maintainer="LitmusChaos"

<<<<<<< HEAD
COPY ./exec /usr/local/bin/exec

ENTRYPOINT ["exec"]
=======
ENV EXECUTOR=/usr/local/bin/chaos-executor

COPY build/_output/bin/chaos-executor ${EXECUTOR}

ENTRYPOINT ["/usr/local/bin/chaos-executor"]
>>>>>>> 7fc58d356d7f488f50b0af0134e6d881b469225b
