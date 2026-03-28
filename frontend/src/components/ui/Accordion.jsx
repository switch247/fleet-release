import React, { useState } from 'react';
import { ChevronDown } from 'lucide-react';
import { cn } from '../../lib/utils';

export default function Accordion({ title, children }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="rounded-xl border border-slate-800 bg-slate-950/40">
      <button className="flex w-full items-center justify-between px-4 py-3 text-left" onClick={() => setOpen((prev) => !prev)}>
        <span className="text-sm font-medium text-slate-100">{title}</span>
        <ChevronDown className={cn('h-4 w-4 text-slate-400 transition', open && 'rotate-180')} />
      </button>
      {open && <div className="border-t border-slate-800 px-4 py-3 text-xs text-slate-300">{children}</div>}
    </div>
  );
}
