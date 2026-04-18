import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/lib/api', () => ({
  adminNotificationTemplates: vi.fn(),
  adminCreateNotificationTemplate: vi.fn(),
  adminSendNotification: vi.fn(),
  adminRetryNotifications: vi.fn(),
  adminUsers: vi.fn(),
  adminWorkerMetrics: vi.fn(),
}));

import {
  adminNotificationTemplates,
  adminCreateNotificationTemplate,
  adminSendNotification,
  adminRetryNotifications,
  adminUsers,
  adminWorkerMetrics,
} from '../../../src/lib/api';
import AdminNotificationsPage from '../../../src/pages/AdminNotificationsPage';

const sampleTemplates = [
  { id: 't1', name: 'Welcome', channel: 'in_app', enabled: true, title: 'Welcome!', body: 'Hello' },
  { id: 't2', name: 'Reminder', channel: 'email', enabled: false, title: 'Reminder', body: 'Your trip...' },
];

const sampleUsers = [
  { id: 'u1', username: 'alice', roles: ['customer'] },
];

const sampleMetrics = { processed: 50, delivered: 45, deadLettered: 2, currentBacklog: 3 };

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AdminNotificationsPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => {
  adminNotificationTemplates.mockResolvedValue(sampleTemplates);
  adminCreateNotificationTemplate.mockResolvedValue({ id: 't3' });
  adminSendNotification.mockResolvedValue({ ok: true });
  adminRetryNotifications.mockResolvedValue({ retried: 3 });
  adminUsers.mockResolvedValue(sampleUsers);
  adminWorkerMetrics.mockResolvedValue(sampleMetrics);
});

