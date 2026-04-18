import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('../../../src/auth/AuthProvider', () => ({
  useAuth: vi.fn(),
}));

vi.mock('../../../src/lib/api', () => ({
  loginHistory: vi.fn(),
  updateMe: vi.fn(),
}));

import { useAuth } from '../../../src/auth/AuthProvider';
import { loginHistory, updateMe } from '../../../src/lib/api';
import ProfilePage from '../../../src/pages/ProfilePage';

const mockRefreshMe = vi.fn();

function renderPage(user = { username: 'alice', email: 'alice@example.com', roles: ['customer'] }) {
  useAuth.mockReturnValue({ user, refreshMe: mockRefreshMe });
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <ProfilePage />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  loginHistory.mockResolvedValue([]);
  updateMe.mockResolvedValue({});
  mockRefreshMe.mockResolvedValue({});
});

describe('ProfilePage', () => {
  test('renders Profile heading', () => {
    renderPage();
    expect(screen.getByText('Profile')).toBeInTheDocument();
  });

  test('shows current username (disabled)', () => {
    renderPage();
    expect(screen.getByDisplayValue('alice')).toBeInTheDocument();
    expect(screen.getByDisplayValue('alice')).toBeDisabled();
  });

  test('shows current email in editable input', () => {
    renderPage();
    expect(screen.getByDisplayValue('alice@example.com')).toBeInTheDocument();
  });

  test('renders Save Email button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /save email/i })).toBeInTheDocument();
  });

  test('renders Login History card', () => {
    renderPage();
    expect(screen.getByText('Login History (Audit Log)')).toBeInTheDocument();
  });

  test('shows empty history message when no events', async () => {
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No login history available')).toBeInTheDocument()
    );
  });

  test('renders history rows', async () => {
    loginHistory.mockResolvedValue([
      { eventType: 'login_success', ip: '127.0.0.1', createdAt: '2024-01-15T10:00:00Z' },
    ]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('login_success')).toBeInTheDocument()
    );
    expect(screen.getByText('127.0.0.1')).toBeInTheDocument();
  });

  test('allows editing email', () => {
    renderPage();
    const emailInput = screen.getByDisplayValue('alice@example.com');
    fireEvent.change(emailInput, { target: { value: 'new@example.com' } });
    expect(emailInput.value).toBe('new@example.com');
  });

  test('calls updateMe and refreshMe on save', async () => {
    renderPage();
    const emailInput = screen.getByDisplayValue('alice@example.com');
    fireEvent.change(emailInput, { target: { value: 'new@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: /save email/i }));
    await waitFor(() =>
      expect(updateMe).toHaveBeenCalledWith({ email: 'new@example.com' })
    );
    expect(mockRefreshMe).toHaveBeenCalled();
  });

  test('shows success status message after save', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /save email/i }));
    await waitFor(() =>
      expect(screen.getByText('Profile updated.')).toBeInTheDocument()
    );
  });

  test('shows error message when update fails', async () => {
    updateMe.mockRejectedValue(new Error('Email already taken'));
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /save email/i }));
    await waitFor(() =>
      expect(screen.getByText('Email already taken')).toBeInTheDocument()
    );
  });

  test('handles null user gracefully', () => {
    useAuth.mockReturnValue({ user: null, refreshMe: vi.fn() });
    loginHistory.mockResolvedValue([]);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={client}>
        <ProfilePage />
      </QueryClientProvider>
    );
    // Should not crash
    expect(screen.getByText('Profile')).toBeInTheDocument();
  });
});
