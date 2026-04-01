#!/bin/sh
set -e

# Start services
docker compose up -d --build

# Wait for services to be healthy
docker compose exec w1-t1-ti1-db pg_isready -U fleetlease
docker compose exec w1-t1-ti1-backend curl -f http://localhost:8080/health || echo "Backend not ready, continuing..."

# Run backend tests
docker compose run --rm w1-t1-ti1-backend go test ./...

# Run frontend unit tests
docker compose run --rm w1-t1-ti1-frontend npm run test
# Run frontend e2e tests
docker compose exec w1-t1-ti1-frontend npx playwright test
# Stop services
docker compose down
