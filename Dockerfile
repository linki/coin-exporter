FROM golang:1.8-alpine

ENV PROJECT_ROOT github.com/earthcoinproject/coin-exporter

COPY . /go/src/${PROJECT_ROOT}
RUN go install -v ${PROJECT_ROOT}

ENTRYPOINT ["/go/bin/coin-exporter"]
