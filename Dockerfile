# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/desa-agent ./cmd/app


FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -g '' appuser

RUN mkdir -p /app/data && chown -R appuser:appuser /app

COPY --from=builder /app/desa-agent /app/desa-agent

USER appuser

ENV STORAGE_PATH=/app/data

EXPOSE 50051

ENTRYPOINT ["/app/desa-agent"]

