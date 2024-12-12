# Stage 1: Build stage
FROM golang:1.23-alpine AS builder

# Set working directory dalam container
WORKDIR /app

# Salin file go.mod dan go.sum untuk menginstal dependensi terlebih dahulu
COPY go.mod go.sum ./

# Instal dependensi
RUN go mod tidy

# Salin kode sumber ke dalam container
COPY . .

# Build aplikasi Go
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o app .

# Gunakan base image yang minimal untuk runtime
FROM alpine:latest

# Install dependensi yang dibutuhkan oleh aplikasi Go
RUN apk add --no-cache ca-certificates

# Set working directory untuk aplikasi runtime
WORKDIR /root/

# Salin binary hasil build dari stage builder
COPY --from=builder /app/app .

# Copy the config directory into the container
COPY ./config /root/config

# Change the permissions of the .env.k8s file
RUN chmod 644 /root/config/.env.k8s

# Verifikasi apakah file .env.k8s ada di dalam container
RUN ls -l /root/config

# Define port yang akan digunakan
EXPOSE 8082

# Set environment variable default (bisa di-overwrite saat runtime)
ENV ENV=k8s

# Menjalankan aplikasi dengan environment tertentu
CMD ["/root/app"]