# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency definition files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o note-taker-backend .

# Run stage
FROM alpine:latest

# Install ffmpeg, tzdata, and ca-certificates
RUN apk add --no-cache ffmpeg tzdata ca-certificates

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/note-taker-backend .

# Expose the default backend port
EXPOSE 8000

# Set environment variables
ENV PORT=8000
ENV GIN_MODE=release

# Run the backend binary
CMD ["./note-taker-backend"]
