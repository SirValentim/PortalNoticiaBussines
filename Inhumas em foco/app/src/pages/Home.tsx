import { Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Clock, MapPin, Tag, ArrowRight } from 'lucide-react';

export default function Home() {
  const { posts, stores, promotions, banners, neighborhoods } = useAppStore();

  const publishedPosts = posts.filter(p => p.status === 'published');
  const headline = publishedPosts[0];
  const latestNews = publishedPosts.slice(1, 7);

  const catBastidores = posts.filter(p => p.categoryName === 'Política & Bastidores' && p.status === 'published').slice(0, 4);

  const heroBanner = banners.find(b => b.position === 'hero' && b.active);
  const inFeedBanner = banners.find(b => b.position === 'in_feed' && b.active);
  const stickyBanner = banners.find(b => b.position === 'sticky_footer' && b.active);

  const featuredStores = stores.filter(s => s.isFeatured && s.active).slice(0, 6);
  const activePromos = promotions.filter(p => p.status === 'active').slice(0, 3);

  const formatDate = (date?: string) => {
    if (!date) return '';
    return new Date(date).toLocaleDateString('pt-BR');
  };

  return (
    <div>
      {/* Hero Section */}
      <div className="bg-gradient-to-br from-slate-900 to-slate-800 text-white">
        <div className="max-w-7xl mx-auto px-4 py-8">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {headline && (
              <div className="lg:col-span-2 relative rounded-xl overflow-hidden group cursor-pointer">
                <Link to={`/noticia/${headline.slug}`}>
                  <div className="aspect-video bg-slate-700 flex items-center justify-center">
                    <span className="text-slate-500 text-lg">Imagem: {headline.title}</span>
                  </div>
                  <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent" />
                  <div className="absolute bottom-0 left-0 right-0 p-6">
                    <Badge className="mb-2 bg-amber-500 text-slate-900 hover:bg-amber-500">{headline.categoryName}</Badge>
                    <h1 className="text-2xl md:text-3xl font-bold leading-tight mb-2 group-hover:text-amber-400 transition">{headline.title}</h1>
                    <p className="text-slate-300 text-sm line-clamp-2">{headline.excerpt}</p>
                  </div>
                </Link>
              </div>
            )}
            <div className="flex flex-col gap-4">
              {latestNews.slice(0, 3).map(post => (
                <Link key={post.id} to={`/noticia/${post.slug}`} className="bg-white/5 rounded-lg p-4 hover:bg-white/10 transition">
                  <div className="text-xs text-amber-400 font-semibold uppercase mb-1">{post.categoryName}</div>
                  <h3 className="font-semibold text-sm line-clamp-2">{post.title}</h3>
                  <div className="text-xs text-slate-400 mt-1">{formatDate(post.publishedAt)}</div>
                </Link>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Hero Banner */}
      {heroBanner && (
        <div className="max-w-7xl mx-auto px-4 mt-6">
          <Link to={heroBanner.linkUrl} className="block rounded-xl overflow-hidden bg-amber-100 hover:opacity-90 transition">
            <div className="h-32 md:h-40 flex items-center justify-center bg-gradient-to-r from-amber-200 to-orange-200">
              <span className="text-amber-800 font-bold text-lg">{heroBanner.name}</span>
            </div>
          </Link>
        </div>
      )}

      {/* Latest News */}
      <section className="max-w-7xl mx-auto px-4 py-10">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-slate-900">Últimas Notícias</h2>
          <Link to="/categoria/noticias" className="text-sm text-blue-600 hover:underline flex items-center gap-1">
            Ver todas <ArrowRight className="h-4 w-4" />
          </Link>
        </div>

        {inFeedBanner && (
          <div className="mb-6 rounded-xl overflow-hidden bg-blue-50 hover:opacity-90 transition">
            <Link to={inFeedBanner.linkUrl} className="block h-24 flex items-center justify-center">
              <span className="text-blue-800 font-semibold">{inFeedBanner.name}</span>
            </Link>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {latestNews.map(post => (
            <Card key={post.id} className="overflow-hidden hover:shadow-lg transition">
              <Link to={`/noticia/${post.slug}`}>
                <div className="aspect-video bg-slate-200 flex items-center justify-center">
                  <span className="text-slate-400 text-sm">Imagem</span>
                </div>
              </Link>
              <CardContent className="p-4">
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="secondary" className="text-xs">{post.categoryName}</Badge>
                  {post.isSponsored && <Badge className="text-xs bg-amber-500 text-slate-900 hover:bg-amber-500">Patrocinado</Badge>}
                </div>
                <h3 className="font-bold text-lg leading-snug mb-2">
                  <Link to={`/noticia/${post.slug}`} className="hover:text-blue-600 transition">{post.title}</Link>
                </h3>
                <p className="text-sm text-slate-600 line-clamp-2">{post.excerpt}</p>
                <div className="text-xs text-slate-400 mt-3">{formatDate(post.publishedAt)} {post.authorName && `por ${post.authorName}`}</div>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      {/* Política & Bastidores */}
      {catBastidores.length > 0 && (
        <section className="bg-slate-50 py-10">
          <div className="max-w-7xl mx-auto px-4">
            <h2 className="text-2xl font-bold text-slate-900 mb-6">Política & Bastidores</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {catBastidores.map(post => (
                <Card key={post.id} className="overflow-hidden hover:shadow-lg transition">
                  <CardContent className="p-4">
                    <Badge variant="secondary" className="mb-2">Política & Bastidores</Badge>
                    <h3 className="font-bold mb-2">
                      <Link to={`/noticia/${post.slug}`} className="hover:text-blue-600 transition">{post.title}</Link>
                    </h3>
                    <p className="text-sm text-slate-600 line-clamp-3">{post.excerpt}</p>
                    <div className="text-xs text-slate-400 mt-2">{formatDate(post.publishedAt)}</div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* Featured Stores */}
      <section className="max-w-7xl mx-auto px-4 py-10">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-slate-900">Lojas em Destaque</h2>
          <Link to="/lojas" className="text-sm text-blue-600 hover:underline flex items-center gap-1">
            Ver todas <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {featuredStores.map(store => (
            <Card key={store.id} className="hover:shadow-lg transition">
              <CardContent className="p-4 flex gap-4">
                <div className="w-20 h-20 rounded-lg bg-slate-200 flex items-center justify-center shrink-0">
                  <Tag className="h-6 w-6 text-slate-400" />
                </div>
                <div className="flex-1 min-w-0">
                  <h4 className="font-bold truncate">{store.name} {store.isSponsored && <span className="text-xs bg-amber-400 text-slate-900 px-1.5 py-0.5 rounded">Patrocinado</span>}</h4>
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
      </section>

      {/* Promotions */}
      {activePromos.length > 0 && (
        <section className="bg-amber-50 py-10">
          <div className="max-w-7xl mx-auto px-4">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-bold text-slate-900">Promoções do Dia</h2>
              <Link to="/promocoes" className="text-sm text-blue-600 hover:underline flex items-center gap-1">
                Ver todas <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {activePromos.map(promo => (
                <Card key={promo.id} className="overflow-hidden hover:shadow-lg transition">
                  <Link to={`/promocao/${promo.slug}`}>
                    <div className="aspect-video bg-orange-200 flex items-center justify-center">
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
                      <Clock className="h-3 w-3" /> Até {formatDate(promo.endDate)}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* Neighborhoods */}
      <section className="max-w-7xl mx-auto px-4 py-10">
        <h2 className="text-2xl font-bold text-slate-900 mb-6">Bairros de Inhumas</h2>
        <div className="flex flex-wrap gap-3">
          {neighborhoods.map(n => (
            <Link key={n.id} to={`/bairro/${n.slug}`}>
              <Badge variant="secondary" className="px-4 py-2 text-sm hover:bg-slate-200 cursor-pointer">{n.name}</Badge>
            </Link>
          ))}
        </div>
      </section>

      {/* Sticky Banner Mobile */}
      {stickyBanner && (
        <div className="md:hidden fixed bottom-0 left-0 right-0 bg-white shadow-lg z-50 border-t">
          <div className="relative">
            <Link to={stickyBanner.linkUrl} className="block h-16 bg-pink-100 flex items-center justify-center">
              <span className="text-pink-800 font-semibold text-sm">{stickyBanner.name}</span>
            </Link>
            <button
              onClick={() => { /* hide banner */ }}
              className="absolute top-1 right-1 w-6 h-6 bg-black/50 text-white rounded-full text-xs flex items-center justify-center"
            >
              ✕
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
