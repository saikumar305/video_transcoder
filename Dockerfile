# Build stage
FROM golang:1.25-alpine AS builder

# Install FFmpeg
RUN apk add --no-cache ffmpeg git

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install FFmpeg and ca-certificates
RUN apk --no-cache add ca-certificates ffmpeg

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Create directories for video storage
RUN mkdir -p /root/videos/uploads /root/videos/transcoded

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release
ENV DB_HOST=localhost
ENV DB_PORT=5432
ENV DB_USER=postgres
ENV DB_PASSWORD=password
ENV DB_NAME=video_transcoder
ENV DB_SSLMODE=disable

CMD ["./main"]