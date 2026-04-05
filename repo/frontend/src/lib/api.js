const API_BASE = import.meta.env.VITE_API_BASE || 'https://localhost:8080/api/v1';

function buildHeaders(extra = {}) {
  const token = localStorage.getItem('token');
  const headers = { 'Content-Type': 'application/json', ...extra };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
}

export async function api(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: buildHeaders(options.headers || {}),
  });
  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: `request failed: ${response.status}` }));
    throw new Error(payload.error || 'request failed');
  }
  if (response.status === 204) {
    return null;
  }
  return response.json();
}

export async function login(payload) {
  const data = await api('/auth/login', { method: 'POST', body: JSON.stringify(payload) });
  localStorage.setItem('token', data.token);
  localStorage.setItem('user', JSON.stringify(data.user));
  return data;
}

export async function logout() {
  try {
    await api('/auth/logout', { method: 'POST' });
  } finally {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  }
}

// Revoke token in background with retries. Accepts an explicit token
// because callers may clear local storage before revoking (optimistic logout).
export async function revokeToken(token, attempts = 3) {
  if (!token) return;
  let lastErr = null;
  for (let i = 0; i < attempts; i++) {
    try {
      const headers = { 'Content-Type': 'application/json' };
      if (token) headers.Authorization = `Bearer ${token}`;
      const resp = await fetch(`${API_BASE}/auth/logout`, { method: 'POST', headers });
      if (resp.ok) return;
      const payload = await resp.json().catch(() => ({}));
      lastErr = new Error(payload.error || `logout failed: ${resp.status}`);
    } catch (err) {
      lastErr = err;
    }
    // exponential-ish backoff
    await new Promise((r) => setTimeout(r, 300 * (i + 1)));
  }
  // don't throw — caller should not block UI. Log for diagnostics.
  // eslint-disable-next-line no-console
  console.warn('revokeToken: failed to revoke token after retries', lastErr);
}

export const me = () => api('/auth/me');
export const updateMe = (payload) => api('/auth/me', { method: 'PATCH', body: JSON.stringify(payload) });
export const loginHistory = () => api('/auth/login-history');
export const statsSummary = () => api('/stats/summary');
export const listings = () => api('/listings');
export const categories = (view = '') => api(`/categories${view ? `?view=${encodeURIComponent(view)}` : ''}`);
export const bookings = () => api('/bookings');
export const estimateBooking = (payload) => api('/bookings/estimate', { method: 'POST', body: JSON.stringify(payload) });
export const createBooking = (payload) => api('/bookings', { method: 'POST', body: JSON.stringify(payload) });
export const inspectionsVerify = (bookingID) => api(`/inspections/verify/${bookingID}`);
export const listInspections = async (bookingId) => {
  const url = `${API_BASE}/inspections?bookingId=${encodeURIComponent(bookingId)}`;
  const headers = buildHeaders();
  const resp = await fetch(url, { headers });
  if (resp.status === 404) return [];
  if (!resp.ok) {
    const payload = await resp.json().catch(() => ({ error: `request failed: ${resp.status}` }));
    throw new Error(payload.error || 'request failed');
  }
  if (resp.status === 204) return null;
  return resp.json();
};
export const submitInspection = (payload) => api('/inspections', { method: 'POST', body: JSON.stringify(payload) });
export const attachmentInit = (payload) => api('/attachments/chunk/init', { method: 'POST', body: JSON.stringify(payload) });
export const attachmentChunk = (payload) => api('/attachments/chunk/upload', { method: 'POST', body: JSON.stringify(payload) });
export const attachmentComplete = (payload) => api('/attachments/chunk/complete', { method: 'POST', body: JSON.stringify(payload) });

