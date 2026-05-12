import { Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { MapPin, Tag } from 'lucide-react';

export default function StoreList() {
  const { stores } = useAppStore();
  const activeStores = stores.filter(s => s.active);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold text-slate-900">Lojas e Comércios</h1>
        <span className="text-sm text-slate-500">{activeStores.length} lojas cadastradas</span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {activeStores.map(store => (
          <Card key={store.id} className="hover:shadow-lg transition">
            <CardContent className="p-4 flex gap-4">
              <div className="w-20 h-20 rounded-lg bg-slate-200 flex items-center justify-center shrink-0">
                <Tag className="h-6 w-6 text-slate-400" />
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="font-bold truncate">
                  {store.name}
                  {store.isSponsored && <span className="ml-2 text-xs bg-amber-400 text-slate-900 px-1.5 py-0.5 rounded">Patrocinado</span>}
                </h4>
                <p className="text-sm text-slate-500">{store.category}</p>
                {store.address && (
                  <p className="text-xs text-slate-400 flex items-center gap-1 mt-1">
                    <MapPin className="h-3 w-3" /> {store.address}
                  </p>
                )}
                <Link to={`/loja/${store.slug}`}>
                  <Button size="sm" variant="outline" className="mt-2">Ver loja</Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {activeStores.length === 0 && (
        <div className="text-center py-16 text-slate-500">Nenhuma loja cadastrada</div>
      )}
    </div>
  );
}
