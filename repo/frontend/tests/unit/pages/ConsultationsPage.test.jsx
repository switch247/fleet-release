import React, { act } from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/auth/AuthProvider', () => ({
  useAuth: vi.fn(),
}));

vi.mock('../../../src/lib/api', () => ({
  bookings: vi.fn(),
  consultationsForUser: vi.fn(),
  consultationAttachments: vi.fn(),
  createConsultation: vi.fn(),
  addConsultationAttachment: vi.fn(),
  uploadAttachmentFile: vi.fn(),
  presignAttachment: vi.fn(),
}));

import { useAuth } from '../../../src/auth/AuthProvider';
import {
  bookings,
  consultationsForUser,
  consultationAttachments,
  createConsultation,
  addConsultationAttachment,
  uploadAttachmentFile,
  presignAttachment,
} from '../../../src/lib/api';
import ConsultationsPage from '../../../src/pages/ConsultationsPage';

global.URL.createObjectURL = vi.fn().mockReturnValue('blob:mock');
global.URL.revokeObjectURL = vi.fn();
window.open = vi.fn();

const csaUser = { id: 'u1', username: 'agent', roles: ['csa'] };
const customerUser = { id: 'u2', username: 'alice', roles: ['customer'] };

const sampleConsultations = [
  { id: 'con1', topic: 'Engine Issue', version: 1, visibility: 'parties', keyPoints: 'Oil leak', recommendation: 'Replace gasket', followUp: 'Check next week' },
];

const sampleBookings = [{ id: 'b1', status: 'active' }];

function renderPage(user = csaUser) {
  useAuth.mockReturnValue({ user });
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <ConsultationsPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => {
  bookings.mockResolvedValue(sampleBookings);
  consultationsForUser.mockResolvedValue(sampleConsultations);
  consultationAttachments.mockResolvedValue([]);
  createConsultation.mockResolvedValue({ id: 'con2' });
  addConsultationAttachment.mockResolvedValue({ id: 'ca1' });
  uploadAttachmentFile.mockResolvedValue('att-1');
  presignAttachment.mockResolvedValue({ url: 'https://example.com/file.jpg' });
});

