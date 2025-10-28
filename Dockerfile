FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ./k8s-monitoring-app ./cmd/main.go

FROM alpine:3.22 AS release
WORKDIR /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/k8s-monitoring-app .
COPY database/migrations /app/database/migrations

RUN chmod +x ./k8s-monitoring-app
RUN chown -R appuser:appgroup /app
USER appuser

CMD ["/app/k8s-monitoring-app"]