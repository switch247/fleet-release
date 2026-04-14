const KEY = 'fleetlease_offline_queue_v1';

export function getQueue() {
  return JSON.parse(localStorage.getItem(KEY) || '[]');
}

export function enqueue(item) {
  const queue = getQueue();
  queue.push({ ...item, idempotencyKey: crypto.randomUUID(), createdAt: new Date().toISOString() });
  localStorage.setItem(KEY, JSON.stringify(queue));
}

export function clearQueue() {
  localStorage.removeItem(KEY);
}

export function removeFromQueue(predicate) {
  const queue = getQueue();
  localStorage.setItem(KEY, JSON.stringify(queue.filter((item) => !predicate(item))));
}

/**
 * Flush the queue by dispatching each item to a per-type handler function.
 * apiCallMap: { booking: fn, inspection: fn, complaint: fn, ... }
 */
export async function flushQueue(apiCallMap) {
  const queue = getQueue();
  for (const item of queue) {
    if (typeof apiCallMap === 'function') {
      // Legacy single-function signature
      await apiCallMap(item);
    } else {
      const handler = apiCallMap[item.type];
      if (handler) await handler(item.payload, item);
    }
  }
  clearQueue();
}

/**
 * Batch-reconcile the entire queue via the server-side POST /sync/reconcile endpoint.
 * apiFetch(path, options) must be an authenticated fetch wrapper returning parsed JSON.
 * apiFallbackMap is used for per-type fallback when the reconcile endpoint is unavailable.
 */
export async function reconcileQueue(apiFetch, apiFallbackMap = {}) {
  const queue = getQueue();
  if (queue.length === 0) return { applied: 0, total: 0, results: [] };

  const operations = queue.map((item) => ({
    idempotencyKey: item.idempotencyKey,
    type: item.type,
    payload: item.payload,
  }));

  try {
    const result = await apiFetch('/sync/reconcile', {
      method: 'POST',
      body: JSON.stringify({ operations }),
    });
    clearQueue();
    return result;
  } catch (_) {
    await flushQueue(apiFallbackMap);
    return { applied: queue.length, total: queue.length, results: [] };
  }
}
