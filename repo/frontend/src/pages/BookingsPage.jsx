import React, { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { enqueue, getQueue } from '../offline/queue';
import { bookings, createBooking, estimateBooking, listings } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import Modal from '../components/ui/Modal';
import DataTable from '../components/ui/DataTable';
import Badge from '../components/ui/Badge';
import EstimateSummary from '../components/EstimateSummary';

export default function BookingsPage() {
  const queryClient = useQueryClient();
  const bookingsQuery = useQuery({ queryKey: ['bookings'], queryFn: bookings });
  const listingsQuery = useQuery({ queryKey: ['listings'], queryFn: listings });
  const [open, setOpen] = useState(false);
  const [queueSize, setQueueSize] = useState(getQueue().length);
  const [form, setForm] = useState({ listingId: '', couponCode: '', startAt: '', endAt: '', odoStart: '0', odoEnd: '0' });
  const [estimatePreview, setEstimatePreview] = useState(null);
  const [estimateError, setEstimateError] = useState('');

  const createMutation = useMutation({
    mutationFn: createBooking,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bookings'] });
      setOpen(false);
    },
  });

  const estimateMutation = useMutation({
    mutationFn: estimateBooking,
    onSuccess: (data) => {
      setEstimatePreview(data?.estimate || null);
      setEstimateError('');
    },
    onError: (err) => {
      setEstimatePreview(null);
      setEstimateError(err.message);
    },
  });

  const submitBooking = async () => {
    const payload = {
      ...form,
      // datetime-local inputs keep local-format string (YYYY-MM-DDTHH:MM). Convert to ISO for the API.
      startAt: form.startAt ? new Date(form.startAt).toISOString() : null,
      endAt: form.endAt ? new Date(form.endAt).toISOString() : null,
      odoStart: Number(form.odoStart),
      odoEnd: Number(form.odoEnd),
    };
    if (!navigator.onLine) {
      enqueue({ type: 'booking', path: '/bookings', method: 'POST', payload });
      setQueueSize(getQueue().length);
      setOpen(false);
      return;
    }
    if (!estimatePreview) {
      setEstimateError('Preview estimate before confirming the booking.');
      return;
    }
    await createMutation.mutateAsync(payload);
  };

  const previewEstimate = async () => {
    const payload = {
      listingId: form.listingId,
      startAt: form.startAt ? new Date(form.startAt).toISOString() : null,
      endAt: form.endAt ? new Date(form.endAt).toISOString() : null,
      odoStart: Number(form.odoStart),
      odoEnd: Number(form.odoEnd),
    };
    await estimateMutation.mutateAsync(payload);
  };

  const selectedListing = (listingsQuery.data || []).find((l) => l.id === form.listingId);
  const isValid = form.listingId && form.startAt && form.endAt && (new Date(form.startAt) < new Date(form.endAt));

  useEffect(() => {
    setEstimatePreview(null);
    setEstimateError('');
  }, [form.listingId, form.startAt, form.endAt, form.odoStart, form.odoEnd]);

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

      <Modal
        open={open}
        onClose={() => setOpen(false)}
        title="Create Booking"
        footer={(
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setOpen(false)}>Cancel</Button>
          <Button onClick={submitBooking} disabled={!isValid || createMutation.isPending}>Create Booking</Button>
        </div>
      )}
    >
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-slate-200 mb-1">Listing</label>
            <select
              className="w-full rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm"
              value={form.listingId}
              onChange={(e) => setForm((prev) => ({ ...prev, listingId: e.target.value }))}
            >
              <option value="">Select listing</option>
              {(listingsQuery.data || []).map((listing) => (
                <option key={listing.id} value={listing.id}>{listing.name} ({listing.spu}/{listing.sku})</option>
              ))}
            </select>
            {selectedListing && (
              <p className="mt-2 text-xs text-slate-400">Selected: {selectedListing.name} — {selectedListing.spu}/{selectedListing.sku}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">Coupon code</label>
            <Input placeholder="Optional" value={form.couponCode} onChange={(e) => setForm((prev) => ({ ...prev, couponCode: e.target.value }))} />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">Odometer (start)</label>
            <Input placeholder="0" value={form.odoStart} onChange={(e) => setForm((prev) => ({ ...prev, odoStart: e.target.value }))} />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">Odometer (end)</label>
            <Input placeholder="0" value={form.odoEnd} onChange={(e) => setForm((prev) => ({ ...prev, odoEnd: e.target.value }))} />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">Start</label>
            <Input type="datetime-local" value={form.startAt} onChange={(e) => setForm((prev) => ({ ...prev, startAt: e.target.value }))} />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-200 mb-1">End</label>
            <Input type="datetime-local" value={form.endAt} onChange={(e) => setForm((prev) => ({ ...prev, endAt: e.target.value }))} />
          </div>

          <div className="md:col-span-2">
            {!isValid && <p className="text-xs text-rose-400">Please select a listing and a valid start/end period.</p>}
            {isValid && (
              <div className="flex items-center gap-2 mt-2">
                <Button variant="outline" onClick={previewEstimate} disabled={estimateMutation.isPending}>
                  {estimateMutation.isPending ? 'Calculating...' : 'Preview Estimate'}
                </Button>
              </div>
            )}
            {estimateError && <p className="mt-2 text-xs text-rose-400">{estimateError}</p>}
            {estimatePreview && <EstimateSummary estimate={estimatePreview} />}
          </div>
        </div>
      </Modal>
    </div>
  );
}
