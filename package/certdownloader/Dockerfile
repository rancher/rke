FROM alpine:3.4

RUN apk update && apk add bash

COPY entrypoint.sh /tmp/entrypoint.sh
CMD ["/tmp/entrypoint.sh"]
