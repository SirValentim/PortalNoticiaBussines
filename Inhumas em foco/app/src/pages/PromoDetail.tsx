import { useParams, Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { MessageCircle, Clock, ArrowLeft } from 'lucide-react';

export default function PromoDetail() {
  const { slug } = useParams<{ slug: string }>();
  const { promotions, stores } = useAppStore();

  const promo = promotions.find(p => p.slug === slug);
  if (!promo) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <h1 className="text-2xl font-bold text-slate-900 mb-4">Promoção não encontrada</h1>
        <Link to="/promocoes" className="text-blue-600 hover:underline">Ver todas as promoções</Link>
      </div>
    );
  }

  const store = stores.find(s => s.id === promo.storeId);

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <Link to="/promocoes" className="text-sm text-slate-500 hover:text-blue-600 flex items-center gap-1 mb-6">
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Link>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <div className="aspect-video bg-orange-100 rounded-xl flex items-center justify-center mb-6">
            <span className="text-orange-800 font-bold text-lg">{promo.storeName}</span>
          </div>

          <Badge variant="outline" className="mb-2">{promo.storeName}</Badge>
          <h1 className="text-3xl font-bold text-slate-900 mb-4">
            {promo.title}
            {promo.isSponsored && <span className="ml-2 text-sm bg-amber-400 text-slate-900 px-2 py-0.5 rounded">Patrocinado</span>}
          </h1>

          <p className="text-2xl text-orange-600 font-bold mb-6">{promo.priceDisplay}</p>

          {promo.description && (
            <div className="prose prose-slate max-w-none mb-6" dangerouslySetInnerHTML={{ __html: promo.description }} />
          )}

          <div className="bg-slate-50 rounded-xl p-6 mb-6">
            <p className="text-sm text-slate-600 mb-2">
              <strong>Validade:</strong> {new Date(promo.startDate).toLocaleDateString('pt-BR')} até {new Date(promo.endDate).toLocaleDateString('pt-BR')}
            </p>
            <p className="text-sm text-slate-600">
              <strong>Loja:</strong>{' '}
              <Link to={`/loja/${promo.storeSlug}`} className="text-blue-600 hover:underline">{promo.storeName}</Link>
            </p>
          </div>

          {store?.whatsapp && (
            <a href={`https://wa.me/${store.whatsapp}`} target="_blank" rel="noopener noreferrer">
              <Button size="lg" className="bg-green-600 hover:bg-green-700">
                <MessageCircle className="h-5 w-5 mr-2" /> Resgatar no WhatsApp
              </Button>
            </a>
          )}
        </div>

        <aside className="space-y-4">
          <div className="bg-white border rounded-xl p-6">
            <h4 className="font-bold mb-4">Detalhes</h4>
            <div className="text-sm text-slate-600 space-y-2">
              <p className="flex items-center gap-2">
                <Clock className="h-4 w-4" /> Válido até {new Date(promo.endDate).toLocaleDateString('pt-BR')}
              </p>
            </div>
          </div>

          {promo.isSponsored && (
            <div className="bg-amber-50 border border-amber-200 rounded-xl p-4">
              <Badge className="bg-amber-500 text-slate-900 mb-2">Conteúdo Patrocinado</Badge>
              <p className="text-xs text-slate-600">Esta promoção é uma parceria do Inhumas em Foco.</p>
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}
