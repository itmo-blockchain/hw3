# syntax=docker/dockerfile:1

FROM golang:1.19.1-alpine

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
RUN go mod tidy

COPY . ./

RUN go build -o /monitor ./

ENTRYPOINT [ "/monitor" ]