# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Using static linking for compatibility with distroless
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o api ./cmd/api

# Final stage using Google's distroless base image
FROM gcr.io/distroless/static-debian12

# Copy the binary from builder
COPY --from=builder /app/api /api

# Expose the port the API will run on
EXPOSE 8080

# Command to run (arguments need to be specified at container runtime)
ENTRYPOINT ["/api"]

# Default arguments
CMD ["--host", "0.0.0.0", "--port", "8080"]
