# FROM golang:1.19-alpine AS builder

# ARG VERSION="nightly"

# RUN apk add --update git
# RUN mkdir -p src/github.com/nats-io && \
#     cd src/github.com/nats-io/ && \
#     git clone https://github.com/nats-io/natscli.git && \
#     cd natscli/nats && \
#     go build -ldflags "-w -X main.version=${VERSION}" -o /nats

# RUN go install github.com/nats-io/nsc/v2@latest

# FROM alpine:latest

# RUN apk add --update ca-certificates && mkdir -p /nats/bin && mkdir /nats/conf

# COPY docker/nats-server.conf /nats/conf/nats-server.conf
# COPY nats-server /bin/nats-server
# COPY --from=builder /nats /bin/nats
# COPY --from=builder /go/bin/nsc /bin/nsc

# EXPOSE 4222 8222 6222 5222

# ENTRYPOINT ["/bin/nats-server"]
# CMD ["-c", "/nats/conf/nats-server.conf"]

FROM golang:1.18-bullseye as base

WORKDIR /app
ENV GO111MODULE=on CGO_ENABLED=0
COPY . .
RUN go build -o /app/main /app/main.go

FROM alpine:3.16

WORKDIR /app
COPY --from=base /app/main /app/main
EXPOSE 1323
ENTRYPOINT [ "/bin/sh", "-c", "/app/main"]