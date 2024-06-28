# Define the build stage
FROM golang:1.22.4-alpine as builder

RUN apk add --no-cache git

# Set the Current Working Directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum to download dependencies.
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files do not change.
RUN go mod download

# Copy the source code into the container.
COPY . .

# Build the Go app as a static binary.
# 'CGO_ENABLED=0' is required to build a statically-linked executable that is fully self-contained.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o miraibox ./cmd/http/

# Start a new, final image to reduce size.
FROM alpine:latest  

# Add ca-certificates in case you need HTTPS.
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container for the runtime image.
WORKDIR /root/

# Copy the pre-built binary file from the previous stage.
COPY --from=builder /app/miraibox .

# Expose port 8080 to the outside world.
EXPOSE 8080

# Command to run the executable, use the array form to ensure signals are properly handled.
CMD ["./miraibox"]