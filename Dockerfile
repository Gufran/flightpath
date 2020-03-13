FROM golang:1.14 as base
ENV CGO_ENABLED=0

ADD . /src
WORKDIR /src

RUN /bin/bash build.sh native

FROM alpine:latest as certs
RUN apk update && apk add ca-certificates

FROM scratch
COPY --from=base /src/flightpath /flightpath
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/flightpath"]
