CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  email TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL,
  government_id_enc TEXT,
  payment_reference_enc TEXT,
  address_enc TEXT,
  failed_attempts INT NOT NULL DEFAULT 0,
  locked_until TIMESTAMPTZ,
  totp_secret TEXT,
  totp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS payment_reference_enc TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS address_enc TEXT;

CREATE TABLE IF NOT EXISTS user_roles (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL,
  PRIMARY KEY (user_id, role)
);

CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  issued_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL,
  absolute_exp TIMESTAMPTZ NOT NULL,
  revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS auth_events (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  username TEXT,
  ip TEXT,
  event_type TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  parent_id UUID REFERENCES categories(id)
);
ALTER TABLE categories ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES categories(id);

CREATE TABLE IF NOT EXISTS listings (
  id UUID PRIMARY KEY,
  category_id UUID NOT NULL REFERENCES categories(id),
  provider_id UUID REFERENCES users(id),
  spu TEXT NOT NULL,
  sku TEXT NOT NULL,
  name TEXT NOT NULL,
  included_miles NUMERIC(10,2) NOT NULL,
  deposit NUMERIC(10,2) NOT NULL,
  available BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS bookings (
  id UUID PRIMARY KEY,
  customer_id UUID NOT NULL REFERENCES users(id),
  provider_id UUID REFERENCES users(id),
  listing_id UUID NOT NULL REFERENCES listings(id),
  coupon_code TEXT,
  status TEXT NOT NULL,
  estimated_amount NUMERIC(10,2) NOT NULL,
  deposit_amount NUMERIC(10,2) NOT NULL,
  start_at TIMESTAMPTZ NOT NULL,
  end_at TIMESTAMPTZ NOT NULL,
  odo_start NUMERIC(10,2) NOT NULL,
  odo_end NUMERIC(10,2) NOT NULL
);

ALTER TABLE bookings ADD COLUMN IF NOT EXISTS coupon_code TEXT;
ALTER TABLE listings ADD COLUMN IF NOT EXISTS provider_id UUID REFERENCES users(id);
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS provider_id UUID REFERENCES users(id);

CREATE TABLE IF NOT EXISTS inspection_revisions (
  id UUID PRIMARY KEY,
  booking_id UUID NOT NULL REFERENCES bookings(id),
  stage TEXT NOT NULL,
  payload JSONB NOT NULL,
  prev_hash TEXT,
  hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by UUID NOT NULL REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS attachments (
  id UUID PRIMARY KEY,
  booking_id UUID NOT NULL REFERENCES bookings(id),
  type TEXT NOT NULL,
  path TEXT NOT NULL,
  size_bytes BIGINT NOT NULL,
  checksum TEXT NOT NULL,
  fingerprint TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ledger_entries (
  id UUID PRIMARY KEY,
  booking_id UUID NOT NULL REFERENCES bookings(id),
  entry_type TEXT NOT NULL,
  amount NUMERIC(10,2) NOT NULL,
  description TEXT NOT NULL,
  prev_hash TEXT,
  hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS complaints (
  id UUID PRIMARY KEY,
  booking_id UUID NOT NULL REFERENCES bookings(id),
  opened_by UUID NOT NULL REFERENCES users(id),
  status TEXT NOT NULL,
  outcome TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS consultation_versions (
  id UUID PRIMARY KEY,
  consultation_thread_id TEXT NOT NULL,
  consultation_key TEXT,
  booking_id UUID REFERENCES bookings(id),
  version INT NOT NULL,
  topic TEXT NOT NULL,
  key_points TEXT,
  recommendation TEXT,
  follow_up TEXT,
  visibility TEXT NOT NULL DEFAULT 'csa_admin',
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  change_reason TEXT
);

ALTER TABLE consultation_versions ADD COLUMN IF NOT EXISTS consultation_thread_id TEXT;
UPDATE consultation_versions
SET consultation_thread_id = COALESCE(consultation_thread_id, consultation_key, COALESCE(booking_id::text, '') || '::' || topic)
WHERE consultation_thread_id IS NULL OR consultation_thread_id = '';
CREATE UNIQUE INDEX IF NOT EXISTS consultation_thread_version_idx ON consultation_versions (consultation_thread_id, version);
CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  template_id UUID,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'queued',
  attempts INT NOT NULL DEFAULT 0,
  fingerprint TEXT NOT NULL,
  delivered_at TIMESTAMPTZ,
  UNIQUE (user_id, fingerprint)
);
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS template_id UUID;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'queued';
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS coupon_redemptions (
  id UUID PRIMARY KEY,
  code TEXT NOT NULL,
  booking_id UUID NOT NULL REFERENCES bookings(id),
  status TEXT NOT NULL,
  fingerprint TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS consultation_attachments (
  id UUID PRIMARY KEY,
  consultation_id UUID NOT NULL REFERENCES consultation_versions(id) ON DELETE CASCADE,
  attachment_id UUID NOT NULL REFERENCES attachments(id),
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_templates (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  channel TEXT NOT NULL DEFAULT 'in_app',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_by UUID NOT NULL REFERENCES users(id),
  modified_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ratings (
  id UUID PRIMARY KEY,
  booking_id UUID NOT NULL REFERENCES bookings(id),
  from_user_id UUID NOT NULL REFERENCES users(id),
  to_user_id UUID NOT NULL REFERENCES users(id),
  score INT NOT NULL CHECK (score BETWEEN 1 AND 5),
  comment TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS backup_jobs (
  id UUID PRIMARY KEY,
  job_type TEXT NOT NULL,
  status TEXT NOT NULL,
  artifact TEXT,
  requested_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  finished_at TIMESTAMPTZ,
  error_message TEXT
);

CREATE TABLE IF NOT EXISTS retention_jobs (
  id UUID PRIMARY KEY,
  attachments_deleted INT NOT NULL,
  ledger_deleted INT NOT NULL,
  file_delete_errors INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS password_reset_evidence (
  id UUID PRIMARY KEY,
  target_user_id UUID NOT NULL REFERENCES users(id),
  checked_by UUID NOT NULL REFERENCES users(id),
  method TEXT NOT NULL,
  evidence_ref TEXT NOT NULL,
  reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


