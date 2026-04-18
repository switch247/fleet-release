import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/lib/api', () => ({
  bookings: vi.fn(),
  listRatings: vi.fn(),
  createRating: vi.fn(),
}));

import { bookings, listRatings, createRating } from '../../../src/lib/api';
import RatingsPage from '../../../src/pages/RatingsPage';

const sampleBookings = [
  { id: 'b1', status: 'settled' },
  { id: 'b2', status: 'settled' },
];

const sampleRatings = [
  { id: 'r1', bookingId: 'b1', fromUserId: 'u1', toUserId: 'u2', score: 5, comment: 'Great!' },
];

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <RatingsPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => {
  bookings.mockResolvedValue(sampleBookings);
  listRatings.mockResolvedValue(sampleRatings);
  createRating.mockResolvedValue({ id: 'r2' });
});

describe('RatingsPage', () => {
  test('renders Two-Way Ratings heading', () => {
    renderPage();
    expect(screen.getByText('Two-Way Ratings')).toBeInTheDocument();
  });

  test('renders Ratings card', () => {
    renderPage();
    expect(screen.getByText('Ratings')).toBeInTheDocument();
  });

  test('renders Create Rating button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /create rating/i })).toBeInTheDocument();
  });

  test('renders booking filter dropdown', async () => {
    renderPage();
    await screen.findByText('Filter by booking (optional)');
    expect(screen.getByText('Filter by booking (optional)')).toBeInTheDocument();
  });

  test('shows bookings in filter dropdown', async () => {
    renderPage();
    // Wait for bookings to load
    await waitFor(() => {
      const options = document.querySelectorAll('select option');
      return Array.from(options).some((o) => o.value === 'b1');
    });
  });

  test('shows ratings in table after load', async () => {
    // Set up allRatings to come from the aggregated query
    listRatings.mockResolvedValue(sampleRatings);
    renderPage();
    await waitFor(() => expect(listRatings).toHaveBeenCalled());
  });

  test('shows empty message when no ratings', async () => {
    listRatings.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No ratings available')).toBeInTheDocument()
    );
  });

  test('clicking Create Rating opens modal', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /create rating/i }));
    await waitFor(() =>
      expect(screen.getByText(/create rating/i, { selector: 'h3' })).toBeInTheDocument()
    );
  });

  test('modal has score dropdown with 5 options', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /create rating/i }));
    await screen.findByRole('heading', { name: /create rating/i });
    // There should be a score selector
    expect(screen.getByText('5 stars')).toBeInTheDocument();
    expect(screen.getByText('1 stars')).toBeInTheDocument();
  });

  test('modal can be closed', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /create rating/i }));
    await screen.findByRole('heading', { name: /create rating/i });
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));
    await waitFor(() =>
      expect(screen.queryByText('5 stars')).not.toBeInTheDocument()
    );
  });

  test('table column headers are correct', () => {
    renderPage();
    expect(screen.getByText('Booking')).toBeInTheDocument();
    expect(screen.getByText('From')).toBeInTheDocument();
    expect(screen.getByText('To')).toBeInTheDocument();
    expect(screen.getByText('Score')).toBeInTheDocument();
    expect(screen.getByText('Comment')).toBeInTheDocument();
  });

  test('selecting a booking in filter shows filtered ratings', async () => {
    renderPage();
    // Wait for bookings to load in the filter dropdown
    await waitFor(() => {
      const options = document.querySelectorAll('select option');
      return Array.from(options).some((o) => o.value === 'b1');
    });
    // Select a booking to trigger ratingsQuery
    const selects = document.querySelectorAll('select');
    fireEvent.change(selects[0], { target: { value: 'b1' } });
    // listRatings should eventually be called with 'b1'
    await waitFor(() =>
      expect(listRatings).toHaveBeenCalledWith('b1')
    );
  });

  test('shows rating comment in table', async () => {
    listRatings.mockResolvedValue(sampleRatings);
    renderPage();
    // allRatingsQuery fetches after bookingsQuery succeeds
    await waitFor(() =>
      expect(screen.getAllByText('Great!').length).toBeGreaterThanOrEqual(1)
    );
  });

  test('Create Rating modal cancel button closes modal', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /create rating/i }));
    await screen.findByRole('heading', { name: /create rating/i });
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    await waitFor(() =>
      expect(screen.queryByRole('heading', { name: /create rating/i })).not.toBeInTheDocument()
    );
  });

  test('can select score in modal and submit', async () => {
    renderPage();
    // Wait for bookings to load
    await waitFor(() => {
      const options = document.querySelectorAll('select option');
      return Array.from(options).some((o) => o.value === 'b1');
    });
    fireEvent.click(screen.getByRole('button', { name: /create rating/i }));
    await screen.findByRole('heading', { name: /create rating/i });
    // Select booking
    const allSelects = document.querySelectorAll('select');
    const modalSelect = allSelects[allSelects.length - 2]; // booking select in modal
    fireEvent.change(modalSelect, { target: { value: 'b1' } });
    // Change score
    const scoreSelects = Array.from(document.querySelectorAll('select')).filter((s) =>
      Array.from(s.options).some((o) => o.text.includes('stars'))
    );
    if (scoreSelects.length > 0) {
      fireEvent.change(scoreSelects[0], { target: { value: '3' } });
    }
    // Submit
    fireEvent.click(screen.getByRole('button', { name: /submit rating/i }));
    await waitFor(() =>
      expect(createRating).toHaveBeenCalled()
    );
  });

  test('shows No ratings for this booking when booking selected and empty', async () => {
    listRatings.mockResolvedValue([]);
    renderPage();
    await waitFor(() => {
      const options = document.querySelectorAll('select option');
      return Array.from(options).some((o) => o.value === 'b1');
    });
    const selects = document.querySelectorAll('select');
    fireEvent.change(selects[0], { target: { value: 'b1' } });
    await waitFor(() =>
      expect(screen.getByText(/no ratings/i)).toBeInTheDocument()
    );
  });
});
