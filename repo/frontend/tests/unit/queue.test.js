/**
 * Unit tests for the offline queue module (src/offline/queue.js).
 * Uses jsdom localStorage (provided by vitest/jest jsdom environment).
 */

// Stub crypto.randomUUID if not available in test environment
if (typeof globalThis.crypto === 'undefined' || typeof globalThis.crypto.randomUUID !== 'function') {
  let counter = 0;
  globalThis.crypto = { randomUUID: () => `test-uuid-${++counter}` };
}

import { getQueue, enqueue, clearQueue, flushQueue, reconcileQueue, removeFromQueue } from '../../src/offline/queue';

beforeEach(() => {
  clearQueue();
});

// ── getQueue / enqueue / clearQueue ─────────────────────────────────────────

test('getQueue returns empty array when queue is empty', () => {
  expect(getQueue()).toEqual([]);
});

test('enqueue adds an item with idempotencyKey and createdAt', () => {
  enqueue({ type: 'booking', payload: { listingId: 'l1' } });
  const queue = getQueue();
  expect(queue).toHaveLength(1);
  expect(queue[0].type).toBe('booking');
  expect(queue[0].payload).toEqual({ listingId: 'l1' });
  expect(typeof queue[0].idempotencyKey).toBe('string');
  expect(typeof queue[0].createdAt).toBe('string');
});

test('enqueue assigns distinct idempotency keys for each item', () => {
  enqueue({ type: 'booking', payload: {} });
  enqueue({ type: 'complaint', payload: {} });
  const [a, b] = getQueue();
  expect(a.idempotencyKey).not.toBe(b.idempotencyKey);
});

test('clearQueue empties the queue', () => {
  enqueue({ type: 'booking', payload: {} });
  clearQueue();
  expect(getQueue()).toEqual([]);
});

// ── flushQueue (per-type handler map) ───────────────────────────────────────

test('flushQueue dispatches each item to its type handler', async () => {
  enqueue({ type: 'booking', payload: { id: 1 } });
  enqueue({ type: 'complaint', payload: { id: 2 } });
  enqueue({ type: 'booking', payload: { id: 3 } });

  const bookingHandler = vi.fn().mockResolvedValue(undefined);
  const complaintHandler = vi.fn().mockResolvedValue(undefined);

  await flushQueue({ booking: bookingHandler, complaint: complaintHandler });

  expect(bookingHandler).toHaveBeenCalledTimes(2);
  expect(complaintHandler).toHaveBeenCalledTimes(1);
  expect(bookingHandler).toHaveBeenCalledWith({ id: 1 }, expect.objectContaining({ type: 'booking' }));
  expect(complaintHandler).toHaveBeenCalledWith({ id: 2 }, expect.objectContaining({ type: 'complaint' }));
  expect(getQueue()).toEqual([]);
});

test('flushQueue supports legacy single-function signature', async () => {
  enqueue({ type: 'inspection', payload: { bookingId: 'b1' } });
  const legacyFn = vi.fn().mockResolvedValue(undefined);
  await flushQueue(legacyFn);
  expect(legacyFn).toHaveBeenCalledTimes(1);
  expect(getQueue()).toEqual([]);
});

test('flushQueue skips items with no registered handler (no throw)', async () => {
  enqueue({ type: 'unknown_type', payload: {} });
  await expect(flushQueue({ booking: vi.fn() })).resolves.not.toThrow();
  expect(getQueue()).toEqual([]);
});

// ── reconcileQueue ───────────────────────────────────────────────────────────

test('reconcileQueue returns zero totals and clears nothing when queue is empty', async () => {
  const apiFetch = vi.fn();
  const result = await reconcileQueue(apiFetch);
  expect(result).toEqual({ applied: 0, total: 0, results: [] });
  expect(apiFetch).not.toHaveBeenCalled();
});

test('reconcileQueue POSTs operations to /sync/reconcile and clears queue on success', async () => {
  enqueue({ type: 'booking', payload: { listingId: 'l1' } });
  enqueue({ type: 'complaint', payload: { bookingId: 'b2' } });

  const serverResponse = { status: 'ok', applied: 2, total: 2, results: [] };
  const apiFetch = vi.fn().mockResolvedValue(serverResponse);

  const result = await reconcileQueue(apiFetch);

  expect(apiFetch).toHaveBeenCalledWith('/sync/reconcile', expect.objectContaining({ method: 'POST' }));
  const body = JSON.parse(apiFetch.mock.calls[0][1].body);
  expect(body.operations).toHaveLength(2);
  expect(body.operations[0].type).toBe('booking');
  expect(body.operations[1].type).toBe('complaint');
  expect(body.operations.every((op) => typeof op.idempotencyKey === 'string')).toBe(true);

  expect(result).toEqual(serverResponse);
  expect(getQueue()).toEqual([]);
});

// ── removeFromQueue ──────────────────────────────────────────────────────────

test('removeFromQueue removes items matching predicate', () => {
  enqueue({ type: 'booking', payload: { id: 1 } });
  enqueue({ type: 'complaint', payload: { id: 2 } });
  enqueue({ type: 'booking', payload: { id: 3 } });
  removeFromQueue((item) => item.type === 'booking');
  const queue = getQueue();
  expect(queue).toHaveLength(1);
  expect(queue[0].type).toBe('complaint');
});

test('removeFromQueue leaves items not matching predicate', () => {
  enqueue({ type: 'inspection', payload: {} });
  enqueue({ type: 'inspection', payload: {} });
  removeFromQueue((item) => item.type === 'booking');
  expect(getQueue()).toHaveLength(2);
});

test('reconcileQueue falls back to flushQueue when apiFetch throws', async () => {
  enqueue({ type: 'booking', payload: { listingId: 'l2' } });
  const apiFetch = vi.fn().mockRejectedValue(new Error('network error'));
  const fallbackFn = vi.fn().mockResolvedValue(undefined);

  await reconcileQueue(apiFetch, { booking: fallbackFn });

  expect(fallbackFn).toHaveBeenCalledTimes(1);
  expect(getQueue()).toEqual([]);
});
