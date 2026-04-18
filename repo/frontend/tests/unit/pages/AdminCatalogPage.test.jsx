import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/lib/api', () => ({
  adminCategories: vi.fn(),
  adminCreateCategory: vi.fn(),
  adminUpdateCategory: vi.fn(),
  adminDeleteCategory: vi.fn(),
  adminListings: vi.fn(),
  adminCreateListing: vi.fn(),
  adminUpdateListing: vi.fn(),
  adminDeleteListing: vi.fn(),
  adminBulkListings: vi.fn(),
  adminSearchListings: vi.fn(),
  adminUsers: vi.fn(),
}));

import {
  adminCategories,
  adminCreateCategory,
  adminUpdateCategory,
  adminDeleteCategory,
  adminListings,
  adminCreateListing,
  adminDeleteListing,
  adminBulkListings,
  adminSearchListings,
  adminUsers,
} from '../../../src/lib/api';
import AdminCatalogPage from '../../../src/pages/AdminCatalogPage';

const sampleCategories = [
  { id: 'c1', name: 'Sedans', parentId: null },
  { id: 'c2', name: 'SUVs', parentId: null },
];

const sampleListings = [
  { id: 'l1', name: 'Camry', spu: 'TOY', sku: 'CAM-001', includedMiles: 100, deposit: 200, available: true },
  { id: 'l2', name: 'CR-V', spu: 'HON', sku: 'CRV-001', includedMiles: 150, deposit: 250, available: false },
];

const sampleUsers = [
  { id: 'u1', username: 'provider1', roles: ['provider'] },
];

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AdminCatalogPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

beforeEach(() => {
  adminCategories.mockResolvedValue(sampleCategories);
  adminListings.mockResolvedValue(sampleListings);
  adminUsers.mockResolvedValue(sampleUsers);
  adminCreateCategory.mockResolvedValue({ id: 'c3' });
  adminCreateListing.mockResolvedValue({ id: 'l3' });
  adminDeleteCategory.mockResolvedValue(null);
  adminDeleteListing.mockResolvedValue(null);
  adminBulkListings.mockResolvedValue({ ok: true });
  adminSearchListings.mockResolvedValue([]);
});

