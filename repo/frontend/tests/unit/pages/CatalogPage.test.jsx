import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../src/lib/api', () => ({
  categories: vi.fn(),
  listings: vi.fn(),
}));

import { categories, listings } from '../../../src/lib/api';
import CatalogPage from '../../../src/pages/CatalogPage';

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <CatalogPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

const sampleCategories = [
  { id: 'c1', name: 'Sedans', children: [] },
  { id: 'c2', name: 'SUVs', children: [{ id: 'c3', name: 'Compact SUV', children: [] }] },
];

const sampleListings = [
  { id: 'l1', name: 'Toyota Camry', sku: 'CAM-001', spu: 'TOY-CAM', categoryId: 'c1', available: true, deposit: 200, includedMiles: 100 },
  { id: 'l2', name: 'Honda CR-V', sku: 'CRV-001', spu: 'HON-CRV', categoryId: 'c2', available: false, deposit: 250, includedMiles: 150 },
];

beforeEach(() => {
  categories.mockResolvedValue(sampleCategories);
  listings.mockResolvedValue(sampleListings);
});

describe('CatalogPage', () => {
  test('renders Catalog heading', () => {
    renderPage();
    expect(screen.getByText('Catalog')).toBeInTheDocument();
  });

  test('renders Category Tree card', () => {
    renderPage();
    expect(screen.getByText('Category Tree')).toBeInTheDocument();
  });

  test('renders Listings card', () => {
    renderPage();
    expect(screen.getByText('Listings')).toBeInTheDocument();
  });

  test('shows category names after load', async () => {
    renderPage();
    expect(await screen.findByText('Sedans')).toBeInTheDocument();
    expect(await screen.findByText('SUVs')).toBeInTheDocument();
  });

  test('shows nested child category', async () => {
    renderPage();
    expect(await screen.findByText('Compact SUV')).toBeInTheDocument();
  });

  test('shows listing names after load', async () => {
    renderPage();
    expect(await screen.findByText(/Toyota Camry/)).toBeInTheDocument();
    expect(await screen.findByText(/Honda CR-V/)).toBeInTheDocument();
  });

  test('shows availability status for listings', async () => {
    renderPage();
    await screen.findByText(/Toyota Camry/);
    expect(screen.getByText(/Available/)).toBeInTheDocument();
    expect(screen.getByText(/Unavailable/)).toBeInTheDocument();
  });

  test('shows All Categories button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /all categories/i })).toBeInTheDocument();
  });

  test('has Compare button to toggle compare mode', async () => {
    renderPage();
    const compareBtn = screen.getByRole('button', { name: /compare/i });
    expect(compareBtn).toBeInTheDocument();
  });

  test('toggling Compare changes card title to Variant Comparison', async () => {
    renderPage();
    await screen.findByText(/Toyota Camry/);
    fireEvent.click(screen.getByRole('button', { name: /compare/i }));
    expect(screen.getByText('Variant Comparison')).toBeInTheDocument();
  });

  test('compare mode shows SKU column headers', async () => {
    renderPage();
    await screen.findByText(/Toyota Camry/);
    fireEvent.click(screen.getByRole('button', { name: /compare/i }));
    // SKU and Price (Deposit) headers appear once per SPU group
    const skuHeaders = screen.getAllByText('SKU');
    expect(skuHeaders.length).toBeGreaterThanOrEqual(1);
    const priceHeaders = screen.getAllByText('Price (Deposit)');
    expect(priceHeaders.length).toBeGreaterThanOrEqual(1);
  });

  test('clicking a category button filters listings', async () => {
    renderPage();
    const sedansBtn = await screen.findByRole('button', { name: 'Sedans' });
    fireEvent.click(sedansBtn);
    // Only Sedans category listing (c1) should be visible
    await waitFor(() =>
      expect(screen.queryByText(/Honda CR-V/)).not.toBeInTheDocument()
    );
  });

  test('clicking All Categories resets filter', async () => {
    renderPage();
    const sedansBtn = await screen.findByRole('button', { name: 'Sedans' });
    fireEvent.click(sedansBtn);
    fireEvent.click(screen.getByRole('button', { name: /all categories/i }));
    await waitFor(() =>
      expect(screen.getByText(/Honda CR-V/)).toBeInTheDocument()
    );
  });

  test('shows no listings message when category has no matches', async () => {
    listings.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('No listings for the selected category.')).toBeInTheDocument()
    );
  });

  test('groups listings by SPU', async () => {
    renderPage();
    await screen.findByText(/Toyota Camry/);
    expect(screen.getByText(/TOY-CAM/)).toBeInTheDocument();
    expect(screen.getByText(/HON-CRV/)).toBeInTheDocument();
  });
});
