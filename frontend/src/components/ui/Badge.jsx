import React from 'react';
import { cn } from '../../lib/utils';

export default function Badge({ children, variant = 'neutral' }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2 py-1 text-xs font-medium',
        variant === 'success' && 'bg-emerald-500/20 text-emerald-300',
        variant === 'warning' && 'bg-amber-500/20 text-amber-200',
        variant === 'danger' && 'bg-rose-500/20 text-rose-200',
        variant === 'neutral' && 'bg-slate-700/40 text-slate-200'
      )}
    >
      {children}
    </span>
  );
}
