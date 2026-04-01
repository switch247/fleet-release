import React, { useMemo, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import {
  attachmentChunk,
  attachmentComplete,
  attachmentInit,
  bookings,
  closeSettlement,
  submitInspection,
  listInspections,
} from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import Badge from '../components/ui/Badge';
import DataTable from '../components/ui/DataTable';
import Modal from '../components/ui/Modal';

const BASE_ITEMS = [
  { key: 'exterior', name: 'Exterior bodywork' },
  { key: 'tires', name: 'Tires and wheels' },
  { key: 'interior', name: 'Interior and dashboard' },
  { key: 'fuel', name: 'Fuel and fluids' },
];

function toHex(buffer) {
  return Array.from(new Uint8Array(buffer)).map((b) => b.toString(16).padStart(2, '0')).join('');
}

async function fileChecksum(file) {
  const bytes = await file.arrayBuffer();
  const digest = await crypto.subtle.digest('SHA-256', bytes);
  return { checksum: toHex(digest), bytes };
}

export default function InspectionsPage() {
  const bookingsQuery = useQuery({ queryKey: ['inspection-bookings'], queryFn: bookings });
  const [bookingID, setBookingID] = useState('');
  const [stage, setStage] = useState('handover');
  const [notes, setNotes] = useState('');
  const [items, setItems] = useState(() => BASE_ITEMS.map((item) => ({ ...item, condition: 'good', file: null })));
  const [settlement, setSettlement] = useState(null);
  const [status, setStatus] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [step, setStep] = useState(1); // 1=setup,2=checklist,3=review

  const submitMutation = useMutation({
    mutationFn: async () => {
      if (!bookingID) throw new Error('Select a booking first');
      const payloadItems = [];
      for (const item of items) {
        if (!item.file) throw new Error(`Evidence file required for ${item.name}`);
        const { checksum, bytes } = await fileChecksum(item.file);
        const fingerprint = `${bookingID}:${item.key}:${checksum}`;
        const init = await attachmentInit({
          bookingId: bookingID,
          type: item.file.type.startsWith('video') ? 'video' : 'photo',
          sizeBytes: item.file.size,
          checksum,
          fingerprint,
        });
        const uploadId = init.uploadId || init.attachment?.id;
        if (!init.deduplicated) {
          const chunkBase64 = btoa(String.fromCharCode(...new Uint8Array(bytes)));
          await attachmentChunk({ uploadId, chunkBase64 });
          await attachmentComplete({ uploadId });
        }
        payloadItems.push({ name: item.name, condition: item.condition, evidenceIds: [uploadId] });
      }
      await submitInspection({ bookingId: bookingID, stage, items: payloadItems, notes });
    },
    onSuccess: () => setStatus('Inspection submitted successfully.'),
    onError: (error) => setStatus(error.message),
  });

  const settleMutation = useMutation({
    mutationFn: () => closeSettlement(bookingID),
    onSuccess: (data) => setSettlement(data),
    onError: (error) => setStatus(error.message),
  });

  const deductions = useMemo(() => {
    const pricing = { good: 0, minor: 20, major: 80 };
    const list = items.map((item) => ({ item: item.name, deduction: pricing[item.condition] || 0 })).filter((entry) => entry.deduction > 0);
    const total = list.reduce((sum, entry) => sum + entry.deduction, 0);
    return { list, total };
  }, [items]);

  const inspectionsQuery = useQuery({
    queryKey: ['inspections', bookingID],
    queryFn: () => listInspections(bookingID),
    enabled: !!bookingID,
  });

  return (
    <div className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold">Guided Inspection & Settlement</h2>
        <p className="text-sm text-slate-400">Camera evidence is mandatory per checklist item before submission.</p>
      </header>

      <div className="flex justify-between items-center">
        <div>
          {bookingID ? (
            <div className="text-sm">Selected booking: <strong>{bookingID}</strong></div>
          ) : (
            <div className="text-sm text-slate-400">No booking selected — choose from the list below</div>
          )}
        </div>
        <div>
          <Button onClick={() => { if (bookingID) { setModalOpen(true); setStep(1); } }} disabled={!bookingID}>Open Inspection Modal</Button>
        </div>
      </div>

      <Card>
        <CardTitle>Bookings</CardTitle>
        <div className="mt-3 space-y-2">
          {bookingsQuery.isLoading && <p className="text-sm text-slate-400">Loading bookings...</p>}
          {bookingsQuery.isError && <p className="text-sm text-red-400">Failed to load bookings</p>}
          {(bookingsQuery.data || []).map((b) => (
            <div key={b.id} className="rounded-lg border border-slate-800 p-3 flex items-center justify-between">
              <div>
                <div className="font-medium">{b.id} — {b.status}</div>
                <div className="text-xs text-slate-400">{new Date(b.startAt).toLocaleDateString()} → {new Date(b.endAt).toLocaleDateString()}</div>
              </div>
              <div className="flex items-center gap-2">
                <Button size="sm" onClick={() => { setBookingID(b.id); setModalOpen(true); setStep(1); }}>Start Inspection</Button>
                <Button size="sm" variant="outline" onClick={() => setBookingID(b.id)}>View Inspections</Button>
                {bookingID === b.id && <Button size="sm" variant="ghost" onClick={() => setBookingID('')}>Clear</Button>}
              </div>
            </div>
          ))}
          {((bookingsQuery.data || []).length === 0) && <p className="text-sm text-slate-300">No bookings available.</p>}
        </div>
      </Card>

      {/* Completed inspections for selected booking */}
      {bookingID && (
        <Card>
          <CardTitle>Completed Inspections</CardTitle>
          <div className="mt-3 space-y-2">
            {inspectionsQuery.isLoading && <p className="text-sm text-slate-400">Loading...</p>}
            {inspectionsQuery.isError && <p className="text-sm text-red-400">Failed to load inspections</p>}
            {inspectionsQuery.data && inspectionsQuery.data.length === 0 && <p className="text-sm text-slate-300">No inspections recorded for this booking.</p>}
            {inspectionsQuery.data?.map((rev) => (
              <div key={rev.revisionId} className="rounded-lg border border-slate-800 p-3">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-medium">{rev.stage} — {new Date(rev.createdAt).toLocaleString()}</div>
                    <div className="text-xs text-slate-400">Revision: {rev.revisionId}</div>
                  </div>
                  <div>
                    <span className="px-2 py-1 rounded bg-slate-700 text-white">Recorded</span>
                  </div>
                </div>
                <div className="mt-2 space-y-1">
                  {rev.items.map((it, i) => (
                    <div key={i} className="text-sm">
                      <div className="font-medium">{it.name}</div>
                      <div className="text-xs text-slate-400">Evidence: {it.evidenceIds.join(', ')}</div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Settlement button */}
      {bookingID && inspectionsQuery.data && inspectionsQuery.data.length > 0 && !settlement && (
        <Card>
          <CardTitle>Close Settlement</CardTitle>
          <p className="text-sm text-slate-400 mt-2">Finalize the trip with charge adjustments and deposit refund/deduction.</p>
          <Button className="mt-3" onClick={() => settleMutation.mutate()}>Settle Trip</Button>
        </Card>
      )}

      {/* Multi-step modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="Inspection Wizard" footer={(
        <div className="flex items-center gap-2">
          {step > 1 && <Button variant="outline" onClick={() => setStep((s) => s - 1)}>Back</Button>}
          {step < 3 && <Button onClick={() => setStep((s) => s + 1)}>Next</Button>}
          {step === 3 && <Button onClick={() => submitMutation.mutate()}>Submit Inspection</Button>}
        </div>
      )}>
        {step === 1 && (
          <div className="space-y-3">
            <div className="grid gap-3 md:grid-cols-2">
              <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={stage} onChange={(e) => setStage(e.target.value)}>
                <option value="handover">Handover</option>
                <option value="return">Return</option>
              </select>
              <input className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" placeholder="Notes" value={notes} onChange={(e) => setNotes(e.target.value)} />
            </div>
            <p className="text-sm text-slate-400">Step 1 of 3 — Setup inspection basics. Booking is chosen before opening this modal.</p>
          </div>
        )}

        {step === 2 && (
          <div className="space-y-3">
            <div className="space-y-3">
              {items.map((item, index) => (
                <div key={item.key} className="grid gap-2 rounded-xl border border-slate-800 p-3 md:grid-cols-[1.2fr,0.7fr,1fr]">
                  <p className="text-sm">{item.name}</p>
                  <select className="rounded-lg border border-slate-700 bg-slate-900 px-2 py-1 text-sm" value={item.condition} onChange={(e) => setItems((prev) => prev.map((row, i) => i === index ? { ...row, condition: e.target.value } : row))}>
                    <option value="good">Good</option>
                    <option value="minor">Minor Wear</option>
                    <option value="major">Major Damage</option>
                  </select>
                  <input
                    type="file"
                    accept="image/*,video/*"
                    capture="environment"
                    className="text-xs"
                    onChange={(e) => setItems((prev) => prev.map((row, i) => i === index ? { ...row, file: e.target.files?.[0] || null } : row))}
                  />
                </div>
              ))}
            </div>
            <p className="text-sm text-slate-400">Step 2 of 3 — Capture evidence for each checklist item.</p>
          </div>
        )}

        {step === 3 && (
          <div className="space-y-3">
            <CardTitle>Review & Proposed Deductions</CardTitle>
            <div className="mt-2 space-y-2">
              {deductions.list.map((entry) => (
                <div key={entry.item} className="flex items-center justify-between rounded-lg border border-slate-800 px-3 py-2">
                  <span>{entry.item}</span>
                  <Badge variant="warning">-${entry.deduction.toFixed(2)}</Badge>
                </div>
              ))}
              <p className="text-sm text-slate-300">Proposed deductions total: ${deductions.total.toFixed(2)}</p>
            </div>
            <p className="text-sm text-slate-400">Step 3 of 3 — submit when ready.</p>
          </div>
        )}
      </Modal>

      {settlement && (
        <Card>
          <CardTitle>Settlement Summary</CardTitle>
          <div className="mt-2 space-y-2">
            {settlement.entries.map((entry, i) => (
              <div key={i} className="flex justify-between text-sm">
                <span>{entry.description}</span>
                <span>${entry.amount.toFixed(2)}</span>
              </div>
            ))}
            <div className="border-t border-slate-800 pt-2 flex justify-between font-semibold">
              <span>Total</span>
              <span>${settlement.entries.reduce((sum, e) => sum + e.amount, 0).toFixed(2)}</span>
            </div>
          </div>
        </Card>
      )}

      {status && <p className="text-sm text-cyan-300">{status}</p>}
    </div>
  );
}
