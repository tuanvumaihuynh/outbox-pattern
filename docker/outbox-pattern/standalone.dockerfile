FROM golang:1.25 AS builder

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build -o main cmd/op-standalone/main.go


FROM alpine:3.22 AS prod

WORKDIR /app

RUN addgroup -S op && adduser -S op -G op

COPY --chown=op:op --from=builder /app/main /app/main

USER op

ENTRYPOINT ["/app/main"]