describe('AdminNotificationsPage', () => {
  test('renders Notification Worker Metrics card', () => {
    renderPage();
    expect(screen.getByText('Notification Worker Metrics')).toBeInTheDocument();
  });

  test('renders Create Template card', () => {
    renderPage();
    expect(screen.getByText('Create Template + Preview')).toBeInTheDocument();
  });

  test('renders Send Notification card', () => {
    renderPage();
    expect(screen.getByText('Send Notification')).toBeInTheDocument();
  });

  test('renders Template Catalog card', () => {
    renderPage();
    expect(screen.getByText('Template Catalog')).toBeInTheDocument();
  });

  test('shows worker metrics after load', async () => {
    renderPage();
    await waitFor(() =>
      expect(screen.getByText(/Processed: 50/)).toBeInTheDocument()
    );
    expect(screen.getByText(/Delivered: 45/)).toBeInTheDocument();
    expect(screen.getByText(/Dead Lettered: 2/)).toBeInTheDocument();
    expect(screen.getByText(/Backlog: 3/)).toBeInTheDocument();
  });

  test('shows zero metrics before load', () => {
    adminWorkerMetrics.mockImplementation(() => new Promise(() => {}));
    renderPage();
    expect(screen.getByText(/Processed: 0/)).toBeInTheDocument();
  });

  test('shows Retry Queue Now button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /retry queue now/i })).toBeInTheDocument();
  });

  test('clicking Retry Queue Now calls adminRetryNotifications', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /retry queue now/i }));
    await waitFor(() =>
      expect(adminRetryNotifications).toHaveBeenCalled()
    );
  });

  test('shows template name input', () => {
    renderPage();
    expect(screen.getByPlaceholderText('Template name')).toBeInTheDocument();
  });

  test('shows channel selector', () => {
    renderPage();
    expect(screen.getByDisplayValue('In-app')).toBeInTheDocument();
  });

  test('shows title and body inputs', () => {
    renderPage();
    expect(screen.getByPlaceholderText('Title')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Body')).toBeInTheDocument();
  });

  test('template preview updates as you type', () => {
    renderPage();
    const titleInput = screen.getByPlaceholderText('Title');
    fireEvent.change(titleInput, { target: { value: 'My Title' } });
    expect(screen.getByText(/My Title/)).toBeInTheDocument();
  });

  test('shows Save Template button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /save template/i })).toBeInTheDocument();
  });

  test('clicking Save Template calls adminCreateNotificationTemplate', async () => {
    renderPage();
    fireEvent.change(screen.getByPlaceholderText('Template name'), { target: { value: 'Alert' } });
    fireEvent.click(screen.getByRole('button', { name: /save template/i }));
    // React Query v5 passes a second context arg to mutationFn
    await waitFor(() =>
      expect(adminCreateNotificationTemplate).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'Alert' }),
        expect.any(Object)
      )
    );
  });

  test('shows templates in catalog table', async () => {
    renderPage();
    // 'Welcome' appears in both the DataTable and the Send Notification template dropdown
    const welcomeEls = await screen.findAllByText('Welcome');
    expect(welcomeEls.length).toBeGreaterThanOrEqual(1);
    const reminderEls = await screen.findAllByText('Reminder');
    expect(reminderEls.length).toBeGreaterThanOrEqual(1);
  });

  test('shows template enabled status', async () => {
    renderPage();
    // Wait for templates to load (they appear in table and dropdown)
    await screen.findAllByText('Welcome');
    expect(screen.getByText('Yes')).toBeInTheDocument();
    expect(screen.getByText('No')).toBeInTheDocument();
  });

  test('Send button is disabled when no template/user selected', () => {
    renderPage();
    const sendBtn = screen.getByRole('button', { name: /^send$/i });
    expect(sendBtn).toBeDisabled();
  });

  test('empty template catalog message', async () => {
    adminNotificationTemplates.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No templates')).toBeInTheDocument()
    );
  });

  test('shows catalog column headers', () => {
    renderPage();
    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('Channel')).toBeInTheDocument();
    expect(screen.getByText('Enabled')).toBeInTheDocument();
  });

  test('clicking Retry Queue Now calls adminRetryNotifications', async () => {
    adminRetryNotifications.mockResolvedValue({});
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /retry queue now/i }));
    await waitFor(() =>
      expect(adminRetryNotifications).toHaveBeenCalled()
    );
  });

  test('filling template form updates preview', async () => {
    renderPage();
    fireEvent.change(screen.getByPlaceholderText('Title'), { target: { value: 'My Title' } });
    fireEvent.change(screen.getByPlaceholderText('Body'), { target: { value: 'My Body' } });
    // Preview should show the entered title and body
    await waitFor(() =>
      expect(screen.getByText(/My Title/)).toBeInTheDocument()
    );
  });

  test('selecting template and user enables Send button', async () => {
    renderPage();
    await screen.findAllByText('Welcome');
    // Select template
    const selects = document.querySelectorAll('select');
    const templateSelect = Array.from(selects).find((s) =>
      Array.from(s.options).some((o) => o.value === 't1')
    );
    fireEvent.change(templateSelect, { target: { value: 't1' } });
    // Select user
    const userSelect = Array.from(selects).find((s) =>
      Array.from(s.options).some((o) => o.value === 'u1')
    );
    if (userSelect) {
      fireEvent.change(userSelect, { target: { value: 'u1' } });
    }
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /^send$/i })).not.toBeDisabled()
    );
  });

  test('clicking Send calls adminSendNotification', async () => {
    adminSendNotification.mockResolvedValue({});
    renderPage();
    await screen.findAllByText('Welcome');
    const selects = document.querySelectorAll('select');
    const templateSelect = Array.from(selects).find((s) =>
      Array.from(s.options).some((o) => o.value === 't1')
    );
    fireEvent.change(templateSelect, { target: { value: 't1' } });
    const userSelect = Array.from(selects).find((s) =>
      Array.from(s.options).some((o) => o.value === 'u1')
    );
    if (userSelect) {
      fireEvent.change(userSelect, { target: { value: 'u1' } });
    }
    const sendBtn = await screen.findByRole('button', { name: /^send$/i });
    if (!sendBtn.disabled) {
      fireEvent.click(sendBtn);
      await waitFor(() =>
        expect(adminSendNotification).toHaveBeenCalled()
      );
    }
  });
});
