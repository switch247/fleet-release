# Audit Report 2 - Fix Verification

Date: 2026-04-13
Method: Static code review only (no runtime execution in this pass)
Source baseline: `.tmp/audit_report-2.md`

## Overall Re-check Verdict
- Status: Fully Fixed (6 fixed, 0 partially fixed, 0 still open as originally described)

## Issue-by-Issue Verification

### 1) Coupon workflow does not affect fare estimation or settlement economics
- Previous severity: Blocker
- Current status: Fixed
- Evidence:
  - Coupon discount is loaded and passed into estimate input:
    - `repo/backend/internal/handlers/booking_handlers.go:63`
    - `repo/backend/internal/handlers/booking_handlers.go:67`
    - `repo/backend/internal/handlers/booking_handlers.go:104`
    - `repo/backend/internal/handlers/booking_handlers.go:116`
  - Pricing engine applies percentage discount to total and returns discount amount:
    - `repo/backend/internal/services/pricing.go:80`
    - `repo/backend/internal/services/pricing.go:81`
    - `repo/backend/internal/services/pricing.go:82`
    - `repo/backend/internal/services/pricing.go:90`
  - Booking persists discount amount:
    - `repo/backend/internal/handlers/booking_handlers.go:69`
  - Settlement ledger writes explicit coupon discount entry:
    - `repo/backend/internal/handlers/booking_handlers.go:180`
    - `repo/backend/internal/handlers/booking_handlers.go:182`
    - `repo/backend/internal/handlers/booking_handlers.go:184`
  - Frontend estimate UI renders coupon discount breakdown:
    - `repo/frontend/src/components/EstimateSummary.jsx` (couponDiscountAmount row present)

### 2) Offline inspection queue bypasses mandatory evidence requirement
- Previous severity: High
- Current status: Fixed
- Evidence:
  - Backend still enforces evidence required per checklist item:
    - `repo/backend/internal/handlers/inspection_handlers.go:62`
    - `repo/backend/internal/handlers/inspection_handlers.go:63`
  - Frontend offline path now blocks queueing without local evidence files:
    - `repo/frontend/src/pages/InspectionsPage.jsx:98`
  - Offline queue stores `_offlineFile` payload and sync step uploads it before submitting inspection:
    - `repo/frontend/src/pages/InspectionsPage.jsx:67`
    - `repo/frontend/src/pages/InspectionsPage.jsx:68`
    - `repo/frontend/src/pages/InspectionsPage.jsx:108`
  - `evidenceIds: []` is only an intermediate queued shape before sync upload replacement:
    - `repo/frontend/src/pages/InspectionsPage.jsx:106`

### 3) Customer inspection flow is hidden in primary navigation
- Previous severity: High
- Current status: Fixed
- Evidence:
  - App nav now explicitly shows Inspections for customer role:
    - `repo/frontend/src/components/layout/AppShell.jsx:13`
  - E2E test now aligns with current behavior (customer sees inspections nav):
    - `repo/frontend/tests/e2e/app.ui.spec.js:142`
    - `repo/frontend/tests/e2e/app.ui.spec.js:145`
    - `repo/frontend/tests/e2e/app.ui.spec.js:146`

### 4) Booking creation does not enforce listing availability
- Previous severity: High
- Current status: Fixed
- Evidence:
  - Server-side availability guard added at booking creation:
    - `repo/backend/internal/handlers/booking_handlers.go:45`
    - `repo/backend/internal/handlers/booking_handlers.go:46`

### 5) Repo-local documentation path references are inconsistent
- Previous severity: Medium
- Current status: Fixed
- Evidence:
  - README now adds explicit clarification that docs are in workspace root above repo and links using `../docs/...`:
    - `repo/README.md:11`
    - `repo/README.md:75`
    - `repo/README.md:76`
    - `repo/README.md:77`
    - `repo/README.md:78`
    - `repo/README.md:79`
    - `repo/README.md:80`
    - `repo/README.md:81`

### 6) Login page includes hardcoded default credentials
- Previous severity: Medium
- Current status: Fixed
- Evidence:
  - Login form initializes empty username/password/totp values:
    - `repo/frontend/src/pages/LoginPage.jsx:11`

## Delta Summary
- Fixed since prior audit:
  - Coupon economics integrated into estimate + booking persistence + settlement ledger.
  - Booking availability enforcement added.
  - Offline inspection flow now requires evidence before queueing and uploads on sync.
  - Login defaults removed.
  - Customer inspections nav enabled.
  - Customer inspections e2e assertion aligned with current role policy.

## Suggested Follow-up (small)
No remaining follow-up required for the six issues from `.tmp/audit_report-2.md`.
