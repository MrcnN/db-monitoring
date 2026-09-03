# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
ENV GOTOOLCHAIN=auto
# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata
# Copy project source
COPY . .
# Download dependencies and build binary
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o /app/bin/api ./cmd/api

# Stage 2: Runtime
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations
RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/api"]
