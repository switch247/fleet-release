import React from 'react';
import { render, screen } from '@testing-library/react';
import { Card, CardTitle } from '../../../../src/components/ui/Card';

describe('Card', () => {
  test('renders children', () => {
    render(<Card>Card content</Card>);
    expect(screen.getByText('Card content')).toBeInTheDocument();
  });

  test('renders as a section element', () => {
    const { container } = render(<Card>x</Card>);
    expect(container.firstChild.tagName.toLowerCase()).toBe('section');
  });

  test('has rounded-2xl class', () => {
    const { container } = render(<Card>x</Card>);
    expect(container.firstChild.className).toContain('rounded-2xl');
  });

  test('accepts and merges custom className', () => {
    const { container } = render(<Card className="custom-card">x</Card>);
    expect(container.firstChild.className).toContain('custom-card');
    expect(container.firstChild.className).toContain('rounded-2xl');
  });

  test('renders multiple children', () => {
    render(
      <Card>
        <span>First</span>
        <span>Second</span>
      </Card>
    );
    expect(screen.getByText('First')).toBeInTheDocument();
    expect(screen.getByText('Second')).toBeInTheDocument();
  });
});

describe('CardTitle', () => {
  test('renders children', () => {
    render(<CardTitle>My Title</CardTitle>);
    expect(screen.getByText('My Title')).toBeInTheDocument();
  });

  test('renders as an h3 element', () => {
    render(<CardTitle>Title</CardTitle>);
    expect(screen.getByRole('heading', { level: 3 })).toBeInTheDocument();
  });

  test('has font-semibold styling', () => {
    render(<CardTitle>Title</CardTitle>);
    expect(screen.getByRole('heading').className).toContain('font-semibold');
  });
});
