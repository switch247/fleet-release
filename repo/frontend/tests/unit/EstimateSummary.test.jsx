import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import EstimateSummary from '../../src/components/EstimateSummary';

test('renders each line of the estimate breakdown', () => {
  const estimate = {
    baseAmount: 12.5,
    mileageAmount: 3.75,
    timeAmount: 4.25,
    nightSurcharge: 1.5,
    deposit: 45,
    total: 67,
  };
  render(<EstimateSummary estimate={estimate} />);
  expect(screen.getByText('Pre-trip estimate breakdown')).toBeInTheDocument();
  expect(screen.getByText('Base fare')).toBeInTheDocument();
  expect(screen.getByText('Mileage')).toBeInTheDocument();
  expect(screen.getByText('Time')).toBeInTheDocument();
  expect(screen.getByText('Night surcharge')).toBeInTheDocument();
  expect(screen.getByText('Deposit')).toBeInTheDocument();
  expect(screen.getByText('$67.00')).toBeInTheDocument();
});
