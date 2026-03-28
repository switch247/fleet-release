import React, { useMemo, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import {
  adminCreateNotificationTemplate,
  adminNotificationTemplates,
  adminRetryNotifications,
  adminSendNotification,
  adminUsers,
  adminWorkerMetrics,
} from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Input from '../components/ui/Input';
import Button from '../components/ui/Button';
import DataTable from '../components/ui/DataTable';

export default function AdminNotificationsPage() {
  const templatesQuery = useQuery({ queryKey: ['admin-notification-templates'], queryFn: adminNotificationTemplates });
  const usersQuery = useQuery({ queryKey: ['admin-notification-users'], queryFn: adminUsers });
  const metricsQuery = useQuery({ queryKey: ['admin-worker-metrics'], queryFn: adminWorkerMetrics, refetchInterval: 5000 });

  const [templateForm, setTemplateForm] = useState({ name: '', title: '', body: '', channel: 'in_app' });
  const [sendForm, setSendForm] = useState({ templateId: '', userId: '' });

  const createTemplate = useMutation({
    mutationFn: adminCreateNotificationTemplate,
    onSuccess: () => {
      templatesQuery.refetch();
      setTemplateForm({ name: '', title: '', body: '', channel: 'in_app' });
    },
  });

  const sendMutation = useMutation({ mutationFn: adminSendNotification });
  const retryMutation = useMutation({
    mutationFn: adminRetryNotifications,
    onSuccess: () => metricsQuery.refetch(),
  });

  const templatePreview = useMemo(() => `${templateForm.title || 'Template title preview'}\n\n${templateForm.body || 'Template body preview'}`, [templateForm]);

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Notification Worker Metrics</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-4 text-sm">
          <div className="rounded-xl border border-slate-800 p-3">Processed: {metricsQuery.data?.processed ?? 0}</div>
          <div className="rounded-xl border border-slate-800 p-3">Delivered: {metricsQuery.data?.delivered ?? 0}</div>
          <div className="rounded-xl border border-slate-800 p-3">Dead Lettered: {metricsQuery.data?.deadLettered ?? 0}</div>
          <div className="rounded-xl border border-slate-800 p-3">Backlog: {metricsQuery.data?.currentBacklog ?? 0}</div>
        </div>
        <Button className="mt-3" onClick={() => retryMutation.mutate()} disabled={retryMutation.isPending}>
          Retry Queue Now
        </Button>
      </Card>

      <Card>
        <CardTitle>Create Template + Preview</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-2">
          <Input placeholder="Template name" value={templateForm.name} onChange={(e) => setTemplateForm((prev) => ({ ...prev, name: e.target.value }))} />
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={templateForm.channel} onChange={(e) => setTemplateForm((prev) => ({ ...prev, channel: e.target.value }))}>
            <option value="in_app">In-app</option>
            <option value="email">Email (disabled offline)</option>
            <option value="sms">SMS (disabled offline)</option>
          </select>
          <Input placeholder="Title" value={templateForm.title} onChange={(e) => setTemplateForm((prev) => ({ ...prev, title: e.target.value }))} />
          <Input placeholder="Body" value={templateForm.body} onChange={(e) => setTemplateForm((prev) => ({ ...prev, body: e.target.value }))} />
        </div>
        <div className="mt-3 rounded-xl border border-dashed border-slate-700 bg-slate-900/60 p-3 whitespace-pre-wrap text-sm text-slate-200">
          {templatePreview}
        </div>
        <Button className="mt-3" onClick={() => createTemplate.mutate(templateForm)}>
          Save Template
        </Button>
      </Card>

      <Card>
        <CardTitle>Send Notification</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-2">
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={sendForm.templateId} onChange={(e) => setSendForm((prev) => ({ ...prev, templateId: e.target.value }))}>
            <option value="">Select template</option>
            {(templatesQuery.data || []).map((template) => <option key={template.id} value={template.id}>{template.name}</option>)}
          </select>
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={sendForm.userId} onChange={(e) => setSendForm((prev) => ({ ...prev, userId: e.target.value }))}>
            <option value="">Select user</option>
            {(usersQuery.data || []).map((user) => <option key={user.id} value={user.id}>{user.username}</option>)}
          </select>
        </div>
        <Button className="mt-3" onClick={() => sendMutation.mutate(sendForm)} disabled={!sendForm.templateId || !sendForm.userId}>
          Send
        </Button>
      </Card>

      <Card>
        <CardTitle>Template Catalog</CardTitle>
        <DataTable
          columns={[
            { key: 'name', title: 'Name' },
            { key: 'channel', title: 'Channel' },
            { key: 'enabled', title: 'Enabled', render: (row) => (row.enabled ? 'Yes' : 'No') },
            { key: 'title', title: 'Title' },
          ]}
          rows={templatesQuery.data || []}
          empty="No templates"
        />
      </Card>
    </div>
  );
}
