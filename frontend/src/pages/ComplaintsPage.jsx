import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { arbitrateComplaint, bookings, complaints, createComplaint } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import DataTable from '../components/ui/DataTable';
import Badge from '../components/ui/Badge';
import { useAuth } from '../auth/AuthProvider';

export default function ComplaintsPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const bookingsQuery = useQuery({ queryKey: ['complaint-bookings'], queryFn: bookings });
  const complaintsQuery = useQuery({ queryKey: ['complaints'], queryFn: () => complaints('') });
  const [form, setForm] = useState({ bookingId: '', outcome: '' });
  const [decision, setDecision] = useState({ status: 'resolved', outcome: '' });

  const createMutation = useMutation({
    mutationFn: createComplaint,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['complaints'] }),
  });

  const arbitrateMutation = useMutation({
    mutationFn: ({ id, payload }) => arbitrateComplaint(id, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['complaints'] }),
  });

  const canArbitrate = user?.roles?.includes('csa') || user?.roles?.includes('admin');

  const rows = complaintsQuery.data || [];
  const columns = useMemo(() => [
    { key: 'id', title: 'Complaint ID' },
    { key: 'bookingId', title: 'Booking' },
    { key: 'openedBy', title: 'Opened By' },
    {
      key: 'status',
      title: 'Status',
      render: (row) => <Badge variant={row.status === 'open' ? 'warning' : 'success'}>{row.status}</Badge>,
    },
    { key: 'outcome', title: 'Details' },
    {
      key: 'actions',
      title: 'Actions',
      render: (row) => canArbitrate ? (
        <div className="flex items-center gap-2">
          <select
            className="rounded-lg border border-slate-700 bg-slate-900 px-2 py-1 text-xs"
            value={decision.status}
            onChange={(e) => setDecision((prev) => ({ ...prev, status: e.target.value }))}
          >
            <option value="resolved">Resolved</option>
            <option value="dismissed">Dismissed</option>
            <option value="under_review">Under Review</option>
          </select>
          <Input
            className="h-8"
            placeholder="Outcome"
            value={decision.outcome}
            onChange={(e) => setDecision((prev) => ({ ...prev, outcome: e.target.value }))}
          />
          <Button
            className="h-8"
            onClick={() => arbitrateMutation.mutate({ id: row.id, payload: { status: decision.status, outcome: decision.outcome } })}
          >
            Apply
          </Button>
        </div>
      ) : 'N/A',
    },
  ], [arbitrateMutation, canArbitrate, decision.outcome, decision.status]);

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Open Complaint</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-[1fr,2fr,auto]">
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={form.bookingId} onChange={(e) => setForm((prev) => ({ ...prev, bookingId: e.target.value }))}>
            <option value="">Select booking</option>
            {(bookingsQuery.data || []).map((booking) => <option key={booking.id} value={booking.id}>{booking.id}</option>)}
          </select>
          <Input placeholder="Complaint details" value={form.outcome} onChange={(e) => setForm((prev) => ({ ...prev, outcome: e.target.value }))} />
          <Button onClick={() => createMutation.mutate(form)}>Submit</Button>
        </div>
      </Card>

      <Card>
        <CardTitle>Complaints</CardTitle>
        <div className="mt-4">
          <DataTable columns={columns} rows={rows} empty="No complaints" />
        </div>
      </Card>
    </div>
  );
}
