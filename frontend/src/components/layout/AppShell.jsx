import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { BarChart3, BellRing, CalendarClock, ClipboardCheck, MessageSquare, Shield, Star, User, FileText } from 'lucide-react';
import { useAuth } from '../../auth/AuthProvider';
import Button from '../ui/Button';
import IdleWarning from './IdleWarning';

function RoleNav({ roles }) {
  const items = [
    { to: '/overview', label: 'Overview', icon: BarChart3, show: true },
    { to: '/bookings', label: 'Bookings', icon: CalendarClock, show: roles.includes('customer') || roles.includes('provider') || roles.includes('admin') },
    { to: '/inspections', label: 'Inspections', icon: ClipboardCheck, show: roles.includes('provider') || roles.includes('csa') || roles.includes('admin') },
    { to: '/complaints', label: 'Complaints', icon: MessageSquare, show: roles.includes('customer') || roles.includes('provider') || roles.includes('csa') || roles.includes('admin') },
    { to: '/consultations', label: 'Consultations', icon: FileText, show: roles.includes('customer') || roles.includes('provider') || roles.includes('csa') || roles.includes('admin') },
    { to: '/ratings', label: 'Ratings', icon: Star, show: roles.includes('customer') || roles.includes('provider') || roles.includes('admin') },
    { to: '/notifications', label: 'Inbox', icon: BellRing, show: true },
    { to: '/profile', label: 'Profile', icon: User, show: true },
    { to: '/admin/users', label: 'Admin Users', icon: Shield, show: roles.includes('admin') },
    { to: '/admin/catalog', label: 'Admin Catalog', icon: Shield, show: roles.includes('admin') },
    { to: '/admin/notifications', label: 'Admin Notify', icon: BellRing, show: roles.includes('admin') },
  ];

  return (
    <nav className="space-y-2">
      {items.filter((item) => item.show).map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          className={({ isActive }) =>
            `flex items-center gap-2 rounded-xl px-3 py-2 text-sm ${isActive ? 'bg-cyan-500 text-slate-950' : 'text-slate-200 hover:bg-slate-800'}`
          }
        >
          <item.icon className="h-4 w-4" />
          {item.label}
        </NavLink>
      ))}
    </nav>
  );
}

export default function AppShell() {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_right,#1f2937,#020617_45%)] text-slate-100">
      <div className="grid min-h-screen grid-cols-1 md:grid-cols-[260px,1fr]">
        <aside className="border-r border-slate-800 bg-slate-950/70 p-5">
          <p className="text-xs uppercase tracking-[0.3em] text-cyan-300">FleetLease</p>
          <h1 className="mt-2 text-2xl font-semibold">Operations Suite</h1>
          <p className="mt-1 text-xs text-slate-400">{user?.username} · {user?.roles?.join(', ')}</p>
          <div className="mt-6">
            <RoleNav roles={user?.roles || []} />
          </div>
          <Button variant="outline" className="mt-6 w-full" onClick={logout}>Logout</Button>
        </aside>
        <main className="p-5 md:p-8">
          <Outlet />
        </main>
      </div>
      <IdleWarning />
    </div>
  );
}
