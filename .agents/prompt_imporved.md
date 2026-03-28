Improved Prompt: FleetLease Rental & Fare Operations Suite

Purpose
Provide a single, authoritative prompt that contains the original product prompt, an extracted and structured requirements list (business + code), a step-by-step build plan, and a clear Definition of Done. The prompt below is intended to be used by an implementation agent (or developer) that will produce code, tests, artifacts and documentation following the project's code-generation standards (see general_guide.md).

---

1) ORIGINAL PROMPT

Set up a FleetLease Rental & Fare Operations Suite that enables offline-first vehicle rental providers to publish inventory, run delivery and return inspections, calculate trip charges and deposits, and manage customer service outcomes with auditable records. The React web interface (English UI) supports four roles: Customer (renter), Provider (vehicle owner/operator), Customer Service Agent, and Administrator. Customers can browse multi-level categories and vehicle listings, compare SPU/SKU variants and availability, apply offline coupon codes, and confirm a booking with a clear pre-trip price estimate showing base fare, included miles, night surcharge windows (10:00 PM–5:59 AM), and refundable deposit. Providers and Customers complete standardized handover and return checklists in a guided flow that requires item-by-item confirmation and forces photo/video evidence capture from the device camera before completion; the UI immediately shows a wear-and-tear assessment with proposed deductions and a one-click settlement statement summarizing trip charges, adjustments, and deposit refund or deduction. After closure, both sides submit two-way ratings and can open a complaint that a Customer Service Agent arbitrates with structured outcomes and proof export (PDF download) for disputes; the system also supports structured consultation records (topic, key points, recommendations, follow-up plan) with attachments, version history, and role-based visibility. Administrators manage product/category publishing, bulk updates, and searchable listing fields, and can subscribe users to in-app notifications (e.g., “inspection required,” “settlement ready”) with reusable templates and delivery status within the app’s inbox.

The backend uses Go with Echo to expose REST-style APIs consumed by the React client, with PostgreSQL as the local system-of-record for users, roles/permissions, inventory, bookings, inspections, ledgers, ratings, complaints, consultations, and notifications. Authentication is fully offline with username/password only; passwords must be at least 12 characters with complexity checks, sessions use signed JWTs with a 30-minute idle timeout and a 12-hour absolute timeout, and password recovery is handled via Administrator-initiated reset after identity checks (no email/SMS). Security and privacy controls include TLS (HTTPS) even on internal networks, least-privilege RBAC, configurable IP allowlists for admin endpoints, and login-failure lockout for 15 minutes after 5 failed attempts. Optional MFA is supported via offline TOTP (RFC 6238) using an authenticator app without any network dependency. Sensitive fields (government ID, payment reference placeholders, addresses) are encrypted at rest with AES-256 and masked in UI and logs; attachments are stored on the local server file system with checksum validation and size limits (photos up to 10 MB, videos up to 100 MB). Inspection records are tamper-evident by chaining cryptographic hashes across revisions, while consultation notes maintain versioning with editor, timestamp, and change reason. The pricing and fare settlement engine computes charges from configurable rules (e.g., $1.80 base fare, $0.65 per mile, $0.22 per minute, 20% night surcharge) using odometer start/end and timestamps entered during checklists; refunds and adjustments post to an immutable ledger, and deposits auto-settle at trip end. Backups are local, scheduled nightly with a 30-day retention policy, plus one-click restore and configurable deletion rules (e.g., purge attachments after 365 days, retain financial ledgers for 7 years). The messaging center provides in-app delivery receipts, deduplication by message fingerprint, and retry queues for temporary write failures, while SMS/email channels remain disabled in offline mode but can be retained as templates for future on-prem connectors without external services.

---

2) EXTRACTED REQUIREMENTS (High-level)

Business requirements
- Roles: Customer, Provider, Customer Service Agent (CSA), Administrator. Support multi-role assignment per user if required.
- Browse multi-level categories and SPU/SKU vehicle variants with availability and comparisons.
- Offline coupon codes supported (device-first entry, server reconciliation).
- Booking flow: clear pre-trip price estimate (base fare, included miles, time charges, night surcharge windows, deposit), confirm booking, complete trip lifecycle.
- Handover and return inspections: guided item-by-item checklists, required camera capture for each item, immediate wear-and-tear assessment, proposed deductions, settlement statement and deposit adjustments.
- Ratings and two-way feedback; complaints with CSA arbitration, structured outcomes, and PDF proof export including audit hashes.
- Consultations: structured notes with version history and role-based visibility.
- Admin: product/category publishing, bulk updates, search fields, notification template management and in-app inbox.

