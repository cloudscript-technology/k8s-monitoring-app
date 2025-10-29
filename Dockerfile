FROM golang:1.24-alpine AS builder
WORKDIR /app
# Enable CGO and install toolchain needed for github.com/mattn/go-sqlite3
RUN apk add --no-cache build-base
ENV CGO_ENABLED=1
COPY . .
RUN go mod download
RUN go build -o ./k8s-monitoring-app ./cmd/main.go

FROM alpine:3.22 AS release
WORKDIR /app
# Ensure TLS works (OAuth, k8s client, etc.)
RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/k8s-monitoring-app .
COPY database/migrations /app/database/migrations
# Copy web UI templates and static assets used by the handler
COPY web/templates /app/web/templates
COPY web/static /app/web/static

RUN chmod +x ./k8s-monitoring-app
RUN chown -R appuser:appgroup /app
USER appuser

CMD ["/app/k8s-monitoring-app"]