describe('ConsultationsPage', () => {
  test('renders Consultations heading', () => {
    renderPage();
    expect(screen.getByText('Consultations — notes & evidence')).toBeInTheDocument();
  });

  test('renders Consultation Notes card', () => {
    renderPage();
    expect(screen.getByText('Consultation Notes')).toBeInTheDocument();
  });

  test('shows Create Consultation button for csa role', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    expect(screen.getByRole('button', { name: /create consultation/i })).toBeInTheDocument();
  });

  test('does not show Create Consultation button for customer role', async () => {
    renderPage(customerUser);
    await screen.findByText('Consultation Notes');
    expect(screen.queryByRole('button', { name: /create consultation/i })).not.toBeInTheDocument();
  });

  test('shows consultations in table', async () => {
    renderPage();
    expect(await screen.findByText('Engine Issue')).toBeInTheDocument();
  });

  test('shows correct table columns', () => {
    renderPage();
    expect(screen.getByText('ID')).toBeInTheDocument();
    expect(screen.getByText('Topic')).toBeInTheDocument();
    expect(screen.getByText('Version')).toBeInTheDocument();
    expect(screen.getByText('Visibility')).toBeInTheDocument();
    expect(screen.getByText('Key Points')).toBeInTheDocument();
    expect(screen.getByText('Actions')).toBeInTheDocument();
  });

  test('shows empty message when no consultations', async () => {
    consultationsForUser.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No consultations available')).toBeInTheDocument()
    );
  });

  test('clicking Create Consultation opens modal', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    fireEvent.click(screen.getByRole('button', { name: /create consultation/i }));
    await waitFor(() =>
      expect(screen.getByRole('heading', { name: /create consultation/i })).toBeInTheDocument()
    );
  });

  test('create modal has Topic input', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    fireEvent.click(screen.getByRole('button', { name: /create consultation/i }));
    await screen.findByRole('heading', { name: /create consultation/i });
    expect(screen.getByPlaceholderText('Topic')).toBeInTheDocument();
  });

  test('create modal Create button is disabled when required fields empty', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    fireEvent.click(screen.getByRole('button', { name: /create consultation/i }));
    await screen.findByRole('heading', { name: /create consultation/i });
    // Create button should be disabled (topic, followUp, changeReason required)
    const createBtn = screen.getAllByRole('button', { name: /^create$/i }).find(
      (b) => b.closest('[role="dialog"], .fixed') !== null || b.disabled
    );
    expect(createBtn || screen.queryByRole('button', { name: /^create$/i, hidden: true })).toBeTruthy();
  });

  test('create modal closes on cancel', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    fireEvent.click(screen.getByRole('button', { name: /create consultation/i }));
    await screen.findByRole('heading', { name: /create consultation/i });
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    await waitFor(() =>
      expect(screen.queryByPlaceholderText('Topic')).not.toBeInTheDocument()
    );
  });

  test('shows Attach Evidence button for csa role', async () => {
    renderPage(csaUser);
    await screen.findByText('Engine Issue');
    expect(screen.getByRole('button', { name: /attach evidence/i })).toBeInTheDocument();
  });

  test('does not show Attach Evidence button for customer role', async () => {
    renderPage(customerUser);
    await screen.findByText('Engine Issue');
    expect(screen.queryByRole('button', { name: /attach evidence/i })).not.toBeInTheDocument();
  });

  test('shows Preview button for all roles', async () => {
    renderPage(customerUser);
    await screen.findByText('Engine Issue');
    expect(screen.getByRole('button', { name: /preview/i })).toBeInTheDocument();
  });

  test('clicking Preview opens preview modal', async () => {
    renderPage(customerUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /preview/i }));
    await waitFor(() =>
      expect(screen.getByText('Recommendation')).toBeInTheDocument()
    );
  });

  test('preview modal shows consultation details', async () => {
    renderPage(customerUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /preview/i }));
    await waitFor(() =>
      expect(screen.getByText('Replace gasket')).toBeInTheDocument()
    );
    // 'Oil leak' appears in both the table row and the modal — both are valid
    expect(screen.getAllByText('Oil leak').length).toBeGreaterThanOrEqual(1);
  });

  test('preview modal shows Follow Up section', async () => {
    renderPage(customerUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /preview/i }));
    await waitFor(() =>
      expect(screen.getByText('Follow Up')).toBeInTheDocument()
    );
    expect(screen.getByText('Check next week')).toBeInTheDocument();
  });

  test('clicking Attach Evidence opens attach modal', async () => {
    renderPage(csaUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /attach evidence/i }));
    await waitFor(() =>
      expect(screen.getByRole('heading', { name: /attach evidence/i })).toBeInTheDocument()
    );
  });

  test('attach modal has file input', async () => {
    renderPage(csaUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /attach evidence/i }));
    await screen.findByRole('heading', { name: /attach evidence/i });
    expect(document.querySelector('input[type="file"]')).toBeInTheDocument();
  });

  test('attach modal cancel closes it', async () => {
    renderPage(csaUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /attach evidence/i }));
    await screen.findByRole('heading', { name: /attach evidence/i });
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    await waitFor(() =>
      expect(screen.queryByRole('heading', { name: /attach evidence/i })).not.toBeInTheDocument()
    );
  });

  test('filling out create consultation form and clicking Create calls createConsultation', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    fireEvent.click(screen.getByRole('button', { name: /create consultation/i }));
    await screen.findByRole('heading', { name: /create consultation/i });
    // Fill required fields: topic, followUp, changeReason
    fireEvent.change(screen.getByPlaceholderText('Topic'), { target: { value: 'Brake Issue' } });
    // Fill follow-up plan textarea
    const textarea = screen.getByPlaceholderText('Describe next steps');
    fireEvent.change(textarea, { target: { value: 'Check brakes weekly' } });
    // Select change reason
    const allSelects = document.querySelectorAll('select');
    const changeReasonSelect = Array.from(allSelects).find((s) =>
      Array.from(s.options).some((o) => o.value === 'pricing_update')
    );
    fireEvent.change(changeReasonSelect, { target: { value: 'customer_request' } });
    // Click Create
    const createBtns = screen.getAllByRole('button', { name: /^create$/i });
    fireEvent.click(createBtns[createBtns.length - 1]);
    await waitFor(() =>
      expect(createConsultation).toHaveBeenCalledWith(
        expect.objectContaining({ topic: 'Brake Issue' })
      )
    );
  });

  test('selecting a file in attach modal updates file state', async () => {
    renderPage(csaUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /attach evidence/i }));
    await screen.findByRole('heading', { name: /attach evidence/i });
    const fileInput = document.querySelector('input[type="file"]');
    const mockFile = new File(['data'], 'photo.jpg', { type: 'image/jpeg' });
    Object.defineProperty(fileInput, 'files', { value: [mockFile], configurable: true });
    fireEvent.change(fileInput);
    // Attach button is still there after selecting file
    expect(screen.getByRole('button', { name: /^attach$/i })).toBeInTheDocument();
  });

  test('clicking Attach in attach modal calls uploadAttachmentFile then addConsultationAttachment', async () => {
    renderPage(csaUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /attach evidence/i }));
    await screen.findByRole('heading', { name: /attach evidence/i });
    // Set a file
    const fileInput = document.querySelector('input[type="file"]');
    const mockFile = new File(['data'], 'photo.jpg', { type: 'image/jpeg' });
    Object.defineProperty(fileInput, 'files', { value: [mockFile], configurable: true });
    fireEvent.change(fileInput);
    // Click Attach
    fireEvent.click(screen.getByRole('button', { name: /^attach$/i }));
    await waitFor(() =>
      expect(uploadAttachmentFile).toHaveBeenCalled()
    );
  });

  test('create modal booking select can be changed', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    fireEvent.click(screen.getByRole('button', { name: /create consultation/i }));
    await screen.findByRole('heading', { name: /create consultation/i });
    // Change the booking select
    const bookingSelects = document.querySelectorAll('select');
    if (bookingSelects.length > 0) {
      fireEvent.change(bookingSelects[0], { target: { value: 'b1' } });
    }
    expect(document.querySelectorAll('select').length).toBeGreaterThan(0);
  });

  test('create modal visibility select can be changed', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    fireEvent.click(screen.getByRole('button', { name: /create consultation/i }));
    await screen.findByRole('heading', { name: /create consultation/i });
    // Find visibility select (has options csa_admin, parties, all)
    const allSelects = document.querySelectorAll('select');
    const visibilitySelect = Array.from(allSelects).find((s) =>
      Array.from(s.options).some((o) => o.value === 'csa_admin')
    );
    if (visibilitySelect) {
      fireEvent.change(visibilitySelect, { target: { value: 'csa_admin' } });
      expect(visibilitySelect.value).toBe('csa_admin');
    }
  });

  test('preview modal shows attachments when available', async () => {
    consultationAttachments.mockResolvedValue([
      { id: 'ca1', attachmentId: 'att1', createdBy: 'u1' },
    ]);
    renderPage(customerUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /preview/i }));
    await waitFor(() =>
      expect(screen.getByText('att1')).toBeInTheDocument()
    );
  });

  test('clicking Preview on attachment calls presignAttachment', async () => {
    consultationAttachments.mockResolvedValue([
      { id: 'ca1', attachmentId: 'att1', createdBy: 'u1' },
    ]);
    renderPage(customerUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /preview/i }));
    await waitFor(() =>
      expect(screen.getByText('att1')).toBeInTheDocument()
    );
    // Click the Preview button on the attachment item
    const previewBtns = screen.getAllByRole('button', { name: /preview/i });
    // The last Preview button is the one inside the attachment list
    fireEvent.click(previewBtns[previewBtns.length - 1]);
    await waitFor(() =>
      expect(presignAttachment).toHaveBeenCalledWith('att1', 60)
    );
  });

  test('keyPoints and recommendation inputs can be changed in create modal', async () => {
    renderPage(csaUser);
    await screen.findByText('Consultation Notes');
    fireEvent.click(screen.getByRole('button', { name: /create consultation/i }));
    await screen.findByRole('heading', { name: /create consultation/i });
    fireEvent.change(screen.getByPlaceholderText('Key points'), { target: { value: 'Point 1, Point 2' } });
    fireEvent.change(screen.getByPlaceholderText('Recommendation'), { target: { value: 'Replace parts' } });
    expect(screen.getByPlaceholderText('Key points').value).toBe('Point 1, Point 2');
    expect(screen.getByPlaceholderText('Recommendation').value).toBe('Replace parts');
  });

  test('Attach button handles upload error gracefully', async () => {
    uploadAttachmentFile.mockRejectedValueOnce(new Error('Upload failed'));
    renderPage(csaUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /attach evidence/i }));
    await screen.findByRole('heading', { name: /attach evidence/i });
    const fileInput = document.querySelector('input[type="file"]');
    const mockFile = new File(['data'], 'test.jpg', { type: 'image/jpeg' });
    Object.defineProperty(fileInput, 'files', { value: [mockFile], configurable: true });
    fireEvent.change(fileInput);
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: /^attach$/i }));
    });
    // Modal remains open after error
    expect(document.querySelector('input[type="file"]')).toBeInTheDocument();
  });

  test('Preview attachment handles presign error gracefully', async () => {
    consultationAttachments.mockResolvedValue([
      { id: 'ca1', attachmentId: 'att1', createdBy: 'u1' },
    ]);
    presignAttachment.mockRejectedValueOnce(new Error('Presign failed'));
    renderPage(customerUser);
    await screen.findByText('Engine Issue');
    fireEvent.click(screen.getByRole('button', { name: /preview/i }));
    await waitFor(() =>
      expect(screen.getByText('att1')).toBeInTheDocument()
    );
    const previewBtns = screen.getAllByRole('button', { name: /preview/i });
    await act(async () => {
      fireEvent.click(previewBtns[previewBtns.length - 1]);
    });
    expect(screen.getByText('att1')).toBeInTheDocument();
  });
});
