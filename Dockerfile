FROM golang:alpine as BUILD

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/metrumresearchgroup/gridengine_prometheus

COPY . .

RUN go get -d -v && \
    go build -o /go/bin/gridengine_exporter

EXPOSE 9081
ENTRYPOINT ["/go/bin/gridengine_exporter"]