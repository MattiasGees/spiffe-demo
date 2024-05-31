FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine

RUN apk add --no-cache bash
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