Security & privacy
- Offline authentication: username/password (min 12 chars + complexity), optional offline TOTP (RFC 6238). JWT sessions: 30-minute idle timeout, 12-hour absolute timeout.
- Login lockout: 5 failed attempts -> 15-minute lockout.
- TLS everywhere, AES-256 for sensitive fields at rest, PII masking in UI and logs, admin IP allowlists, least-privilege RBAC.

Offline & sync
- Fully offline-capable core flows: bookings, inspections, attachment capture, coupon acceptance; reliable sync/resolution on reconnect.
- Deduplication by fingerprint, provisional redemptions for coupons, retry queues and chunked uploads for attachments.

Attachments & storage
- Photos <= 10 MB, videos <= 100 MB. Server stores attachments on local filesystem with checksum validation and fingerprint deduplication. Retention rules configurable (default purge 365 days for attachments; financial ledgers retain 7 years).

Pricing & settlement
- Configurable pricing engine: base fare, per-mile, per-minute, included miles, night surcharges (10:00 PM–5:59 AM), deposits, rounding rules, minimum fare, ledger of immutable transactions, atomic trip settlement with deposit auto-settle.

Audit & data model
- Inspection revisions must be chained with cryptographic hashes (tamper-evident). Ledgers append-only and versioned. Consultation notes versioned with editor, timestamp, reason.

Backups & retention
- Nightly local backups, 30-day retention, one-click restore UI. Configurable deletion/purge jobs.

Messaging
- In-app messages with delivery receipts, deduplication, retry queues; email/SMS channels disabled by default but templates stored.

Non-functional
- Target OS/deploy: on-prem local server (Linux preferred). Expect modest scale per deployment; define expected vehicles/concurrency per customer.
- Performance: sync must be robust; server authoritative for final reconciliations.

---

3) CODE & ARCHITECTURE REQUIREMENTS

- Backend: Go using Echo (REST APIs). Provide a minimal, modular service layout: auth, users, inventory, bookings, inspections, attachments, pricing, ledger, notifications, admin.
- Database: PostgreSQL (relational model with revision tables and append-only ledger). Define migrations and seed data for demo.
- Frontend: React (English UI). Offline-first client storage (IndexedDB or local file store), camera integration for evidence capture, offline queue for sync.
- Authentication: username/password, optional TOTP (RFC 6238), JWT signed tokens with configured timeouts, server-side session revocation.
- Encryption: AES-256 for sensitive DB fields; attachment checksums; secure key handling documented (local key file or KMS option).
- Attachments: local filesystem storage with path referencing in DB, checksum validation, chunked uploads for large files, retry queues.
- Pricing engine: configurable rule-based engine with deterministic evaluation and unit tests for pricing scenarios.
- Auditability: revision history tables with cryptographic hash chaining; exportable PDFs including evidence and hash chain.
- Backups: implement DB dump + attachments snapshot process; retention policy and restore script/endpoint for admins.
- Testing: unit tests for business logic, integration tests for API endpoints, end-to-end tests for critical flows (booking → inspections → settlement).
- Dev ergonomics: Docker/dev-compose for easy local runs, Makefile or scripts for build/test, and a README with quickstart.
- Follow project's code generation standards in general_guide.md: produce runnable minimal implementation, tests, README, and dependency manifests.

---

4) STEP-BY-STEP BUILD PLAN (for the implementing agent)

Phase 0 — Confirm & Plan
- Confirm open questions and acceptance criteria with stakeholders. Lock MVP scope.
- Deliverable: final spec document and prioritized feature list.

Phase 1 — Data model & API surface
- Design DB schema (users, roles, products, spu/sku, inventory, bookings, inspections, attachments, ledger, notifications, complaints, consultations).
- Design REST endpoints and request/response shapes.
- Deliverable: ER diagram, migration files, API spec (OpenAPI/Swagger).

Phase 2 — Auth, Security & Infra
- Implement auth (password policy, JWT, TOTP enrollment flow stub), RBAC middleware, IP allowlist middleware, encryption support for sensitive fields, login lockout.
- Deliverable: auth endpoints, middleware, tests for security rules.

