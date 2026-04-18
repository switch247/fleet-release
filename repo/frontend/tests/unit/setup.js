import '@testing-library/jest-dom';

// Stub ResizeObserver (used by recharts, not available in jsdom)
if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
}

// Stub crypto.randomUUID for offline queue and crypto.subtle.digest for file checksum
{
  let _counter = 0;
  const needsRandomUUID = typeof globalThis.crypto === 'undefined' || typeof globalThis.crypto?.randomUUID !== 'function';
  const needsDigest = typeof globalThis.crypto?.subtle?.digest !== 'function';
  if (needsRandomUUID || needsDigest) {
    globalThis.crypto = {
      ...(globalThis.crypto || {}),
      randomUUID: needsRandomUUID ? () => `test-uuid-${++_counter}` : globalThis.crypto.randomUUID,
      subtle: {
        ...(globalThis.crypto?.subtle || {}),
        digest: needsDigest
          ? (_alg, _data) => Promise.resolve(new ArrayBuffer(32))
          : globalThis.crypto.subtle.digest.bind(globalThis.crypto.subtle),
      },
    };
  }
}
