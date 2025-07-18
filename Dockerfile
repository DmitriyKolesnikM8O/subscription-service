# Stage 1: Modules caching
FROM golang:1.24-alpine as modules
WORKDIR /modules
COPY go.mod go.sum ./
RUN go mod download

# Stage 2: Builder
FROM golang:1.24-alpine as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

# Build the application
RUN go build -o /bin/subscription-service ./cmd/main

FROM scratch
COPY --from=builder /bin/subscription-service .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config ./config
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY .env .

CMD ["./subscription-service"]