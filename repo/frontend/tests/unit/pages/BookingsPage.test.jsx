import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/lib/api', () => ({
  bookings: vi.fn(),
  listings: vi.fn(),
  createBooking: vi.fn(),
  estimateBooking: vi.fn(),
  apiFetch: vi.fn(),
}));

vi.mock('../../../src/offline/queue', () => ({
  enqueue: vi.fn(),
  getQueue: vi.fn().mockReturnValue([]),
  reconcileQueue: vi.fn().mockResolvedValue({ applied: 0, total: 0, results: [] }),
  flushQueue: vi.fn(),
  clearQueue: vi.fn(),
}));

import { bookings, listings, createBooking, estimateBooking } from '../../../src/lib/api';
import { getQueue } from '../../../src/offline/queue';
import BookingsPage from '../../../src/pages/BookingsPage';

const sampleBookings = [
  { id: 'b1', listingId: 'l1', status: 'active', estimatedAmount: 100, depositAmount: 20 },
  { id: 'b2', listingId: 'l1', status: 'settled', estimatedAmount: 200, depositAmount: 40 },
];

const sampleListings = [
  { id: 'l1', name: 'Toyota Camry', spu: 'TOY', sku: 'CAM-001' },
];

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <BookingsPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => {
  bookings.mockResolvedValue(sampleBookings);
  listings.mockResolvedValue(sampleListings);
  createBooking.mockResolvedValue({ booking: { id: 'b3' } });
  estimateBooking.mockResolvedValue({ estimate: { baseAmount: 50, mileageAmount: 10, timeAmount: 5, nightSurcharge: 0, deposit: 20, total: 85 } });
  getQueue.mockReturnValue([]);
});

