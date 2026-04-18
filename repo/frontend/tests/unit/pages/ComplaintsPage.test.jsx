import React, { act } from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/auth/AuthProvider', () => ({
  useAuth: vi.fn(),
}));

vi.mock('../../../src/lib/api', () => ({
  bookings: vi.fn(),
  complaints: vi.fn(),
  createComplaint: vi.fn(),
  arbitrateComplaint: vi.fn(),
  exportDisputePDF: vi.fn(),
}));

vi.mock('../../../src/offline/queue', () => ({
  enqueue: vi.fn(),
  getQueue: vi.fn().mockReturnValue([]),
}));

import { useAuth } from '../../../src/auth/AuthProvider';
import { bookings, complaints, createComplaint, arbitrateComplaint, exportDisputePDF } from '../../../src/lib/api';
import ComplaintsPage from '../../../src/pages/ComplaintsPage';

// Stub URL.createObjectURL
global.URL.createObjectURL = vi.fn().mockReturnValue('blob:mock');
global.URL.revokeObjectURL = vi.fn();

const customerUser = { id: 'u1', username: 'alice', roles: ['customer'] };
const csaUser = { id: 'u2', username: 'agent', roles: ['csa'] };

const sampleBookings = [{ id: 'b1', status: 'active' }];
const sampleComplaints = [
  { id: 'c1', bookingId: 'b1', openedBy: 'u1', status: 'open', outcome: 'vehicle scratch' },
  { id: 'c2', bookingId: 'b1', openedBy: 'u1', status: 'resolved', outcome: 'vehicle damage confirmed' },
];

function renderPage(user = customerUser) {
  useAuth.mockReturnValue({ user });
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <ComplaintsPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => {
  bookings.mockResolvedValue(sampleBookings);
  complaints.mockResolvedValue(sampleComplaints);
  createComplaint.mockResolvedValue({ id: 'c3' });
  arbitrateComplaint.mockResolvedValue({ id: 'c1', status: 'resolved' });
  exportDisputePDF.mockResolvedValue(new Blob(['%PDF'], { type: 'application/pdf' }));
});

