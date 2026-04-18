import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import Button from '../../../../src/components/ui/Button';

describe('Button', () => {
  test('renders children', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument();
  });

  test('default variant has cyan background', () => {
    render(<Button>Default</Button>);
    expect(screen.getByRole('button').className).toContain('bg-cyan-500');
  });

  test('outline variant has border styling', () => {
    render(<Button variant="outline">Outline</Button>);
    expect(screen.getByRole('button').className).toContain('border-slate-700');
  });

  test('danger variant has rose background', () => {
    render(<Button variant="danger">Danger</Button>);
    expect(screen.getByRole('button').className).toContain('bg-rose-600');
  });

  test('calls onClick handler when clicked', () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Click</Button>);
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  test('renders as disabled', () => {
    render(<Button disabled>Disabled</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });

  test('disabled class applied when disabled', () => {
    render(<Button disabled>Disabled</Button>);
    expect(screen.getByRole('button').className).toContain('disabled:opacity-50');
  });

  test('accepts extra className and merges it', () => {
    render(<Button className="extra-class">X</Button>);
    expect(screen.getByRole('button').className).toContain('extra-class');
  });

  test('passes through type prop', () => {
    render(<Button type="submit">Submit</Button>);
    expect(screen.getByRole('button')).toHaveAttribute('type', 'submit');
  });

  test('passes through aria-label', () => {
    render(<Button aria-label="close dialog">X</Button>);
    expect(screen.getByRole('button', { name: 'close dialog' })).toBeInTheDocument();
  });
});
