#!/bin/sh
set -e

# Start all services in the background.
docker compose up -d --build

# Wait for the database to accept connections.
echo "Waiting for database..."
TRIES=0
until docker compose exec w1-t1-ti1-db pg_isready -U fleetlease > /dev/null 2>&1; do
  TRIES=$((TRIES + 1))
  if [ "$TRIES" -ge 30 ]; then
    echo "ERROR: database did not become ready in time"
    docker compose logs w1-t1-ti1-db
    exit 1
  fi
  sleep 2
done
echo "Database is ready."

# Wait for the backend to serve the health endpoint (up to 90 s).
echo "Waiting for backend..."
TRIES=0
until docker compose exec w1-t1-ti1-backend \
    curl -sfk https://localhost:8080/health > /dev/null 2>&1; do
  TRIES=$((TRIES + 1))
  if [ "$TRIES" -ge 45 ]; then
    echo "ERROR: backend did not become healthy in time"
    docker compose logs w1-t1-ti1-backend
    exit 1
  fi
  sleep 2
done
echo "Backend is ready."

# Wait for the frontend dev server to become reachable over HTTP.
echo "Waiting for frontend..."
TRIES=0
until docker compose exec w1-t1-ti1-frontend \
  curl -sf http://127.0.0.1:5173 > /dev/null 2>&1; do
  TRIES=$((TRIES + 1))
  if [ "$TRIES" -ge 45 ]; then
    echo "ERROR: frontend did not become ready in time"
    docker compose logs w1-t1-ti1-frontend
    exit 1
  fi
  sleep 2
done
echo "Frontend is ready."

# Run ALL backend tests inside a fresh container that shares the compose
# network.  TEST_SERVER_URL points to the already-running backend service so
# that the live/* tests can make real network HTTP calls against it.
docker compose run --rm \
  -e TEST_SERVER_URL=https://w1-t1-ti1-backend:8080 \
  -e TEST_ADMIN_TOTP_SECRET=JBSWY3DPEHPK3PXP \
  w1-t1-ti1-backend \
  go test -v -count=1 -timeout 300s ./...

# Run frontend unit tests.
docker compose run --rm w1-t1-ti1-frontend npm run test

# Run frontend e2e tests with explicit project/spec targeting.
# This avoids accidental cross-project matching and enforces HTTP for the Vite UI server.
docker compose exec w1-t1-ti1-frontend \
  npx playwright test tests/e2e/booking.spec.js --project=api

docker compose exec -e FRONTEND_BASE_URL=http://127.0.0.1:5173 w1-t1-ti1-frontend \
  npx playwright test tests/e2e/app.ui.spec.js --project=ui

# Tear down all services.
docker compose down