describe('ComplaintsPage', () => {
  test('renders Open Complaint form card', () => {
    renderPage();
    expect(screen.getByText('Open Complaint')).toBeInTheDocument();
  });

  test('renders Complaints table card', () => {
    renderPage();
    expect(screen.getByText('Complaints')).toBeInTheDocument();
  });

  test('shows booking selector', async () => {
    renderPage();
    await screen.findByText('Select booking');
    expect(screen.getByText('Select booking')).toBeInTheDocument();
  });

  test('shows complaint details input', () => {
    renderPage();
    expect(screen.getByPlaceholderText('Complaint details')).toBeInTheDocument();
  });

  test('shows Submit button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /submit/i })).toBeInTheDocument();
  });

  test('shows complaints in table', async () => {
    renderPage();
    expect(await screen.findByText('c1')).toBeInTheDocument();
    expect(await screen.findByText('c2')).toBeInTheDocument();
  });

  test('shows status badges for complaints', async () => {
    renderPage();
    await screen.findByText('c1');
    expect(screen.getByText('open')).toBeInTheDocument();
    expect(screen.getByText('resolved')).toBeInTheDocument();
  });

  test('shows Export PDF buttons for each complaint', async () => {
    renderPage();
    await screen.findByText('c1');
    const pdfBtns = screen.getAllByRole('button', { name: /export pdf/i });
    expect(pdfBtns).toHaveLength(2);
  });

  test('does not show Arbitrate button for customer role', async () => {
    renderPage(customerUser);
    await screen.findByText('c1');
    expect(screen.queryByRole('button', { name: /arbitrate/i })).not.toBeInTheDocument();
  });

  test('shows Arbitrate button for csa role', async () => {
    renderPage(csaUser);
    await screen.findByText('c1');
    expect(screen.getAllByRole('button', { name: /arbitrate/i }).length).toBeGreaterThan(0);
  });

  test('calls createComplaint on submit', async () => {
    renderPage();
    // Wait for booking option to appear in the select (bookings load async)
    await screen.findByRole('option', { name: 'b1' });
    const select = screen.getAllByRole('combobox')[0];
    fireEvent.change(select, { target: { value: 'b1' } });
    fireEvent.change(screen.getByPlaceholderText('Complaint details'), { target: { value: 'damage found' } });
    fireEvent.click(screen.getByRole('button', { name: /submit/i }));
    // React Query v5 passes a second context arg to mutationFn
    await waitFor(() =>
      expect(createComplaint).toHaveBeenCalledWith(
        expect.objectContaining({ bookingId: 'b1', outcome: 'damage found' }),
        expect.any(Object)
      )
    );
  });

  test('arbitrate modal opens for csa user', async () => {
    renderPage(csaUser);
    await screen.findByText('c1');
    fireEvent.click(screen.getAllByRole('button', { name: /arbitrate/i })[0]);
    await waitFor(() =>
      expect(screen.getByText(/arbitrate/i, { selector: 'h3' })).toBeInTheDocument()
    );
  });

  test('empty complaints message shown when no complaints', async () => {
    complaints.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No complaints')).toBeInTheDocument()
    );
  });

  test('table has correct column headers', () => {
    renderPage();
    expect(screen.getByText('Complaint ID')).toBeInTheDocument();
    expect(screen.getByText('Booking')).toBeInTheDocument();
    expect(screen.getByText('Opened By')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
  });

  test('clicking Apply in arbitrate modal calls arbitrateComplaint', async () => {
    renderPage(csaUser);
    await screen.findByText('c1');
    // Open arbitrate modal for first complaint
    fireEvent.click(screen.getAllByRole('button', { name: /arbitrate/i })[0]);
    await waitFor(() =>
      expect(screen.getByText(/arbitrate/i, { selector: 'h3' })).toBeInTheDocument()
    );
    // Click Apply
    fireEvent.click(screen.getByRole('button', { name: /apply/i }));
    await waitFor(() =>
      expect(arbitrateComplaint).toHaveBeenCalledWith('c1', expect.any(Object))
    );
  });

  test('Details column is shown in table', async () => {
    renderPage();
    await screen.findByText('c1');
    expect(screen.getByText('Details')).toBeInTheDocument();
  });

  test('clicking Export PDF button calls exportDisputePDF', async () => {
    renderPage();
    await screen.findByText('c1');
    const pdfBtns = screen.getAllByRole('button', { name: /export pdf/i });
    await act(async () => {
      fireEvent.click(pdfBtns[0]);
    });
    await waitFor(() =>
      expect(exportDisputePDF).toHaveBeenCalled()
    );
  });

  test('modal decision status and outcome fields can be changed', async () => {
    renderPage(csaUser);
    await screen.findByText('c1');
    fireEvent.click(screen.getAllByRole('button', { name: /arbitrate/i })[0]);
    await waitFor(() =>
      expect(screen.getByText(/arbitrate/i, { selector: 'h3' })).toBeInTheDocument()
    );
    // Change status select
    const statusSelect = screen.getByDisplayValue('Resolved');
    fireEvent.change(statusSelect, { target: { value: 'dismissed' } });
    expect(screen.getByDisplayValue('Dismissed')).toBeInTheDocument();
    // Change outcome input
    fireEvent.change(screen.getByPlaceholderText('Outcome'), { target: { value: 'Claim denied' } });
    expect(screen.getByPlaceholderText('Outcome').value).toBe('Claim denied');
  });

  test('Submit enqueues when offline', async () => {
    const originalOnLine = Object.getOwnPropertyDescriptor(navigator, 'onLine');
    Object.defineProperty(navigator, 'onLine', { value: false, configurable: true });
    try {
      renderPage();
      await screen.findByRole('option', { name: 'b1' });
      const select = screen.getAllByRole('combobox')[0];
      fireEvent.change(select, { target: { value: 'b1' } });
      fireEvent.change(screen.getByPlaceholderText('Complaint details'), { target: { value: 'scratch' } });
      fireEvent.click(screen.getByRole('button', { name: /submit/i }));
      // When offline, enqueue is called instead of createComplaint
      await waitFor(() =>
        expect(createComplaint).not.toHaveBeenCalled()
      );
    } finally {
      if (originalOnLine) {
        Object.defineProperty(navigator, 'onLine', originalOnLine);
      } else {
        Object.defineProperty(navigator, 'onLine', { value: true, configurable: true });
      }
    }
  });
});
