# Build Geth in a stock Go builder container
FROM golang:1.17-alpine3.16 as builder

RUN apk add --no-cache make cmake gcc musl-dev linux-headers git build-base libc-dev bash

ADD . /go-ethereum

RUN cd /go-ethereum && make geth


# Pull all binaries into a second stage deploy alpine container
FROM alpine:3.16

RUN apk add ca-certificates jq  bash bind-tools tini grep curl sed gcc

COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp
ENTRYPOINT ["geth"]
