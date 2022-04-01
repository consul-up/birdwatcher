# syntax=docker/dockerfile:1

FROM golang:1.17-alpine as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ./frontend

FROM alpine:3.15.0
RUN apk add --no-cache curl
RUN apk add --no-cache openssl
COPY --from=builder /app/frontend /bin/frontend
ENV GIN_MODE=release
ENTRYPOINT ["/bin/frontend"]
