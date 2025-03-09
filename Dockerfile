# syntax=docker/dockerfile:1

# Stage 1: Build the Go binary.
FROM golang:1.24 AS builder

# Set the working directory inside the container.
WORKDIR /app

# Cache dependencies by copying go.mod and go.sum.
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container.
COPY . .

# Build the binary from the main package in the cmd directory.
RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o echo-api ./cmd

# Stage 2: Create a lean production image.
FROM alpine:latest

# Install CA certificates (if your API uses HTTPS or calls external services).
RUN apk --no-cache add ca-certificates

# Set the working directory in the final image.
WORKDIR /root/

# Copy the compiled binary from the builder stage.
COPY --from=builder /app/echo-api .

# Expose the port on which your API listens (adjust if needed).
EXPOSE 8080

# Start the application.
CMD ["./echo-api"]
