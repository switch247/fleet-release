import React from 'react';
import { cn } from '../../lib/utils';

export function Card({ className, children }) {
  return <section className={cn('rounded-2xl border border-slate-800 bg-slate-900/70 p-5', className)}>{children}</section>;
}

export function CardTitle({ children }) {
  return <h3 className="text-lg font-semibold text-slate-100">{children}</h3>;
}
