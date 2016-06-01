FROM alpine:3.3

RUN \
  apk add --update go git make gcc musl-dev gmp-dev gmp && \
  git clone https://github.com/shiftcurrency/shift     && \
  (cd shift && make gshift)                         && \
  cp shift/build/bin/gshift /gshift                   && \
  apk del go git make gcc musl-dev gmp-dev              && \
  rm -rf /shift && rm -rf /var/cache/apk/*

EXPOSE 53901
EXPOSE 53900

ENTRYPOINT ["/gshift"]
