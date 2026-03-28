import React, { useEffect, useState } from 'react';
import { format } from 'date-fns';
import { useAuth } from '../auth/AuthProvider';
import { loginHistory, updateMe } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Input from '../components/ui/Input';
import Button from '../components/ui/Button';
import DataTable from '../components/ui/DataTable';

export default function ProfilePage() {
  const { user, refreshMe } = useAuth();
  const [email, setEmail] = useState(user?.email || '');
  const [historyRows, setHistoryRows] = useState([]);
  const [status, setStatus] = useState('');

  useEffect(() => {
    setEmail(user?.email || '');
  }, [user]);

  useEffect(() => {
    loginHistory().then(setHistoryRows).catch(() => setHistoryRows([]));
  }, []);

  const save = async () => {
    try {
      await updateMe({ email });
      await refreshMe();
      setStatus('Profile updated.');
    } catch (err) {
      setStatus(err.message);
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Profile</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-2">
          <Input value={user?.username || ''} disabled />
          <Input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email" />
        </div>
        <div className="mt-3 flex items-center gap-3">
          <Button onClick={save}>Save Email</Button>
          {status && <p className="text-sm text-slate-300">{status}</p>}
        </div>
      </Card>

      <Card>
        <CardTitle>Login History (Audit Log)</CardTitle>
        <div className="mt-4">
          <DataTable
            columns={[
              { key: 'eventType', title: 'Event' },
              { key: 'ip', title: 'IP Address' },
              { key: 'createdAt', title: 'Time', render: (row) => format(new Date(row.createdAt), 'yyyy-MM-dd HH:mm:ss') },
            ]}
            rows={historyRows}
            empty="No login history available"
          />
        </div>
      </Card>
    </div>
  );
}
