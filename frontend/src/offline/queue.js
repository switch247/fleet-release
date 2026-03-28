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

export async function flushQueue(apiCall) {
  const queue = getQueue();
  for (const item of queue) {
    await apiCall(item);
  }
  clearQueue();
}
