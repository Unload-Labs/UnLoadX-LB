FROM golang

ENV GOPATH=/go
WORKDIR $GOPATH

RUN mkdir -p $GOPATH/src/github.com/aebrow4/unloadx-lb
WORKDIR $GOPATH/src/github.com/aebrow4/unloadx-lb
ADD . .
RUN go build loadbalancer/loadbalancer.go
RUN go install
RUN go get -d -v

EXPOSE 9000
EXPOSE 9090
