import React from 'react';
import { render, screen } from '@testing-library/react';
import DataTable from '../../../../src/components/ui/DataTable';

const columns = [
  { key: 'name', title: 'Name' },
  { key: 'status', title: 'Status' },
  { key: 'amount', title: 'Amount', render: (row) => `$${row.amount}` },
];

const rows = [
  { id: '1', name: 'Alice', status: 'active', amount: 100 },
  { id: '2', name: 'Bob', status: 'pending', amount: 200 },
];

describe('DataTable', () => {
  test('renders column headers', () => {
    render(<DataTable columns={columns} rows={[]} />);
    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
    expect(screen.getByText('Amount')).toBeInTheDocument();
  });

  test('renders row data', () => {
    render(<DataTable columns={columns} rows={rows} />);
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
    expect(screen.getByText('active')).toBeInTheDocument();
    expect(screen.getByText('pending')).toBeInTheDocument();
  });

  test('uses render function for custom cells', () => {
    render(<DataTable columns={columns} rows={rows} />);
    expect(screen.getByText('$100')).toBeInTheDocument();
    expect(screen.getByText('$200')).toBeInTheDocument();
  });

  test('shows custom empty message when no rows', () => {
    render(<DataTable columns={columns} rows={[]} empty="Nothing here" />);
    expect(screen.getByText('Nothing here')).toBeInTheDocument();
  });

  test('shows default empty message when not provided', () => {
    render(<DataTable columns={columns} rows={[]} />);
    expect(screen.getByText('No data')).toBeInTheDocument();
  });

  test('empty cell spans all columns', () => {
    render(<DataTable columns={columns} rows={[]} />);
    const cell = screen.getByText('No data').closest('td');
    expect(cell).toHaveAttribute('colspan', '3');
  });

  test('renders correct number of table cells', () => {
    render(<DataTable columns={columns} rows={rows} />);
    // 3 columns × 2 rows = 6 cells
    const cells = document.querySelectorAll('tbody td');
    expect(cells).toHaveLength(6);
  });

  test('uses row.id as key when available', () => {
    const { container } = render(<DataTable columns={columns} rows={rows} />);
    expect(container.querySelectorAll('tbody tr')).toHaveLength(2);
  });

  test('falls back to index key when row has no id', () => {
    const rowsWithoutId = [{ name: 'X', status: 'a', amount: 1 }];
    const { container } = render(<DataTable columns={columns} rows={rowsWithoutId} />);
    expect(container.querySelectorAll('tbody tr')).toHaveLength(1);
  });
});
