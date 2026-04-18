/**
 * Unit tests for src/lib/api.js
 * Mocks global fetch; tests the request/response helpers and named exports.
 */

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

import {
  api,
  login,
  logout,
  revokeToken,
  me,
  updateMe,
  bookings,
  listings,
  categories,
  createBooking,
  estimateBooking,
  createRating,
  listRatings,
  inboxNotifications,
  adminUsers,
  uploadAttachmentFile,
  exportDisputePDF,
  loginHistory,
  statsSummary,
  listInspections,
  submitInspection,
  closeSettlement,
  complaints,
  consultationsForUser,
  // admin users
  adminCreateUser,
  adminUpdateUser,
  adminDeleteUser,
  // admin categories
  adminCategories,
  adminCreateCategory,
  adminUpdateCategory,
  adminDeleteCategory,
  // admin listings
  adminListings,
  adminCreateListing,
  adminUpdateListing,
  adminDeleteListing,
  adminSearchListings,
  adminBulkListings,
  // admin notifications
  adminNotificationTemplates,
  adminCreateNotificationTemplate,
  adminSendNotification,
  adminRetryNotifications,
  // complaints / consultations
  arbitrateComplaint,
  createComplaint,
  consultations,
  createConsultation,
  consultationAttachments,
  addConsultationAttachment,
  presignAttachment,
  // misc
  inspectionsVerify,
  apiFetch,
} from '../../../src/lib/api';

beforeEach(() => {
  localStorage.clear();
});

// ── helpers ──────────────────────────────────────────────────────────────────

function ok(data, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: vi.fn().mockResolvedValue(data),
    blob: vi.fn().mockResolvedValue(new Blob()),
  };
}

function fail(status, error = 'something went wrong') {
  return {
    ok: false,
    status,
    json: vi.fn().mockResolvedValue({ error }),
    blob: vi.fn().mockResolvedValue(new Blob()),
  };
}

// ── api() core ───────────────────────────────────────────────────────────────

describe('api()', () => {
  test('builds correct URL from path', async () => {
    mockFetch.mockResolvedValue(ok({ ok: true }));
    await api('/test-path');
    expect(mockFetch.mock.calls[0][0]).toContain('/test-path');
  });

  test('sends Content-Type: application/json header', async () => {
    mockFetch.mockResolvedValue(ok({}));
    await api('/x');
    expect(mockFetch.mock.calls[0][1].headers['Content-Type']).toBe('application/json');
  });

  test('adds Authorization header when token is in localStorage', async () => {
    localStorage.setItem('token', 'bearer-tok');
    mockFetch.mockResolvedValue(ok({}));
    await api('/secure');
    expect(mockFetch.mock.calls[0][1].headers.Authorization).toBe('Bearer bearer-tok');
  });

  test('does not add Authorization header when no token', async () => {
    mockFetch.mockResolvedValue(ok({}));
    await api('/public');
    expect(mockFetch.mock.calls[0][1].headers.Authorization).toBeUndefined();
  });

  test('returns parsed JSON on success', async () => {
    mockFetch.mockResolvedValue(ok({ data: 42 }));
    const result = await api('/data');
    expect(result).toEqual({ data: 42 });
  });

  test('returns null on 204 No Content', async () => {
    mockFetch.mockResolvedValue({ ok: true, status: 204, json: vi.fn() });
    const result = await api('/empty');
    expect(result).toBeNull();
  });

  test('throws error from response body on failure', async () => {
    mockFetch.mockResolvedValue(fail(404, 'Not found'));
    await expect(api('/missing')).rejects.toThrow('Not found');
  });

  test('throws fallback message when error body cannot be parsed', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      json: vi.fn().mockRejectedValue(new Error('bad json')),
    });
    await expect(api('/boom')).rejects.toThrow('request failed: 500');
  });

  test('passes method and body options through to fetch', async () => {
    mockFetch.mockResolvedValue(ok({ id: '1' }));
    await api('/items', { method: 'POST', body: '{"name":"x"}' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][1].body).toBe('{"name":"x"}');
  });
});

// ── login() ──────────────────────────────────────────────────────────────────

