# ---------- Stage 1: Build ----------
FROM golang:1.23-alpine AS builder

# Required for some builds (especially CGO off)
ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app

# Only copy go mod files first (this layer is cacheable)
COPY go.mod go.sum ./
RUN go mod download

# Copy only the necessary source files
COPY . .

# Build the binary
RUN go build -trimpath -ldflags="-s -w" -o storage-service ./cmd/main.go

# ---------- Stage 2: Runtime ----------
FROM alpine:latest

WORKDIR /root/

# Copy binary only — small image
COPY --from=builder /app/storage-service .

# Expose the app port
EXPOSE 8080

# Run the app
CMD ["./storage-service"]
