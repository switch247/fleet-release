import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { categories, listings } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';

export default function CatalogPage() {
  const categoriesQuery = useQuery({ queryKey: ['categories'], queryFn: categories });
  const listingsQuery = useQuery({ queryKey: ['listings'], queryFn: listings });

  return (
    <div className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold">Catalog</h2>
        <p className="text-sm text-slate-400">Browse categories and listings</p>
      </header>

      <Card>
        <CardTitle>Categories</CardTitle>
        <div className="mt-4 space-y-2">
          {categoriesQuery.data?.map((cat) => (
            <div key={cat.id} className="p-2 border border-slate-800 rounded">
              {cat.name}
            </div>
          ))}
        </div>
      </Card>

      <Card>
        <CardTitle>Listings</CardTitle>
        <div className="mt-4 space-y-2">
          {listingsQuery.data?.map((listing) => (
            <div key={listing.id} className="p-2 border border-slate-800 rounded">
              {listing.name} - {listing.categoryId}
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}