import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { bookings, createRating, listRatings } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Input from '../components/ui/Input';
import Button from '../components/ui/Button';
import DataTable from '../components/ui/DataTable';

export default function RatingsPage() {
  const queryClient = useQueryClient();
  const bookingsQuery = useQuery({ queryKey: ['ratings-bookings'], queryFn: bookings });
  const [bookingId, setBookingId] = useState('');
  const [form, setForm] = useState({ score: 5, comment: '' });

  const ratingsQuery = useQuery({
    queryKey: ['ratings', bookingId],
    queryFn: () => listRatings(bookingId),
    enabled: Boolean(bookingId),
  });

  const createMutation = useMutation({
    mutationFn: (payload) => createRating(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['ratings', bookingId] }),
  });

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Two-Way Ratings</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-[1fr,0.5fr,1fr,auto]">
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={bookingId} onChange={(e) => setBookingId(e.target.value)}>
            <option value="">Select booking</option>
            {(bookingsQuery.data || []).map((booking) => <option key={booking.id} value={booking.id}>{booking.id}</option>)}
          </select>
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={form.score} onChange={(e) => setForm((prev) => ({ ...prev, score: Number(e.target.value) }))}>
            {[5, 4, 3, 2, 1].map((score) => <option key={score} value={score}>{score} stars</option>)}
          </select>
          <Input placeholder="Comment" value={form.comment} onChange={(e) => setForm((prev) => ({ ...prev, comment: e.target.value }))} />
          <Button onClick={() => createMutation.mutate({ bookingId, score: form.score, comment: form.comment })}>Submit</Button>
        </div>
      </Card>

      <Card>
        <CardTitle>Ratings History</CardTitle>
        <div className="mt-4">
          <DataTable
            columns={[
              { key: 'fromUserId', title: 'From' },
              { key: 'toUserId', title: 'To' },
              { key: 'score', title: 'Score' },
              { key: 'comment', title: 'Comment' },
            ]}
            rows={ratingsQuery.data || []}
            empty={bookingId ? 'No ratings yet' : 'Select a booking'}
          />
        </div>
      </Card>
    </div>
  );
}
