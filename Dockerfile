# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scheduler ./cmd/scheduler

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/scheduler .
COPY --from=builder /app/config ./config

EXPOSE 8080 9090

CMD ["./scheduler"]