describe('login()', () => {
  test('stores token in localStorage', async () => {
    mockFetch.mockResolvedValue(ok({ token: 'tok-abc', user: { username: 'alice' } }));
    await login({ username: 'alice', password: 'p' });
    expect(localStorage.getItem('token')).toBe('tok-abc');
  });

  test('stores serialised user in localStorage', async () => {
    const user = { id: '1', username: 'alice', roles: ['customer'] };
    mockFetch.mockResolvedValue(ok({ token: 't', user }));
    await login({ username: 'alice', password: 'p' });
    expect(JSON.parse(localStorage.getItem('user'))).toEqual(user);
  });

  test('returns the full response', async () => {
    const resp = { token: 't', user: { username: 'bob' } };
    mockFetch.mockResolvedValue(ok(resp));
    expect(await login({ username: 'bob', password: 'p' })).toEqual(resp);
  });
});

// ── logout() ─────────────────────────────────────────────────────────────────

describe('logout()', () => {
  test('clears token from localStorage after API call', async () => {
    localStorage.setItem('token', 'old-tok');
    mockFetch.mockResolvedValue({ ok: true, status: 204, json: vi.fn() });
    await logout();
    expect(localStorage.getItem('token')).toBeNull();
  });

  test('clears user from localStorage after API call', async () => {
    localStorage.setItem('user', '{"username":"alice"}');
    mockFetch.mockResolvedValue({ ok: true, status: 204, json: vi.fn() });
    await logout();
    expect(localStorage.getItem('user')).toBeNull();
  });

  test('still clears storage even when API call throws', async () => {
    localStorage.setItem('token', 'tok');
    mockFetch.mockRejectedValue(new Error('network error'));
    // logout() uses try/finally, so it re-throws after clearing storage
    await expect(logout()).rejects.toThrow('network error');
    expect(localStorage.getItem('token')).toBeNull();
  });
});

// ── revokeToken() ─────────────────────────────────────────────────────────────

describe('revokeToken()', () => {
  test('does nothing when token is null', async () => {
    await revokeToken(null);
    expect(mockFetch).not.toHaveBeenCalled();
  });

  test('does nothing when token is empty string', async () => {
    await revokeToken('');
    expect(mockFetch).not.toHaveBeenCalled();
  });

  test('POSTs to /auth/logout with the given token', async () => {
    mockFetch.mockResolvedValue({ ok: true, status: 204, json: vi.fn() });
    await revokeToken('my-token');
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/auth/logout'),
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ Authorization: 'Bearer my-token' }),
      })
    );
  });

  test('logs a warning (not throw) after all retries fail', async () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
    mockFetch.mockResolvedValue({ ok: false, status: 500, json: vi.fn().mockResolvedValue({ error: 'err' }) });
    await expect(revokeToken('tok', 1)).resolves.not.toThrow();
    expect(warnSpy).toHaveBeenCalled();
    warnSpy.mockRestore();
  });
});

// ── named endpoint wrappers ──────────────────────────────────────────────────

