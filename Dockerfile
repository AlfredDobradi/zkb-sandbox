FROM golang:bookworm AS builder

WORKDIR /src
COPY . .
RUN go build -o dist/ticker ./client/...

FROM debian:bookworm

RUN apt-get update && apt-get install -y \
    ca-certificates

COPY --from=builder /src/dist/ticker /usr/local/bin/ticker
