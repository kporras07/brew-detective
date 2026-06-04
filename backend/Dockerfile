# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

RUN adduser -D -u 1000 appuser
WORKDIR /home/appuser

# Copy the binary from builder stage
COPY --from=builder /app/main .

USER appuser

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]