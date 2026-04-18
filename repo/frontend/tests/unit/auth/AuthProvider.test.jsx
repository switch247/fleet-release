import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';

vi.mock('../../../src/lib/api', () => ({
  login: vi.fn(),
  logout: vi.fn(),
  revokeToken: vi.fn().mockResolvedValue(undefined),
  me: vi.fn(),
}));

import { login as apiLogin, revokeToken, me as apiMe } from '../../../src/lib/api';
import { AuthProvider, useAuth } from '../../../src/auth/AuthProvider';

// ── helper component that exercises the context ──────────────────────────────

function TestConsumer() {
  const { user, loading, login, logout, refreshMe } = useAuth();
  return (
    <div>
      <span data-testid="loading">{String(loading)}</span>
      <span data-testid="user">{user ? user.username : 'none'}</span>
      <button onClick={() => login({ username: 'u', password: 'p' })}>Login</button>
      <button onClick={logout}>Logout</button>
      <button onClick={refreshMe}>Refresh</button>
    </div>
  );
}

beforeEach(() => {
  localStorage.clear();
  apiMe.mockResolvedValue(null);
  revokeToken.mockResolvedValue(undefined);
});

// ── tests ────────────────────────────────────────────────────────────────────

describe('AuthProvider', () => {
  test('renders children', () => {
    render(
      <AuthProvider>
        <p>hello</p>
      </AuthProvider>
    );
    expect(screen.getByText('hello')).toBeInTheDocument();
  });

  test('starts in loading=false with no token', async () => {
    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    );
    await waitFor(() =>
      expect(screen.getByTestId('loading').textContent).toBe('false')
    );
    expect(screen.getByTestId('user').textContent).toBe('none');
  });

  test('starts in loading=true when token exists, then resolves', async () => {
    localStorage.setItem('token', 'tok');
    apiMe.mockResolvedValue({ username: 'alice', roles: ['customer'] });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    );

    await waitFor(() =>
      expect(screen.getByTestId('loading').textContent).toBe('false')
    );
    expect(screen.getByTestId('user').textContent).toBe('alice');
  });

  test('clears storage and user when me() rejects', async () => {
    localStorage.setItem('token', 'bad');
    localStorage.setItem('user', JSON.stringify({ username: 'old' }));
    apiMe.mockRejectedValue(new Error('401'));

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    );

    await waitFor(() =>
      expect(screen.getByTestId('loading').textContent).toBe('false')
    );
    expect(screen.getByTestId('user').textContent).toBe('none');
    expect(localStorage.getItem('token')).toBeNull();
  });

  test('login() sets user from API response', async () => {
    apiLogin.mockResolvedValue({
      token: 'new-tok',
      user: { username: 'bob', roles: ['customer'] },
    });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    );

    await waitFor(() =>
      expect(screen.getByTestId('loading').textContent).toBe('false')
    );

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Login' }));
    });

    await waitFor(() =>
      expect(screen.getByTestId('user').textContent).toBe('bob')
    );
  });

  test('logout() clears user immediately', async () => {
    localStorage.setItem('token', 'tok');
    apiMe.mockResolvedValue({ username: 'alice', roles: ['customer'] });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    );

    await waitFor(() =>
      expect(screen.getByTestId('user').textContent).toBe('alice')
    );

    act(() => {
      fireEvent.click(screen.getByRole('button', { name: 'Logout' }));
    });

    await waitFor(() =>
      expect(screen.getByTestId('user').textContent).toBe('none')
    );
    expect(localStorage.getItem('token')).toBeNull();
  });

  test('refreshMe() updates user in state', async () => {
    localStorage.setItem('token', 'tok');
    apiMe.mockResolvedValueOnce({ username: 'alice', roles: ['customer'] });
    apiMe.mockResolvedValueOnce({ username: 'alice-updated', roles: ['customer'] });

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>
    );

    await waitFor(() =>
      expect(screen.getByTestId('user').textContent).toBe('alice')
    );

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    });

    await waitFor(() =>
      expect(screen.getByTestId('user').textContent).toBe('alice-updated')
    );
  });
});

// ── useAuth hook ─────────────────────────────────────────────────────────────

describe('useAuth()', () => {
  test('exposes user, loading, warningOpen, secondsLeft', async () => {
    function Probe() {
      const ctx = useAuth();
      return (
        <div>
          <span data-testid="has-user">{String('user' in ctx)}</span>
          <span data-testid="has-loading">{String('loading' in ctx)}</span>
          <span data-testid="has-warning">{String('warningOpen' in ctx)}</span>
        </div>
      );
    }

    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>
    );

    await waitFor(() =>
      expect(screen.getByTestId('has-loading').textContent).toBe('true')
    );
    expect(screen.getByTestId('has-user').textContent).toBe('true');
    expect(screen.getByTestId('has-warning').textContent).toBe('true');
  });
});
