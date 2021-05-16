FROM golang:1.16

RUN apt-get update && apt-get install -y upx

ADD . ${GOPATH}/src/github.com/cbrand/mdnsforwarder

WORKDIR ${GOPATH}/src/github.com/cbrand/mdnsforwarder

