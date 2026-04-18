// --- Ensure crypto.subtle.digest always resolves for SHA-256 in jsdom tests ---
beforeAll(() => {
  if (typeof globalThis.crypto !== 'undefined' && globalThis.crypto.subtle) {
    if (!globalThis.crypto.subtle.__mockedForTest) {
      vi.spyOn(globalThis.crypto.subtle, 'digest').mockImplementation(async (alg, data) => {
        return new Uint8Array(32).buffer;
      });
      globalThis.crypto.subtle.__mockedForTest = true;
    }
  } else if (typeof globalThis.crypto !== 'undefined') {
    globalThis.crypto.subtle = {
      digest: async (algorithm, data) => {
        return new Uint8Array(32).buffer;
      },
    };
  }
});
// Polyfill File.prototype.arrayBuffer for jsdom if missing
if (typeof File !== 'undefined' && !File.prototype.arrayBuffer) {
  File.prototype.arrayBuffer = function () {
    return Promise.resolve(new Uint8Array([]).buffer);
  };
}

// Polyfill crypto.subtle.digest for jsdom to accept Uint8Array/ArrayBuffer
if (typeof globalThis.crypto !== 'undefined' && globalThis.crypto.subtle && !globalThis.crypto.subtle.__patchedForTest) {
  const origDigest = globalThis.crypto.subtle.digest;
  globalThis.crypto.subtle.digest = function (algorithm, data) {
    // Accept Uint8Array, Buffer, DataView, or ArrayBuffer
    if (data instanceof Uint8Array) {
      data = data.buffer;
    } else if (ArrayBuffer.isView(data)) {
      data = data.buffer;
    }
    return origDigest.call(this, algorithm, data);
  };
  globalThis.crypto.subtle.digest.__patchedForTest = true;
}
// Force-override File.prototype.arrayBuffer to prevent cross-realm ArrayBuffer
// issues in jsdom where the native SubtleCrypto rejects jsdom's internal Blob buffer.
if (typeof File !== 'undefined') {
  File.prototype.arrayBuffer = function () {
    return Promise.resolve(new ArrayBuffer(this.size || 9));
  };
}
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/lib/api', () => ({
  bookings: vi.fn(),
  listInspections: vi.fn(),
  submitInspection: vi.fn(),
  closeSettlement: vi.fn(),
  attachmentInit: vi.fn(),
  attachmentChunk: vi.fn(),
  attachmentComplete: vi.fn(),
}));

vi.mock('../../../src/offline/queue', () => ({
  enqueue: vi.fn(),
  getQueue: vi.fn().mockReturnValue([]),
  removeFromQueue: vi.fn(),
}));

import { bookings, listInspections, submitInspection, closeSettlement, attachmentInit, attachmentChunk, attachmentComplete } from '../../../src/lib/api';
import { getQueue } from '../../../src/offline/queue';
import InspectionsPage from '../../../src/pages/InspectionsPage';

const sampleBookings = [
  { id: 'b1', status: 'active', startAt: '2024-01-10T10:00:00Z', endAt: '2024-01-11T10:00:00Z' },
];

const sampleInspections = [
  {
    revisionId: 'rev1',
    stage: 'handover',
    createdAt: '2024-01-10T10:00:00Z',
    items: [{ name: 'Exterior bodywork', condition: 'good', evidenceIds: ['att1'] }],
  },
];

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <InspectionsPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => {
  bookings.mockResolvedValue(sampleBookings);
  listInspections.mockResolvedValue([]);
  submitInspection.mockResolvedValue({ ok: true });
  closeSettlement.mockResolvedValue({ booking: { id: 'b1' }, ledger: [] });
  getQueue.mockReturnValue([]);
  attachmentInit.mockResolvedValue({ uploadId: 'up1', deduplicated: false });
  attachmentChunk.mockResolvedValue({ ok: true });
  attachmentComplete.mockResolvedValue({ ok: true });
});

