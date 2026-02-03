FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate

RUN CGO_ENABLED=0 GOOS=linux go build -o bot main.go

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ca-certificates \
    zlib1g \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bot .
COPY --from=builder /app/libtdjson.so.* ./

CMD ["./bot"]