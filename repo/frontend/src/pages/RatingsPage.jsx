import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { bookings, createRating, listRatings } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Input from '../components/ui/Input';
import Button from '../components/ui/Button';
import DataTable from '../components/ui/DataTable';
import Modal from '../components/ui/Modal';

export default function RatingsPage() {
  const queryClient = useQueryClient();
  const bookingsQuery = useQuery({ queryKey: ['ratings-bookings'], queryFn: bookings });
  const [bookingId, setBookingId] = useState('');
  const [form, setForm] = useState({ score: 5, comment: '' });
  const [open, setOpen] = useState(false);

  const ratingsQuery = useQuery({
    queryKey: ['ratings', bookingId],
    queryFn: () => listRatings(bookingId),
    enabled: Boolean(bookingId),
  });

  const createMutation = useMutation({
    mutationFn: (payload) => createRating(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['ratings', bookingId] }),
  });

  // aggregate ratings across bookings the user can access so we can show
  // a list even when no booking is selected.
  const allRatingsQuery = useQuery({
    queryKey: ['ratings', 'all'],
    queryFn: async () => {
      const b = bookingsQuery.data || [];
      const sets = await Promise.all(b.map((bk) => listRatings(bk.id).catch(() => [])));
      return sets.flat();
    },
    enabled: bookingsQuery.isSuccess,
  });

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-2xl font-semibold">Two-Way Ratings</h2>
          <p className="text-sm text-slate-400">View and send ratings between booking parties</p>
        </div>
        <div className="flex items-center gap-2">
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={bookingId} onChange={(e) => setBookingId(e.target.value)}>
            <option value="">Filter by booking (optional)</option>
            {(bookingsQuery.data || []).map((booking) => <option key={booking.id} value={booking.id}>{booking.id}</option>)}
          </select>
          <Button onClick={() => setOpen(true)}>Create Rating</Button>
        </div>
      </div>

      <Card>
        <CardTitle>Ratings</CardTitle>
        <div className="mt-4">
          <DataTable
            columns={[
              { key: 'bookingId', title: 'Booking' },
              { key: 'fromUserId', title: 'From' },
              { key: 'toUserId', title: 'To' },
              { key: 'score', title: 'Score' },
              { key: 'comment', title: 'Comment' },
            ]}
            rows={(bookingId ? ratingsQuery.data : allRatingsQuery.data) || []}
            empty={(bookingId ? 'No ratings yet for this booking' : 'No ratings available')}
          />
        </div>
      </Card>

      <Modal open={open} onClose={() => setOpen(false)} title="Create Rating" footer={<div className="flex justify-end"><Button variant="secondary" onClick={() => setOpen(false)}>Cancel</Button><Button onClick={() => { createMutation.mutate({ bookingId, score: form.score, comment: form.comment }); setOpen(false); }}>Submit Rating</Button></div>}>
        <div className="grid grid-cols-1 gap-3">
          <label className="text-sm text-slate-200">Booking</label>
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={bookingId} onChange={(e) => setBookingId(e.target.value)}>
            <option value="">Select booking</option>
            {(bookingsQuery.data || []).map((booking) => <option key={booking.id} value={booking.id}>{booking.id}</option>)}
          </select>

          <label className="text-sm text-slate-200">Score</label>
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={form.score} onChange={(e) => setForm((prev) => ({ ...prev, score: Number(e.target.value) }))}>
            {[5, 4, 3, 2, 1].map((score) => <option key={score} value={score}>{score} stars</option>)}
          </select>

          <label className="text-sm text-slate-200">Comment</label>
          <Input placeholder="Optional comment" value={form.comment} onChange={(e) => setForm((prev) => ({ ...prev, comment: e.target.value }))} />
        </div>
      </Modal>
    </div>
  );
}