describe('named API wrappers', () => {
  beforeEach(() => {
    mockFetch.mockResolvedValue(ok([]));
  });

  test('me() calls /auth/me', async () => {
    mockFetch.mockResolvedValue(ok({ username: 'alice' }));
    await me();
    expect(mockFetch.mock.calls[0][0]).toContain('/auth/me');
  });

  test('updateMe() sends PATCH to /auth/me', async () => {
    mockFetch.mockResolvedValue(ok({ username: 'alice' }));
    await updateMe({ email: 'a@b.com' });
    expect(mockFetch.mock.calls[0][1].method).toBe('PATCH');
    expect(mockFetch.mock.calls[0][0]).toContain('/auth/me');
  });

  test('bookings() calls /bookings', async () => {
    await bookings();
    expect(mockFetch.mock.calls[0][0]).toContain('/bookings');
  });

  test('listings() calls /listings', async () => {
    await listings();
    expect(mockFetch.mock.calls[0][0]).toContain('/listings');
  });

  test('categories() calls /categories', async () => {
    await categories();
    expect(mockFetch.mock.calls[0][0]).toContain('/categories');
  });

  test('categories() appends view param when provided', async () => {
    await categories('tree');
    expect(mockFetch.mock.calls[0][0]).toContain('view=tree');
  });

  test('createBooking() sends POST to /bookings', async () => {
    mockFetch.mockResolvedValue(ok({ booking: { id: 'b1' } }));
    await createBooking({ listingId: 'l1' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/bookings');
  });

  test('estimateBooking() sends POST to /bookings/estimate', async () => {
    mockFetch.mockResolvedValue(ok({ estimate: {} }));
    await estimateBooking({ listingId: 'l1' });
    expect(mockFetch.mock.calls[0][0]).toContain('/bookings/estimate');
  });

  test('inboxNotifications() calls /notifications', async () => {
    await inboxNotifications();
    expect(mockFetch.mock.calls[0][0]).toContain('/notifications');
  });

  test('adminUsers() calls /admin/users', async () => {
    await adminUsers();
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/users');
  });

  test('listRatings() appends bookingId param', async () => {
    await listRatings('booking-123');
    expect(mockFetch.mock.calls[0][0]).toContain('booking-123');
  });

  test('createRating() sends POST to /ratings', async () => {
    mockFetch.mockResolvedValue(ok({ id: 'r1' }));
    await createRating({ score: 5 });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/ratings');
  });

  test('loginHistory() calls /auth/login-history', async () => {
    await loginHistory();
    expect(mockFetch.mock.calls[0][0]).toContain('/auth/login-history');
  });

  test('statsSummary() calls /stats/summary', async () => {
    await statsSummary();
    expect(mockFetch.mock.calls[0][0]).toContain('/stats/summary');
  });

  test('listInspections() calls /inspections', async () => {
    mockFetch.mockResolvedValue(ok({ revisions: [] }));
    await listInspections('b1');
    expect(mockFetch.mock.calls[0][0]).toContain('/inspections');
  });

  test('submitInspection() sends POST to /inspections', async () => {
    mockFetch.mockResolvedValue(ok({ id: 'i1' }));
    await submitInspection({ bookingId: 'b1', stage: 'handover', items: [] });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/inspections');
  });

  test('closeSettlement() sends POST to /settlements/close/:id', async () => {
    mockFetch.mockResolvedValue(ok({ booking: {}, ledger: [] }));
    await closeSettlement('b1');
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/settlements/close/b1');
  });

  test('complaints() calls /complaints with optional bookingId', async () => {
    await complaints('b1');
    expect(mockFetch.mock.calls[0][0]).toContain('bookingId=b1');
  });

  test('complaints() calls /complaints without param when empty', async () => {
    await complaints();
    expect(mockFetch.mock.calls[0][0]).toContain('/complaints');
    expect(mockFetch.mock.calls[0][0]).not.toContain('bookingId');
  });

  test('consultationsForUser() calls /consultations', async () => {
    await consultationsForUser();
    expect(mockFetch.mock.calls[0][0]).toContain('/consultations');
  });
});

// ── uploadAttachmentFile() ────────────────────────────────────────────────────

describe('uploadAttachmentFile()', () => {
  // jsdom Blob does not implement arrayBuffer(); use a plain mock file object instead.
  function mockFile(content = 'hello') {
    const buf = new TextEncoder().encode(content).buffer;
    return {
      size: content.length,
      type: 'image/jpeg',
      arrayBuffer: vi.fn().mockResolvedValue(buf),
    };
  }

  beforeEach(() => {
    mockFetch.mockReset();
    // init → upload → complete
    mockFetch
      .mockResolvedValueOnce(ok({ uploadId: 'up-123' }))    // attachmentInit
      .mockResolvedValueOnce(ok({}))                          // attachmentChunk
      .mockResolvedValueOnce(ok({ id: 'up-123' }));           // attachmentComplete
  });

  test('returns the uploadId on success', async () => {
    const result = await uploadAttachmentFile(mockFile('hello'), 'b1', 'photo');
    expect(result).toBe('up-123');
  });

  test('calls attachmentInit with bookingId and type', async () => {
    await uploadAttachmentFile(mockFile('test'), 'booking-abc', 'photo');
    expect(mockFetch.mock.calls[0][0]).toContain('/attachments/chunk/init');
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body.bookingId).toBe('booking-abc');
    expect(body.type).toBe('photo');
  });

  test('calls attachmentChunk after init', async () => {
    await uploadAttachmentFile(mockFile('data'), 'b1');
    expect(mockFetch.mock.calls[1][0]).toContain('/attachments/chunk/upload');
  });

  test('calls attachmentComplete after chunking', async () => {
    await uploadAttachmentFile(mockFile('data'), 'b1');
    expect(mockFetch.mock.calls[2][0]).toContain('/attachments/chunk/complete');
  });

  test('skips upload when server deduplicates', async () => {
    mockFetch.mockReset();
    // init returns deduplicated=true
    mockFetch.mockResolvedValueOnce(ok({ deduplicated: true, attachment: { id: 'existing-id' } }));
    const result = await uploadAttachmentFile(mockFile('data'), 'b1');
    expect(result).toBe('existing-id');
    // Only init was called, no chunk or complete
    expect(mockFetch).toHaveBeenCalledTimes(1);
  });
});

// ── exportDisputePDF() ────────────────────────────────────────────────────────

describe('exportDisputePDF()', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  test('returns a Blob on success', async () => {
    const blob = new Blob(['%PDF-1.4'], { type: 'application/pdf' });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      blob: vi.fn().mockResolvedValue(blob),
      json: vi.fn(),
    });
    const result = await exportDisputePDF('complaint-1');
    expect(result).toBeInstanceOf(Blob);
  });

  test('calls the correct URL', async () => {
    const blob = new Blob(['%PDF'], { type: 'application/pdf' });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      blob: vi.fn().mockResolvedValue(blob),
      json: vi.fn(),
    });
    await exportDisputePDF('c-abc');
    expect(mockFetch.mock.calls[0][0]).toContain('/exports/dispute-pdf/c-abc');
  });

  test('throws when response is not ok', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: vi.fn().mockResolvedValue({ error: 'not found' }),
    });
    await expect(exportDisputePDF('missing')).rejects.toThrow('not found');
  });

  test('sends Authorization header when token present', async () => {
    localStorage.setItem('token', 'pdf-tok');
    const blob = new Blob(['%PDF'], { type: 'application/pdf' });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      blob: vi.fn().mockResolvedValue(blob),
      json: vi.fn(),
    });
    await exportDisputePDF('x');
    expect(mockFetch.mock.calls[0][1].headers.Authorization).toBe('Bearer pdf-tok');
  });
});

