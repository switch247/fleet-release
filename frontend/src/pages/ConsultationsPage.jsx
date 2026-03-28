import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  addConsultationAttachment,
  bookings,
  consultationAttachments,
  consultations,
  createConsultation,
} from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import DataTable from '../components/ui/DataTable';
import Badge from '../components/ui/Badge';
import { useAuth } from '../auth/AuthProvider';

export default function ConsultationsPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const bookingsQuery = useQuery({ queryKey: ['consult-bookings'], queryFn: bookings });
  const [selectedBooking, setSelectedBooking] = useState('');

  const consultationsQuery = useQuery({
    queryKey: ['consultations', selectedBooking],
    queryFn: () => consultations(selectedBooking),
    enabled: Boolean(selectedBooking),
  });

  const [form, setForm] = useState({ topic: '', keyPoints: '', recommendation: '', followUp: '', visibility: 'parties', changeReason: '' });
  const [attachment, setAttachment] = useState({ consultationId: '', attachmentId: '' });

  const attachmentListQuery = useQuery({
    queryKey: ['consultation-attachments', attachment.consultationId],
    queryFn: () => consultationAttachments(attachment.consultationId),
    enabled: Boolean(attachment.consultationId),
  });

  const createMutation = useMutation({
    mutationFn: (payload) => createConsultation(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['consultations', selectedBooking] }),
  });

  const attachMutation = useMutation({
    mutationFn: addConsultationAttachment,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['consultation-attachments', attachment.consultationId] }),
  });

  const canManage = user?.roles?.includes('csa') || user?.roles?.includes('admin');

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Consultation Management</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-[1fr,auto]">
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={selectedBooking} onChange={(e) => setSelectedBooking(e.target.value)}>
            <option value="">Select booking</option>
            {(bookingsQuery.data || []).map((booking) => <option key={booking.id} value={booking.id}>{booking.id}</option>)}
          </select>
          {selectedBooking && <Badge variant="neutral">Booking: {selectedBooking}</Badge>}
        </div>
      </Card>

      {canManage && selectedBooking && (
        <Card>
          <CardTitle>Create Consultation Note</CardTitle>
          <div className="mt-4 grid gap-3 md:grid-cols-2">
            <Input placeholder="Topic" value={form.topic} onChange={(e) => setForm((prev) => ({ ...prev, topic: e.target.value }))} />
            <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={form.visibility} onChange={(e) => setForm((prev) => ({ ...prev, visibility: e.target.value }))}>
              <option value="csa_admin">CSA/Admin</option>
              <option value="parties">Parties</option>
              <option value="all">All</option>
            </select>
            <Input placeholder="Key points" value={form.keyPoints} onChange={(e) => setForm((prev) => ({ ...prev, keyPoints: e.target.value }))} />
            <Input placeholder="Recommendation" value={form.recommendation} onChange={(e) => setForm((prev) => ({ ...prev, recommendation: e.target.value }))} />
            <Input placeholder="Follow-up" value={form.followUp} onChange={(e) => setForm((prev) => ({ ...prev, followUp: e.target.value }))} />
            <Input placeholder="Change reason" value={form.changeReason} onChange={(e) => setForm((prev) => ({ ...prev, changeReason: e.target.value }))} />
          </div>
          <Button className="mt-3" onClick={() => createMutation.mutate({ bookingId: selectedBooking, ...form })}>Create Note</Button>
        </Card>
      )}

      <Card>
        <CardTitle>Consultation Notes</CardTitle>
        <div className="mt-4">
          <DataTable
            columns={[
              { key: 'id', title: 'ID' },
              { key: 'topic', title: 'Topic' },
              { key: 'version', title: 'Version' },
              { key: 'visibility', title: 'Visibility' },
              { key: 'keyPoints', title: 'Key Points' },
            ]}
            rows={consultationsQuery.data || []}
            empty={selectedBooking ? 'No consultations yet' : 'Select a booking first'}
          />
        </div>
      </Card>

      {canManage && (
        <Card>
          <CardTitle>Attach Evidence to Consultation</CardTitle>
          <div className="mt-4 grid gap-3 md:grid-cols-[1fr,1fr,auto]">
            <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={attachment.consultationId} onChange={(e) => setAttachment((prev) => ({ ...prev, consultationId: e.target.value }))}>
              <option value="">Consultation</option>
              {(consultationsQuery.data || []).map((item) => <option key={item.id} value={item.id}>{item.topic} (v{item.version})</option>)}
            </select>
            <Input placeholder="Attachment ID" value={attachment.attachmentId} onChange={(e) => setAttachment((prev) => ({ ...prev, attachmentId: e.target.value }))} />
            <Button onClick={() => attachMutation.mutate(attachment)}>Attach</Button>
          </div>

          {attachment.consultationId && (
            <div className="mt-4">
              <DataTable
                columns={[
                  { key: 'id', title: 'Link ID' },
                  { key: 'attachmentId', title: 'Attachment ID' },
                  { key: 'createdBy', title: 'Linked By' },
                ]}
                rows={attachmentListQuery.data || []}
                empty="No evidence links"
              />
            </div>
          )}
        </Card>
      )}
    </div>
  );
}
