import React, { useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import {
  adminBulkListings,
  adminCategories,
  adminCreateCategory,
  adminCreateListing,
  adminDeleteCategory,
  adminDeleteListing,
  adminListings,
  adminSearchListings,
  adminUpdateCategory,
  adminUpdateListing,
  adminUsers,
} from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Input from '../components/ui/Input';
import Button from '../components/ui/Button';
import DataTable from '../components/ui/DataTable';

export default function AdminCatalogPage() {
  const categoriesQuery = useQuery({ queryKey: ['admin-categories'], queryFn: adminCategories });
  const usersQuery = useQuery({ queryKey: ['admin-provider-users'], queryFn: adminUsers });
  const listingsQuery = useQuery({ queryKey: ['admin-listings'], queryFn: adminListings });
  const [search, setSearch] = useState('');
  const [results, setResults] = useState([]);
  const [selectedIds, setSelectedIds] = useState([]);
  const [categoryName, setCategoryName] = useState('');
  const [editCategory, setEditCategory] = useState({ id: '', name: '' });
  const [editListing, setEditListing] = useState({ id: '', name: '', includedMiles: '', deposit: '', available: true });

  const createCategoryMutation = useMutation({
    mutationFn: adminCreateCategory,
    onSuccess: () => {
      categoriesQuery.refetch();
      setCategoryName('');
    },
  });

  const updateCategoryMutation = useMutation({
    mutationFn: ({ id, payload }) => adminUpdateCategory(id, payload),
    onSuccess: () => {
      categoriesQuery.refetch();
      setEditCategory({ id: '', name: '' });
    },
  });

  const deleteCategoryMutation = useMutation({
    mutationFn: adminDeleteCategory,
    onSuccess: () => categoriesQuery.refetch(),
  });

  const createListingMutation = useMutation({
    mutationFn: adminCreateListing,
    onSuccess: () => listingsQuery.refetch(),
  });

  const updateListingMutation = useMutation({
    mutationFn: ({ id, payload }) => adminUpdateListing(id, payload),
    onSuccess: () => {
      listingsQuery.refetch();
      setEditListing({ id: '', name: '', includedMiles: '', deposit: '', available: true });
    },
  });

  const deleteListingMutation = useMutation({
    mutationFn: adminDeleteListing,
    onSuccess: () => listingsQuery.refetch(),
  });

  const bulkMutation = useMutation({
    mutationFn: adminBulkListings,
    onSuccess: () => {
      listingsQuery.refetch();
      setSelectedIds([]);
    },
  });

  const [listingForm, setListingForm] = useState({ categoryId: '', providerId: '', name: '', spu: '', sku: '', includedMiles: '0', deposit: '0', available: true });

  return (
    <div className="space-y-6">
      <Card>
        <CardTitle>Category CRUD</CardTitle>
        <div className="mt-4 flex gap-2">
          <Input placeholder="New category name" value={categoryName} onChange={(e) => setCategoryName(e.target.value)} />
          <Button onClick={() => createCategoryMutation.mutate({ name: categoryName })}>Create</Button>
        </div>
        <div className="mt-4">
          <DataTable
            columns={[
              { key: 'name', title: 'Name' },
              {
                key: 'actions',
                title: 'Actions',
                render: (row) => (
                  <div className="flex gap-2">
                    <Button variant="outline" onClick={() => setEditCategory({ id: row.id, name: row.name })}>Edit</Button>
                    <Button variant="danger" onClick={() => deleteCategoryMutation.mutate(row.id)}>Delete</Button>
                  </div>
                ),
              },
            ]}
            rows={categoriesQuery.data || []}
            empty="No categories"
          />
        </div>
        {editCategory.id && (
          <div className="mt-3 flex gap-2">
            <Input value={editCategory.name} onChange={(e) => setEditCategory((prev) => ({ ...prev, name: e.target.value }))} />
            <Button onClick={() => updateCategoryMutation.mutate({ id: editCategory.id, payload: { name: editCategory.name } })}>Save</Button>
          </div>
        )}
      </Card>

      <Card>
        <CardTitle>Create Listing</CardTitle>
        <div className="mt-4 grid gap-3 md:grid-cols-3">
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={listingForm.categoryId} onChange={(e) => setListingForm((prev) => ({ ...prev, categoryId: e.target.value }))}>
            <option value="">Category</option>
            {(categoriesQuery.data || []).map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}
          </select>
          <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={listingForm.providerId} onChange={(e) => setListingForm((prev) => ({ ...prev, providerId: e.target.value }))}>
            <option value="">Provider</option>
            {(usersQuery.data || []).filter((user) => user.roles.includes('provider')).map((user) => <option key={user.id} value={user.id}>{user.username}</option>)}
          </select>
          <Input placeholder="Name" value={listingForm.name} onChange={(e) => setListingForm((prev) => ({ ...prev, name: e.target.value }))} />
          <Input placeholder="SPU" value={listingForm.spu} onChange={(e) => setListingForm((prev) => ({ ...prev, spu: e.target.value }))} />
          <Input placeholder="SKU" value={listingForm.sku} onChange={(e) => setListingForm((prev) => ({ ...prev, sku: e.target.value }))} />
          <Input placeholder="Included miles" value={listingForm.includedMiles} onChange={(e) => setListingForm((prev) => ({ ...prev, includedMiles: e.target.value }))} />
          <Input placeholder="Deposit" value={listingForm.deposit} onChange={(e) => setListingForm((prev) => ({ ...prev, deposit: e.target.value }))} />
        </div>
        <Button
          className="mt-3"
          onClick={() =>
            createListingMutation.mutate({
              ...listingForm,
              includedMiles: Number(listingForm.includedMiles),
              deposit: Number(listingForm.deposit),
            })
          }
        >
          Create Listing
        </Button>
      </Card>

      <Card>
        <CardTitle>Listing CRUD + Bulk Controls</CardTitle>
        <div className="mt-4 flex gap-2">
          <Input placeholder="Search name/SPU/SKU" value={search} onChange={(e) => setSearch(e.target.value)} />
          <Button variant="outline" onClick={() => adminSearchListings(search).then(setResults)}>Search</Button>
          <Button onClick={() => bulkMutation.mutate({ listingIds: selectedIds, available: false })} disabled={selectedIds.length === 0}>Mark Unavailable</Button>
          <Button variant="outline" onClick={() => bulkMutation.mutate({ listingIds: selectedIds, available: true })} disabled={selectedIds.length === 0}>Mark Available</Button>
        </div>
        <div className="mt-4">
          <DataTable
            columns={[
              {
                key: 'select',
                title: '',
                render: (row) => (
                  <input
                    type="checkbox"
                    checked={selectedIds.includes(row.id)}
                    onChange={(e) => setSelectedIds((prev) => (e.target.checked ? [...prev, row.id] : prev.filter((id) => id !== row.id)))}
                  />
                ),
              },
              { key: 'name', title: 'Name' },
              { key: 'spu', title: 'SPU' },
              { key: 'sku', title: 'SKU' },
              { key: 'includedMiles', title: 'Included Miles' },
              { key: 'deposit', title: 'Deposit' },
              { key: 'available', title: 'Available', render: (row) => (row.available ? 'Yes' : 'No') },
              {
                key: 'actions',
                title: 'Actions',
                render: (row) => (
                  <div className="flex gap-2">
                    <Button variant="outline" onClick={() => setEditListing({ id: row.id, name: row.name, includedMiles: String(row.includedMiles), deposit: String(row.deposit), available: row.available })}>Edit</Button>
                    <Button variant="danger" onClick={() => deleteListingMutation.mutate(row.id)}>Delete</Button>
                  </div>
                ),
              },
            ]}
            rows={results.length ? results : listingsQuery.data || []}
            empty="No listings"
          />
        </div>

        {editListing.id && (
          <div className="mt-4 grid gap-2 md:grid-cols-5">
            <Input value={editListing.name} onChange={(e) => setEditListing((prev) => ({ ...prev, name: e.target.value }))} />
            <Input value={editListing.includedMiles} onChange={(e) => setEditListing((prev) => ({ ...prev, includedMiles: e.target.value }))} />
            <Input value={editListing.deposit} onChange={(e) => setEditListing((prev) => ({ ...prev, deposit: e.target.value }))} />
            <select className="rounded-xl border border-slate-700 bg-slate-900/70 px-3 py-2 text-sm" value={String(editListing.available)} onChange={(e) => setEditListing((prev) => ({ ...prev, available: e.target.value === 'true' }))}>
              <option value="true">Available</option>
              <option value="false">Unavailable</option>
            </select>
            <Button
              onClick={() =>
                updateListingMutation.mutate({
                  id: editListing.id,
                  payload: {
                    name: editListing.name,
                    includedMiles: Number(editListing.includedMiles),
                    deposit: Number(editListing.deposit),
                    available: editListing.available,
                  },
                })
              }
            >
              Save Listing
            </Button>
          </div>
        )}
      </Card>
    </div>
  );
}
