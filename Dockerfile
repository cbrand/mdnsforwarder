FROM golang:1.23

RUN printf "deb http://httpredir.debian.org/debian bookworm-backports main non-free\ndeb-src http://httpredir.debian.org/debian bookworm-backports main non-free" > /etc/apt/sources.list.d/backports.list
RUN apt-get update && apt-get install -y upx-ucl

ADD . ${GOPATH}/src/github.com/cbrand/mdnsforwarder

WORKDIR ${GOPATH}/src/github.com/cbrand/mdnsforwarder

