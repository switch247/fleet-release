import React from 'react';
import { cn } from '../../lib/utils';

export default function Button({ className, variant = 'default', ...props }) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center rounded-xl px-4 py-2 text-sm font-medium transition disabled:opacity-50 disabled:cursor-not-allowed',
        variant === 'default' && 'bg-cyan-500 text-slate-950 hover:bg-cyan-400',
        variant === 'outline' && 'border border-slate-700 bg-slate-900/80 text-slate-100 hover:bg-slate-800',
        variant === 'danger' && 'bg-rose-600 text-white hover:bg-rose-500',
        className
      )}
      {...props}
    />
  );
}
