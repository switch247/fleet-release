import React from 'react';
import { cn } from '../../lib/utils';

export default function Input(props) {
  const { type, onChange } = props;

  // For native date/datetime-local inputs some browsers keep the picker
  // open after selection when the input remains focused. To provide a
  // smoother UX we forward the onChange and then blur the input so the
  // picker closes. Use setTimeout to let the original handler run first.
  const handleChange = (e) => {
    if (onChange) onChange(e);
    if (type === 'date' || type === 'datetime-local' || type === 'time') {
      // schedule blur after event handlers complete
      setTimeout(() => {
        try {
          e.target.blur();
        } catch (err) {
          // ignore
        }
      }, 0);
    }
  };

  return (
    <input
      {...props}
      onChange={handleChange}
      className={cn(
        'w-full rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-cyan-500 focus:outline-none',
        props.className
      )}
    />
  );
}
