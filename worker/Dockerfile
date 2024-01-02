# Start from a Golang base image
FROM golang:1.21 AS builder

# Set the working directory in the container
WORKDIR /app

# Copy the go.mod and go.sum and download the dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Use a minimal alpine image to run the API service
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080
EXPOSE 8080

# Run the binary executable
CMD ["./worker"]
