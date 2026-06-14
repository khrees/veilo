# Stage 1: Build stage
FROM golang:1.22-alpine AS builder

# Install git and certificates for dependency resolution
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency specifications and download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build the statically compiled binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o veilo .

# Stage 2: Final lightweight execution environment
FROM alpine:latest

# Install certificates and timezone data (critical for secure connections and logs)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/veilo .

# Expose port (Render/Railway injects PORT env to bind to dynamically)
EXPOSE 8084

# Run the app as server by default
ENTRYPOINT ["./veilo"]
CMD ["server"]
