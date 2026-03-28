import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { inboxNotifications } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import DataTable from '../components/ui/DataTable';
import Badge from '../components/ui/Badge';

export default function NotificationsPage() {
  const notificationsQuery = useQuery({ queryKey: ['inbox-notifications'], queryFn: inboxNotifications });

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Inbox & Delivery Status</CardTitle>
        <div className="mt-4">
          <DataTable
            columns={[
              { key: 'title', title: 'Title' },
              { key: 'body', title: 'Body' },
              { key: 'attempts', title: 'Attempts' },
              {
                key: 'status',
                title: 'Status',
                render: (row) => (
                  <Badge variant={row.status === 'delivered' ? 'success' : row.status === 'disabled_offline' ? 'warning' : 'neutral'}>
                    {row.status}
                  </Badge>
                ),
              },
            ]}
            rows={notificationsQuery.data || []}
            empty="No inbox notifications"
          />
        </div>
      </Card>
    </div>
  );
}
