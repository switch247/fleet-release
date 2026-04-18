import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import Input from '../../../../src/components/ui/Input';

describe('Input', () => {
  test('renders an input element', () => {
    render(<Input placeholder="Enter text" />);
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
  });

  test('calls onChange handler', () => {
    const onChange = vi.fn();
    render(<Input onChange={onChange} />);
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'hello' } });
    expect(onChange).toHaveBeenCalledTimes(1);
  });

  test('displays the current value', () => {
    render(<Input value="initial" onChange={vi.fn()} />);
    expect(screen.getByDisplayValue('initial')).toBeInTheDocument();
  });

  test('has rounded-xl border styling', () => {
    render(<Input />);
    const input = screen.getByRole('textbox');
    expect(input.className).toContain('rounded-xl');
    expect(input.className).toContain('border-slate-700');
  });

  test('accepts type="password"', () => {
    const { container } = render(<Input type="password" />);
    expect(container.querySelector('input[type="password"]')).toBeInTheDocument();
  });

  test('accepts custom className and merges with base', () => {
    render(<Input className="my-custom" />);
    expect(screen.getByRole('textbox').className).toContain('my-custom');
    expect(screen.getByRole('textbox').className).toContain('rounded-xl');
  });

  test('can be disabled', () => {
    render(<Input disabled />);
    expect(screen.getByRole('textbox')).toBeDisabled();
  });

  test('accepts placeholder text', () => {
    render(<Input placeholder="Search..." />);
    expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument();
  });

  test('handles date type input', () => {
    const onChange = vi.fn();
    const { container } = render(<Input type="date" onChange={onChange} />);
    const input = container.querySelector('input[type="date"]');
    expect(input).toBeInTheDocument();
    // blur is scheduled via setTimeout for date inputs; just verify onChange fires
    fireEvent.change(input, { target: { value: '2024-01-01' } });
    expect(onChange).toHaveBeenCalled();
  });

  test('handles datetime-local type input', () => {
    const onChange = vi.fn();
    const { container } = render(<Input type="datetime-local" onChange={onChange} />);
    const input = container.querySelector('input[type="datetime-local"]');
    fireEvent.change(input, { target: { value: '2024-01-01T10:00' } });
    expect(onChange).toHaveBeenCalled();
  });
});
