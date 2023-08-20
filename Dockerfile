# Stage 1: Compiling the application
FROM golang:1.17 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Navigate to the cmd directory where the main.go file resides
WORKDIR /app/cmd

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Stage 2: Build the minimal docker image
FROM scratch

WORKDIR /root/

# Copy the pre-built binary from the previous stage
COPY --from=builder /app/cmd/app .

# Command to run
CMD ["./app"]
