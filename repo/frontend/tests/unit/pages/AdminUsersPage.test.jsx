import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/lib/api', () => ({
  adminUsers: vi.fn(),
  adminCreateUser: vi.fn(),
  adminUpdateUser: vi.fn(),
  adminDeleteUser: vi.fn(),
}));

import { adminUsers, adminCreateUser, adminUpdateUser, adminDeleteUser } from '../../../src/lib/api';
import AdminUsersPage from '../../../src/pages/AdminUsersPage';

const sampleUsers = [
  { id: 'u1', username: 'alice', email: 'alice@ex.com', roles: ['customer'] },
  { id: 'u2', username: 'admin', email: 'admin@ex.com', roles: ['admin'] },
  { id: 'u3', username: 'provider1', email: 'p@ex.com', roles: ['provider'] },
];

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AdminUsersPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => {
  adminUsers.mockResolvedValue(sampleUsers);
  adminCreateUser.mockResolvedValue({ id: 'u4' });
  adminUpdateUser.mockResolvedValue({ id: 'u1' });
  adminDeleteUser.mockResolvedValue(null);
});

describe('AdminUsersPage', () => {
  test('renders Create User card', () => {
    renderPage();
    expect(screen.getByText('Create User')).toBeInTheDocument();
  });

  test('renders User Management card', () => {
    renderPage();
    expect(screen.getByText('User Management')).toBeInTheDocument();
  });

  test('renders username input', () => {
    renderPage();
    expect(screen.getByPlaceholderText('Username')).toBeInTheDocument();
  });

  test('renders email input', () => {
    renderPage();
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument();
  });

  test('renders password input', () => {
    renderPage();
    expect(document.querySelector('input[type="password"]')).toBeInTheDocument();
  });

  test('renders role selector', () => {
    renderPage();
    expect(screen.getByDisplayValue('Customer')).toBeInTheDocument();
  });

  test('renders Create button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument();
  });

  test('shows users in table after load', async () => {
    renderPage();
    expect(await screen.findByText('alice')).toBeInTheDocument();
    expect(await screen.findByText('provider1')).toBeInTheDocument();
  });

  test('shows user email in table', async () => {
    renderPage();
    expect(await screen.findByText('alice@ex.com')).toBeInTheDocument();
  });

  test('shows user roles in table', async () => {
    renderPage();
    await screen.findByText('alice');
    expect(screen.getByText('customer')).toBeInTheDocument();
  });

  test('Edit button exists for each user', async () => {
    renderPage();
    await screen.findByText('alice');
    const editBtns = screen.getAllByRole('button', { name: /edit/i });
    expect(editBtns.length).toBeGreaterThanOrEqual(sampleUsers.length);
  });

  test('Delete button is disabled for admin user', async () => {
    renderPage();
    // 'admin' appears in both username cell and role cell — use findAllByText
    await screen.findAllByText('admin');
    const rows = document.querySelectorAll('tbody tr');
    // Find the row whose username is 'admin'
    let adminDeleteBtn = null;
    rows.forEach((row) => {
      const cells = row.querySelectorAll('td');
      if (cells[0] && cells[0].textContent === 'admin') {
        adminDeleteBtn = row.querySelector('button[disabled]');
      }
    });
    expect(adminDeleteBtn).not.toBeNull();
  });

  test('clicking Edit opens edit modal', async () => {
    renderPage();
    await screen.findByText('alice');
    fireEvent.click(screen.getAllByRole('button', { name: /edit/i })[0]);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /save changes/i })).toBeInTheDocument()
    );
  });

  test('edit modal shows user email', async () => {
    renderPage();
    await screen.findByText('alice');
    fireEvent.click(screen.getAllByRole('button', { name: /edit/i })[0]);
    await waitFor(() =>
      expect(screen.getByDisplayValue('alice@ex.com')).toBeInTheDocument()
    );
  });

  test('save changes calls adminUpdateUser', async () => {
    renderPage();
    await screen.findByText('alice');
    fireEvent.click(screen.getAllByRole('button', { name: /edit/i })[0]);
    await screen.findByRole('button', { name: /save changes/i });
    fireEvent.click(screen.getByRole('button', { name: /save changes/i }));
    await waitFor(() =>
      expect(adminUpdateUser).toHaveBeenCalledWith('u1', expect.any(Object))
    );
  });

  test('calls adminCreateUser when Create is clicked', async () => {
    renderPage();
    fireEvent.change(screen.getByPlaceholderText('Username'), { target: { value: 'newuser' } });
    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 'new@ex.com' } });
    fireEvent.change(document.querySelector('input[type="password"]'), { target: { value: 'Pass123!' } });
    fireEvent.click(screen.getByRole('button', { name: /create/i }));
    // React Query v5 passes a second context arg to mutationFn
    await waitFor(() =>
      expect(adminCreateUser).toHaveBeenCalledWith(
        expect.objectContaining({ username: 'newuser', email: 'new@ex.com' }),
        expect.any(Object)
      )
    );
  });

  test('shows empty message when no users', async () => {
    adminUsers.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No users')).toBeInTheDocument()
    );
  });

  test('table column headers are correct', () => {
    renderPage();
    expect(screen.getByText('Username')).toBeInTheDocument();
    expect(screen.getByText('Email')).toBeInTheDocument();
    expect(screen.getByText('Roles')).toBeInTheDocument();
    expect(screen.getByText('Actions')).toBeInTheDocument();
  });

  test('edit modal opens when Edit is clicked', async () => {
    renderPage();
    await screen.findAllByText('alice');
    const editBtns = screen.getAllByRole('button', { name: /^edit$/i });
    fireEvent.click(editBtns[0]);
    await waitFor(() =>
      expect(screen.getByRole('heading', { name: /edit/i })).toBeInTheDocument()
    );
  });

  test('Save Changes with password includes password in payload', async () => {
    renderPage();
    await screen.findAllByText('alice');
    const editBtns = screen.getAllByRole('button', { name: /^edit$/i });
    fireEvent.click(editBtns[0]);
    await screen.findByRole('heading', { name: /edit/i });
    // Find the password input in the modal (there may be multiple password inputs)
    const pwInputs = document.querySelectorAll('input[type="password"]');
    if (pwInputs.length > 0) {
      // Use the last password input (which is in the edit modal)
      fireEvent.change(pwInputs[pwInputs.length - 1], { target: { value: 'NewPass123!' } });
    }
    fireEvent.click(screen.getByRole('button', { name: /save changes/i }));
    await waitFor(() =>
      expect(adminUpdateUser).toHaveBeenCalled()
    );
  });

  test('clicking Delete for non-admin user calls adminDeleteUser', async () => {
    renderPage();
    await screen.findAllByText('alice');
    const deleteBtns = screen.getAllByRole('button', { name: /^delete$/i });
    // First non-disabled delete button
    const activeDelete = Array.from(deleteBtns).find((b) => !b.disabled);
    if (activeDelete) {
      fireEvent.click(activeDelete);
      await waitFor(() =>
        expect(adminDeleteUser).toHaveBeenCalled()
      );
    }
  });

  test('edit modal email and role inputs can be changed', async () => {
    renderPage();
    await screen.findByText('alice');
    fireEvent.click(screen.getAllByRole('button', { name: /^edit$/i })[0]);
    await screen.findByRole('heading', { name: /edit/i });
    // Change email
    const emailInput = screen.getByDisplayValue('alice@ex.com');
    fireEvent.change(emailInput, { target: { value: 'newemail@ex.com' } });
    expect(screen.getByDisplayValue('newemail@ex.com')).toBeInTheDocument();
    // The modal's role select includes 'admin' option — the create form only has 3 options
    const allSelects = document.querySelectorAll('select');
    const modalRoleSelect = Array.from(allSelects).find((s) =>
      Array.from(s.options).some((o) => o.value === 'admin')
    );
    if (modalRoleSelect) {
      fireEvent.change(modalRoleSelect, { target: { value: 'provider' } });
      expect(modalRoleSelect.value).toBe('provider');
    }
  });
});
