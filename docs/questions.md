
**Requirements — Scope & Roles**
1) Question: What is the minimal end-to-end booking flow that must be demoable for MVP?
   My understanding: Must allow browsing, selecting a vehicle variant, applying offline coupon, confirming booking, and completing handover and return inspections with settlement.
   Suggested solution: Define an MVP story with required screens/data fields and list any optional features (complaints, consultations) as future milestones.

2) Question: Are the four roles strictly separated, and can one user have multiple roles simultaneously?
   My understanding: Roles are distinct but users (e.g., a provider) might also act as a customer in other contexts.
   Suggested solution: Clarify whether role assignment is many-to-many and whether a user can switch contexts in-app.

3) Question: Who can create coupons and set coupon rules (Provider, Admin)? Are coupons globally unique or provider-scoped?
   My understanding: Admins likely create reusable coupons; providers may have local coupons.
   Suggested solution: Define coupon scope and rule fields (percent/flat, validity window, usage limits, offline code format).

---

**Tech Stack, Deployment & Ops Constraints**
4) Question: Confirm backend tech: Go + Echo + PostgreSQL — are there version constraints or preferred deployment targets (Windows/Linux) for on-prem installs?
   My understanding: Linux servers are typical for production, but Windows may be used in some customers.
   Suggested solution: Document supported OS, Go and Postgres minimum versions, and build/distribution method (single binary, installer, container).

5) Question: Are there constraints on third-party services (no external cloud dependencies)?
   My understanding: System must run fully offline with no external SaaS dependencies for core flows.
   Suggested solution: Explicitly list allowed optional connectors (e.g., future SMS gateway) and mark them disabled for MVP.

---

**Security, Auth & Privacy**
6) Question: Password policy specifics beyond length (special chars, upper/lower, disallowed substrings)?
   My understanding: Min 12 chars plus complexity checks; exact rules needed for validation UI and password strength meter.
   Suggested solution: Provide a formal regex or a checklist used by client-side validator.

7) Question: How are AES-256 keys managed for at-rest encryption (local KMS, config file, hardware module)?
   My understanding: Keys likely stored on the server and must be exportable/importable for backups.
   Suggested solution: Specify key rotation policy, storage format, and where to place keys during install.

8) Question: For TOTP MFA, do we need QR provisioning and recovery codes, or strictly seed export/import for offline use?
   My understanding: Offline TOTP should work via seed; recovery flows must be admin-driven.
   Suggested solution: Confirm required UX for enrolling and recovering MFA.

9) Question: Admin endpoints: what IP allowlist format and management UX is required?
   My understanding: Admins can configure CIDR ranges and individual IPs via Admin UI.
   Suggested solution: Specify UI requirements and storage schema for allowlists.

---

**Offline-first, Sync & Conflict Handling**
10) Question: What is the expected sync model between client and server (last-write-wins, operational transforms, CRDTs)?
	My understanding: Primary workflows (bookings, inspections, ledgers) require strong consistency and auditable reconciliation.
	Suggested solution: Use a transactional server-side authoritative model with client shadow copies and conflict detection rules (reject conflicting settlement edits; require reconciliation flag).

11) Question: How should offline coupons be validated once the device reconnects (redeem-on-sync vs reserve-on-booking)?
	My understanding: For offline use, codes can be accepted locally and validated later; must prevent double-redemption.
	Suggested solution: Implement locally-issued provisional redemptions with server reconciliation and fingerprint-based deduplication.

12) Question: For attachments captured offline (photos/videos), how to handle upload retry, partial uploads, and checksum validation?
	My understanding: Client stores files locally and uploads on connectivity with retry and checksum verification.
	Suggested solution: Define fingerprint and chunked upload approach plus a retry queue with exponential backoff.

---

**Data Model, Auditing & Retention**
13) Question: What fields must be immutable/append-only (ledger entries, settlement transactions)?
	My understanding: Financial ledger must be append-only; bookings and inspections need revision history with hashes.
	Suggested solution: Mark ledgers immutable and implement revision metadata (editor, timestamp, reason, prev-hash).

14) Question: Attachment retention policies: confirm defaults (photos 365 days? attachments purge example given) and per-tenant overrides.
	My understanding: Default purge after 365 days, but financial ledgers retained 7 years.
	Suggested solution: Allow admin-configurable retention rules per attachment type and tenant.

15) Question: What PII fields require masking in logs and UI (gov ID, addresses, payment refs)? Any fields excluded from backups?
	My understanding: Gov ID, payment references, full addresses are sensitive and must be masked; backups may include encrypted PII.
	Suggested solution: Provide canonical list of PII fields and masking rules.

---

**Pricing, Fare Rules & Settlement**
16) Question: Confirm pricing formula and edge rules: base fare, per-mile, per-minute, night surcharge window, included miles, deposit calculation, rounding rules, minimum fare.
	My understanding: Prompt lists example rates; we need all parameters and precedence for surcharges and discounts.
	Suggested solution: Provide a complete pricing rule schema and examples for typical trips (day/night, long distance, early return).

