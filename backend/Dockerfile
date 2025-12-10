# syntax=docker/dockerfile:1

FROM golang:1.17-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ./backend

FROM alpine:3.15.0
RUN apk add --no-cache curl openssl ca-certificates
WORKDIR /app
COPY --from=builder /app/backend /bin/backend
COPY --from=builder /app/*.json ./
ENV GIN_MODE=release
EXPOSE 7000
ENTRYPOINT ["/bin/backend"]
