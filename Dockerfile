# === STAGE 1: Build the Go Application ===
FROM golang:1.23-alpine AS builder

# Set necessary environment variables
ENV CGO_ENABLED=0
ENV GOOS=linux

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files and download dependencies
# This is cached, so rebuilds are faster if code changes but dependencies don't
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# The output binary is named 'main'
RUN go build -ldflags="-w -s" -o /goapp main.go

# === STAGE 2: Create the Final Production Image ===
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /goapp .

# Expose the port your application listens on
# (The Go code sets this to 8080)
EXPOSE 8080

# The command to run when the container starts
CMD ["/app/goapp"]