// ── listInspections() error paths ─────────────────────────────────────────────

describe('listInspections() error paths', () => {
  beforeEach(() => { mockFetch.mockReset(); });

  test('returns [] on 404', async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 404, json: vi.fn().mockResolvedValue({}) });
    const result = await listInspections('b1');
    expect(result).toEqual([]);
  });

  test('throws on non-ok non-404 response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: vi.fn().mockResolvedValue({ error: 'server error' }),
    });
    await expect(listInspections('b1')).rejects.toThrow('server error');
  });

  test('returns null on 204', async () => {
    mockFetch.mockResolvedValueOnce({ ok: true, status: 204, json: vi.fn() });
    const result = await listInspections('b1');
    expect(result).toBeNull();
  });
});

// ── apiFetch alias ────────────────────────────────────────────────────────────

describe('apiFetch', () => {
  beforeEach(() => { mockFetch.mockReset(); });

  test('apiFetch is same as api', async () => {
    mockFetch.mockResolvedValueOnce(ok({ ok: true }));
    const result = await apiFetch('/test');
    expect(result).toEqual({ ok: true });
  });
});

// ── admin user endpoints ──────────────────────────────────────────────────────

describe('admin user endpoints', () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockFetch.mockResolvedValue(ok({}));
  });

  test('adminCreateUser() sends POST to /admin/users', async () => {
    await adminCreateUser({ username: 'u', email: 'e@e.com', password: 'p', roles: ['customer'] });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/users');
  });

  test('adminUpdateUser() sends PATCH to /admin/users/:id', async () => {
    await adminUpdateUser('uid1', { email: 'new@e.com' });
    expect(mockFetch.mock.calls[0][1].method).toBe('PATCH');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/users/uid1');
  });

  test('adminDeleteUser() sends DELETE to /admin/users/:id', async () => {
    await adminDeleteUser('uid1');
    expect(mockFetch.mock.calls[0][1].method).toBe('DELETE');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/users/uid1');
  });
});

// ── admin category endpoints ──────────────────────────────────────────────────

describe('admin category endpoints', () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockFetch.mockResolvedValue(ok([]));
  });

  test('adminCategories() calls /admin/categories', async () => {
    await adminCategories();
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/categories');
  });

  test('adminCreateCategory() sends POST to /admin/categories', async () => {
    await adminCreateCategory({ name: 'Trucks' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/categories');
  });

  test('adminUpdateCategory() sends PATCH to /admin/categories/:id', async () => {
    await adminUpdateCategory('c1', { name: 'Big Trucks' });
    expect(mockFetch.mock.calls[0][1].method).toBe('PATCH');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/categories/c1');
  });

  test('adminDeleteCategory() sends DELETE to /admin/categories/:id', async () => {
    await adminDeleteCategory('c1');
    expect(mockFetch.mock.calls[0][1].method).toBe('DELETE');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/categories/c1');
  });
});

