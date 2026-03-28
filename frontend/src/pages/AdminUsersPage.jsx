import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pencil, Trash2, UserPlus } from 'lucide-react';
import { adminCreateUser, adminDeleteUser, adminUpdateUser, adminUsers } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Input from '../components/ui/Input';
import Button from '../components/ui/Button';
import DataTable from '../components/ui/DataTable';
import Modal from '../components/ui/Modal';

export default function AdminUsersPage() {
  const queryClient = useQueryClient();
  const usersQuery = useQuery({ queryKey: ['admin-users'], queryFn: adminUsers });
  const [form, setForm] = useState({ username: '', email: '', password: '', role: 'customer' });
  const [editing, setEditing] = useState(null);
  const [editForm, setEditForm] = useState({ email: '', role: 'customer', password: '' });

  const createMutation = useMutation({
    mutationFn: adminCreateUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setForm({ username: '', email: '', password: '', role: 'customer' });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, payload }) => adminUpdateUser(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setEditing(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: adminDeleteUser,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Create User</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-4">
          <Input placeholder="Username" value={form.username} onChange={(e) => setForm((prev) => ({ ...prev, username: e.target.value }))} />
          <Input placeholder="Email" value={form.email} onChange={(e) => setForm((prev) => ({ ...prev, email: e.target.value }))} />
          <Input type="password" placeholder="Password" value={form.password} onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))} />
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={form.role} onChange={(e) => setForm((prev) => ({ ...prev, role: e.target.value }))}>
            <option value="customer">Customer</option>
            <option value="provider">Provider</option>
            <option value="csa">Customer Service Agent</option>
          </select>
        </div>
        <Button className="mt-3" onClick={() => createMutation.mutate({ username: form.username, email: form.email, password: form.password, roles: [form.role] })}>
          <UserPlus className="mr-2 h-4 w-4" />
          Create
        </Button>
      </Card>

      <Card>
        <CardTitle>User Management</CardTitle>
        <div className="mt-4">
          <DataTable
            columns={[
              { key: 'username', title: 'Username' },
              { key: 'email', title: 'Email' },
              { key: 'roles', title: 'Roles', render: (row) => row.roles.join(', ') },
              {
                key: 'actions',
                title: 'Actions',
                render: (row) => (
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      onClick={() => {
                        setEditing(row);
                        setEditForm({ email: row.email || '', role: row.roles?.[0] || 'customer', password: '' });
                      }}
                    >
                      <Pencil className="mr-2 h-4 w-4" />
                      Edit
                    </Button>
                    <Button variant="danger" disabled={row.roles.includes('admin')} onClick={() => deleteMutation.mutate(row.id)}>
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </Button>
                  </div>
                ),
              },
            ]}
            rows={usersQuery.data || []}
            empty="No users"
          />
        </div>
      </Card>

      <Modal
        open={!!editing}
        onClose={() => setEditing(null)}
        title={editing ? `Edit ${editing.username}` : 'Edit User'}
        footer={
          <Button
            onClick={() => {
              if (!editing) return;
              const payload = { email: editForm.email, roles: [editForm.role] };
              if (editForm.password) {
                payload.password = editForm.password;
              }
              updateMutation.mutate({ id: editing.id, payload });
            }}
          >
            Save Changes
          </Button>
        }
      >
        <Input placeholder="Email" value={editForm.email} onChange={(e) => setEditForm((prev) => ({ ...prev, email: e.target.value }))} />
        <select className="w-full rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={editForm.role} onChange={(e) => setEditForm((prev) => ({ ...prev, role: e.target.value }))}>
          <option value="customer">Customer</option>
          <option value="provider">Provider</option>
          <option value="csa">Customer Service Agent</option>
          <option value="admin">Admin</option>
        </select>
        <Input type="password" placeholder="New Password (optional)" value={editForm.password} onChange={(e) => setEditForm((prev) => ({ ...prev, password: e.target.value }))} />
      </Modal>
    </div>
  );
}
