FROM golang:1.14 AS build-env
ENV GOPROXY=direct
WORKDIR /go/src/go-sqs-bench
ADD . /go/src/go-sqs-bench

RUN go get -d -v ./...
RUN go build -o /go/bin/sqs-bench

FROM ubuntu:18.04
RUN apt-get update && apt-get -y install ca-certificates
COPY --from=build-env /go/bin/sqs-bench /
CMD ["/sqs-bench"]
