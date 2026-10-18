# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p /app/logs

CMD ["./server"]
