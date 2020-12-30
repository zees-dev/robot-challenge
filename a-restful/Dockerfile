FROM golang:alpine as build
LABEL maintainer="github.com/zees-dev"

WORKDIR /go/src/app

COPY . .
RUN go mod download


WORKDIR /go/src/app/a-restful
RUN go build -o ./app

EXPOSE 8000

ENTRYPOINT ["./app"]
