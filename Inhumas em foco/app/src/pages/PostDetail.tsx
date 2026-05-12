import { useParams, Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { ArrowLeft, Calendar, User } from 'lucide-react';

export default function PostDetail() {
  const { slug } = useParams<{ slug: string }>();
  const { posts } = useAppStore();

  const post = posts.find(p => p.slug === slug);
  if (!post) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <h1 className="text-2xl font-bold text-slate-900 mb-4">Notícia não encontrada</h1>
        <Link to="/" className="text-blue-600 hover:underline">Voltar para Home</Link>
      </div>
    );
  }

  const related = posts.filter(p => p.categoryId === post.categoryId && p.id !== post.id && p.status === 'published').slice(0, 3);

  const formatDate = (date?: string) => {
    if (!date) return '';
    return new Date(date).toLocaleDateString('pt-BR', { day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <Link to="/" className="text-sm text-slate-500 hover:text-blue-600 flex items-center gap-1 mb-6">
        <ArrowLeft className="h-4 w-4" /> Voltar
      </Link>

      <article>
        <div className="aspect-video bg-slate-200 rounded-xl flex items-center justify-center mb-6">
          <span className="text-slate-400">Imagem: {post.title}</span>
        </div>

        <div className="flex items-center gap-2 mb-4">
          <Badge variant="secondary">{post.categoryName}</Badge>
          {post.isSponsored && <Badge className="bg-amber-500 text-slate-900 hover:bg-amber-500">Conteúdo Patrocinado</Badge>}
        </div>

        <h1 className="text-3xl md:text-4xl font-bold text-slate-900 mb-4">{post.title}</h1>

        <div className="flex items-center gap-4 text-sm text-slate-500 mb-6">
          {post.publishedAt && (
            <span className="flex items-center gap-1"><Calendar className="h-4 w-4" /> {formatDate(post.publishedAt)}</span>
          )}
          {post.authorName && (
            <span className="flex items-center gap-1"><User className="h-4 w-4" /> {post.authorName}</span>
          )}
        </div>

        {post.excerpt && (
          <p className="text-lg text-slate-600 italic mb-6 border-l-4 border-amber-400 pl-4">{post.excerpt}</p>
        )}

        <div className="prose prose-slate max-w-none mb-8" dangerouslySetInnerHTML={{ __html: post.content }} />

        {post.editorialNotes && (
          <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-8">
            <h4 className="text-amber-800 font-bold mb-2">Notas Editoriais</h4>
            <p className="text-sm text-slate-700">{post.editorialNotes}</p>
            <p className="text-xs text-slate-500 mt-2">Responsável: {post.editorResponsible}</p>
          </div>
        )}
      </article>

      {related.length > 0 && (
        <div className="mt-12">
          <h2 className="text-2xl font-bold text-slate-900 mb-6">Leia também</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {related.map(r => (
              <Card key={r.id} className="overflow-hidden hover:shadow-lg transition">
                <CardContent className="p-4">
                  <Badge variant="secondary" className="mb-2">{r.categoryName}</Badge>
                  <h3 className="font-bold mb-2">
                    <Link to={`/noticia/${r.slug}`} className="hover:text-blue-600 transition">{r.title}</Link>
                  </h3>
                  <p className="text-sm text-slate-600 line-clamp-2">{r.excerpt}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
