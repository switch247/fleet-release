import React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { useAuth } from './auth/AuthProvider';
import LoginPage from './pages/LoginPage';
import AppShell from './components/layout/AppShell';
import OverviewPage from './pages/OverviewPage';
import BookingsPage from './pages/BookingsPage';
import InspectionsPage from './pages/InspectionsPage';
import ProfilePage from './pages/ProfilePage';
import AdminUsersPage from './pages/AdminUsersPage';
import AdminCatalogPage from './pages/AdminCatalogPage';
import ComplaintsPage from './pages/ComplaintsPage';
import ConsultationsPage from './pages/ConsultationsPage';
import RatingsPage from './pages/RatingsPage';
import NotificationsPage from './pages/NotificationsPage';
import AdminNotificationsPage from './pages/AdminNotificationsPage';

function Guard({ children, roles }) {
  const { user, loading } = useAuth();
  if (loading) {
    return <div className="min-h-screen grid place-items-center text-slate-200">Loading session...</div>;
  }
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  if (roles && roles.length > 0 && !roles.some((role) => user.roles.includes(role))) {
    return <Navigate to="/overview" replace />;
  }
  return children;
}

export default function App() {
  const { user } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={user ? <Navigate to="/overview" replace /> : <LoginPage />} />
      <Route
        path="/"
        element={
          <Guard>
            <AppShell />
          </Guard>
        }
      >
        <Route index element={<Navigate to="/overview" replace />} />
        <Route path="overview" element={<OverviewPage />} />
        <Route path="bookings" element={<BookingsPage />} />
        <Route path="inspections" element={<InspectionsPage />} />
        <Route path="complaints" element={<ComplaintsPage />} />
        <Route path="consultations" element={<ConsultationsPage />} />
        <Route path="ratings" element={<RatingsPage />} />
        <Route path="notifications" element={<NotificationsPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route
          path="admin/users"
          element={
            <Guard roles={['admin']}>
              <AdminUsersPage />
            </Guard>
          }
        />
        <Route
          path="admin/catalog"
          element={
            <Guard roles={['admin']}>
              <AdminCatalogPage />
            </Guard>
          }
        />
        <Route
          path="admin/notifications"
          element={
            <Guard roles={['admin']}>
              <AdminNotificationsPage />
            </Guard>
          }
        />
      </Route>
      <Route path="*" element={<Navigate to={user ? '/overview' : '/login'} replace />} />
    </Routes>
  );
}
