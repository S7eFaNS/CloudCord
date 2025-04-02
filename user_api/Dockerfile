
# Stage 1: Build the Go binary
FROM golang:1.24.1 AS builder

WORKDIR /user_api

COPY go.mod ./
RUN go mod download

COPY *.go ./

# Build the binary with CGO disabled and stripped of debugging symbols to reduce size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /user

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

# Install any necessary dependencies (if needed, e.g., certificates)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the Go binary from the builder stage to the final image
COPY --from=builder /user /user

EXPOSE 8081

CMD ["/user"]

