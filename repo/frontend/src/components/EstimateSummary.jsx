import React from 'react';

const labelClasses = 'text-xs uppercase tracking-wide text-slate-400';

export default function EstimateSummary({ estimate }) {
  if (!estimate) {
    return null;
  }

  const format = (value) => `$${Number(value || 0).toFixed(2)}`;
  const formatMiles = (value) => `${Number(value || 0).toFixed(0)} mi`;

  return (
    <div className="mt-3 rounded-xl border border-slate-700 bg-slate-900/70 p-3 text-sm space-y-1">
      <p className="font-semibold text-slate-100">Pre-trip estimate breakdown</p>
      <div className="space-y-0.5">
        <div className="flex justify-between">
          <span className={labelClasses}>Included Miles</span>
          <span className="text-slate-300">{formatMiles(estimate.includedMiles)}</span>
        </div>
        <div className="flex justify-between">
          <span className={labelClasses}>Base fare</span>
          <span className="text-slate-300">{format(estimate.baseAmount)}</span>
        </div>
        <div className="flex justify-between">
          <span className={labelClasses}>Mileage</span>
          <span className="text-slate-300">{format(estimate.mileageAmount)}</span>
        </div>
        <div className="flex justify-between">
          <span className={labelClasses}>Time</span>
          <span className="text-slate-300">{format(estimate.timeAmount)}</span>
        </div>
        <div className="flex justify-between">
          <span className={labelClasses}>Night surcharge</span>
          <span className="text-slate-300">{format(estimate.nightSurcharge)}</span>
        </div>
        <div className="flex justify-between">
          <span className={labelClasses}>Deposit</span>
          <span className="text-slate-300">{format(estimate.deposit)}</span>
        </div>
      </div>
      <p className="text-xs text-slate-400 mt-2">Night surcharge window applies: 10:00 PM - 5:59 AM.</p>
      <div className="pt-2 border-t border-slate-800 flex justify-between text-cyan-300 font-semibold">
        <span>Total</span>
        <span>{format(estimate.total)}</span>
      </div>
    </div>
  );
}
