import React from 'react';
import { render, screen } from '@testing-library/react';
import Badge from '../../../../src/components/ui/Badge';

describe('Badge', () => {
  test('renders children', () => {
    render(<Badge>Active</Badge>);
    expect(screen.getByText('Active')).toBeInTheDocument();
  });

  test('neutral variant (default) has slate styling', () => {
    render(<Badge>Neutral</Badge>);
    expect(screen.getByText('Neutral').className).toContain('bg-slate-700');
  });

  test('success variant has emerald styling', () => {
    render(<Badge variant="success">OK</Badge>);
    expect(screen.getByText('OK').className).toContain('bg-emerald-500');
  });

  test('warning variant has amber styling', () => {
    render(<Badge variant="warning">Warn</Badge>);
    expect(screen.getByText('Warn').className).toContain('bg-amber-500');
  });

  test('danger variant has rose styling', () => {
    render(<Badge variant="danger">Error</Badge>);
    expect(screen.getByText('Error').className).toContain('bg-rose-500');
  });

  test('renders as a span element', () => {
    render(<Badge>Tag</Badge>);
    expect(screen.getByText('Tag').tagName.toLowerCase()).toBe('span');
  });

  test('always has rounded-full class', () => {
    render(<Badge variant="success">Rounded</Badge>);
    expect(screen.getByText('Rounded').className).toContain('rounded-full');
  });
});
