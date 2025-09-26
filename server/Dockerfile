# Build stage
FROM golang:1.25.1-alpine AS builder
WORKDIR /app
# Copy files needed for package manager first for more layers
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main ./
RUN chmod +x ./main
CMD ["./main"]
