# --- Stage 1: The "builder" ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /linkpulse-worker .

# --- Stage 2: The "final" image ---
# Change FROM scratch to FROM alpine:latest
FROM alpine:latest

WORKDIR /

# Copy the ca-certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary from the builder stage
COPY --from=builder /linkpulse-worker /linkpulse-worker

ENTRYPOINT ["/linkpulse-worker"]