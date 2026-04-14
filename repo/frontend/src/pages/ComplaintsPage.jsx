import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { arbitrateComplaint, bookings, complaints, createComplaint, exportDisputePDF } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import DataTable from '../components/ui/DataTable';
import Badge from '../components/ui/Badge';
import Modal from '../components/ui/Modal';
import { useAuth } from '../auth/AuthProvider';
import { enqueue } from '../offline/queue';

export default function ComplaintsPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const bookingsQuery = useQuery({ queryKey: ['complaint-bookings'], queryFn: bookings });
  const complaintsQuery = useQuery({ queryKey: ['complaints'], queryFn: () => complaints('') });
  const [form, setForm] = useState({ bookingId: '', outcome: '' });
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedComplaint, setSelectedComplaint] = useState(null);
  const [modalDecision, setModalDecision] = useState({ status: 'resolved', outcome: '' });

  const createMutation = useMutation({
    mutationFn: createComplaint,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['complaints'] }),
  });

  const arbitrateMutation = useMutation({
    mutationFn: ({ id, payload }) => arbitrateComplaint(id, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['complaints'] }),
  });
  const exportMutation = useMutation({ mutationFn: exportDisputePDF });

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
      render: (row) => (
        <div className="flex items-center gap-2">
          {canArbitrate ? (
            <Button size="sm" onClick={() => { setSelectedComplaint(row); setModalDecision({ status: 'resolved', outcome: '' }); setModalOpen(true); }}>Arbitrate</Button>
          ) : null}
          <Button
            size="sm"
            variant="outline"
            onClick={async () => {
              const blob = await exportMutation.mutateAsync(row.id);
              const url = URL.createObjectURL(blob);
              const a = document.createElement('a');
              a.href = url;
              a.download = `dispute-${row.id}.pdf`;
              a.click();
              URL.revokeObjectURL(url);
            }}
          >
            Export PDF
          </Button>
        </div>
      ),
    },
  ], [canArbitrate, exportMutation]);

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
          <Button onClick={() => {
              if (!navigator.onLine) {
                enqueue({ type: 'complaint', payload: { ...form } });
                return;
              }
              createMutation.mutate(form);
            }}>Submit</Button>
        </div>
      </Card>

      <Card>
        <CardTitle>Complaints</CardTitle>
        <div className="mt-4">
          <DataTable columns={columns} rows={rows} empty="No complaints" />
        </div>
      </Card>

      {/* Arbitrate modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={selectedComplaint ? `Arbitrate ${selectedComplaint.id}` : 'Arbitrate'} footer={(
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => setModalOpen(false)}>Cancel</Button>
          <Button onClick={() => {
            if (!selectedComplaint) return;
            arbitrateMutation.mutate({ id: selectedComplaint.id, payload: modalDecision }, {
              onSuccess: () => {
                setModalOpen(false);
                setSelectedComplaint(null);
              }
            });
          }}>Apply</Button>
        </div>
      )}>
        <div className="space-y-3">
          <select
            className="rounded-lg border border-slate-700 bg-slate-900 px-2 py-1 text-sm"
            value={modalDecision.status}
            onChange={(e) => setModalDecision((p) => ({ ...p, status: e.target.value }))}
          >
            <option value="resolved">Resolved</option>
            <option value="dismissed">Dismissed</option>
            <option value="under_review">Under Review</option>
          </select>
          <Input placeholder="Outcome" value={modalDecision.outcome} onChange={(e) => setModalDecision((p) => ({ ...p, outcome: e.target.value }))} />
        </div>
      </Modal>
    </div>
  );
}
