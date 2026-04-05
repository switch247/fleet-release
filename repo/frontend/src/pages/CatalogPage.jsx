import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { categories, listings } from '../lib/api';
import { Card, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';

function CategoryTree({ nodes, depth = 0, selectedCategory, onSelect }) {
  if (!nodes?.length) return null;
  return (
    <div className="space-y-2">
      {nodes.map((node) => (
        <div key={node.id}>
          <button
            type="button"
            className={`w-full text-left rounded border px-3 py-2 text-sm ${selectedCategory === node.id ? 'border-cyan-500 bg-cyan-950/30' : 'border-slate-800 bg-slate-900/40'}`}
            style={{ marginLeft: `${depth * 12}px` }}
            onClick={() => onSelect(node.id)}
          >
            {node.name}
          </button>
          <CategoryTree nodes={node.children} depth={depth + 1} selectedCategory={selectedCategory} onSelect={onSelect} />
        </div>
      ))}
    </div>
  );
}

export default function CatalogPage() {
  const categoriesQuery = useQuery({ queryKey: ['categories', 'tree'], queryFn: () => categories('tree') });
  const listingsQuery = useQuery({ queryKey: ['listings'], queryFn: listings });
  const [selectedCategory, setSelectedCategory] = useState('');
  const [compareMode, setCompareMode] = useState(false);

  const filteredListings = useMemo(() => {
    const rows = listingsQuery.data || [];
    if (!selectedCategory) return rows;
    return rows.filter((listing) => listing.categoryId === selectedCategory);
  }, [listingsQuery.data, selectedCategory]);

  const groupedBySPU = useMemo(() => {
    const map = new Map();
    for (const listing of filteredListings) {
      const key = listing.spu || 'NO-SPU';
      if (!map.has(key)) map.set(key, []);
      map.get(key).push(listing);
    }
    return Array.from(map.entries());
  }, [filteredListings]);

  return (
    <div className="space-y-6">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold">Catalog</h2>
          <p className="text-sm text-slate-400">Tree navigation and SPU/SKU comparison for available variants.</p>
        </div>
        <Button variant={compareMode ? 'secondary' : 'outline'} onClick={() => setCompareMode((v) => !v)}>
          {compareMode ? 'Comparison On' : 'Compare'}
        </Button>
      </header>

      <Card>
        <CardTitle>Category Tree</CardTitle>
        <div className="mt-4 space-y-3">
          <Button size="sm" variant="outline" onClick={() => setSelectedCategory('')}>All Categories</Button>
          <CategoryTree nodes={categoriesQuery.data || []} selectedCategory={selectedCategory} onSelect={setSelectedCategory} />
        </div>
      </Card>

      <Card>
        <CardTitle>{compareMode ? 'Variant Comparison' : 'Listings'}</CardTitle>
        <div className="mt-4 space-y-4">
          {groupedBySPU.map(([spu, variants]) => (
            <div key={spu} className="rounded border border-slate-800 p-3">
              <p className="font-medium">SPU: {spu}</p>
              {compareMode ? (
                <div className="mt-2 overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="text-left text-slate-400">
                        <th className="py-1 pr-3">SKU</th>
                        <th className="py-1 pr-3">Name</th>
                        <th className="py-1 pr-3">Price (Deposit)</th>
                        <th className="py-1 pr-3">Availability</th>
                        <th className="py-1 pr-3">Features</th>
                      </tr>
                    </thead>
                    <tbody>
                      {variants.map((listing) => (
                        <tr key={listing.id} className="border-t border-slate-800">
                          <td className="py-2 pr-3">{listing.sku}</td>
                          <td className="py-2 pr-3">{listing.name}</td>
                          <td className="py-2 pr-3">${Number(listing.deposit || 0).toFixed(2)}</td>
                          <td className="py-2 pr-3">{listing.available ? 'Available' : 'Unavailable'}</td>
                          <td className="py-2 pr-3">Included miles: {Number(listing.includedMiles || 0).toFixed(0)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="mt-2 space-y-2">
                  {variants.map((listing) => (
                    <div key={listing.id} className="rounded border border-slate-800 p-2 text-sm">
                      {listing.name} ({listing.sku}) - {listing.available ? 'Available' : 'Unavailable'}
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
          {groupedBySPU.length === 0 && <p className="text-sm text-slate-400">No listings for the selected category.</p>}
        </div>
      </Card>
    </div>
  );
}