17) Question: When is deposit captured vs reserved vs auto-settled? Are settlements batched or per-trip, and can adjustments be applied after settlement?
	My understanding: Deposit auto-settles at trip end; ledger must record adjustments and final settlement atomicity.
	Suggested solution: Define deposit lifecycle states and allowed post-settlement adjustments (with audit trail).

---

**Inspection Flows & Evidence**
18) Question: Detailed item checklist model: are checklists per-SPU/SKU, per-vehicle, or templated per-provider?
	My understanding: Providers maintain templates; vehicles have specific items that can be added/removed.
	Suggested solution: Support checklist templates with per-vehicle overrides and versioning.

19) Question: Photo/video capture requirements: mandatory count per item, max size, enforced camera-only vs gallery allowed, and file types.
	My understanding: Camera capture is required; size limits: photos 10 MB, videos 100 MB.
	Suggested solution: Enforce camera capture in the client and implement client-side compression and validation.

20) Question: Wear-and-tear assessment rules: who proposes deductions and final arbiter for contested deductions?
	My understanding: The UI proposes deductions automatically; final settlement can be adjusted by Provider or Customer with agent arbitration.
	Suggested solution: Capture proposed deductions, allow counter-proposals, and require CSA approval for disputes.

---

**Complaints, Consultations & Exporting Proof**
21) Question: Complaint lifecycle and structured outcomes: what are allowed outcomes and time-to-resolution SLAs?
	My understanding: Outcomes include refund, partial refund, no action, consult, escalate; SLAs TBD.
	Suggested solution: Define a finite outcome set and SLA defaults.

22) Question: PDF export requirements: which records must be included (attachments, hashes, signatures)?
	My understanding: Export should include full evidence, timestamps, signatures, and audit hash chain.
	Suggested solution: Create a PDF template and required fields checklist.

---

**Notifications & Messaging**
23) Question: In-app notification templates and delivery semantics: must support delivery receipts, retries, deduplication, and per-role inbox visibility?
	My understanding: Yes — templates, delivery status, retries, and deduplication required.
	Suggested solution: Define template variables, inbox model, and fingerprinting strategy for deduplication.

24) Question: Are email/SMS templates stored even when channels are disabled? Any export/import needs for templates?
	My understanding: Templates kept for future connectors.
	Suggested solution: Allow export/import of templates (JSON) and flag channels enabled/disabled.

---

**Backups, Restore & Retention**
25) Question: Backup format and restore expectations: full DB dumps + attachments or filesystem snapshot? One-click restore semantics?
	My understanding: Nightly local backups with 30-day retention; need restore UI for admins.
	Suggested solution: Document backup artifacts, encryption for backups, and a tested restore procedure.

26) Question: Data deletion rules: soft-delete vs physical delete; cascading deletes for related attachments and ledger constraints?
	My understanding: Soft-delete preferred for audit; physical purge per retention policy.
	Suggested solution: Implement `deleted_at` with purge jobs and exceptions for legal retention.

---

**Performance, Scale & Edge Cases**
27) Question: Expected scale per deployment (number of vehicles, concurrent clients, booking throughput)?
	My understanding: Unknown — requires vendor/customer input.
	Suggested solution: Request target scale to size DB, backup windows, and local resource requirements.

28) Question: Conflict scenarios during simultaneous inspections or rapid odometer edits — what rules resolve them?
	My understanding: Server should reject conflicting submissions and require manual reconciliation.
	Suggested solution: Use optimistic locking with revision IDs and human reconciliation workflow.

---

**Acceptance Criteria (Proposed — please confirm/adjust)**
AC-1: End-to-end booking demo: browse categories, select SPU/SKU, apply offline coupon, complete booking, perform handover and return inspections, and show settlement with deposit adjustment. Successful demo includes photo/video evidence for both inspections.
AC-2: Offline-first behavior: Client can complete booking and both inspections while offline; on reconnect, attachments and events sync with server with no data loss; deduplication prevents double-redemption of coupons in 99.99% tested cases.
AC-3: Security: Password policy enforced (min 12 chars + complexity), JWT sessions expire after 30 minutes idle and 12-hour absolute timeout; account lockout after 5 failed attempts for 15 minutes; AES-256 encrypts sensitive fields at rest.
AC-4: Auditability: All inspection revisions and ledger entries are versioned and include cryptographic hashes; exported dispute PDFs contain evidence and hash chain.
AC-5: Attachments & limits: Client enforces photo <= 10 MB and video <= 100 MB; server validates checksums and stores attachments with fingerprint deduplication.
AC-6: Backups & retention: Nightly local backup runs successfully and retains 30 days; financial ledgers retained a minimum of 7 years and are not purged by automated retention.
AC-7: Admin controls: Admins can publish categories/products, manage retention rules, configure IP allowlist, and download templates; RBAC enforces least privilege.

