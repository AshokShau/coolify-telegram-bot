FROM golang:1.24.4-alpine3.22 AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY . .

RUN go build -ldflags="-w -s" -o myapp .

FROM alpine:3.20.2

RUN apk --no-cache add ca-certificates && \
    apk update && apk upgrade --available && sync

COPY --from=builder /app/myapp /myapp

ENTRYPOINT ["/myapp"]
