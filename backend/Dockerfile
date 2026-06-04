# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /app
ENV GOTOOLCHAIN=auto

# Install git for downloading dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build a statically linked, optimized Go executable
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o sprintgpt-server ./cmd/server

# Stage 2: Final lightweight execution container
FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/sprintgpt-server .

# Expose the API port
EXPOSE 8080

# Run the pre-compiled Go server binary
CMD ["./sprintgpt-server"]
