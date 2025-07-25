# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates for go mod download
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o policy-bot-tests .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/policy-bot-tests .

# Copy any necessary config files
COPY --from=builder /app/.policy.yml ./
COPY --from=builder /app/.policy-tests.yml ./

# Make the binary executable
RUN chmod +x policy-bot-tests

# Set the entrypoint
ENTRYPOINT ["./policy-bot-tests"] 