Phase 3 — Core booking flows
- Implement product listing, SPU/SKU variant selection, pre-trip estimate calculation (pricing engine), coupon acceptance (offline handling), booking creation.
- Deliverable: API + minimal UI pages for browsing and booking, pricing engine unit tests.

Phase 4 — Inspections & Evidence
- Implement checklist templates, per-vehicle overrides, camera-only capture enforcement, attachment storage and chunked uploads, evidence validation, wear-and-tear proposals, settlement computation.
- Deliverable: UI checklist flows (handover/return), attachment upload with checksum and retry, settlement API and UI.

Phase 5 — Ledger, Settlement & Disputes
- Implement immutable ledger entries, deposit lifecycle, settlement finalization, complaint lifecycle and CSA arbitration UI, PDF export with audit hashes.
- Deliverable: ledger model, settlement endpoints, complaint handling, PDF export tests.

Phase 6 — Offline sync & deduplication
- Implement client-side queueing, provisional local redemptions, fingerprint dedup, conflict resolution/reconciliation endpoints.
- Deliverable: sync protocol docs, client sync implementation, integration tests.

Phase 7 — Admin & Backups
- Admin UI for categories/products, retention rules, IP allowlist, notification templates; nightly backup job + restore UI.
- Deliverable: admin pages, backup/restore scripts, backup verification tests.

Phase 8 — QA, E2E & Documentation
- Add E2E tests for AC flows, security test cases, performance checks for sync and attachment upload; finalize README and runbooks.
- Deliverable: test suite, README, deployment notes.

Phase 9 — Demo & Handover
- Prepare a demo that satisfies Acceptance Criteria and provide a checklist for reviewers.

---

5) DEFINITION OF DONE (Concrete Acceptance Criteria)

- DO-1: End-to-end booking demoable: browse categories, select SPU/SKU, apply offline coupon, confirm booking, perform both handover and return inspections with required camera evidence, and show final settlement including deposit adjustment. Demo runs offline for client actions and syncs after reconnect.
- DO-2: Security & Auth: Password policy enforced (≥12 chars + complexity rules defined), JWT timeouts enforced (30m idle, 12h absolute), account lockout after 5 failed attempts for 15 minutes, optional offline TOTP flows implemented.
- DO-3: Attachments: Client enforces photo ≤10 MB and video ≤100 MB; server validates checksums, stores files on local filesystem, and deduplicates by fingerprint.
- DO-4: Auditability: Inspection revisions and ledger entries are versioned; each revision includes hash of previous revision; exported dispute PDF contains evidence, timestamps, and the hash chain.
- DO-5: Offline behavior: Client can complete booking and inspections while offline; attachments queue and metadata sync reliably on reconnect; coupon provisional redemptions reconcile without loss; deduplication prevents double-redemption in normal scenarios.
- DO-6: Backups & retention: Nightly backups run and retain 30 days; financial ledgers preserved for 7 years by default and excluded from automated purge; restore procedure validated.
- DO-7: Admin: Admin UI supports publishing categories/products, configuring retention rules and IP allowlist, managing notification templates; RBAC enforced for admin endpoints.
- DO-8: Test Coverage: Unit tests for business logic (≥70% for critical modules), integration tests for API, and at least one E2E test covering the full AC-1 flow.

---

6) IMPLEMENTATION NOTES & CONSTRAINTS

- No external cloud services are used for core flows. Email/SMS integrations remain templated and disabled in default offline mode.
- Sensitive keys and encryption secrets must be documented in the deployment README; support local key file and an optional KMS configuration.
- Packaging: produce a runnable local deployment (Docker compose and single-binary option for backend) plus a minimal seed dataset for demo.
- The implementing agent must follow the project's code generation guidelines in `general_guide.md` when producing code and artifacts: produce runnable examples, tests, migrations, README, and manifests.

---

7) NEXT ACTIONS (for you or the implementing agent)

- Review and confirm or correct extracted requirements and Definition of Done.
- If confirmed, proceed with Phase 1 (data model + API spec). Produce an OpenAPI file, migrations, and a seed dataset for the demo.

---

If any requirement above is incorrect or needs further precision, annotate the specific numbered item and provide the replacement text or values.
