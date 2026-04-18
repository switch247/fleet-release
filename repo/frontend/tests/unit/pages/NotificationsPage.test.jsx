import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('../../../src/lib/api', () => ({
  inboxNotifications: vi.fn(),
}));

import { inboxNotifications } from '../../../src/lib/api';
import NotificationsPage from '../../../src/pages/NotificationsPage';

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <NotificationsPage />
    </QueryClientProvider>
  );
}

describe('NotificationsPage', () => {
  test('renders Inbox & Delivery Status heading', () => {
    inboxNotifications.mockResolvedValue([]);
    renderPage();
    expect(screen.getByText('Inbox & Delivery Status')).toBeInTheDocument();
  });

  test('shows empty message when no notifications', async () => {
    inboxNotifications.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No inbox notifications')).toBeInTheDocument()
    );
  });

  test('renders notification rows', async () => {
    inboxNotifications.mockResolvedValue([
      { id: 'n1', title: 'Trip Alert', body: 'Your trip starts soon', attempts: 1, status: 'delivered' },
      { id: 'n2', title: 'Payment', body: 'Receipt ready', attempts: 2, status: 'pending' },
    ]);
    renderPage();
    expect(await screen.findByText('Trip Alert')).toBeInTheDocument();
    expect(await screen.findByText('Payment')).toBeInTheDocument();
  });

  test('renders correct column headers', async () => {
    inboxNotifications.mockResolvedValue([]);
    renderPage();
    expect(screen.getByText('Title')).toBeInTheDocument();
    expect(screen.getByText('Body')).toBeInTheDocument();
    expect(screen.getByText('Attempts')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
  });

  test('renders delivered badge for delivered status', async () => {
    inboxNotifications.mockResolvedValue([
      { id: 'n1', title: 'T', body: 'B', attempts: 1, status: 'delivered' },
    ]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('delivered')).toBeInTheDocument()
    );
  });

  test('renders warning badge for disabled_offline status', async () => {
    inboxNotifications.mockResolvedValue([
      { id: 'n1', title: 'T', body: 'B', attempts: 1, status: 'disabled_offline' },
    ]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('disabled_offline')).toBeInTheDocument()
    );
  });
});
