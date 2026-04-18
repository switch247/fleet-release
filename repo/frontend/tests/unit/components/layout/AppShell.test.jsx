import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../../src/auth/AuthProvider', () => ({
  useAuth: vi.fn(),
}));

import { useAuth } from '../../../../src/auth/AuthProvider';
import AppShell from '../../../../src/components/layout/AppShell';

function renderShell(user = { username: 'alice', roles: ['customer'] }, logout = vi.fn()) {
  useAuth.mockReturnValue({ user, logout });
  return render(
    <MemoryRouter initialEntries={['/overview']}>
      <AppShell />
    </MemoryRouter>
  );
}

describe('AppShell', () => {
  test('renders FleetLease brand name', () => {
    renderShell();
    expect(screen.getByText('FleetLease')).toBeInTheDocument();
  });

  test('shows Operations Suite title', () => {
    renderShell();
    expect(screen.getByText('Operations Suite')).toBeInTheDocument();
  });

  test('shows username in sidebar', () => {
    renderShell();
    expect(screen.getByText(/alice/)).toBeInTheDocument();
  });

  test('shows user roles in sidebar', () => {
    renderShell();
    expect(screen.getByText(/customer/)).toBeInTheDocument();
  });

  test('Overview nav link is always visible', () => {
    renderShell();
    expect(screen.getByText('Overview')).toBeInTheDocument();
  });

  test('Inbox nav link is always visible', () => {
    renderShell();
    expect(screen.getByText('Inbox')).toBeInTheDocument();
  });

  test('Profile nav link is always visible', () => {
    renderShell();
    expect(screen.getByText('Profile')).toBeInTheDocument();
  });

  test('Bookings link shown for customer role', () => {
    renderShell({ username: 'alice', roles: ['customer'] });
    expect(screen.getByText('Bookings')).toBeInTheDocument();
  });

  test('Catalog link shown for customer role', () => {
    renderShell({ username: 'alice', roles: ['customer'] });
    expect(screen.getByText('Catalog')).toBeInTheDocument();
  });

  test('admin links hidden for customer role', () => {
    renderShell({ username: 'alice', roles: ['customer'] });
    expect(screen.queryByText('Admin Users')).not.toBeInTheDocument();
    expect(screen.queryByText('Admin Catalog')).not.toBeInTheDocument();
    expect(screen.queryByText('Admin Notify')).not.toBeInTheDocument();
  });

  test('admin links shown for admin role', () => {
    renderShell({ username: 'admin', roles: ['admin'] });
    expect(screen.getByText('Admin Users')).toBeInTheDocument();
    expect(screen.getByText('Admin Catalog')).toBeInTheDocument();
    expect(screen.getByText('Admin Notify')).toBeInTheDocument();
  });

  test('Catalog link hidden for provider role', () => {
    renderShell({ username: 'prov', roles: ['provider'] });
    expect(screen.queryByText('Catalog')).not.toBeInTheDocument();
  });

  test('Inspections link shown for csa role', () => {
    renderShell({ username: 'agent', roles: ['csa'] });
    expect(screen.getByText('Inspections')).toBeInTheDocument();
  });

  test('Logout button is rendered', () => {
    renderShell();
    expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument();
  });

  test('clicking Logout calls the logout function', () => {
    const logout = vi.fn();
    renderShell({ username: 'alice', roles: ['customer'] }, logout);
    fireEvent.click(screen.getByRole('button', { name: /logout/i }));
    expect(logout).toHaveBeenCalledTimes(1);
  });
});
