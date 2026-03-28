FROM golang:1.23-alpine AS builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go build -o /fleetlease-server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /fleetlease-server /fleetlease-server
COPY backend/migrations /app/migrations
COPY run_tests.sh /app/run_tests.sh
RUN chmod +x /app/run_tests.sh
EXPOSE 8080
CMD ["/fleetlease-server"]
