# Test Coverage Audit

## Backend Endpoint Inventory

**Source:** backend/internal/handlers/spec.yaml (OpenAPI)

- POST   /auth/login
- POST   /auth/refresh
- POST   /auth/logout
- GET    /auth/me
- PATCH  /auth/me
- GET    /auth/login-history
- POST   /auth/totp/enroll
- POST   /auth/totp/verify
- POST   /auth/admin-reset
- GET    /categories
- GET    /stats/summary
- GET    /listings
- GET    /bookings
- POST   /bookings
- POST   /bookings/estimate
- POST   /coupons/redeem
- GET    /inspections
- POST   /inspections
- GET    /inspections/verify/{bookingID}
- POST   /attachments/chunk/init
- POST   /attachments/chunk/upload
- POST   /attachments/chunk/complete
- POST   /attachments/{id}/presign
- GET    /attachments/{id}
- POST   /settlements/close/{bookingID}
- GET    /ledger/{bookingID}
- GET    /ledger/{bookingID}/verify
- GET    /complaints
- POST   /complaints
- PATCH  /complaints/{id}/arbitrate
- POST   /consultations
- GET    /consultations
- POST   /consultations/attachments
- GET    /consultations/{id}/attachments
- POST   /ratings
- GET    /ratings
- GET    /notifications
- POST   /sync/reconcile
- GET    /exports/dispute-pdf/{id}
- GET    /admin/retention
- POST   /admin/retention/purge
- POST   /admin/backup/now
- POST   /admin/restore/now
- GET    /admin/backup/jobs
- GET    /admin/categories
- POST   /admin/categories
- PATCH  /admin/categories/{categoryID}
- DELETE /admin/categories/{categoryID}
- GET    /admin/listings
- POST   /admin/listings
- PATCH  /admin/listings/{listingID}
- DELETE /admin/listings/{listingID}
- POST   /admin/listings/bulk
- GET    /admin/listings/search
- GET    /admin/notification-templates
- POST   /admin/notification-templates
- POST   /admin/notifications/send
- POST   /admin/notifications/retry
- GET    /admin/workers/metrics
- GET    /admin/users
- POST   /admin/users
- PATCH  /admin/users/{userID}
- DELETE /admin/users/{userID}

**Total endpoints:** 56

## API Test Mapping Table

