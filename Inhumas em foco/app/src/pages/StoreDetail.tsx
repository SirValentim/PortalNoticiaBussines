import { useParams, Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { MapPin, Phone, MessageCircle, ArrowLeft, Tag } from 'lucide-react';

export default function StoreDetail() {
  const { slug } = useParams<{ slug: string }>();
  const { stores, promotions } = useAppStore();

  const store = stores.find(s => s.slug === slug);
  if (!store) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <h1 className="text-2xl font-bold text-slate-900 mb-4">Loja não encontrada</h1>
        <Link to="/lojas" className="text-blue-600 hover:underline">Ver todas as lojas</Link>
      </div>
    );
  }

  const storePromos = promotions.filter(p => p.storeId === store.id && p.status === 'active');

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      <Link to="/lojas" className="text-sm text-slate-500 hover:text-blue-600 flex items-center gap-1 mb-6">
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Link>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          {store.coverImageKey && (
            <div className="aspect-video bg-slate-200 rounded-xl flex items-center justify-center mb-6">
              <span className="text-slate-400">Imagem da loja</span>
            </div>
          )}

          <div className="flex items-center gap-4 mb-6">
            {store.logoKey && (
              <div className="w-20 h-20 rounded-lg bg-slate-200 flex items-center justify-center">
                <Tag className="h-8 w-8 text-slate-400" />
              </div>
            )}
            <div>
              <h1 className="text-3xl font-bold text-slate-900">
                {store.name}
                {store.isSponsored && <span className="ml-2 text-sm bg-amber-400 text-slate-900 px-2 py-0.5 rounded">Patrocinado</span>}
              </h1>
              <p className="text-slate-500">{store.category}</p>
            </div>
          </div>

          <p className="text-slate-700 mb-6">{store.description}</p>

          <div className="bg-slate-50 rounded-xl p-6 mb-8">
            <h3 className="font-bold text-lg mb-4">Informações</h3>
            {store.address && (
              <p className="flex items-center gap-2 text-sm text-slate-600 mb-2">
                <MapPin className="h-4 w-4" /> {store.address}
              </p>
            )}
            {store.phone && (
              <p className="flex items-center gap-2 text-sm text-slate-600 mb-2">
                <Phone className="h-4 w-4" /> {store.phone}
              </p>
            )}
            {store.whatsapp && (
              <p className="flex items-center gap-2 text-sm text-slate-600">
                <MessageCircle className="h-4 w-4" />
                <a href={`https://wa.me/${store.whatsapp}`} target="_blank" rel="noopener noreferrer" className="text-green-600 hover:underline">
                  {store.whatsapp}
                </a>
              </p>
            )}
          </div>

          {storePromos.length > 0 && (
            <div>
              <h2 className="text-2xl font-bold text-slate-900 mb-4">Promoções</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {storePromos.map(promo => (
                  <Card key={promo.id} className="overflow-hidden hover:shadow-lg transition">
                    <CardContent className="p-4">
                      <h3 className="font-bold mb-2">
                        <Link to={`/promocao/${promo.slug}`} className="hover:text-blue-600 transition">{promo.title}</Link>
                      </h3>
                      <p className="text-orange-600 font-bold">{promo.priceDisplay}</p>
                      <p className="text-xs text-slate-400 mt-2">Até {new Date(promo.endDate).toLocaleDateString('pt-BR')}</p>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </div>
          )}
        </div>

        <aside className="space-y-4">
          <div className="bg-white border rounded-xl p-6">
            <h4 className="font-bold mb-4">Ações</h4>
            {store.whatsapp && (
              <a href={`https://wa.me/${store.whatsapp}`} target="_blank" rel="noopener noreferrer">
                <Button className="w-full mb-2 bg-green-600 hover:bg-green-700">
                  <MessageCircle className="h-4 w-4 mr-2" /> WhatsApp
                </Button>
              </a>
            )}
            {store.phone && (
              <a href={`tel:${store.phone}`}>
                <Button variant="outline" className="w-full">
                  <Phone className="h-4 w-4 mr-2" /> Ligar
                </Button>
              </a>
            )}
          </div>

          {store.isSponsored && (
            <div className="bg-amber-50 border border-amber-200 rounded-xl p-4">
              <Badge className="bg-amber-500 text-slate-900 mb-2">Conteúdo Patrocinado</Badge>
              <p className="text-xs text-slate-600">Esta loja é uma parceira do Inhumas em Foco.</p>
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}
