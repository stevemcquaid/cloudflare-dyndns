FROM BASEIMAGE

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY run /

USER 1001:1001

CMD ["/run"]
