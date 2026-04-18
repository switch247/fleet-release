import React from 'react';
import { render, screen } from '@testing-library/react';
import EstimateSummary from '../../src/components/EstimateSummary';

const baseEstimate = {
  baseAmount: 12.5,
  mileageAmount: 3.75,
  timeAmount: 4.25,
  nightSurcharge: 1.5,
  deposit: 45,
  total: 67,
  includedMiles: 100,
};

describe('EstimateSummary', () => {
  test('renders each line of the estimate breakdown', () => {
    render(<EstimateSummary estimate={baseEstimate} />);
    expect(screen.getByText('Pre-trip estimate breakdown')).toBeInTheDocument();
    expect(screen.getByText('Base fare')).toBeInTheDocument();
    expect(screen.getByText('Mileage')).toBeInTheDocument();
    expect(screen.getByText('Time')).toBeInTheDocument();
    expect(screen.getByText('Night surcharge')).toBeInTheDocument();
    expect(screen.getByText('Deposit')).toBeInTheDocument();
    expect(screen.getByText('$67.00')).toBeInTheDocument();
  });

  test('renders null when estimate prop is null', () => {
    const { container } = render(<EstimateSummary estimate={null} />);
    expect(container.firstChild).toBeNull();
  });

  test('renders null when estimate prop is undefined', () => {
    const { container } = render(<EstimateSummary />);
    expect(container.firstChild).toBeNull();
  });

  test('shows included miles', () => {
    render(<EstimateSummary estimate={baseEstimate} />);
    expect(screen.getByText('100 mi')).toBeInTheDocument();
  });

  test('formats monetary values to 2 decimal places', () => {
    render(<EstimateSummary estimate={baseEstimate} />);
    expect(screen.getByText('$12.50')).toBeInTheDocument();
    expect(screen.getByText('$3.75')).toBeInTheDocument();
    expect(screen.getByText('$4.25')).toBeInTheDocument();
    expect(screen.getByText('$45.00')).toBeInTheDocument();
  });

  test('does not show coupon discount row when couponDiscountAmount is 0', () => {
    render(<EstimateSummary estimate={{ ...baseEstimate, couponDiscountAmount: 0 }} />);
    expect(screen.queryByText('Coupon discount')).not.toBeInTheDocument();
  });

  test('shows coupon discount row when couponDiscountAmount > 0', () => {
    render(<EstimateSummary estimate={{ ...baseEstimate, couponDiscountAmount: 10.5 }} />);
    expect(screen.getByText('Coupon discount')).toBeInTheDocument();
    expect(screen.getByText('-$10.50')).toBeInTheDocument();
  });

  test('shows night surcharge window note', () => {
    render(<EstimateSummary estimate={baseEstimate} />);
    expect(screen.getByText(/10:00 PM - 5:59 AM/)).toBeInTheDocument();
  });

  test('shows Total label', () => {
    render(<EstimateSummary estimate={baseEstimate} />);
    expect(screen.getByText('Total')).toBeInTheDocument();
  });

  test('handles zero values gracefully', () => {
    const zeroEstimate = {
      baseAmount: 0,
      mileageAmount: 0,
      timeAmount: 0,
      nightSurcharge: 0,
      deposit: 0,
      total: 0,
      includedMiles: 0,
    };
    render(<EstimateSummary estimate={zeroEstimate} />);
    // Should show $0.00 for all values
    const zeros = screen.getAllByText('$0.00');
    expect(zeros.length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('0 mi')).toBeInTheDocument();
  });

  test('handles missing optional fields', () => {
    render(<EstimateSummary estimate={{ total: 50 }} />);
    // Should not crash, should render $50.00 for total
    expect(screen.getByText('$50.00')).toBeInTheDocument();
  });
});
