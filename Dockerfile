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

# Use a minimal alpine image to run the application
FROM alpine:latest
#RUN apk --nocache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage and environment variable file
COPY --from=builder /app/main .
COPY .env /root/.env
COPY /pkg/db/migrations /migrations

# Expose port 8080 for the API service
EXPOSE 9090

# The ENTRYPOINT defines the initial command that gets executed when the container starts
ENTRYPOINT ["./main"]

# Default command if no arguments are supplied to docker run
CMD ["api"]  # By default it starts the API service
