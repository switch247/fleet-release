import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('../../../../src/auth/AuthProvider', () => ({
  useAuth: vi.fn(),
}));

import { useAuth } from '../../../../src/auth/AuthProvider';
import IdleWarning from '../../../../src/components/layout/IdleWarning';

beforeEach(() => {
  useAuth.mockReturnValue({
    warningOpen: false,
    secondsLeft: 0,
    staySignedIn: vi.fn(),
    logout: vi.fn(),
  });
});

describe('IdleWarning', () => {
  test('renders nothing when warningOpen is false', () => {
    const { container } = render(<IdleWarning />);
    expect(container.firstChild).toBeNull();
  });

  test('renders warning banner when warningOpen is true', () => {
    useAuth.mockReturnValue({
      warningOpen: true,
      secondsLeft: 60,
      staySignedIn: vi.fn(),
      logout: vi.fn(),
    });
    render(<IdleWarning />);
    expect(screen.getByText('Session timeout warning')).toBeInTheDocument();
  });

  test('shows seconds remaining', () => {
    useAuth.mockReturnValue({
      warningOpen: true,
      secondsLeft: 45,
      staySignedIn: vi.fn(),
      logout: vi.fn(),
    });
    render(<IdleWarning />);
    expect(screen.getByText(/45s/)).toBeInTheDocument();
  });

  test('calls staySignedIn when Stay Signed In button clicked', () => {
    const staySignedIn = vi.fn();
    useAuth.mockReturnValue({
      warningOpen: true,
      secondsLeft: 30,
      staySignedIn,
      logout: vi.fn(),
    });
    render(<IdleWarning />);
    fireEvent.click(screen.getByRole('button', { name: /stay signed in/i }));
    expect(staySignedIn).toHaveBeenCalledTimes(1);
  });

  test('calls logout when Logout button clicked', () => {
    const logout = vi.fn();
    useAuth.mockReturnValue({
      warningOpen: true,
      secondsLeft: 10,
      staySignedIn: vi.fn(),
      logout,
    });
    render(<IdleWarning />);
    fireEvent.click(screen.getByRole('button', { name: /logout/i }));
    expect(logout).toHaveBeenCalledTimes(1);
  });

  test('shows both action buttons', () => {
    useAuth.mockReturnValue({
      warningOpen: true,
      secondsLeft: 20,
      staySignedIn: vi.fn(),
      logout: vi.fn(),
    });
    render(<IdleWarning />);
    expect(screen.getAllByRole('button')).toHaveLength(2);
  });
});
