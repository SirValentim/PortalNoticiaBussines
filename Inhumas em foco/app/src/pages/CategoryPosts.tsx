import { useParams, Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { categories } from '@/data/mock';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { ArrowLeft } from 'lucide-react';

export default function CategoryPosts() {
  const { slug } = useParams<{ slug: string }>();
  const { posts } = useAppStore();

  const category = categories.find(c => c.slug === slug);
  const catPosts = posts.filter(p => p.categoryId === category?.id && p.status === 'published');

  if (!category) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <h1 className="text-2xl font-bold text-slate-900 mb-4">Categoria não encontrada</h1>
        <Link to="/" className="text-blue-600 hover:underline">Voltar para Home</Link>
      </div>
    );
  }

  const formatDate = (date?: string) => {
    if (!date) return '';
    return new Date(date).toLocaleDateString('pt-BR');
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <Link to="/" className="text-sm text-slate-500 hover:text-blue-600 flex items-center gap-1 mb-6">
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Link>

      <h1 className="text-3xl font-bold text-slate-900 mb-2">{category.name}</h1>
      <p className="text-slate-600 mb-8">{category.description}</p>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {catPosts.map(post => (
          <Card key={post.id} className="overflow-hidden hover:shadow-lg transition">
            <CardContent className="p-4">
              <Badge variant="secondary" className="mb-2">{post.categoryName}</Badge>
              <h3 className="font-bold text-lg mb-2">
                <Link to={`/noticia/${post.slug}`} className="hover:text-blue-600 transition">{post.title}</Link>
              </h3>
              <p className="text-sm text-slate-600 line-clamp-2">{post.excerpt}</p>
              <div className="text-xs text-slate-400 mt-2">{formatDate(post.publishedAt)}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      {catPosts.length === 0 && (
        <div className="text-center py-16 text-slate-500">Nenhuma notícia nesta categoria</div>
      )}
    </div>
  );
}
