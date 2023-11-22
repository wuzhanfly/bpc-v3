# Build Geth in a stock Go builder container
FROM golang:1.17-alpine3.16 as builder

RUN apk add --no-cache make cmake gcc musl-dev linux-headers git build-base libc-dev bash

ADD . /go-ethereum
RUN cd /go-ethereum && make geth

# Pull Geth into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates curl jq bash bind-tools grep sed gcc tini
COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

EXPOSE 8545 8546 8547 30303 30303/udp
ENTRYPOINT ["geth"]