const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/v1';

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

export const me = () => api('/auth/me');
export const updateMe = (payload) => api('/auth/me', { method: 'PATCH', body: JSON.stringify(payload) });
export const loginHistory = () => api('/auth/login-history');
export const statsSummary = () => api('/stats/summary');
export const listings = () => api('/listings');
export const categories = () => api('/categories');
export const bookings = () => api('/bookings');
export const createBooking = (payload) => api('/bookings', { method: 'POST', body: JSON.stringify(payload) });
export const inspectionsVerify = (bookingID) => api(`/inspections/verify/${bookingID}`);
export const submitInspection = (payload) => api('/inspections', { method: 'POST', body: JSON.stringify(payload) });
export const attachmentInit = (payload) => api('/attachments/chunk/init', { method: 'POST', body: JSON.stringify(payload) });
export const attachmentChunk = (payload) => api('/attachments/chunk/upload', { method: 'POST', body: JSON.stringify(payload) });
export const attachmentComplete = (payload) => api('/attachments/chunk/complete', { method: 'POST', body: JSON.stringify(payload) });
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
export const consultations = (bookingId) => api(`/consultations?bookingId=${encodeURIComponent(bookingId)}`);
export const createConsultation = (payload) => api('/consultations', { method: 'POST', body: JSON.stringify(payload) });
export const consultationAttachments = (id) => api(`/consultations/${id}/attachments`);
export const addConsultationAttachment = (payload) => api('/consultations/attachments', { method: 'POST', body: JSON.stringify(payload) });
