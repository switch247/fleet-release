import React, { useMemo, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import {
  attachmentChunk,
  attachmentComplete,
  attachmentInit,
  bookings,
  closeSettlement,
  submitInspection,
} from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import Badge from '../components/ui/Badge';
import DataTable from '../components/ui/DataTable';

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

  return (
    <div className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold">Guided Inspection & Settlement</h2>
        <p className="text-sm text-slate-400">Camera evidence is mandatory per checklist item before submission.</p>
      </header>

      <Card>
        <CardTitle>Inspection Setup</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-3">
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={bookingID} onChange={(e) => setBookingID(e.target.value)}>
            <option value="">Select booking</option>
            {(bookingsQuery.data || []).map((booking) => <option key={booking.id} value={booking.id}>{booking.id} ({booking.status})</option>)}
          </select>
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={stage} onChange={(e) => setStage(e.target.value)}>
            <option value="handover">Handover</option>
            <option value="return">Return</option>
          </select>
          <input className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" placeholder="Notes" value={notes} onChange={(e) => setNotes(e.target.value)} />
        </div>
      </Card>

      <Card>
        <CardTitle>Checklist (Camera Capture Required)</CardTitle>
        <div className="mt-4 space-y-3">
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
        <Button className="mt-4" onClick={() => submitMutation.mutate()} disabled={submitMutation.isPending}>
          {submitMutation.isPending ? 'Submitting...' : 'Submit Inspection'}
        </Button>
      </Card>

      <Card>
        <CardTitle>Wear & Tear Assessment</CardTitle>
        <div className="mt-4 space-y-2">
          {deductions.list.map((entry) => (
            <div key={entry.item} className="flex items-center justify-between rounded-lg border border-slate-800 px-3 py-2">
              <span>{entry.item}</span>
              <Badge variant="warning">-${entry.deduction.toFixed(2)}</Badge>
            </div>
          ))}
          <p className="text-sm text-slate-300">Proposed deductions total: ${deductions.total.toFixed(2)}</p>
        </div>
      </Card>

      <Card>
        <CardTitle>One-Click Settlement Statement</CardTitle>
        <Button onClick={() => settleMutation.mutate()} disabled={!bookingID || settleMutation.isPending}>
          {settleMutation.isPending ? 'Closing Trip...' : 'Close Trip & Generate Statement'}
        </Button>
        {settlement?.ledger && (
          <div className="mt-4">
            <DataTable
              columns={[
                { key: 'type', title: 'Entry Type' },
                { key: 'amount', title: 'Amount', render: (row) => `$${Number(row.amount).toFixed(2)}` },
                { key: 'description', title: 'Description' },
              ]}
              rows={settlement.ledger}
              empty="No settlement entries"
            />
          </div>
        )}
      </Card>

      {status && <p className="text-sm text-cyan-300">{status}</p>}
    </div>
  );
}
