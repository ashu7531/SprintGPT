FROM golang:alpine

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

# Expose the API port
EXPOSE 8080

# Run the Go server
CMD ["go", "run", "./cmd/server"]
