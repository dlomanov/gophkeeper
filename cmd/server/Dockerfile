# Use an official Go runtime as a parent image
FROM golang:1.22.3 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.* ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the local code to the container's workspace
COPY /internal ./internal
COPY /cmd/server ./cmd/server
COPY /cmd/server/cert_test ./build

## Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o ./build/server ./cmd/server/main.go

# Use a smaller image for the runtime
FROM alpine:latest

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/build/ ./
RUN chmod +x /server

ENTRYPOINT ["/server"]