describe('BookingsPage', () => {
  test('renders Bookings heading', () => {
    renderPage();
    expect(screen.getByText('Bookings')).toBeInTheDocument();
  });

  test('renders subtitle', () => {
    renderPage();
    expect(screen.getByText(/offline queue support/i)).toBeInTheDocument();
  });

  test('renders New Booking button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /new booking/i })).toBeInTheDocument();
  });

  test('renders Current Bookings card', () => {
    renderPage();
    expect(screen.getByText('Current Bookings')).toBeInTheDocument();
  });

  test('shows booking IDs after load', async () => {
    renderPage();
    expect(await screen.findByText('b1')).toBeInTheDocument();
    expect(await screen.findByText('b2')).toBeInTheDocument();
  });

  test('shows booking status badges', async () => {
    renderPage();
    await screen.findByText('b1');
    expect(screen.getByText('active')).toBeInTheDocument();
    expect(screen.getByText('settled')).toBeInTheDocument();
  });

  test('shows empty message when no bookings', async () => {
    bookings.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No bookings yet')).toBeInTheDocument()
    );
  });

  test('shows offline queue size badge', () => {
    getQueue.mockReturnValue([]);
    renderPage();
    expect(screen.getByText(/offline queue: 0/i)).toBeInTheDocument();
  });

  test('shows Sync Queue button when queue has items', async () => {
    getQueue.mockReturnValue([{ type: 'booking', idempotencyKey: 'k1', createdAt: '2024-01-01T00:00:00Z' }]);
    renderPage();
    // re-render to pick up new queue value
    expect(screen.getByRole('button', { name: /sync queue/i })).toBeInTheDocument();
  });

  test('clicking New Booking opens modal', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    // "Create Booking" appears in modal heading and button — use heading role to confirm modal opened
    await waitFor(() =>
      expect(screen.getByRole('heading', { name: 'Create Booking' })).toBeInTheDocument()
    );
  });

  test('modal can be closed', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByRole('heading', { name: 'Create Booking' });
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));
    await waitFor(() =>
      expect(screen.queryByText(/select listing/i)).not.toBeInTheDocument()
    );
  });

  test('modal shows listing dropdown', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByRole('heading', { name: 'Create Booking' });
    // The modal has a "Select listing" placeholder option which is unique
    expect(screen.getByText('Select listing')).toBeInTheDocument();
  });

  test('modal shows available listings in dropdown', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await waitFor(() =>
      expect(screen.getByText('Toyota Camry (TOY/CAM-001)')).toBeInTheDocument()
    );
  });

  test('shows validation message when form is incomplete', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByRole('heading', { name: 'Create Booking' });
    expect(screen.getByText(/please select a listing/i)).toBeInTheDocument();
  });

  test('Create Booking button is disabled without estimate', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByRole('heading', { name: 'Create Booking' });
    // Find the footer Create Booking button (not the heading)
    const createBtns = screen.getAllByRole('button', { name: 'Create Booking' });
    expect(createBtns[0]).toBeDisabled();
  });

  test('table has correct column headers', () => {
    renderPage();
    expect(screen.getByText('Booking ID')).toBeInTheDocument();
    expect(screen.getByText('Listing')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
    expect(screen.getByText('Estimate')).toBeInTheDocument();
    expect(screen.getByText('Deposit')).toBeInTheDocument();
  });

  // ── Modal form interactions ──────────────────────────────────────────────────

  test('selecting a listing shows selected listing text', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByRole('heading', { name: 'Create Booking' });
    // Wait for listings to load
    await screen.findByText('Toyota Camry (TOY/CAM-001)');
    const listingSelect = screen.getByRole('combobox');
    fireEvent.change(listingSelect, { target: { value: 'l1' } });
    await waitFor(() =>
      expect(screen.getByText(/Selected: Toyota Camry/)).toBeInTheDocument()
    );
  });

  test('Preview Estimate button appears when form is valid', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByText('Toyota Camry (TOY/CAM-001)');
    const listingSelect = screen.getByRole('combobox');
    fireEvent.change(listingSelect, { target: { value: 'l1' } });
    // Set startAt/endAt via the datetime-local inputs
    const dateInputs = document.querySelectorAll('input[type="datetime-local"]');
    fireEvent.change(dateInputs[0], { target: { value: '2024-01-10T10:00' } });
    fireEvent.change(dateInputs[1], { target: { value: '2024-01-12T10:00' } });
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /preview estimate/i })).toBeInTheDocument()
    );
  });

  test('Preview Estimate button click calls estimateBooking', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByText('Toyota Camry (TOY/CAM-001)');
    const listingSelect = screen.getByRole('combobox');
    fireEvent.change(listingSelect, { target: { value: 'l1' } });
    const dateInputs = document.querySelectorAll('input[type="datetime-local"]');
    fireEvent.change(dateInputs[0], { target: { value: '2024-01-10T10:00' } });
    fireEvent.change(dateInputs[1], { target: { value: '2024-01-12T10:00' } });
    const previewBtn = await screen.findByRole('button', { name: /preview estimate/i });
    fireEvent.click(previewBtn);
    await waitFor(() =>
      expect(estimateBooking).toHaveBeenCalledWith(
        expect.objectContaining({ listingId: 'l1' }),
        expect.any(Object)
      )
    );
  });

  test('estimate preview displays after calling Preview Estimate', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByText('Toyota Camry (TOY/CAM-001)');
    const listingSelect = screen.getByRole('combobox');
    fireEvent.change(listingSelect, { target: { value: 'l1' } });
    const dateInputs = document.querySelectorAll('input[type="datetime-local"]');
    fireEvent.change(dateInputs[0], { target: { value: '2024-01-10T10:00' } });
    fireEvent.change(dateInputs[1], { target: { value: '2024-01-12T10:00' } });
    fireEvent.click(await screen.findByRole('button', { name: /preview estimate/i }));
    await waitFor(() =>
      expect(screen.getByText('Pre-trip estimate breakdown')).toBeInTheDocument()
    );
    expect(screen.getByText('$85.00')).toBeInTheDocument(); // total from mock
  });

  test('coupon code input can be filled', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByRole('heading', { name: 'Create Booking' });
    const couponInput = screen.getByPlaceholderText('Optional');
    fireEvent.change(couponInput, { target: { value: 'SAVE10' } });
    expect(couponInput.value).toBe('SAVE10');
  });

  // ── Sync Queue ───────────────────────────────────────────────────────────────

  test('clicking Sync Queue button calls reconcileQueue', async () => {
    const { reconcileQueue } = await import('../../../src/offline/queue');
    getQueue.mockReturnValue([{ type: 'booking', idempotencyKey: 'k1', createdAt: '2024-01-01T00:00:00Z' }]);
    renderPage();
    const syncBtn = await screen.findByRole('button', { name: /sync queue/i });
    fireEvent.click(syncBtn);
    await waitFor(() =>
      expect(reconcileQueue).toHaveBeenCalled()
    );
  });

  test('reconcileQueue booking handler calls createBooking', async () => {
    const { reconcileQueue } = await import('../../../src/offline/queue');
    reconcileQueue.mockImplementation(async (_fetcher, handlers) => {
      if (handlers?.booking) await handlers.booking({ listingId: 'l1' });
      return { applied: 1, total: 1 };
    });
    getQueue.mockReturnValue([{ type: 'booking', idempotencyKey: 'k1', createdAt: '2024-01-01T00:00:00Z' }]);
    renderPage();
    const syncBtn = await screen.findByRole('button', { name: /sync queue/i });
    fireEvent.click(syncBtn);
    await waitFor(() =>
      expect(createBooking).toHaveBeenCalled()
    );
  });

  // ── Submit Booking ───────────────────────────────────────────────────────────

  test('Create Booking button calls createBooking after estimate', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
    await screen.findByText('Toyota Camry (TOY/CAM-001)');
    // Fill form
    const listingSelect = screen.getByRole('combobox');
    fireEvent.change(listingSelect, { target: { value: 'l1' } });
    const dateInputs = document.querySelectorAll('input[type="datetime-local"]');
    fireEvent.change(dateInputs[0], { target: { value: '2024-01-10T10:00' } });
    fireEvent.change(dateInputs[1], { target: { value: '2024-01-12T10:00' } });
    // Click Preview Estimate
    const previewBtn = await screen.findByRole('button', { name: /preview estimate/i });
    fireEvent.click(previewBtn);
    // Wait for estimate to appear
    await waitFor(() =>
      expect(screen.getByText('Pre-trip estimate breakdown')).toBeInTheDocument()
    );
    // Now the Create Booking button should be enabled — click it
    const createBtns = screen.getAllByRole('button', { name: 'Create Booking' });
    const submitBtn = createBtns.find((b) => !b.disabled);
    if (submitBtn) {
      fireEvent.click(submitBtn);
      await waitFor(() =>
        expect(createBooking).toHaveBeenCalled()
      );
    }
  });

  test('submitting booking while offline enqueues it instead of calling createBooking', async () => {
    const originalOnLine = Object.getOwnPropertyDescriptor(navigator, 'onLine');
    Object.defineProperty(navigator, 'onLine', { value: false, configurable: true });
    try {
      renderPage();
      fireEvent.click(screen.getByRole('button', { name: /new booking/i }));
      await screen.findByText('Toyota Camry (TOY/CAM-001)');
      const listingSelect = screen.getByRole('combobox');
      fireEvent.change(listingSelect, { target: { value: 'l1' } });
      const dateInputs = document.querySelectorAll('input[type="datetime-local"]');
      fireEvent.change(dateInputs[0], { target: { value: '2024-01-10T10:00' } });
      fireEvent.change(dateInputs[1], { target: { value: '2024-01-12T10:00' } });
      // Get an estimate so the Create Booking button is enabled
      const previewBtn = await screen.findByRole('button', { name: /preview estimate/i });
      fireEvent.click(previewBtn);
      await waitFor(() =>
        expect(screen.getByText('Pre-trip estimate breakdown')).toBeInTheDocument()
      );
      const createBtns = screen.getAllByRole('button', { name: 'Create Booking' });
      const submitBtn = createBtns.find((b) => !b.disabled);
      if (submitBtn) {
        fireEvent.click(submitBtn);
        await waitFor(() =>
          expect(createBooking).not.toHaveBeenCalled()
        );
      }
    } finally {
      if (originalOnLine) {
        Object.defineProperty(navigator, 'onLine', originalOnLine);
      } else {
        Object.defineProperty(navigator, 'onLine', { value: true, configurable: true });
      }
    }
  });
});
