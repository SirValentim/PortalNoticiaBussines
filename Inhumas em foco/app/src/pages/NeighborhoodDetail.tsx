import { useParams, Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ArrowLeft, MapPin } from 'lucide-react';

export default function NeighborhoodDetail() {
  const { slug } = useParams<{ slug: string }>();
  const { neighborhoods, stores } = useAppStore();

  const neighborhood = neighborhoods.find(n => n.slug === slug);
  if (!neighborhood) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <h1 className="text-2xl font-bold text-slate-900 mb-4">Bairro não encontrado</h1>
        <Link to="/" className="text-blue-600 hover:underline">Voltar para Home</Link>
      </div>
    );
  }

  const neighborhoodStores = stores.filter(s => s.active);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <Link to="/" className="text-sm text-slate-500 hover:text-blue-600 flex items-center gap-1 mb-6">
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Link>

      <div className="aspect-video bg-slate-200 rounded-xl flex items-center justify-center mb-6">
        <span className="text-slate-400 text-lg">Imagem do bairro: {neighborhood.name}</span>
      </div>

      <h1 className="text-3xl font-bold text-slate-900 mb-2">{neighborhood.name}</h1>
      <p className="text-slate-600 mb-8">{neighborhood.description}</p>

      <h2 className="text-2xl font-bold text-slate-900 mb-6">Comércios do Bairro</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {neighborhoodStores.map(store => (
          <Card key={store.id} className="hover:shadow-lg transition">
            <CardContent className="p-4 flex gap-4">
              <div className="w-16 h-16 rounded-lg bg-slate-200 flex items-center justify-center shrink-0">
                <MapPin className="h-5 w-5 text-slate-400" />
              </div>
              <div>
                <h4 className="font-bold">{store.name}</h4>
                <p className="text-sm text-slate-500">{store.category}</p>
                <Link to={`/loja/${store.slug}`}>
                  <Button size="sm" variant="outline" className="mt-2">Ver loja</Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
