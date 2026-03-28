import React from 'react';
import { cn } from '../../lib/utils';

export default function Input(props) {
  return (
    <input
      {...props}
      className={cn(
        'w-full rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-cyan-500 focus:outline-none',
        props.className
      )}
    />
  );
}
