import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import Accordion from '../../../../src/components/ui/Accordion';

describe('Accordion', () => {
  test('renders the title', () => {
    render(<Accordion title="Section A">Content</Accordion>);
    expect(screen.getByText('Section A')).toBeInTheDocument();
  });

  test('content is hidden by default', () => {
    render(<Accordion title="Title">Secret content</Accordion>);
    expect(screen.queryByText('Secret content')).not.toBeInTheDocument();
  });

  test('clicking the button reveals content', () => {
    render(<Accordion title="Title">Revealed</Accordion>);
    fireEvent.click(screen.getByRole('button'));
    expect(screen.getByText('Revealed')).toBeInTheDocument();
  });

  test('clicking again hides content', () => {
    render(<Accordion title="Title">Toggle</Accordion>);
    const btn = screen.getByRole('button');
    fireEvent.click(btn);
    expect(screen.getByText('Toggle')).toBeInTheDocument();
    fireEvent.click(btn);
    expect(screen.queryByText('Toggle')).not.toBeInTheDocument();
  });

  test('chevron has rotate-180 class when open', () => {
    render(<Accordion title="T">content</Accordion>);
    fireEvent.click(screen.getByRole('button'));
    // The svg/icon container should have rotate-180
    expect(screen.getByRole('button').querySelector('svg')?.className.baseVal ?? '').toContain('rotate-180');
  });

  test('renders the title in a button', () => {
    render(<Accordion title="My Section">x</Accordion>);
    const btn = screen.getByRole('button');
    expect(btn).toContainElement(screen.getByText('My Section'));
  });
});
