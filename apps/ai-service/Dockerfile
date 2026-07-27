# Multi-stage build for the Go ai-service.
#
# The image carries both entrypoints — the HTTP API and the triage worker —
# because compose runs them as two services from the same image with different
# commands.
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Download dependencies first so the layer is cached when only sources change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/api    ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/ingest ./cmd/ingest

FROM alpine:3.19

RUN apk --no-cache add ca-certificates wget

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app/

COPY --from=builder /out/api /out/worker /out/ingest ./
# The API applies these on startup (kb_chunks and its indexes).
COPY --from=builder /app/migrations ./migrations

RUN chown -R appuser:appgroup /app/

USER appuser

EXPOSE 8081

CMD ["./api"]
