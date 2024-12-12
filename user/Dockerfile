# Stage 1: Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go modules and source code
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the entire project to the container
COPY . .

# Build the Go application
RUN GOARCH=amd64 GOOS=linux go build -o app .

# Stage 2: Final stage (minimal image)
FROM alpine:latest

# Install necessary dependencies for running the app
RUN apk add --no-cache ca-certificates

# Set working directory in the final image
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/app .

# Expose the port your app will run on
EXPOSE 8081

# Menjalankan aplikasi Go
CMD ["/root/app"]