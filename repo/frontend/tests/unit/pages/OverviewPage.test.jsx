import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Mock recharts to avoid canvas/ResizeObserver issues in jsdom
vi.mock('recharts', () => ({
  BarChart: ({ children }) => React.createElement('div', { 'data-testid': 'bar-chart' }, children),
  Bar: () => null,
  ResponsiveContainer: ({ children }) => React.createElement('div', { 'data-testid': 'responsive-container' }, children),
  XAxis: () => null,
  YAxis: () => null,
  Tooltip: () => null,
}));

vi.mock('../../../src/lib/api', () => ({
  statsSummary: vi.fn(),
  bookings: vi.fn(),
}));

import { statsSummary, bookings } from '../../../src/lib/api';
import OverviewPage from '../../../src/pages/OverviewPage';

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
}

function renderPage() {
  return render(
    <QueryClientProvider client={createClient()}>
      <OverviewPage />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  statsSummary.mockResolvedValue({
    activeBookings: 3,
    settledTrips: 10,
    inspectionsDue: 2,
    heldDeposits: 450.75,
  });
  bookings.mockResolvedValue([
    { id: 'b1', status: 'active' },
    { id: 'b2', status: 'settled' },
    { id: 'b3', status: 'active' },
  ]);
});

describe('OverviewPage', () => {
  test('renders heading', () => {
    renderPage();
    expect(screen.getByText('Operational Overview')).toBeInTheDocument();
  });

  test('shows subtitle referencing stats endpoint', () => {
    renderPage();
    expect(screen.getByText(/stats\/summary/i)).toBeInTheDocument();
  });

  test('renders Active Bookings stat card', async () => {
    renderPage();
    expect(await screen.findByText('Active Bookings')).toBeInTheDocument();
  });

  test('renders Settled Trips stat card', async () => {
    renderPage();
    expect(await screen.findByText('Settled Trips')).toBeInTheDocument();
  });

  test('renders Inspections Due stat card', async () => {
    renderPage();
    expect(await screen.findByText('Inspections Due')).toBeInTheDocument();
  });

  test('renders Held Deposits stat card', async () => {
    renderPage();
    expect(await screen.findByText('Held Deposits')).toBeInTheDocument();
  });

  test('displays fetched active bookings count', async () => {
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('3')).toBeInTheDocument()
    );
  });

  test('displays fetched settled trips count', async () => {
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('10')).toBeInTheDocument()
    );
  });

  test('displays held deposits formatted as currency', async () => {
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('$450.75')).toBeInTheDocument()
    );
  });

  test('renders Booking Status Distribution card', () => {
    renderPage();
    expect(screen.getByText('Booking Status Distribution')).toBeInTheDocument();
  });

  test('renders the bar chart container', () => {
    renderPage();
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument();
  });

  test('shows dash placeholder before data loads', () => {
    statsSummary.mockImplementation(() => new Promise(() => {}));
    renderPage();
    // While loading, values show '-'
    expect(screen.getAllByText('-').length).toBeGreaterThanOrEqual(1);
  });
});
