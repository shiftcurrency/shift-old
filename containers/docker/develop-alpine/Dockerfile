FROM alpine:3.3

RUN \
  apk add --update go git make gcc musl-dev         && \
  git clone https://github.com/shiftcurrency/shift && \
  (cd shift && git checkout develop)          && \
  (cd shift && make geth)                     && \
  cp shift/build/bin/gshift /gshift               && \
  apk del go git make gcc musl-dev                  && \
  rm -rf /shift && rm -rf /var/cache/apk/*

EXPOSE 54901
EXPOSE 54900

ENTRYPOINT ["/gshift"]