// ── admin listing endpoints ───────────────────────────────────────────────────

describe('admin listing endpoints', () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockFetch.mockResolvedValue(ok([]));
  });

  test('adminListings() calls /admin/listings', async () => {
    await adminListings();
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/listings');
  });

  test('adminCreateListing() sends POST to /admin/listings', async () => {
    await adminCreateListing({ name: 'Car' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/listings');
  });

  test('adminUpdateListing() sends PATCH to /admin/listings/:id', async () => {
    await adminUpdateListing('l1', { name: 'Updated Car' });
    expect(mockFetch.mock.calls[0][1].method).toBe('PATCH');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/listings/l1');
  });

  test('adminDeleteListing() sends DELETE to /admin/listings/:id', async () => {
    await adminDeleteListing('l1');
    expect(mockFetch.mock.calls[0][1].method).toBe('DELETE');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/listings/l1');
  });

  test('adminSearchListings() appends query param', async () => {
    await adminSearchListings('camry');
    expect(mockFetch.mock.calls[0][0]).toContain('q=camry');
  });

  test('adminSearchListings() works with empty query', async () => {
    await adminSearchListings('');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/listings/search');
    expect(mockFetch.mock.calls[0][0]).not.toContain('?q=');
  });

  test('adminBulkListings() sends POST to /admin/listings/bulk', async () => {
    await adminBulkListings([{ name: 'Car A' }]);
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/listings/bulk');
  });
});

// ── admin notification endpoints ──────────────────────────────────────────────

describe('admin notification endpoints', () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockFetch.mockResolvedValue(ok({}));
  });

  test('adminNotificationTemplates() calls /admin/notification-templates', async () => {
    await adminNotificationTemplates();
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/notification-templates');
  });

  test('adminCreateNotificationTemplate() sends POST', async () => {
    await adminCreateNotificationTemplate({ name: 'Alert', subject: 'Hi' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/notification-templates');
  });

  test('adminSendNotification() sends POST to /admin/notifications/send', async () => {
    await adminSendNotification({ templateId: 't1', userId: 'u1' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/notifications/send');
  });

  test('adminRetryNotifications() sends POST to /admin/notifications/retry', async () => {
    await adminRetryNotifications();
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/admin/notifications/retry');
  });
});

// ── complaint & consultation endpoints ───────────────────────────────────────

describe('complaint & consultation endpoints', () => {
  beforeEach(() => {
    mockFetch.mockReset();
    mockFetch.mockResolvedValue(ok({}));
  });

  test('createComplaint() sends POST to /complaints', async () => {
    await createComplaint({ bookingId: 'b1', outcome: 'scratch' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/complaints');
  });

  test('arbitrateComplaint() sends PATCH to /complaints/:id/arbitrate', async () => {
    await arbitrateComplaint('c1', { resolution: 'upheld' });
    expect(mockFetch.mock.calls[0][1].method).toBe('PATCH');
    expect(mockFetch.mock.calls[0][0]).toContain('/complaints/c1/arbitrate');
  });

  test('consultations() calls /consultations?bookingId=...', async () => {
    await consultations('b1');
    expect(mockFetch.mock.calls[0][0]).toContain('/consultations');
    expect(mockFetch.mock.calls[0][0]).toContain('bookingId=b1');
  });

  test('createConsultation() sends POST to /consultations', async () => {
    await createConsultation({ topic: 'Engine' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/consultations');
  });

  test('consultationAttachments() calls /consultations/:id/attachments', async () => {
    await consultationAttachments('con1');
    expect(mockFetch.mock.calls[0][0]).toContain('/consultations/con1/attachments');
  });

  test('addConsultationAttachment() sends POST to /consultations/attachments', async () => {
    await addConsultationAttachment({ consultationId: 'con1', attachmentId: 'att1' });
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/consultations/attachments');
  });

  test('presignAttachment() sends POST to /attachments/:id/presign', async () => {
    await presignAttachment('att1', 120);
    expect(mockFetch.mock.calls[0][1].method).toBe('POST');
    expect(mockFetch.mock.calls[0][0]).toContain('/attachments/att1/presign');
    expect(JSON.parse(mockFetch.mock.calls[0][1].body).ttlSeconds).toBe(120);
  });

  test('inspectionsVerify() calls /inspections/verify/:id', async () => {
    await inspectionsVerify('b1');
    expect(mockFetch.mock.calls[0][0]).toContain('/inspections/verify/b1');
  });
});
