# Testing Mode Configuration

## Storage Backend
- `TEST_STORE_BACKEND=postgres` tries to seed Postgres migrations and uses a real database for API tests.
- `TEST_STORE_BACKEND=memory` forces the in-memory repository (useful for CI when Postgres is unavailable).
- `TEST_DATABASE_URL` points to a custom Postgres DSN (defaults to the env-driven backend URL).
- When Postgres cannot be reached, the public test harness logs a hint and transparently falls back to memory.

## Admin MFA Simulation
- `TEST_REQUIRE_ADMIN_MFA=true` (default) requires test tokens to satisfy the MFA guard for admin flows.
- Set `TEST_REQUIRE_ADMIN_MFA=false` for faster automation when MFA is not part of the verification path.

## Connectivity Hints
- Run the backend locally before executing Playwright e2e tests (`npm run test:e2e`) or set `API_BASE_URL` to a reachable host.
- Use `playwright:install` to stage browser binaries ahead of CI builds.
