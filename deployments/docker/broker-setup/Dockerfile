FROM alpine:3.14

RUN apk --no-cache add bash
RUN apk --no-cache add curl

COPY deployments/docker/broker-setup/setup.sh /scripts/setup.sh

ENTRYPOINT ["/scripts/setup.sh"]