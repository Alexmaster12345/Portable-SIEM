FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy module file and download deps (go.sum generated automatically if missing)
COPY go.mod ./
RUN go mod download && go mod tidy || true

# Copy full source and build
COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o siem-server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/siem-server .

EXPOSE 8080 514/udp 9000
CMD ["/app/siem-server"]
