FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o twitch-mock ./cmd/main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache wget
COPY --from=builder /app/twitch-mock .
EXPOSE 7777 8081 3333
CMD ["./twitch-mock"]