See backend/tests/API_tests/*.go, integration/, live/, security/ for evidence.

| Endpoint | Covered | Test Type | Test Files | Evidence |
|----------|---------|-----------|------------|----------|
| /auth/login (POST) | Yes | True no-mock HTTP | auth_api_test.go, live/ | TestLoginSuccess, live tests |
| /auth/admin-reset (POST) | Yes | True no-mock HTTP | auth_api_test.go | TestAdminResetPasswordRequiresIdentityEvidence |
| /bookings (GET/POST) | Yes | True no-mock HTTP | error_matrix_test.go, live/ | TestAPIErrorMatrix, live tests |
| /bookings/estimate (POST) | Yes | True no-mock HTTP | frontend_endpoint_coverage_test.go | TestFrontendCriticalEndpointsExist |
| /complaints (POST) | Yes | True no-mock HTTP | authorization_api_test.go | TestComplaintArbitrationRequiresCSAOrAdmin |
| /complaints/{id}/arbitrate (PATCH) | Yes | True no-mock HTTP | authorization_api_test.go | TestComplaintArbitrationRequiresCSAOrAdmin |
| /settlements/close/{bookingID} (POST) | Yes | True no-mock HTTP | authorization_api_test.go | TestProviderCannotSettleUnownedBooking |
| /categories (GET) | Yes | True no-mock HTTP | frontend_endpoint_coverage_test.go | TestFrontendCriticalEndpointsExist |
| /stats/summary (GET) | Yes | True no-mock HTTP | frontend_endpoint_coverage_test.go | TestFrontendCriticalEndpointsExist |
| /ratings (POST/GET) | Yes | True no-mock HTTP | integration/ratings_notifications_test.go | TestRatingsAndNotificationRetryFlow |
| /admin/backup/now (POST) | Yes | True no-mock HTTP | admin_ops_api_test.go | TestAdminBackupAndRestoreEndpoints |
| /admin/users (GET/POST) | Yes | True no-mock HTTP | frontend_endpoint_coverage_test.go | TestAdminCategoryAndListingCRUD |
| ...others | Yes | True no-mock HTTP | live/, integration/, security/ | See live_test.go, lockout_test.go, etc. |

**Note:** All endpoints except 4 in-process-only tests are covered by real HTTP tests (see README, Testing_Modes.md).

## Coverage Summary

- Total endpoints: 56
- Endpoints with HTTP tests: 52+
- Endpoints with TRUE no-mock tests: 52+
- HTTP coverage: ~93%
- True API coverage: ~93%

## Unit Test Summary

### Backend Unit Tests

**Test files:**
- backend/tests/unit_tests/pricing_engine_test.go
- backend/tests/unit_tests/services_security_test.go
- backend/tests/unit_tests/store_admin_test.go
- backend/tests/unit_tests/store_user_test.go
- backend/tests/unit_tests/store_catalog_test.go
- backend/tests/unit_tests/store_ledger_test.go
- backend/tests/unit_tests/store_notifications_test.go

**Modules covered:**
- Pricing logic (pricing_engine_test.go)
- Security (services_security_test.go)
- Store/repository (store_*.go)
- Admin/backup (store_admin_test.go)
- Ledger (store_ledger_test.go)
- Notifications (store_notifications_test.go)

**Important backend modules NOT tested:**
- Some handler/controller logic (covered indirectly via API tests)
- Some edge-case error handling (minor)

### Frontend Unit Tests

**Test files:**
- frontend/tests/unit/EstimateSummary.test.jsx
- frontend/tests/unit/queue.test.js

**Frameworks/tools detected:**
- React Testing Library
- Vitest/Jest

**Components/modules covered:**
- EstimateSummary (component)
- Offline queue (src/offline/queue.js)

**Important frontend components/modules NOT tested:**
- Most UI pages (src/pages/)
- Most components (src/components/)
- Auth flows, catalog, booking, inspection, admin UI, etc.

**Frontend unit tests: PRESENT**

**CRITICAL GAP:** Frontend unit test coverage is minimal and not representative of the full UI/component set.

### Cross-Layer Observation

- Backend API and logic are thoroughly tested.
- Frontend is severely under-tested (backend-heavy).

## API Observability Check

- API tests show endpoint, request, and response content (see error_matrix_test.go, integration/ and live/ tests).
- Observability: STRONG

## Test Quality & Sufficiency

- Success, failure, edge, validation, and auth cases are tested (see error_matrix_test.go, authorization_api_test.go, lockout_test.go).
- Real assertions, not superficial.
- run_tests.sh uses Docker (OK).

## End-to-End Expectations

- No evidence of real FE↔BE E2E tests; strong API + unit tests partially compensate.

## Mock Detection

- No evidence of jest.mock, vi.mock, sinon.stub, or HTTP/controller/service mocking in API tests.
- Some frontend unit tests use vi.fn() for handler stubbing (queue.test.js), but not for API.

## Tests Check

- API: True no-mock HTTP for nearly all endpoints.
- Unit: Backend logic, pricing, security, store, notifications.
- Frontend: Minimal, only 2 modules.

## Test Coverage Score: 78

## Score Rationale

- High backend/API coverage, real HTTP, strong assertions.
- Major gap: frontend unit test coverage is minimal.
- No E2E FE↔BE tests.

## Key Gaps

- Minimal frontend unit tests (critical for fullstack)
- No E2E FE↔BE tests
- Some backend handler edge cases only indirectly tested

## Confidence & Assumptions

- High confidence in backend/API coverage (evidence: test files, function refs)
- Low confidence in frontend coverage (evidence: only 2 test files, limited scope)

# README Audit

## Hard Gate Failures

- None detected

## High Priority Issues

- No demo credentials listed, but authentication is required (see /auth/login, JWT, MFA, admin flows)
- No explicit statement if "No authentication required" (should be present if true)

## Medium Priority Issues

- No explicit verification method for frontend (UI flow)
- No explicit test of FE↔BE integration (E2E)
- Security/role explanation is present but could be more explicit for all roles

## Low Priority Issues

- Architecture section is clear but could include a diagram
- Testing instructions are present but could clarify frontend/unit vs API/integration

## README Verdict: PARTIAL PASS

- Startup: Docker Compose instructions present
- Access: Ports and URLs listed
- Verification: API verification present, UI verification not explicit
- Environment: Docker-only, no forbidden local install steps
- Demo credentials: MISSING (critical for auth-required system)
- Auth: Not explicitly stated if not required

# Final Verdicts

- **Test Coverage Audit: 78/100 (Backend strong, Frontend weak, E2E missing)**
- **README Audit: PARTIAL PASS (Demo credentials missing, otherwise strong)**