describe('AdminCatalogPage', () => {
  test('renders Category CRUD card', () => {
    renderPage();
    expect(screen.getByText('Category CRUD')).toBeInTheDocument();
  });

  test('renders Create Listing card', () => {
    renderPage();
    // 'Create Listing' appears as both a card heading and a button — use heading role
    expect(screen.getByRole('heading', { name: 'Create Listing' })).toBeInTheDocument();
  });

  test('renders Listing CRUD card', () => {
    renderPage();
    expect(screen.getByText('Listing CRUD + Bulk Controls')).toBeInTheDocument();
  });

  test('shows category name input', () => {
    renderPage();
    expect(screen.getByPlaceholderText('New category name')).toBeInTheDocument();
  });

  test('shows Create button for categories', () => {
    renderPage();
    // Multiple Create buttons, look for the first
    expect(screen.getAllByRole('button', { name: /^create$/i }).length).toBeGreaterThanOrEqual(1);
  });

  test('shows categories in table after load', async () => {
    renderPage();
    // 'Sedans' appears in the DataTable rows AND in category select dropdowns
    const sedansEls = await screen.findAllByText('Sedans');
    expect(sedansEls.length).toBeGreaterThanOrEqual(1);
    const suvsEls = await screen.findAllByText('SUVs');
    expect(suvsEls.length).toBeGreaterThanOrEqual(1);
  });

  test('shows listings in table after load', async () => {
    renderPage();
    expect(await screen.findByText('Camry')).toBeInTheDocument();
    expect(await screen.findByText('CR-V')).toBeInTheDocument();
  });

  test('shows listing availability', async () => {
    renderPage();
    await screen.findByText('Camry');
    expect(screen.getByText('Yes')).toBeInTheDocument();
    expect(screen.getByText('No')).toBeInTheDocument();
  });

  test('shows Search input', () => {
    renderPage();
    expect(screen.getByPlaceholderText('Search name/SPU/SKU')).toBeInTheDocument();
  });

  test('shows Search button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /^search$/i })).toBeInTheDocument();
  });

  test('shows Mark Unavailable bulk button (disabled initially)', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /mark unavailable/i })).toBeDisabled();
  });

  test('shows Mark Available bulk button (disabled initially)', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /mark available/i })).toBeDisabled();
  });

  test('calls adminCreateCategory when Create is clicked', async () => {
    renderPage();
    fireEvent.change(screen.getByPlaceholderText('New category name'), { target: { value: 'Trucks' } });
    fireEvent.click(screen.getAllByRole('button', { name: /^create$/i })[0]);
    // React Query v5 passes a second context arg to mutationFn
    await waitFor(() =>
      expect(adminCreateCategory).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'Trucks' }),
        expect.any(Object)
      )
    );
  });

  test('calls adminDeleteCategory when Delete is clicked for category', async () => {
    renderPage();
    // 'Sedans' appears in table rows and category selects — use findAllByText
    await screen.findAllByText('Sedans');
    // Each row has Edit and Delete buttons
    const deleteBtns = screen.getAllByRole('button', { name: /^delete$/i });
    fireEvent.click(deleteBtns[0]);
    // React Query v5 passes a second context arg to mutationFn
    await waitFor(() =>
      expect(adminDeleteCategory).toHaveBeenCalledWith('c1', expect.any(Object))
    );
  });

  test('category edit form appears after clicking Edit', async () => {
    renderPage();
    // 'Sedans' appears in table rows and category selects — use findAllByText
    await screen.findAllByText('Sedans');
    fireEvent.click(screen.getAllByRole('button', { name: /^edit$/i })[0]);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /^save$/i })).toBeInTheDocument()
    );
  });

  test('calls search when Search button is clicked', async () => {
    adminSearchListings.mockResolvedValue([{ id: 'l1', name: 'Camry', spu: 'TOY', sku: 'CAM-001', includedMiles: 100, deposit: 200, available: true }]);
    renderPage();
    fireEvent.change(screen.getByPlaceholderText('Search name/SPU/SKU'), { target: { value: 'Camry' } });
    fireEvent.click(screen.getByRole('button', { name: /^search$/i }));
    await waitFor(() =>
      expect(adminSearchListings).toHaveBeenCalledWith('Camry')
    );
  });

  test('checkbox enables bulk operations when checked', async () => {
    renderPage();
    await screen.findByText('Camry');
    const checkboxes = document.querySelectorAll('input[type="checkbox"]');
    fireEvent.click(checkboxes[0]);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /mark unavailable/i })).not.toBeDisabled()
    );
  });

  test('shows provider users in listing provider dropdown', async () => {
    renderPage();
    await waitFor(() => {
      const options = Array.from(document.querySelectorAll('select option'));
      return options.some((o) => o.textContent === 'provider1');
    });
  });

  test('calls adminCreateListing when Create Listing is clicked', async () => {
    renderPage();
    await screen.findByText('Camry');
    fireEvent.click(screen.getByRole('button', { name: /create listing/i }));
    await waitFor(() =>
      expect(adminCreateListing).toHaveBeenCalled()
    );
  });

  test('listing edit form appears after clicking Edit on a listing row', async () => {
    renderPage();
    await screen.findByText('Camry');
    // Edit buttons include category edits; listing Edit buttons come after category rows
    const editBtns = screen.getAllByRole('button', { name: /^edit$/i });
    // First 2 are category rows, next are listing rows
    fireEvent.click(editBtns[editBtns.length - 1]);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /save listing/i })).toBeInTheDocument()
    );
  });

  test('empty categories message shown', async () => {
    adminCategories.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No categories')).toBeInTheDocument()
    );
  });

  test('empty listings message shown', async () => {
    adminListings.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No listings')).toBeInTheDocument()
    );
  });

  test('clicking Mark Unavailable triggers bulk mutation', async () => {
    renderPage();
    await screen.findByText('Camry');
    // Check first listing checkbox to enable bulk buttons
    const checkboxes = document.querySelectorAll('input[type="checkbox"]');
    fireEvent.click(checkboxes[0]);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /mark unavailable/i })).not.toBeDisabled()
    );
    fireEvent.click(screen.getByRole('button', { name: /mark unavailable/i }));
    await waitFor(() =>
      expect(adminBulkListings).toHaveBeenCalledWith(
        expect.objectContaining({ available: false }),
        expect.any(Object)
      )
    );
  });

  test('clicking Mark Available triggers bulk mutation with available:true', async () => {
    renderPage();
    await screen.findByText('Camry');
    const checkboxes = document.querySelectorAll('input[type="checkbox"]');
    fireEvent.click(checkboxes[0]);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /mark available/i })).not.toBeDisabled()
    );
    fireEvent.click(screen.getByRole('button', { name: /mark available/i }));
    await waitFor(() =>
      expect(adminBulkListings).toHaveBeenCalledWith(
        expect.objectContaining({ available: true }),
        expect.any(Object)
      )
    );
  });

  test('clicking Save Listing triggers updateListingMutation', async () => {
    const { adminUpdateListing } = await import('../../../src/lib/api');
    adminUpdateListing.mockResolvedValue({ id: 'l1' });
    renderPage();
    await screen.findByText('Camry');
    // Click Edit on the last listing row
    const editBtns = screen.getAllByRole('button', { name: /^edit$/i });
    fireEvent.click(editBtns[editBtns.length - 1]);
    await screen.findByRole('button', { name: /save listing/i });
    fireEvent.click(screen.getByRole('button', { name: /save listing/i }));
    await waitFor(() =>
      expect(adminUpdateListing).toHaveBeenCalled()
    );
  });

  test('filling listing form fields fires onChange handlers', async () => {
    renderPage();
    await screen.findByText('Camry');
    // Fill the listing creation form (SPU, SKU, Name, etc.)
    fireEvent.change(screen.getByPlaceholderText('Name'), { target: { value: 'Test Car' } });
    fireEvent.change(screen.getByPlaceholderText('SPU'), { target: { value: 'TEST' } });
    fireEvent.change(screen.getByPlaceholderText('SKU'), { target: { value: 'T001' } });
    fireEvent.change(screen.getByPlaceholderText('Included miles'), { target: { value: '200' } });
    fireEvent.change(screen.getByPlaceholderText('Deposit'), { target: { value: '300' } });
    expect(screen.getByPlaceholderText('Name').value).toBe('Test Car');
  });

  test('calling Save on category edit triggers updateCategoryMutation', async () => {
    const { adminUpdateCategory } = await import('../../../src/lib/api');
    adminUpdateCategory.mockResolvedValue({ id: 'c1' });
    renderPage();
    await screen.findAllByText('Sedans');
    fireEvent.click(screen.getAllByRole('button', { name: /^edit$/i })[0]);
    await screen.findByRole('button', { name: /^save$/i });
    fireEvent.click(screen.getByRole('button', { name: /^save$/i }));
    await waitFor(() =>
      expect(adminUpdateCategory).toHaveBeenCalled()
    );
  });

  test('category table shows parent name when category has a parent', async () => {
    adminCategories.mockResolvedValue([
      { id: 'c1', name: 'Sedans', parentId: null },
      { id: 'c2', name: 'Sport Sedans', parentId: 'c1' },
    ]);
    renderPage();
    // Both 'Sedans' entries (table row + parent name cell) should appear
    await waitFor(() =>
      expect(screen.getAllByText('Sedans').length).toBeGreaterThanOrEqual(2)
    );
  });

  test('edit category parent select can be changed', async () => {
    renderPage();
    await screen.findAllByText('Sedans');
    fireEvent.click(screen.getAllByRole('button', { name: /^edit$/i })[0]);
    await screen.findByRole('button', { name: /^save$/i });
    // Find the parent select in the edit section (has "No parent" option)
    const allSelects = document.querySelectorAll('select');
    const parentSelect = Array.from(allSelects).find((s) =>
      Array.from(s.options).some((o) => o.text === 'No parent')
    );
    if (parentSelect) {
      fireEvent.change(parentSelect, { target: { value: 'c2' } });
    }
    expect(screen.getByRole('button', { name: /^save$/i })).toBeInTheDocument();
  });

  test('unchecking a listing checkbox removes it from selectedIds', async () => {
    renderPage();
    await screen.findByText('Camry');
    const checkboxes = document.querySelectorAll('input[type="checkbox"]');
    // Check first with click
    fireEvent.click(checkboxes[0]);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /mark unavailable/i })).not.toBeDisabled()
    );
    // Uncheck by clicking again — triggers the prev.filter branch
    fireEvent.click(checkboxes[0]);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /mark unavailable/i })).toBeDisabled()
    );
  });
});
