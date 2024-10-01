# syntax=docker/dockerfile:1

FROM golang:1.22.5-alpine AS base

WORKDIR /app

ADD . /app
RUN go mod download
RUN go build -o /docker-commenteer

EXPOSE 8090

CMD ["/docker-commenteer"]