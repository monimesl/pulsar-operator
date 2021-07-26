FROM alpine:3.14

RUN apk --no-cache add bash
RUN apk --no-cache add curl

COPY setup.sh /prepare/setup.sh

ENTRYPOINT ["/prepare/setup.sh"]