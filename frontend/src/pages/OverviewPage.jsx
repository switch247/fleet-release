import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { BarChart, Bar, ResponsiveContainer, XAxis, YAxis, Tooltip } from 'recharts';
import { statsSummary, bookings } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';

function Stat({ label, value }) {
  return (
    <Card className="p-4">
      <p className="text-xs uppercase tracking-[0.2em] text-slate-400">{label}</p>
      <p className="mt-2 text-3xl font-semibold text-cyan-300">{value}</p>
    </Card>
  );
}

export default function OverviewPage() {
  const summaryQuery = useQuery({ queryKey: ['stats-summary'], queryFn: statsSummary });
  const bookingsQuery = useQuery({ queryKey: ['bookings-overview'], queryFn: bookings });

  const chartData = (bookingsQuery.data || []).reduce((acc, booking) => {
    const key = booking.status || 'unknown';
    acc[key] = (acc[key] || 0) + 1;
    return acc;
  }, {});

  const chartRows = Object.entries(chartData).map(([status, count]) => ({ status, count }));

  return (
    <div className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold">Operational Overview</h2>
        <p className="text-sm text-slate-400">Live metrics from /api/v1/stats/summary</p>
      </header>

      <div className="grid gap-4 md:grid-cols-4">
        <Stat label="Active Bookings" value={summaryQuery.data?.activeBookings ?? '-'} />
        <Stat label="Settled Trips" value={summaryQuery.data?.settledTrips ?? '-'} />
        <Stat label="Inspections Due" value={summaryQuery.data?.inspectionsDue ?? '-'} />
        <Stat label="Held Deposits" value={`$${(summaryQuery.data?.heldDeposits ?? 0).toFixed(2)}`} />
      </div>

      <Card>
        <CardTitle>Booking Status Distribution</CardTitle>
        <div className="mt-4 h-72">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartRows}>
              <XAxis dataKey="status" stroke="#94a3b8" />
              <YAxis stroke="#94a3b8" allowDecimals={false} />
              <Tooltip />
              <Bar dataKey="count" fill="#22d3ee" radius={[6, 6, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Card>
    </div>
  );
}
