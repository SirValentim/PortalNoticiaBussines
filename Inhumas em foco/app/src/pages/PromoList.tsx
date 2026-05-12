import { Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Clock } from 'lucide-react';

export default function PromoList() {
  const { promotions } = useAppStore();
  const activePromos = promotions.filter(p => p.status === 'active');

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold text-slate-900">Promoções do Dia</h1>
        <span className="text-sm text-slate-500">{activePromos.length} promoções ativas</span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {activePromos.map(promo => (
          <Card key={promo.id} className="overflow-hidden hover:shadow-lg transition">
            <Link to={`/promocao/${promo.slug}`}>
              <div className="aspect-video bg-orange-100 flex items-center justify-center">
                <span className="text-orange-800 font-bold">{promo.storeName}</span>
              </div>
            </Link>
            <CardContent className="p-4">
              <Badge variant="outline" className="mb-2">{promo.storeName}</Badge>
              <h3 className="font-bold text-lg mb-2">
                <Link to={`/promocao/${promo.slug}`} className="hover:text-blue-600 transition">{promo.title}</Link>
                {promo.isSponsored && <span className="ml-2 text-xs bg-amber-500 text-slate-900 px-1.5 py-0.5 rounded">Patrocinado</span>}
              </h3>
              <p className="text-orange-600 font-bold">{promo.priceDisplay}</p>
              <div className="text-xs text-slate-400 mt-2 flex items-center gap-1">
                <Clock className="h-3 w-3" /> Até {new Date(promo.endDate).toLocaleDateString('pt-BR')}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {activePromos.length === 0 && (
        <div className="text-center py-16 text-slate-500">Nenhuma promoção ativa no momento</div>
      )}
    </div>
  );
}
