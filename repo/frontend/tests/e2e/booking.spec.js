const { test, expect } = require('@playwright/test');
const crypto = require('crypto');

test('invalid login - wrong password', async ({ request }) => {
  const login = await request.post('/api/v1/auth/login', {
    data: { username: 'customer', password: 'wrongpassword' },
  });
  expect(login.status()).toBe(401);
});

test('invalid login - wrong username', async ({ request }) => {
  const login = await request.post('/api/v1/auth/login', {
    data: { username: 'nonexistent', password: 'Customer1234!' },
  });
  expect(login.status()).toBe(401);
});

test('booking to dispute PDF flow', async ({ request }) => {
  const login = await request.post('/api/v1/auth/login', {
    data: { username: 'customer', password: 'Customer1234!' },
  });
  expect(login.ok()).toBe(true);
  const token = (await login.json()).token;
  const headers = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' };

  const startAt = new Date().toISOString();
  const endAt = new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString();
  const listingsRes = await request.get('/api/v1/listings', { headers });
  expect(listingsRes.ok()).toBe(true);
  const listingsData = await listingsRes.json();
  if (!Array.isArray(listingsData) || listingsData.length === 0) {
    throw new Error('no listing found for booking test');
  }
  const bookingRes = await request.post('/api/v1/bookings', {
    headers,
    data: {
      listingId: listingsData[0].id,
      startAt,
      endAt,
      odoStart: 10,
      odoEnd: 30,
    },
  });
  if (!bookingRes.ok()) {
    const errBody = await bookingRes.text();
    console.error('booking creation failed', bookingRes.status(), errBody);
  }
  expect(bookingRes.ok()).toBe(true);
  const booking = await bookingRes.json();
  const bookingId = booking.booking.id;

  const payload = Buffer.from('playwright-evidence');
  const checksum = crypto.createHash('sha256').update(payload).digest('hex');
  const init = await request.post('/api/v1/attachments/chunk/init', {
    headers,
    data: {
      bookingId,
      type: 'photo',
      sizeBytes: payload.length,
      checksum,
      fingerprint: `e2e-${Date.now()}`,
    },
  });
  expect(init.ok()).toBe(true);
  const { uploadId } = await init.json();
  await request.post('/api/v1/attachments/chunk/upload', {
    headers,
    data: { uploadId, chunkBase64: payload.toString('base64') },
  });
  const complete = await request.post('/api/v1/attachments/chunk/complete', {
    headers,
    data: { uploadId },
  });
  expect(complete.ok()).toBe(true);

  const inspection = await request.post('/api/v1/inspections', {
    headers,
    data: {
      bookingId,
      stage: 'handover',
      items: [
        {
          name: 'vehicle',
          condition: 'excellent',
          evidenceIds: [uploadId],
        },
      ],
      notes: 'inspection complete',
    },
  });
  expect(inspection.ok()).toBe(true);

  const settlement = await request.post(`/api/v1/settlements/close/${bookingId}`, { headers });
  expect(settlement.ok()).toBe(true);

  const complaint = await request.post('/api/v1/complaints', {
    headers,
    data: { bookingId, outcome: 'minor scratch' },
  });
  expect(complaint.ok()).toBe(true);
  const complaintId = (await complaint.json()).id;

  const pdf = await request.get(`/api/v1/exports/dispute-pdf/${complaintId}`, { headers });
  expect(pdf.status()).toBe(200);
  expect(pdf.headers()['content-type']).toContain('application/pdf');
});
