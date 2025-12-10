# syntax=docker/dockerfile:1

FROM golang:1.17-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ./frontend

FROM alpine:3.15.0
RUN apk add --no-cache curl openssl ca-certificates
WORKDIR /app
COPY --from=builder /app/frontend /bin/frontend
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
ENV GIN_MODE=release
EXPOSE 6060
ENTRYPOINT ["/bin/frontend"]
