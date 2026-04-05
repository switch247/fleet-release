import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  addConsultationAttachment,
  bookings,
  consultationAttachments,
  consultations,
  consultationsForUser,
  createConsultation,
  uploadAttachmentFile,
  presignAttachment,
} from '../lib/api';
import Modal from '../components/ui/Modal';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import DataTable from '../components/ui/DataTable';
import Badge from '../components/ui/Badge';
import { useAuth } from '../auth/AuthProvider';

const createEmptyForm = () => ({
  bookingId: '',
  topic: '',
  keyPoints: '',
  recommendation: '',
  followUp: '',
  visibility: 'parties',
  changeReason: '',
});

export default function ConsultationsPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const bookingsQuery = useQuery({ queryKey: ['consult-bookings'], queryFn: bookings });
  const [selectedBooking, setSelectedBooking] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [attachOpen, setAttachOpen] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [selectedConsultation, setSelectedConsultation] = useState(null);

  const consultationsQuery = useQuery({
    queryKey: ['consultations'],
    queryFn: () => consultationsForUser(),
  });

  const [form, setForm] = useState(createEmptyForm());
  const [selectedFile, setSelectedFile] = useState(null);

  const attachmentListQuery = useQuery({
    queryKey: ['consultation-attachments', selectedConsultation?.id],
    queryFn: () => consultationAttachments(selectedConsultation?.id || ''),
    enabled: Boolean(selectedConsultation?.id) && previewOpen,
  });

  const createMutation = useMutation({
    mutationFn: (payload) => createConsultation(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['consultations'] });
      setCreateOpen(false);
      setForm(createEmptyForm());
    },
  });

  const attachMutation = useMutation({
    mutationFn: addConsultationAttachment,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['consultation-attachments', selectedConsultation?.id] });
      setAttachOpen(false);
    },
  });

  const canManage = user?.roles?.includes('csa') || user?.roles?.includes('admin');

  const canCreateConsultation = Boolean(form.topic.trim() && form.followUp.trim() && form.changeReason.trim());

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Consultations — notes & evidence</CardTitle>
        <p className="text-sm text-slate-400 mt-2">Shows consultation notes you have access to. Use the actions column to manage notes and evidence.</p>
      </Card>

      <Card>
        <CardTitle>Consultation Notes</CardTitle>
        <div className="mt-4">
          {canManage && <div className="mb-3"><Button onClick={() => { setForm(createEmptyForm()); setCreateOpen(true); }}>Create Consultation</Button></div>}
          <DataTable
            columns={[
              { key: 'id', title: 'ID' },
              { key: 'topic', title: 'Topic' },
              { key: 'version', title: 'Version' },
              { key: 'visibility', title: 'Visibility' },
              { key: 'keyPoints', title: 'Key Points' },
              {
                key: 'actions',
                title: 'Actions',
                render: (row) => (
                  <div className="flex items-center gap-2">
                    {canManage && <Button size="sm" onClick={() => { setSelectedConsultation(row); setAttachOpen(true); }}>Attach Evidence</Button>}
                      <Button size="sm" variant="outline" onClick={() => { setSelectedConsultation(row); setPreviewOpen(true); }}>Preview</Button>
                  </div>
                ),
              },
            ]}
            rows={consultationsQuery.data || []}
            empty="No consultations available"
          />
        </div>
      </Card>

      {/* Duplicate consultation list removed: handled via Preview modal. */}
      
      {/* Create Consultation modal */}
      <Modal open={createOpen} onClose={() => { setCreateOpen(false); setForm(createEmptyForm()); }} title="Create Consultation" footer={(
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => { setCreateOpen(false); setForm(createEmptyForm()); }}>Cancel</Button>
          <Button onClick={() => createMutation.mutate({ bookingId: form.bookingId || selectedBooking, ...form })} disabled={!canCreateConsultation || createMutation.isPending}>Create</Button>
        </div>
      )}>
        <div className="space-y-3">
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={form.bookingId || selectedBooking} onChange={(e) => setForm((prev) => ({ ...prev, bookingId: e.target.value }))}>
            <option value="">Select booking (optional)</option>
            {(bookingsQuery.data || []).map((b) => <option key={b.id} value={b.id}>{b.id}</option>)}
          </select>
          <Input placeholder="Topic" value={form.topic} onChange={(e) => setForm((prev) => ({ ...prev, topic: e.target.value }))} />
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={form.visibility} onChange={(e) => setForm((prev) => ({ ...prev, visibility: e.target.value }))}>
            <option value="csa_admin">CSA/Admin</option>
            <option value="parties">Parties</option>
            <option value="all">All</option>
          </select>
          <Input placeholder="Key points" value={form.keyPoints} onChange={(e) => setForm((prev) => ({ ...prev, keyPoints: e.target.value }))} />
          <Input placeholder="Recommendation" value={form.recommendation} onChange={(e) => setForm((prev) => ({ ...prev, recommendation: e.target.value }))} />
          <label className="block text-sm font-medium text-slate-200 mb-1">Follow-up Plan</label>
          <textarea
            className="w-full rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-cyan-500 focus:outline-none"
            rows={3}
            placeholder="Describe next steps"
            value={form.followUp}
            onChange={(e) => setForm((prev) => ({ ...prev, followUp: e.target.value }))}
          />
          <label className="block text-sm font-medium text-slate-200 mb-1">Change Reason</label>
          <select
            className="w-full rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm"
            value={form.changeReason}
            onChange={(e) => setForm((prev) => ({ ...prev, changeReason: e.target.value }))}
          >
            <option value="">Select reason</option>
            <option value="pricing_update">Pricing update</option>
            <option value="customer_request">Customer request</option>
            <option value="inspection_feedback">Inspection feedback</option>
            <option value="other">Other</option>
          </select>
        </div>
      </Modal>

      {/* Attach Evidence modal */}
      <Modal open={attachOpen} onClose={() => { setAttachOpen(false); setSelectedFile(null); }} title={selectedConsultation ? `Attach Evidence to ${selectedConsultation.topic}` : 'Attach Evidence'} footer={(
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => { setAttachOpen(false); setSelectedFile(null); }}>Cancel</Button>
          <Button onClick={async () => {
            try {
              if (!selectedFile) {
                console.error('no file selected');
                return;
              }
              const bookingId = selectedConsultation?.bookingId || '';
              const attId = await uploadAttachmentFile(selectedFile, bookingId, selectedFile.type.startsWith('video') ? 'video' : 'photo');
              attachMutation.mutate({ consultationId: selectedConsultation?.id, attachmentId: attId });
            } catch (err) {
              console.error('upload failed', err);
            }
          }}>Attach</Button>
        </div>
      )}>
        <div className="space-y-3">
          <div>
            <label className="block text-sm text-slate-400">Choose file to upload</label>
            <input type="file" onChange={(e) => setSelectedFile(e.target.files?.[0] || null)} />
          </div>
          <div className="text-sm text-slate-500">File will be uploaded and linked to the consultation.</div>
        </div>
      </Modal>

      {/* Preview attachments modal */}
        {/* Preview modal shows consultation details and attachments */}
        <Modal open={previewOpen} onClose={() => setPreviewOpen(false)} title={selectedConsultation ? `Consultation — ${selectedConsultation.topic}` : 'Consultation'}>
          <div className="space-y-4">
            {selectedConsultation ? (
              <div className="space-y-2">
                <div className="text-sm text-slate-400">Visibility: <Badge>{selectedConsultation.visibility}</Badge></div>
                <div className="font-medium">Recommendation</div>
                <div className="text-sm text-slate-200">{selectedConsultation.recommendation || '—'}</div>
                <div className="font-medium">Key Points</div>
                <div className="text-sm text-slate-200">{selectedConsultation.keyPoints || '—'}</div>
                {selectedConsultation.followUp && <>
                  <div className="font-medium">Follow Up</div>
                  <div className="text-sm text-slate-200">{selectedConsultation.followUp}</div>
                </>}
              </div>
            ) : <p className="text-sm text-slate-400">No consultation selected</p>}

            <div>
              <div className="font-medium">Attachments</div>
              <div className="space-y-3 mt-2">
                {attachmentListQuery.isLoading && <p className="text-sm text-slate-400">Loading attachments...</p>}
                {attachmentListQuery.isError && <p className="text-sm text-red-400">Failed to load attachments</p>}
                {attachmentListQuery.data?.length === 0 && <p className="text-sm text-slate-300">No attachments linked</p>}
                {attachmentListQuery.data?.map((a) => (
                  <div key={a.id} className="rounded-lg border border-slate-800 p-3">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">{a.attachmentId}</div>
                        <div className="text-xs text-slate-400">Linked by: {a.createdBy}</div>
                      </div>
                      <div>
                        <Button size="sm" variant="outline" onClick={async () => {
                          try {
                            const resp = await presignAttachment(a.attachmentId, 60);
                            const url = resp.url || resp;
                            if (!url) throw new Error('no url returned');
                            window.open(url, '_blank');
                          } catch (err) {
                            console.error('failed to presign', err);
                          }
                        }}>Preview</Button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </Modal>
    </div>
  );
}
