INSERT INTO categories (id, name) VALUES
('00000000-0000-0000-0000-000000000101', 'Cars'),
('00000000-0000-0000-0000-000000000102', 'Vans')
ON CONFLICT DO NOTHING;

INSERT INTO listings (id, category_id, spu, sku, name, included_miles, deposit, available) VALUES
('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000101', 'SEDAN-SPU', 'SEDAN-SKU-A', 'City Sedan', 2.00, 75.00, TRUE),
('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000102', 'VAN-SPU', 'VAN-SKU-X', 'Cargo Van', 3.00, 90.00, TRUE)
ON CONFLICT DO NOTHING;

-- Demo users
INSERT INTO users (id, username, email, password_hash, totp_enabled) VALUES
('00000000-0000-0000-0000-00000000a001', 'admin', 'admin@example.com', 'demo-hash', TRUE),
('00000000-0000-0000-0000-00000000a002', 'provider1', 'provider1@example.com', 'demo-hash', FALSE),
('00000000-0000-0000-0000-00000000a003', 'customer1', 'customer1@example.com', 'demo-hash', FALSE),
('00000000-0000-0000-0000-00000000a004', 'csa1', 'csa1@example.com', 'demo-hash', FALSE)
ON CONFLICT DO NOTHING;

-- Roles
INSERT INTO user_roles (user_id, role) VALUES
('00000000-0000-0000-0000-00000000a001', 'admin'),
('00000000-0000-0000-0000-00000000a002', 'provider'),
('00000000-0000-0000-0000-00000000a003', 'customer'),
('00000000-0000-0000-0000-00000000a004', 'csa')
ON CONFLICT DO NOTHING;

-- Assign provider to existing listings
UPDATE listings SET provider_id = '00000000-0000-0000-0000-00000000a002' WHERE id IN ('00000000-0000-0000-0000-000000000201','00000000-0000-0000-0000-000000000202');

-- Demo bookings (settled, active, cancelled)
INSERT INTO bookings (id, customer_id, provider_id, listing_id, coupon_code, status, estimated_amount, deposit_amount, start_at, end_at, odo_start, odo_end) VALUES
('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-00000000a003', '00000000-0000-0000-0000-00000000a002', '00000000-0000-0000-0000-000000000201', NULL, 'settled', 200.00, 75.00, NOW() - INTERVAL '7 days', NOW() - INTERVAL '5 days', 1000.00, 1100.00),
('00000000-0000-0000-0000-000000000302', '00000000-0000-0000-0000-00000000a003', '00000000-0000-0000-0000-00000000a002', '00000000-0000-0000-0000-000000000202', NULL, 'active', 320.00, 90.00, NOW() + INTERVAL '1 days', NOW() + INTERVAL '3 days', 2000.00, 2000.00),
('00000000-0000-0000-0000-000000000303', '00000000-0000-0000-0000-00000000a003', '00000000-0000-0000-0000-00000000a002', '00000000-0000-0000-0000-000000000201', 'DEMO10', 'cancelled', 150.00, 75.00, NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 days', 5000.00, 5000.00)
ON CONFLICT DO NOTHING;

-- Inspection revision for settled booking
INSERT INTO inspection_revisions (id, booking_id, stage, payload, prev_hash, hash, created_by) VALUES
('00000000-0000-0000-0000-00000000b001', '00000000-0000-0000-0000-000000000301', 'initial', '{}'::jsonb, NULL, 'hash-demo-1', '00000000-0000-0000-0000-00000000a002')
ON CONFLICT DO NOTHING;

-- Attachments for bookings
INSERT INTO attachments (id, booking_id, type, path, size_bytes, checksum, fingerprint) VALUES
('00000000-0000-0000-0000-00000000c001', '00000000-0000-0000-0000-000000000301', 'image', '/app/data/attachments/settled-1.jpg', 12345, 'chk1', 'fp1'),
('00000000-0000-0000-0000-00000000c002', '00000000-0000-0000-0000-000000000302', 'image', '/app/data/attachments/active-1.jpg', 23456, 'chk2', 'fp2')
ON CONFLICT DO NOTHING;

-- Ledger entries for settled booking
INSERT INTO ledger_entries (id, booking_id, entry_type, amount, description, prev_hash, hash) VALUES
('00000000-0000-0000-0000-00000000d001', '00000000-0000-0000-0000-000000000301', 'charge', 200.00, 'Rental charge', NULL, 'lhash1'),
('00000000-0000-0000-0000-00000000d002', '00000000-0000-0000-0000-000000000301', 'deposit_release', -75.00, 'Return deposit', 'lhash1', 'lhash2')
ON CONFLICT DO NOTHING;

-- Ratings for settled booking
INSERT INTO ratings (id, booking_id, from_user_id, to_user_id, score, comment) VALUES
('00000000-0000-0000-0000-00000000e001', '00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-00000000a003', '00000000-0000-0000-0000-00000000a002', 5, 'Great experience')
ON CONFLICT DO NOTHING;

-- Complaint opened for cancelled booking
INSERT INTO complaints (id, booking_id, opened_by, status) VALUES
('00000000-0000-0000-0000-00000000f001', '00000000-0000-0000-0000-000000000303', '00000000-0000-0000-0000-00000000a003', 'open')
ON CONFLICT DO NOTHING;

-- Consultation for active booking
INSERT INTO consultation_versions (id, consultation_key, booking_id, version, topic, key_points, recommendation, visibility, created_by) VALUES
('00000000-0000-0000-0000-000000010001', 'CONSULT-0001', '00000000-0000-0000-0000-000000000302', 1, 'Pre-trip checklist', 'Check tires; Fill fuel', 'Proceed with booking', 'csa_admin', '00000000-0000-0000-0000-00000000a004')
ON CONFLICT DO NOTHING;

-- Notification for provider
INSERT INTO notifications (id, user_id, title, body, fingerprint, status) VALUES
('00000000-0000-0000-0000-000000011001', '00000000-0000-0000-0000-00000000a002', 'New booking', 'You have a new booking (active)', 'notif-fp-1', 'queued')
ON CONFLICT DO NOTHING;

-- Coupon redemption for cancelled booking
INSERT INTO coupon_redemptions (id, code, booking_id, status, fingerprint) VALUES
('00000000-0000-0000-0000-000000012001', 'DEMO10', '00000000-0000-0000-0000-000000000303', 'used', 'cp-fp-1')
ON CONFLICT DO NOTHING;
