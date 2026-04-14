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

  const payload = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8Xw8AAoMBgR5nYeEAAAAASUVORK5CYII=', 'base64');
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

test('notifications endpoint returns 200 for authenticated user', async ({ request }) => {
  const login = await request.post('/api/v1/auth/login', {
    data: { username: 'customer', password: 'Customer1234!' },
  });
  expect(login.ok()).toBe(true);
  const token = (await login.json()).token;
  const headers = { Authorization: `Bearer ${token}` };

  const notifications = await request.get('/api/v1/notifications', { headers });
  expect(notifications.status()).toBe(200);
  const body = await notifications.json();
  expect(Array.isArray(body)).toBe(true);
});

test('unauthenticated access to notifications returns 401', async ({ request }) => {
  const resp = await request.get('/api/v1/notifications');
  expect(resp.status()).toBe(401);
});

test('sync reconcile with empty operations returns reconciled status', async ({ request }) => {
  const login = await request.post('/api/v1/auth/login', {
    data: { username: 'customer', password: 'Customer1234!' },
  });
  expect(login.ok()).toBe(true);
  const token = (await login.json()).token;
  const headers = { Authorization: `Bearer ${token}` };

  const resp = await request.post('/api/v1/sync/reconcile', {
    headers,
    data: { operations: [] },
  });
  expect(resp.status()).toBe(200);
  const body = await resp.json();
  expect(body.status).toBe('reconciled');
  expect(body.applied).toBe(0);
});

test('sync reconcile applies queued booking operation', async ({ request }) => {
  const login = await request.post('/api/v1/auth/login', {
    data: { username: 'customer', password: 'Customer1234!' },
  });
  expect(login.ok()).toBe(true);
  const token = (await login.json()).token;
  const headers = { Authorization: `Bearer ${token}` };

  const listingsRes = await request.get('/api/v1/listings', { headers });
  expect(listingsRes.ok()).toBe(true);
  const listingsData = await listingsRes.json();
  if (!Array.isArray(listingsData) || listingsData.length === 0) {
    throw new Error('no listing found for reconcile test');
  }

  const startAt = new Date().toISOString();
  const endAt = new Date(Date.now() + 3 * 60 * 60 * 1000).toISOString();

  const reconcileResp = await request.post('/api/v1/sync/reconcile', {
    headers,
    data: {
      operations: [
        {
          idempotencyKey: `test-${Date.now()}`,
          type: 'booking',
          payload: {
            listingId: listingsData[0].id,
            startAt,
            endAt,
            odoStart: 100,
            odoEnd: 150,
            couponCode: '',
          },
        },
      ],
    },
  });
  expect(reconcileResp.status()).toBe(200);
  const reconcileBody = await reconcileResp.json();
  expect(reconcileBody.status).toBe('reconciled');
  expect(reconcileBody.applied).toBe(1);
  expect(reconcileBody.results[0].status).toBe('applied');
});

test('cross-user booking access is denied (tenant isolation)', async ({ request }) => {
  // Login as customer to get their booking ID
  const loginA = await request.post('/api/v1/auth/login', {
    data: { username: 'customer', password: 'Customer1234!' },
  });
  expect(loginA.ok()).toBe(true);
  const tokenA = (await loginA.json()).token;
  const headersA = { Authorization: `Bearer ${tokenA}` };

  const startAt = new Date().toISOString();
  const endAt = new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString();
  const listingsRes = await request.get('/api/v1/listings', { headers: headersA });
  const listingsData = await listingsRes.json();
  if (!Array.isArray(listingsData) || listingsData.length === 0) return;

  const bookingRes = await request.post('/api/v1/bookings', {
    headers: headersA,
    data: { listingId: listingsData[0].id, startAt, endAt, odoStart: 0, odoEnd: 10 },
  });
  expect(bookingRes.ok()).toBe(true);
  const bookingId = (await bookingRes.json()).booking.id;

  // Login as provider — should not be able to access customer's ledger
  const loginB = await request.post('/api/v1/auth/login', {
    data: { username: 'provider', password: 'Provider1234!' },
  });
  expect(loginB.ok()).toBe(true);
  const tokenB = (await loginB.json()).token;
  const headersB = { Authorization: `Bearer ${tokenB}` };

  const ledgerResp = await request.get(`/api/v1/ledger/${bookingId}`, { headers: headersB });
  expect(ledgerResp.status()).toBe(403);
});