describe('InspectionsPage', () => {
  test('renders Guided Inspection heading', () => {
    renderPage();
    expect(screen.getByText('Guided Inspection & Settlement')).toBeInTheDocument();
  });

  test('shows camera evidence requirement message', () => {
    renderPage();
    expect(screen.getByText(/camera evidence is mandatory/i)).toBeInTheDocument();
  });

  test('renders Bookings card', () => {
    renderPage();
    expect(screen.getByText('Bookings')).toBeInTheDocument();
  });

  test('shows no booking selected message initially', () => {
    renderPage();
    expect(screen.getByText(/no booking selected/i)).toBeInTheDocument();
  });

  test('shows bookings after load', async () => {
    renderPage();
    expect(await screen.findByText(/b1/)).toBeInTheDocument();
  });

  test('shows Start Inspection button for each booking', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    expect(screen.getByRole('button', { name: /start inspection/i })).toBeInTheDocument();
  });

  test('shows View Inspections button for each booking', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    expect(screen.getByRole('button', { name: /view inspections/i })).toBeInTheDocument();
  });

  test('Open Inspection Modal button is disabled when no booking selected', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /open inspection modal/i })).toBeDisabled();
  });

  test('clicking Start Inspection opens modal', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await waitFor(() =>
      expect(screen.getByText('Inspection Wizard')).toBeInTheDocument()
    );
  });

  test('modal shows step 1 content initially', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText('Inspection Wizard');
    expect(screen.getByText(/step 1 of 3/i)).toBeInTheDocument();
  });

  test('clicking Next advances to step 2', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await waitFor(() =>
      expect(screen.getByText(/step 2 of 3/i)).toBeInTheDocument()
    );
  });

  test('step 2 shows checklist items', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await waitFor(() =>
      expect(screen.getByText('Exterior bodywork')).toBeInTheDocument()
    );
    expect(screen.getByText('Tires and wheels')).toBeInTheDocument();
  });

  test('step 2 has Back button', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /back/i })).toBeInTheDocument()
    );
  });

  test('clicking View Inspections selects the booking', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    // "Selected booking: <strong>b1</strong>" — text is split across elements
    await waitFor(() =>
      expect(screen.getByText('b1', { selector: 'strong' })).toBeInTheDocument()
    );
  });

  test('shows completed inspections card after booking is selected', async () => {
    listInspections.mockResolvedValue(sampleInspections);
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await waitFor(() =>
      expect(screen.getByText('Completed Inspections')).toBeInTheDocument()
    );
  });

  test('shows no inspections message when empty', async () => {
    listInspections.mockResolvedValue([]);
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await waitFor(() =>
      expect(screen.getByText('No inspections recorded for this booking.')).toBeInTheDocument()
    );
  });

  test('shows empty bookings message when none', async () => {
    bookings.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No bookings available.')).toBeInTheDocument()
    );
  });

  // ── Step 1 interactions ─────────────────────────────────────────────────────

  test('stage dropdown in step 1 allows switching to return', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    const stageSelect = screen.getByDisplayValue('Handover');
    fireEvent.change(stageSelect, { target: { value: 'return' } });
    expect(screen.getByDisplayValue('Return')).toBeInTheDocument();
  });

  test('notes input in step 1 accepts text', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    const notesInput = screen.getByPlaceholderText('Notes');
    fireEvent.change(notesInput, { target: { value: 'Good condition overall' } });
    expect(notesInput.value).toBe('Good condition overall');
  });

  // ── Step 2 interactions ─────────────────────────────────────────────────────

  test('condition dropdown in step 2 can be changed to minor wear', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await screen.findByText(/step 2 of 3/i);
    const conditionSelects = screen.getAllByDisplayValue('Good');
    fireEvent.change(conditionSelects[0], { target: { value: 'minor' } });
    expect(screen.getByDisplayValue('Minor Wear')).toBeInTheDocument();
  });

  test('step 2 shows file input per checklist item', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await screen.findByText(/step 2 of 3/i);
    const fileInputs = document.querySelectorAll('input[type="file"]');
    expect(fileInputs.length).toBe(4); // one per BASE_ITEMS entry
  });

  // ── Step 3 ──────────────────────────────────────────────────────────────────

  async function openStep3() {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await screen.findByText(/step 2 of 3/i);
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await screen.findByText(/step 3 of 3/i);
  }

  test('clicking Next on step 2 advances to step 3', async () => {
    await openStep3();
    expect(screen.getByText(/step 3 of 3/i)).toBeInTheDocument();
  });

  test('step 3 shows review heading', async () => {
    await openStep3();
    expect(screen.getByText('Review & Proposed Deductions')).toBeInTheDocument();
  });

  test('step 3 shows proposed deductions total of zero when all items are good', async () => {
    await openStep3();
    expect(screen.getByText(/proposed deductions total: \$0\.00/i)).toBeInTheDocument();
  });

  test('step 3 shows Submit Inspection button', async () => {
    await openStep3();
    expect(screen.getByRole('button', { name: /submit inspection/i })).toBeInTheDocument();
  });

  test('step 3 shows deduction badge when item condition is major', async () => {
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
    await screen.findByText(/step 1 of 3/i);
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await screen.findByText(/step 2 of 3/i);
    // Set first item to major damage
    const conditionSelects = screen.getAllByDisplayValue('Good');
    fireEvent.change(conditionSelects[0], { target: { value: 'major' } });
    fireEvent.click(screen.getByRole('button', { name: /next/i }));
    await screen.findByText(/step 3 of 3/i);
    expect(screen.getByText('-$80.00')).toBeInTheDocument();
    expect(screen.getByText(/proposed deductions total: \$80\.00/i)).toBeInTheDocument();
  });

  test('Submit Inspection shows error when no files attached', async () => {
    await openStep3();
    fireEvent.click(screen.getByRole('button', { name: /submit inspection/i }));
    await waitFor(() =>
      expect(screen.getByText(/evidence file required/i)).toBeInTheDocument()
    );
  });

  test('Back button from step 3 returns to step 2', async () => {
    await openStep3();
    fireEvent.click(screen.getByRole('button', { name: /back/i }));
    await waitFor(() =>
      expect(screen.getByText(/step 2 of 3/i)).toBeInTheDocument()
    );
  });

  // ── Settlement flow ──────────────────────────────────────────────────────────

  test('Settle Trip button appears when inspections exist', async () => {
    listInspections.mockResolvedValue(sampleInspections);
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /settle trip/i })).toBeInTheDocument()
    );
  });

  test('Settle Trip button does NOT appear when no inspections', async () => {
    listInspections.mockResolvedValue([]);
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await waitFor(() =>
      expect(screen.queryByRole('button', { name: /settle trip/i })).not.toBeInTheDocument()
    );
  });

  test('clicking Settle Trip shows Settlement Ledger Statement card', async () => {
    listInspections.mockResolvedValue(sampleInspections);
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await screen.findByRole('button', { name: /settle trip/i });
    fireEvent.click(screen.getByRole('button', { name: /settle trip/i }));
    await waitFor(() =>
      expect(screen.getByText('Settlement Ledger Statement')).toBeInTheDocument()
    );
  });

  test('settlement ledger shows Total Charges and Adjustments rows', async () => {
    listInspections.mockResolvedValue(sampleInspections);
    closeSettlement.mockResolvedValue({
      booking: { id: 'b1' },
      ledger: [
        { id: 'e1', type: 'trip_charge', amount: 120, description: 'Trip charge' },
        { id: 'e2', type: 'deposit_refund', amount: -20, description: 'Deposit refund' },
      ],
    });
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await screen.findByRole('button', { name: /settle trip/i });
    fireEvent.click(screen.getByRole('button', { name: /settle trip/i }));
    await waitFor(() => expect(screen.getByText('Total Charges')).toBeInTheDocument());
    expect(screen.getAllByText('$120.00').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('Trip charge')).toBeInTheDocument();
  });

  test('Mark ledger reviewed button acknowledges settlement', async () => {
    listInspections.mockResolvedValue(sampleInspections);
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await screen.findByRole('button', { name: /settle trip/i });
    fireEvent.click(screen.getByRole('button', { name: /settle trip/i }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /mark ledger reviewed/i })).toBeInTheDocument()
    );
    fireEvent.click(screen.getByRole('button', { name: /mark ledger reviewed/i }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /settlement statement acknowledged/i })).toBeDisabled()
    );
  });

  test('settlement shows deposit label', async () => {
    listInspections.mockResolvedValue(sampleInspections);
    closeSettlement.mockResolvedValue({
      booking: { id: 'b1' },
      ledger: [{ id: 'e1', type: 'deposit_refund', amount: 45, description: 'Deposit refund' }],
    });
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await screen.findByRole('button', { name: /settle trip/i });
    fireEvent.click(screen.getByRole('button', { name: /settle trip/i }));
    await waitFor(() => expect(screen.getByText(/deposit/i)).toBeInTheDocument());
  });

  // ── Completed inspection detail display ──────────────────────────────────────

  test('completed inspection shows revision ID and stage', async () => {
    listInspections.mockResolvedValue(sampleInspections);
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await waitFor(() =>
      expect(screen.getByText(/rev1/)).toBeInTheDocument()
    );
    expect(screen.getByText(/handover/)).toBeInTheDocument();
  });

  test('completed inspection shows evidence IDs for items', async () => {
    listInspections.mockResolvedValue(sampleInspections);
    renderPage();
    await screen.findByText('b1 — active');
    fireEvent.click(screen.getByRole('button', { name: /view inspections/i }));
    await waitFor(() =>
      expect(screen.getByText('Exterior bodywork')).toBeInTheDocument()
    );
    expect(screen.getByText(/att1/)).toBeInTheDocument();
  });

  // ── Offline queue ────────────────────────────────────────────────────────────

  test('shows offline queue badge when inspection items queued', () => {
    getQueue.mockReturnValue([
      { type: 'inspection', idempotencyKey: 'k1', payload: { bookingId: 'b1', stage: 'handover', items: [], notes: '' } },
    ]);
    renderPage();
    expect(screen.getByText(/offline queue: 1/i)).toBeInTheDocument();
  });

  test('shows Sync Offline Inspections button when queue has inspection items', () => {
    getQueue.mockReturnValue([
      { type: 'inspection', idempotencyKey: 'k1', payload: { bookingId: 'b1', stage: 'handover', items: [], notes: '' } },
    ]);
    renderPage();
    expect(screen.getByRole('button', { name: /sync offline inspections/i })).toBeInTheDocument();
  });

  // ── Sync offline inspections ─────────────────────────────────────────────────

  test('clicking Sync Offline Inspections clears queue after syncing', async () => {
    getQueue.mockReturnValue([
      {
        type: 'inspection',
        idempotencyKey: 'k1',
        payload: { bookingId: 'b1', stage: 'handover', items: [], notes: '' },
      },
    ]);
    const { attachmentInit: ai, submitInspection: si } = await import('../../../src/lib/api');
    ai.mockResolvedValue({ uploadId: 'up1', deduplicated: false });
    si.mockResolvedValue({ ok: true });

    renderPage();
    await screen.findByRole('button', { name: /sync offline inspections/i });
    fireEvent.click(screen.getByRole('button', { name: /sync offline inspections/i }));
    await waitFor(() =>
      expect(screen.getByText(/synced/i)).toBeInTheDocument()
    , { timeout: 5000 });
  });

  test('sync offline inspections processes items with _offlineFile', async () => {
    getQueue.mockReturnValue([
      {
        type: 'inspection',
        idempotencyKey: 'k2',
        payload: {
          bookingId: 'b1',
          stage: 'handover',
          items: [
            {
              name: 'Exterior bodywork',
              condition: 'good',
              _offlineFile: { base64: 'dGVzdA==', type: 'image/jpeg', name: 'ext.jpg', checksum: 'abc' },
            },
          ],
          notes: 'offline note',
        },
      },
    ]);

    renderPage();
    await screen.findByRole('button', { name: /sync offline inspections/i });
    fireEvent.click(screen.getByRole('button', { name: /sync offline inspections/i }));
    await waitFor(() =>
      expect(screen.getByText(/synced/i)).toBeInTheDocument()
    , { timeout: 5000 });
    expect(attachmentInit).toHaveBeenCalled();
  });

  test(
    'Submit Inspection with files attached calls submitInspection',
    async () => {
      // Use a real File object for jsdom compatibility
      const fileContent = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9]);
      const realFile = new File([fileContent], 'evidence.jpg', { type: 'image/jpeg' });

      renderPage();
      await screen.findByText('b1 — active');
      fireEvent.click(screen.getByRole('button', { name: /start inspection/i }));
      await screen.findByText(/step 1 of 3/i);
      fireEvent.click(screen.getByRole('button', { name: /next/i }));
      await screen.findByText(/step 2 of 3/i);

      // Attach a real File to each of the 4 checklist inputs
      const fileInputs = document.querySelectorAll('input[type="file"]');
      fileInputs.forEach((input) => {
        Object.defineProperty(input, 'files', { value: [realFile], configurable: true });
        fireEvent.change(input);
      });

      fireEvent.click(screen.getByRole('button', { name: /next/i }));
      await screen.findByText(/step 3 of 3/i);
      fireEvent.click(screen.getByRole('button', { name: /submit inspection/i }));

      await waitFor(() =>
        expect(submitInspection).toHaveBeenCalled()
      , { timeout: 15000 });
    },
    20000 // Increase timeout for this test only
  );
});
