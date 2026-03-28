import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { enqueue, getQueue } from '../offline/queue';
import { bookings, createBooking, listings } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import Modal from '../components/ui/Modal';
import DataTable from '../components/ui/DataTable';
import Badge from '../components/ui/Badge';

export default function BookingsPage() {
  const queryClient = useQueryClient();
  const bookingsQuery = useQuery({ queryKey: ['bookings'], queryFn: bookings });
  const listingsQuery = useQuery({ queryKey: ['listings'], queryFn: listings });
  const [open, setOpen] = useState(false);
  const [queueSize, setQueueSize] = useState(getQueue().length);
  const [form, setForm] = useState({ listingId: '', couponCode: '', startAt: '', endAt: '', odoStart: '0', odoEnd: '0' });

  const createMutation = useMutation({
    mutationFn: createBooking,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bookings'] });
      setOpen(false);
    },
  });

  const submitBooking = async () => {
    const payload = { ...form, odoStart: Number(form.odoStart), odoEnd: Number(form.odoEnd) };
    if (!navigator.onLine) {
      enqueue({ type: 'booking', path: '/bookings', method: 'POST', payload });
      setQueueSize(getQueue().length);
      setOpen(false);
      return;
    }
    await createMutation.mutateAsync(payload);
  };

  const columns = useMemo(() => [
    { key: 'id', title: 'Booking ID' },
    { key: 'listingId', title: 'Listing' },
    {
      key: 'status',
      title: 'Status',
      render: (row) => <Badge variant={row.status === 'settled' ? 'success' : 'warning'}>{row.status}</Badge>,
    },
    { key: 'estimatedAmount', title: 'Estimate', render: (row) => `$${Number(row.estimatedAmount).toFixed(2)}` },
    { key: 'depositAmount', title: 'Deposit', render: (row) => `$${Number(row.depositAmount).toFixed(2)}` },
  ], []);

  return (
    <div className="space-y-6">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold">Bookings</h2>
          <p className="text-sm text-slate-400">Live bookings with offline queue support</p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="neutral">Offline queue: {queueSize}</Badge>
          <Button onClick={() => setOpen(true)}>New Booking</Button>
        </div>
      </header>

      <Card>
        <CardTitle>Current Bookings</CardTitle>
        <div className="mt-4">
          <DataTable columns={columns} rows={bookingsQuery.data || []} empty="No bookings yet" />
        </div>
      </Card>

      <Modal open={open} onClose={() => setOpen(false)} title="Create Booking" footer={<Button onClick={submitBooking}>Create Booking</Button>}>
        <select className="w-full rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={form.listingId} onChange={(e) => setForm((prev) => ({ ...prev, listingId: e.target.value }))}>
          <option value="">Select listing</option>
          {(listingsQuery.data || []).map((listing) => (
            <option key={listing.id} value={listing.id}>{listing.name} ({listing.spu}/{listing.sku})</option>
          ))}
        </select>
        <Input placeholder="Coupon code" value={form.couponCode} onChange={(e) => setForm((prev) => ({ ...prev, couponCode: e.target.value }))} />
        <Input type="datetime-local" value={form.startAt} onChange={(e) => setForm((prev) => ({ ...prev, startAt: new Date(e.target.value).toISOString() }))} />
        <Input type="datetime-local" value={form.endAt} onChange={(e) => setForm((prev) => ({ ...prev, endAt: new Date(e.target.value).toISOString() }))} />
        <div className="grid grid-cols-2 gap-2">
          <Input placeholder="Odometer start" value={form.odoStart} onChange={(e) => setForm((prev) => ({ ...prev, odoStart: e.target.value }))} />
          <Input placeholder="Odometer end" value={form.odoEnd} onChange={(e) => setForm((prev) => ({ ...prev, odoEnd: e.target.value }))} />
        </div>
      </Modal>
    </div>
  );
}
