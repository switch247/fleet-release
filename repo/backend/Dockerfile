FROM golang:1.22-alpine AS builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go build -o /fleetlease-server ./cmd/server

FROM golang:1.22-alpine
WORKDIR /app/backend
RUN apk add --no-cache postgresql-client curl openssl
RUN mkdir -p /app/certs && openssl req -x509 -newkey rsa:2048 -sha256 -nodes -keyout /app/certs/server.key -out /app/certs/server.crt -days 3650 -subj "/CN=localhost"
COPY --from=builder /fleetlease-server /app/fleetlease-server
COPY backend/ .
COPY backend/migrations /app/migrations

EXPOSE 8080
CMD ["/app/fleetlease-server"]
