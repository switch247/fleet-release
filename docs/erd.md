# ERD Draft

## Core Entities
- `users` (1..n) `user_roles`
- `categories` (1..n) `listings`
- `listings` (1..n) `bookings`
- `bookings` (1..n) `inspection_revisions`
- `bookings` (1..n) `attachments`
- `bookings` (1..n) `ledger_entries`
- `bookings` (1..n) `complaints`
- `users` (1..n) `consultation_versions`
- `users` (1..n) `notifications`
- `bookings` (1..n) `coupon_redemptions`

## Auditability
- `inspection_revisions.prev_hash -> inspection_revisions.hash` chain.
- `ledger_entries.prev_hash -> ledger_entries.hash` chain.
- `consultation_versions` keeps version history and change reason.
