# Test Coverage Audit

## Project Type Detection
- Declared in `repo/README.md`: `Project Type: fullstack`.

## Backend Endpoint Inventory
- Source: `repo/backend/internal/api/router.go` (`NewRouter`).
- Total resolved endpoints (`METHOD + PATH`): **66**.

## API Test Mapping Table
- Mapping status: all 66 endpoints have direct HTTP test evidence (from `repo/backend/tests/API_tests/**`).
- Primary evidence hubs:
  - `repo/backend/tests/API_tests/live/coverage_test.go`
  - `repo/backend/tests/API_tests/auth_api_test.go`
  - `repo/backend/tests/API_tests/authorization_api_test.go`
  - `repo/backend/tests/API_tests/frontend_endpoint_coverage_test.go`
  - `repo/backend/tests/API_tests/integration/*.go`
  - `repo/backend/tests/API_tests/security/*.go`

## API Test Classification
1. **True No-Mock HTTP**
- Live `net/http.Client` suites:
  - `repo/backend/tests/API_tests/live_setup_test.go` (`apiCall`)
  - `repo/backend/tests/API_tests/integration/live_setup_test.go` (`intAPI`)
  - `repo/backend/tests/API_tests/security/live_setup_test.go` (`secAPI`)
  - `repo/backend/tests/API_tests/live/live_test.go` (`api`)
- In-process HTTP through real router/handlers (no execution-path mocks):
  - `repo/backend/tests/API_tests/security/transport_test.go`
  - `repo/backend/tests/API_tests/security/admin_allowlist_spoof_test.go`
  - `repo/backend/tests/API_tests/integration/settlement_test.go`
  - `repo/backend/tests/API_tests/integration/postgres_runtime_test.go`

2. **HTTP with Mocking**
- None detected in backend API tests.

3. **Non-HTTP (unit/integration without HTTP transport)**
- `repo/backend/tests/unit_tests/*.go`
- `repo/backend/internal/*/*_test.go`
- Frontend unit suites in `repo/frontend/tests/unit/**`

## Mock Detection Rules
- Backend API tests: no `jest.mock`, `vi.mock`, `sinon.stub`, gomock, or test doubles on controller/service execution path detected.
- Frontend unit tests intentionally use mocks (`vi.mock`) for UI-unit isolation, e.g.:
  - `repo/frontend/tests/unit/auth/AuthProvider.test.jsx`
  - `repo/frontend/tests/unit/pages/BookingsPage.test.jsx`
  - `repo/frontend/tests/unit/pages/OverviewPage.test.jsx`

## Coverage Summary
- Total endpoints: **66**
- Endpoints with HTTP tests: **66**
- Endpoints with TRUE no-mock tests: **66**
- HTTP coverage: **100%**
- True API coverage: **100%**

## Unit Test Summary

### Backend Unit Tests
- Present across:
  - `repo/backend/tests/unit_tests/*.go`
  - `repo/backend/internal/config/config_test.go`
  - `repo/backend/internal/logger/logger_test.go`
  - `repo/backend/internal/middleware/*_test.go`
  - `repo/backend/internal/services/*_test.go`
- Important backend modules not directly unit-tested:
  - `repo/backend/internal/handlers/*.go`
  - `repo/backend/internal/api/router.go`
  - `repo/backend/internal/services/retention.go`

### Frontend Unit Tests
- Required conditions satisfied (files + framework + real component/module imports + render/use):
  - Test files: `repo/frontend/tests/unit/**/*.test.*`
  - Frameworks: Vitest + React Testing Library (`repo/frontend/package.json`, `repo/frontend/vitest.config.js`)
  - Real module imports evidence:
    - `repo/frontend/tests/unit/pages/BookingsPage.test.jsx` -> `src/pages/BookingsPage`
    - `repo/frontend/tests/unit/components/ui/Button.test.jsx` -> `src/components/ui/Button`
    - `repo/frontend/tests/unit/auth/AuthProvider.test.jsx` -> `src/auth/AuthProvider`
    - `repo/frontend/tests/unit/queue.test.js` -> `src/offline/queue`
- Mandatory verdict: **Frontend unit tests: PRESENT**
- Important frontend modules not directly unit-tested:
  - `repo/frontend/src/App.jsx`
  - `repo/frontend/src/main.jsx`

### Cross-Layer Observation
- Backend and frontend both have substantial test presence.
- No backend-heavy/frontend-missing imbalance detected.

## API Observability Check
- Strong: many tests assert request payloads and response bodies (`live/coverage_test.go`, `integration/workflow_test.go`).
- Weak spots:
  - `repo/backend/tests/API_tests/live/coverage_test.go` -> `TestAPICoverageAudit` only checks for non-5xx.
  - Some tests allow broad status windows, reducing strictness.

## Test Quality & Sufficiency
- Strengths: success/failure/auth/authorization/tenant-isolation/integration boundaries covered.
- Gaps: shallow meta-coverage assertion and missing direct handler/router/retention unit tests.
- `run_tests.sh`: Docker-based orchestration confirmed (`repo/run_tests.sh`) -> **OK**.

## End-to-End Expectations
- Fullstack FE<->BE test evidence present:
  - API E2E: `repo/frontend/tests/e2e/booking.spec.js`
  - UI E2E: `repo/frontend/tests/e2e/app.ui.spec.js`

## Tests Check
- Backend Endpoint Inventory: complete
- API Test Mapping Table: complete
- Coverage Summary: complete
- Unit Test Summary: complete

## Test Coverage Score (0�100)
- **84/100**

## Score Rationale
- + 66/66 endpoint HTTP coverage
- + true no-mock API coverage present
- - shallow meta-test assertions in places
- - key orchestration modules lack direct unit tests

## Key Gaps
1. Strengthen semantic assertions in coverage meta-tests.
2. Add direct unit tests for handlers/router/retention.
3. Add dedicated `src/App.jsx` route/guard unit tests.

## Confidence & Assumptions
- Confidence: high.
- Static-only inspection; no test execution performed.

## Test Coverage Verdict
- **PASS with material quality gaps**

---

# README Audit

## README Location
- Present at `repo/README.md`.

## Hard Gates

### Formatting
- PASS: clear markdown structure and sections.

### Startup Instructions (Backend/Fullstack)
- PASS: includes required command literal `docker-compose up --build` in `Start (Docker)`.

### Access Method
- PASS:
  - Backend URL/port declared (`https://localhost:8080`)
  - Frontend URL/port declared (`http://localhost:5173`)

### Verification Method
- PASS: explicit API verification steps with concrete `curl` commands and expected outcomes are present.

### Environment Rules (Docker-contained)
- PASS: startup and primary test flow are Docker-contained (`docker-compose` / `docker compose`).

### Demo Credentials (Auth conditional)
- PASS: authentication is explicitly required, and credentials now include all roles used by the system:
  - `customer`
  - `provider`
  - `agent` (CSA)
  - `admin`

## Engineering Quality
- Tech stack clarity: good.
- Architecture explanation: good.
- Testing instructions: good.
- Security/roles/workflow clarity: good.
- Presentation quality: good.

## High Priority Issues
- None.

## Medium Priority Issues
- None.

## Low Priority Issues
1. External docs are referenced outside repo directory (`../docs/*`), which is acceptable but less self-contained.

## Hard Gate Failures
- None.

## README Verdict
- **PASS**