// Upload a File using the chunked attachment endpoints. Returns the attachment ID (uploadId).
export async function uploadAttachmentFile(file, bookingId, type = 'photo') {
  const sizeBytes = file.size;
  const buf = await file.arrayBuffer();
  const hashBuf = await crypto.subtle.digest('SHA-256', buf);
  const hashArray = Array.from(new Uint8Array(hashBuf));
  const hex = hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
  const checksum = hex;
  const fingerprint = hex;

  const initResp = await attachmentInit({ bookingId, type, sizeBytes, checksum, fingerprint });
  // If server deduplicated the upload, initResp may include { deduplicated: true, attachment: { id: ... } }
  if (initResp && initResp.deduplicated) {
    const existing = initResp.attachment || initResp.att || initResp;
    const existingId = existing && (existing.id || existing.ID || existing.uploadId);
    if (existingId) return existingId;
  }
  const uploadId = initResp.uploadId || initResp.uploadID || initResp.UploadID || initResp.uploadId;
  // chunk size ~512KB
  const chunkSize = 512 * 1024;
  for (let offset = 0; offset < buf.byteLength; offset += chunkSize) {
    const slice = buf.slice(offset, Math.min(offset + chunkSize, buf.byteLength));
    const u8 = new Uint8Array(slice);
    let binary = '';
    for (let i = 0; i < u8.length; i++) binary += String.fromCharCode(u8[i]);
    const chunkBase64 = btoa(binary);
    await attachmentChunk({ uploadId, chunkBase64 });
  }
  await attachmentComplete({ uploadId });
  return uploadId;
}
export const closeSettlement = (bookingID) => api(`/settlements/close/${bookingID}`, { method: 'POST' });
export const listRatings = (bookingId) => api(`/ratings?bookingId=${encodeURIComponent(bookingId)}`);
export const createRating = (payload) => api('/ratings', { method: 'POST', body: JSON.stringify(payload) });
export const inboxNotifications = () => api('/notifications');
export const adminUsers = () => api('/admin/users');
export const adminCreateUser = (payload) => api('/admin/users', { method: 'POST', body: JSON.stringify(payload) });
export const adminUpdateUser = (id, payload) => api(`/admin/users/${id}`, { method: 'PATCH', body: JSON.stringify(payload) });
export const adminDeleteUser = (id) => api(`/admin/users/${id}`, { method: 'DELETE' });
export const adminCreateCategory = (payload) => api('/admin/categories', { method: 'POST', body: JSON.stringify(payload) });
export const adminCategories = () => api('/admin/categories');
export const adminUpdateCategory = (id, payload) => api(`/admin/categories/${id}`, { method: 'PATCH', body: JSON.stringify(payload) });
export const adminDeleteCategory = (id) => api(`/admin/categories/${id}`, { method: 'DELETE' });
export const adminCreateListing = (payload) => api('/admin/listings', { method: 'POST', body: JSON.stringify(payload) });
export const adminListings = () => api('/admin/listings');
export const adminUpdateListing = (id, payload) => api(`/admin/listings/${id}`, { method: 'PATCH', body: JSON.stringify(payload) });
export const adminDeleteListing = (id) => api(`/admin/listings/${id}`, { method: 'DELETE' });
export const adminSearchListings = (q) => api(`/admin/listings/search${q ? `?q=${encodeURIComponent(q)}` : ''}`);
export const adminBulkListings = (payload) => api('/admin/listings/bulk', { method: 'POST', body: JSON.stringify(payload) });
export const adminNotificationTemplates = () => api('/admin/notification-templates');
export const adminCreateNotificationTemplate = (payload) => api('/admin/notification-templates', { method: 'POST', body: JSON.stringify(payload) });
export const adminSendNotification = (payload) => api('/admin/notifications/send', { method: 'POST', body: JSON.stringify(payload) });
export const adminRetryNotifications = () => api('/admin/notifications/retry', { method: 'POST' });
export const adminWorkerMetrics = () => api('/admin/workers/metrics');
export const complaints = (bookingId = '') => api(`/complaints${bookingId ? `?bookingId=${encodeURIComponent(bookingId)}` : ''}`);
export const createComplaint = (payload) => api('/complaints', { method: 'POST', body: JSON.stringify(payload) });
export const arbitrateComplaint = (id, payload) => api(`/complaints/${id}/arbitrate`, { method: 'PATCH', body: JSON.stringify(payload) });
export async function exportDisputePDF(id) {
  const token = localStorage.getItem('token');
  const resp = await fetch(`${API_BASE}/exports/dispute-pdf/${encodeURIComponent(id)}`, {
    method: 'GET',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!resp.ok) {
    const payload = await resp.json().catch(() => ({ error: `request failed: ${resp.status}` }));
    throw new Error(payload.error || 'request failed');
  }
  return resp.blob();
}
export const consultations = (bookingId) => api(`/consultations?bookingId=${encodeURIComponent(bookingId)}`);
export const consultationsForUser = () => api('/consultations');
export const createConsultation = (payload) => api('/consultations', { method: 'POST', body: JSON.stringify(payload) });
export const consultationAttachments = (id) => api(`/consultations/${id}/attachments`);
export const addConsultationAttachment = (payload) => api('/consultations/attachments', { method: 'POST', body: JSON.stringify(payload) });
export const presignAttachment = (id, ttlSeconds = 60) => api(`/attachments/${id}/presign`, { method: 'POST', body: JSON.stringify({ ttlSeconds }) });

