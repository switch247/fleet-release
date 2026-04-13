# FleetLease Follow-up Audit Status (Against audit_report-1.md)

Date: 2026-04-13  
Method: Static verification of current repository state + targeted backend test execution.

## Overall Follow-up Verdict
- Result: Pass (with minor residual risk)
- Previously reported issues reviewed: 9
- Fully fixed: 8
- Partially fixed: 1
- Not fixed: 0

## Targeted Runtime Checks Executed
1. `go test ./tests/API_tests/integration -run "TestSettlementIncludesWearDeductions" -count=1` from `repo/backend` -> `ok`
2. `go test ./tests/API_tests/live -run "TestSyncAndExport" -count=1` from `repo/backend` -> `ok`
3. `npm test -- --run tests/unit/queue.test.js` from `repo/frontend` -> `10 passed`

## Issue-by-Issue Revalidation

### 1) Blocker: Predictable default privileged credentials seeded in runtime startup path
- Previous status: Fail
- Current status: Fixed
- Evidence:
  - `repo/backend/cmd/server/main.go:152` seeding gated by `BOOTSTRAP_SEED=true`.
  - `repo/backend/cmd/server/main.go:160` and `repo/backend/cmd/server/main.go:166` require explicit operator-provided bootstrap passwords.
- Assessment:
  - Hardcoded default credentials are no longer seeded unconditionally at startup.

### 2) High: Sensitive field handling for seeded users not stored as encrypted-at-rest values
- Previous status: Fail
- Current status: Fixed
- Evidence:
  - `repo/backend/cmd/server/main.go:204` stores `GovernmentIDEnc: encGovID` where `encGovID` comes from AES-256 encryption.
- Assessment:
  - Seed path now persists encrypted ciphertext value for government IDs.

### 3) High: Settlement engine ignores inspection deduction adjustments
- Previous status: Fail
- Current status: Fixed
- Evidence:
  - `repo/backend/internal/handlers/booking_handlers.go:163` aggregates `DamageDeductionAmount`.
  - `repo/backend/internal/handlers/booking_handlers.go:171` creates `wear_deduction` ledger entry.
  - `repo/backend/internal/handlers/booking_handlers.go:182` includes deductions in refund/deduction computation.
  - `repo/backend/tests/API_tests/integration/settlement_test.go:81` test validates `wear_deduction` behavior and adjusted deposit result.
- Runtime evidence:
  - `TestSettlementIncludesWearDeductions` passed.

### 4) High: README test command references non-existent compose service
- Previous status: Fail
- Current status: Fixed
- Evidence:
  - `repo/README.md:43` / `repo/README.md:46` now documents `./run_tests.sh` and module-local go test usage; no stale `docker compose run --rm test` command.
- Assessment:
  - Static command mismatch with compose service names is removed.

### 5) High: Testing docs inconsistent with Go module root layout
- Previous status: Partial Fail
- Current status: Fixed
- Evidence:
  - `docs/Testing_Modes.md:5` and `docs/Testing_Modes.md:10` use `cd backend && go test ...`.
- Assessment:
  - Commands are module-aware and aligned with backend `go.mod` placement.

### 6) Medium: Offline reconciliation endpoint was effectively a stub
- Previous status: Partial Fail
- Current status: Fixed
- Evidence:
  - `repo/backend/internal/handlers/disputes_handlers.go:412` enforces idempotency replay behavior (`already_applied`) via `MarkReconcileApplied`.
  - `repo/backend/internal/handlers/disputes_handlers.go:419` applies queued complaint operations.
  - `repo/backend/internal/handlers/disputes_handlers.go:427` applies queued inspection operations.
  - `repo/backend/internal/handlers/disputes_handlers.go:498` and `repo/backend/internal/handlers/disputes_handlers.go:521` contain concrete apply handlers.
  - `repo/backend/internal/store/repository.go:78` declares reconcile idempotency contract.
  - `repo/backend/internal/store/postgres.go:744` persists keys in `reconcile_keys` with conflict-safe insert.
  - `repo/backend/migrations/001_init.sql:234` defines the `reconcile_keys` table.
  - `repo/backend/tests/API_tests/live/coverage_test.go:706` includes sync reconcile coverage.
- Assessment:
  - Endpoint behavior is now materially implemented, including deterministic replay handling and persistence-backed dedup.

### 7) Medium: Frontend offline queue was narrow (booking-focused)
- Previous status: Partial Fail
- Current status: Fixed
- Evidence:
  - `repo/frontend/src/pages/BookingsPage.jsx:68` enqueues `booking`.
  - `repo/frontend/src/pages/InspectionsPage.jsx:70` now enqueues `inspection`.
  - `repo/frontend/src/pages/ComplaintsPage.jsx:87` enqueues `complaint` when offline.
  - Backend reconcile supports `booking`, `complaint`, `inspection` operation types (`repo/backend/internal/handlers/disputes_handlers.go:410-427` logic region).
  - `repo/frontend/tests/unit/queue.test.js:90` validates reconcile batching and payload types.

### 8) Medium: Frontend protocol docs claimed HTTPS while compose frontend dev command was HTTP
- Previous status: Cannot Confirm Statistically
- Current status: Fixed
- Evidence:
  - `docs/Operator_Runbook.md:10` explicitly states frontend dev is plain HTTP and recommends TLS termination in production.
  - `repo/docker-compose.yml:58` frontend runs Vite dev server (`npm run dev`) on port 5173 (HTTP).
- Assessment:
  - Documentation and compose behavior are now aligned.

### 9) Medium: Frontend test suite thin and e2e mostly API-oriented
- Previous status: Partial Fail
- Current status: Partially Fixed
- Evidence:
  - `repo/frontend/tests/e2e/app.ui.spec.js:10` and related cases now use Playwright `page` browser flows (login, role-gated nav, protected route behavior, complaints page access).
  - `repo/frontend/tests/unit/queue.test.js:1` adds dedicated offline queue/reconcile unit coverage.
  - `repo/frontend/tests/unit/EstimateSummary.test.jsx:15` still indicates narrow component-unit scope outside queue and estimate components.
  - `repo/frontend/tests/e2e/booking.spec.js:4` and multiple lines use Playwright `request` API calls to backend endpoints rather than browser `page` UI flows.
- Assessment:
  - The “mostly API-oriented” characterization is no longer accurate, but browser UI depth is still uneven across critical flows (for example booking wizard and inspection modal happy/failure paths).

## Net Change Summary
- Security bootstrap posture significantly improved.
- Settlement business logic and tests now include wear-and-tear deductions.
- Docs/command verifiability issues identified previously are largely resolved.
- Offline reconcile now includes persisted idempotency and deterministic replay behavior.
- Frontend offline queue now covers booking, inspection, and complaint enqueue paths.
- Residual priority work is mainly deeper browser-flow coverage for end-user UI journeys.

## Recommended Next Closures (Priority)
1. Add browser-level Playwright tests for booking wizard completion, inspection modal upload/submission, and complaint submit end-to-end with visual assertions.
2. Add accessibility-focused UI checks (landmarks, keyboard path, alert semantics) on auth, booking, and complaints screens.
3. Add CI split to run API-request tests and browser-UI tests as separate jobs with explicit pass/fail visibility.
