# --- Stage 1: The "builder" ---
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /linkpulse-server .

# --- Stage 2: The "final" image ---
FROM alpine:latest
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /linkpulse-server /linkpulse-server
EXPOSE 8080
ENTRYPOINT ["/linkpulse-server"]