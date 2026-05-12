import { useState } from 'react';
import { Link, useSearchParams } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search } from 'lucide-react';

export default function SearchPage() {
  const [searchParams] = useSearchParams();
  const query = searchParams.get('q') || '';
  const { posts } = useAppStore();
  const [searchQuery, setSearchQuery] = useState(query);

  const results = query
    ? posts.filter(p =>
        p.status === 'published' &&
        (p.title.toLowerCase().includes(query.toLowerCase()) ||
         p.excerpt.toLowerCase().includes(query.toLowerCase()) ||
         p.content.toLowerCase().includes(query.toLowerCase()))
      )
    : [];

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-slate-900 mb-6">Busca</h1>

      <form
        onSubmit={e => { e.preventDefault(); window.location.href = `/busca?q=${encodeURIComponent(searchQuery)}`; }}
        className="flex gap-2 mb-6 max-w-xl"
      >
        <Input
          type="text"
          placeholder="Buscar notícias..."
          className="flex-1"
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
        />
        <Button type="submit"><Search className="h-4 w-4 mr-1" /> Buscar</Button>
      </form>

      {query && (
        <p className="text-sm text-slate-500 mb-4">
          {results.length} resultado{results.length !== 1 ? 's' : ''} para "{query}"
        </p>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {results.map(post => (
          <Card key={post.id} className="overflow-hidden hover:shadow-lg transition">
            <CardContent className="p-4">
              <Badge variant="secondary" className="mb-2">{post.categoryName}</Badge>
              <h3 className="font-bold text-lg mb-2">
                <Link to={`/noticia/${post.slug}`} className="hover:text-blue-600 transition">{post.title}</Link>
              </h3>
              <p className="text-sm text-slate-600 line-clamp-2">{post.excerpt}</p>
              <div className="text-xs text-slate-400 mt-2">
                {new Date(post.publishedAt || '').toLocaleDateString('pt-BR')}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {query && results.length === 0 && (
        <div className="text-center py-16 text-slate-500">Nenhum resultado encontrado</div>
      )}
    </div>
  );
}
