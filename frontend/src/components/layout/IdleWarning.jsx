import React from 'react';
import { useAuth } from '../../auth/AuthProvider';
import Button from '../ui/Button';

export default function IdleWarning() {
  const { warningOpen, secondsLeft, staySignedIn, logout } = useAuth();
  if (!warningOpen) return null;
  return (
    <div className="fixed bottom-4 right-4 z-40 w-80 rounded-xl border border-amber-500/30 bg-amber-950/90 p-4 text-amber-100 shadow-xl">
      <p className="text-sm font-semibold">Session timeout warning</p>
      <p className="mt-1 text-xs">You will be logged out in {secondsLeft}s due to inactivity.</p>
      <div className="mt-3 flex gap-2">
        <Button className="flex-1" onClick={staySignedIn}>Stay Signed In</Button>
        <Button className="flex-1" variant="outline" onClick={logout}>Logout</Button>
      </div>
    </div>
  );
}
