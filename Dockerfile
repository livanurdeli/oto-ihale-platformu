# Çok aşamalı Dockerfile (Multi-stage build)
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auction-worker cmd/auction-worker/main.go

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/server /app/server
COPY --from=builder /app/auction-worker /app/auction-worker
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/web /app/web

EXPOSE 8080
CMD ["/app/server"]

