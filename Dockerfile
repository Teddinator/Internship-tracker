# syntax=docker/dockerfile:1
# Builder image
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -o /app/bin/api \
    ./cmd/api

RUN CGO_ENABLED=0 GOBIN=/app/bin go install \
    github.com/pressly/goose/v3/cmd/goose@v3.27.2

# Runtime image
FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/goose /usr/local/bin/goose
COPY --from=builder /app/migrations /app/migrations

USER app

EXPOSE 8080

CMD ["/app/api"]