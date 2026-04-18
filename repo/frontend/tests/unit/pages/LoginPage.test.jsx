import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/auth/AuthProvider', () => ({
  useAuth: vi.fn(),
}));

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal();
  return { ...actual, useNavigate: () => mockNavigate };
});

import { useAuth } from '../../../src/auth/AuthProvider';
import LoginPage from '../../../src/pages/LoginPage';

const mockLogin = vi.fn();

beforeEach(() => {
  mockLogin.mockResolvedValue({});
  useAuth.mockReturnValue({ login: mockLogin });
});

function renderLogin() {
  return render(<MemoryRouter><LoginPage /></MemoryRouter>);
}

describe('LoginPage', () => {
  test('renders Sign In heading', () => {
    renderLogin();
    expect(screen.getByRole('heading', { name: 'Sign In' })).toBeInTheDocument();
  });

  test('renders username input', () => {
    renderLogin();
    expect(screen.getByPlaceholderText('Username')).toBeInTheDocument();
  });

  test('renders password input', () => {
    renderLogin();
    expect(document.querySelector('input[type="password"]')).toBeInTheDocument();
  });

  test('renders TOTP code input', () => {
    renderLogin();
    expect(screen.getByPlaceholderText('TOTP Code (optional)')).toBeInTheDocument();
  });

  test('renders Sign In button', () => {
    renderLogin();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  test('updates username field on change', () => {
    renderLogin();
    const input = screen.getByPlaceholderText('Username');
    fireEvent.change(input, { target: { value: 'alice' } });
    expect(input.value).toBe('alice');
  });

  test('updates password field on change', () => {
    renderLogin();
    const input = document.querySelector('input[type="password"]');
    fireEvent.change(input, { target: { value: 'Secret1!' } });
    expect(input.value).toBe('Secret1!');
  });

  test('calls login with form values on submit', async () => {
    renderLogin();
    fireEvent.change(screen.getByPlaceholderText('Username'), { target: { value: 'alice' } });
    fireEvent.change(document.querySelector('input[type="password"]'), { target: { value: 'Pass1!' } });
    fireEvent.submit(screen.getByRole('button', { name: /sign in/i }).closest('form'));
    await waitFor(() =>
      expect(mockLogin).toHaveBeenCalledWith(
        expect.objectContaining({ username: 'alice', password: 'Pass1!' })
      )
    );
  });

  test('navigates to /overview after successful login', async () => {
    renderLogin();
    fireEvent.submit(screen.getByRole('button').closest('form'));
    await waitFor(() =>
      expect(mockNavigate).toHaveBeenCalledWith('/overview', { replace: true })
    );
  });

  test('shows error message when login fails', async () => {
    mockLogin.mockRejectedValue(new Error('Invalid credentials'));
    renderLogin();
    fireEvent.submit(screen.getByRole('button').closest('form'));
    await waitFor(() =>
      expect(screen.getByText('Invalid credentials')).toBeInTheDocument()
    );
  });

  test('shows loading text while login is in progress', async () => {
    let resolve;
    mockLogin.mockImplementation(() => new Promise((r) => { resolve = r; }));
    renderLogin();
    fireEvent.submit(screen.getByRole('button').closest('form'));
    await waitFor(() =>
      expect(screen.getByText('Signing in...')).toBeInTheDocument()
    );
    resolve({});
  });

  test('button is disabled while loading', async () => {
    let resolve;
    mockLogin.mockImplementation(() => new Promise((r) => { resolve = r; }));
    renderLogin();
    fireEvent.submit(screen.getByRole('button').closest('form'));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /signing in/i })).toBeDisabled()
    );
    resolve({});
  });

  test('updates TOTP code field on change', () => {
    renderLogin();
    const totpInput = screen.getByPlaceholderText('TOTP Code (optional)');
    fireEvent.change(totpInput, { target: { value: '123456' } });
    expect(totpInput.value).toBe('123456');
  });
});
