import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import Modal from '../../../../src/components/ui/Modal';

describe('Modal', () => {
  test('does not render when open=false', () => {
    render(<Modal open={false} title="Test" onClose={vi.fn()}>Content</Modal>);
    expect(screen.queryByText('Content')).not.toBeInTheDocument();
    expect(screen.queryByText('Test')).not.toBeInTheDocument();
  });

  test('renders children when open=true', () => {
    render(<Modal open title="My Modal" onClose={vi.fn()}>Modal content</Modal>);
    expect(screen.getByText('Modal content')).toBeInTheDocument();
  });

  test('displays the title', () => {
    render(<Modal open title="Important Title" onClose={vi.fn()}>x</Modal>);
    expect(screen.getByText('Important Title')).toBeInTheDocument();
  });

  test('calls onClose when Close button is clicked', () => {
    const onClose = vi.fn();
    render(<Modal open title="T" onClose={onClose}>x</Modal>);
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  test('renders footer when provided', () => {
    render(
      <Modal open title="T" onClose={vi.fn()} footer={<button>Save</button>}>
        content
      </Modal>
    );
    expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument();
  });

  test('no footer section when footer prop is omitted', () => {
    const { container } = render(<Modal open title="T" onClose={vi.fn()}>content</Modal>);
    // Only Close button should be present
    expect(screen.getAllByRole('button')).toHaveLength(1);
  });

  test('renders multiple children', () => {
    render(
      <Modal open title="T" onClose={vi.fn()}>
        <p>Line A</p>
        <p>Line B</p>
      </Modal>
    );
    expect(screen.getByText('Line A')).toBeInTheDocument();
    expect(screen.getByText('Line B')).toBeInTheDocument();
  });
});
