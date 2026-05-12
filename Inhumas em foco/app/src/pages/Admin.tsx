import { useParams, useNavigate, Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import type { Category, Store } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent } from '@/components/ui/tabs';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { toast } from 'sonner';
import {
  LayoutDashboard, FileText, Store as StoreIcon, Image, Percent, MapPin, Users, BarChart3,
  Plus, Trash2, Eye, LogOut
} from 'lucide-react';
import { useState } from 'react';

export default function Admin() {
  const { section } = useParams<{ section?: string }>();
  const { currentUser, logout } = useAppStore();
  const navigate = useNavigate();
  const activeTab = section || 'dashboard';

  if (!currentUser) {
    return (
      <div className="min-h-[80vh] flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-slate-900 mb-4">Acesso Negado</h1>
          <p className="text-slate-600 mb-4">Você precisa estar autenticado para acessar o painel.</p>
          <Button onClick={() => navigate('/login')}>Fazer Login</Button>
        </div>
      </div>
    );
  }

  const navItems = [
    { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { id: 'posts', label: 'Notícias', icon: FileText },
    { id: 'stores', label: 'Lojas', icon: StoreIcon },
    { id: 'banners', label: 'Banners', icon: Image },
    { id: 'promotions', label: 'Promoções', icon: Percent },
    { id: 'neighborhoods', label: 'Bairros', icon: MapPin },
    { id: 'users', label: 'Usuários', icon: Users },
    { id: 'metrics', label: 'Métricas', icon: BarChart3 },
  ];

  return (
    <div className="min-h-[80vh] bg-slate-50">
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="flex flex-col lg:flex-row gap-6">
          {/* Sidebar */}
          <aside className="lg:w-64 shrink-0">
            <div className="bg-slate-900 text-white rounded-xl p-4 sticky top-20">
              <div className="mb-4 pb-4 border-b border-slate-700">
                <p className="font-bold">{currentUser.name}</p>
                <p className="text-xs text-slate-400 capitalize">{currentUser.role}</p>
              </div>
              <nav className="space-y-1">
                {navItems.map(item => {
                  const Icon = item.icon;
                  return (
                    <button
                      key={item.id}
                      onClick={() => navigate(`/painel/7x9k2m/${item.id}`)}
                      className={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition ${
                        activeTab === item.id ? 'bg-white/10 text-white' : 'text-slate-400 hover:text-white hover:bg-white/5'
                      }`}
                    >
                      <Icon className="h-4 w-4" /> {item.label}
                    </button>
                  );
                })}
              </nav>
              <div className="mt-4 pt-4 border-t border-slate-700">
                <Link to="/" className="block text-sm text-slate-400 hover:text-white mb-2">Ver site</Link>
                <button onClick={() => { logout(); navigate('/'); }} className="flex items-center gap-2 text-sm text-red-400 hover:text-red-300">
                  <LogOut className="h-4 w-4" /> Sair
                </button>
              </div>
            </div>
          </aside>

          {/* Content */}
          <main className="flex-1 min-w-0">
            <Tabs value={activeTab} className="w-full">
              <TabsContent value="dashboard" className="mt-0"><DashboardTab /></TabsContent>
              <TabsContent value="posts" className="mt-0"><PostsTab /></TabsContent>
              <TabsContent value="stores" className="mt-0"><StoresTab /></TabsContent>
              <TabsContent value="banners" className="mt-0"><BannersTab /></TabsContent>
              <TabsContent value="promotions" className="mt-0"><PromotionsTab /></TabsContent>
              <TabsContent value="neighborhoods" className="mt-0"><NeighborhoodsTab /></TabsContent>
              <TabsContent value="users" className="mt-0"><UsersTab /></TabsContent>
              <TabsContent value="metrics" className="mt-0"><MetricsTab /></TabsContent>
            </Tabs>
          </main>
        </div>
      </div>
    </div>
  );
}

function DashboardTab() {
  const { posts, stores, promotions, banners } = useAppStore();
  const publishedPosts = posts.filter(p => p.status === 'published').length;
  const activeStores = stores.filter(s => s.active).length;
  const activePromos = promotions.filter(p => p.status === 'active').length;
  const activeBanners = banners.filter(b => b.active).length;
  const navigate = useNavigate();

  const stats = [
    { label: 'Notícias Publicadas', value: publishedPosts, color: 'bg-blue-100 text-blue-800' },
    { label: 'Lojas Ativas', value: activeStores, color: 'bg-green-100 text-green-800' },
    { label: 'Promoções Ativas', value: activePromos, color: 'bg-orange-100 text-orange-800' },
    { label: 'Banners Ativos', value: activeBanners, color: 'bg-purple-100 text-purple-800' },
  ];

  const quickActions = [
    { label: 'Nova Notícia', tab: 'posts', color: 'bg-blue-600 hover:bg-blue-700' },
    { label: 'Nova Loja', tab: 'stores', color: 'bg-green-600 hover:bg-green-700' },
    { label: 'Novo Banner', tab: 'banners', color: 'bg-purple-600 hover:bg-purple-700' },
    { label: 'Nova Promoção', tab: 'promotions', color: 'bg-orange-600 hover:bg-orange-700' },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-slate-900">Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map(stat => (
          <Card key={stat.label}>
            <CardContent className="p-6">
              <p className="text-3xl font-bold">{stat.value}</p>
              <p className="text-sm text-slate-500 mt-1">{stat.label}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardContent className="p-6">
            <h3 className="font-bold text-lg mb-4">Ações Rápidas</h3>
            <div className="grid grid-cols-2 gap-3">
              {quickActions.map(action => (
                <Button
                  key={action.label}
                  className={action.color}
                  onClick={() => navigate(`/painel/7x9k2m/${action.tab}`)}
                >
                  <Plus className="h-4 w-4 mr-1" /> {action.label}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <h3 className="font-bold text-lg mb-4">Status do Sistema</h3>
            <div className="space-y-2 text-sm">
              <p className="flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-green-500" /> Banco de dados: OK</p>
              <p className="flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-green-500" /> Uploads: OK</p>
              <p className="flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-green-500" /> Worker: OK</p>
              <p className="flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-green-500" /> Sitemap: Atualizado</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function PostsTab() {
  const { posts, categories, deletePost, addPost } = useAppStore();
  const [open, setOpen] = useState(false);

  const handleDelete = (id: number) => {
    if (confirm('Tem certeza que deseja excluir?')) {
      deletePost(id);
      toast.success('Notícia excluída');
    }
  };

  const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const catId = parseInt(formData.get('category_id') as string);
    const cat = categories.find(c => c.id === catId);

    const newPost = {
      id: Date.now(),
      title: formData.get('title') as string,
      slug: (formData.get('title') as string).toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, ''),
      excerpt: formData.get('excerpt') as string,
      content: formData.get('content') as string,
      categoryId: catId,
      categoryName: cat?.name,
      status: formData.get('status') as 'draft' | 'published',
      isSponsored: formData.get('is_sponsored') === 'on',
      publishedAt: formData.get('status') === 'published' ? new Date().toISOString() : undefined,
      createdAt: new Date().toISOString(),
    };

    addPost(newPost as any);
    setOpen(false);
    toast.success('Notícia criada com sucesso');
    form.reset();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Notícias</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button><Plus className="h-4 w-4 mr-1" /> Nova Notícia</Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader><DialogTitle>Nova Notícia</DialogTitle></DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div><Label>Título</Label><Input name="title" required /></div>
              <div><Label>Categoria</Label>
                <Select name="category_id" defaultValue="1">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {categories.map((c: Category) => <SelectItem key={c.id} value={c.id.toString()}>{c.name}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
              <div><Label>Resumo</Label><Textarea name="excerpt" rows={3} /></div>
              <div><Label>Conteúdo (HTML)</Label><Textarea name="content" rows={8} required /></div>
              <div><Label>Status</Label>
                <Select name="status" defaultValue="draft">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="draft">Rascunho</SelectItem>
                    <SelectItem value="published">Publicado</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center gap-2">
                <input type="checkbox" name="is_sponsored" id="is_sponsored" />
                <Label htmlFor="is_sponsored">Conteúdo Patrocinado</Label>
              </div>
              <Button type="submit" className="w-full">Salvar</Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Título</TableHead>
            <TableHead>Categoria</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="w-24">Ações</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {posts.map(post => (
            <TableRow key={post.id}>
              <TableCell>
                {post.title}
                {post.isSponsored && <Badge className="ml-2 bg-amber-500 text-slate-900">Patrocinado</Badge>}
              </TableCell>
              <TableCell>{post.categoryName}</TableCell>
              <TableCell>
                {post.status === 'published' ? <span className="text-green-600 font-medium">Publicado</span> :
                 post.status === 'draft' ? <span className="text-slate-500">Rascunho</span> :
                 <span className="text-amber-600">Agendado</span>}
              </TableCell>
              <TableCell>
                <div className="flex gap-1">
                  <Link to={`/noticia/${post.slug}`} target="_blank"><Button size="icon" variant="ghost"><Eye className="h-4 w-4" /></Button></Link>
                  <Button size="icon" variant="ghost" className="text-red-600" onClick={() => handleDelete(post.id)}><Trash2 className="h-4 w-4" /></Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function StoresTab() {
  const { stores, deleteStore, addStore } = useAppStore();
  const [open, setOpen] = useState(false);

  const handleDelete = (id: number) => {
    if (confirm('Tem certeza?')) { deleteStore(id); toast.success('Loja excluída'); }
  };

  const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const store = {
      id: Date.now(),
      name: formData.get('name') as string,
      slug: (formData.get('name') as string).toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, ''),
      description: formData.get('description') as string,
      category: formData.get('category') as string,
      address: formData.get('address') as string,
      phone: formData.get('phone') as string,
      whatsapp: formData.get('whatsapp') as string,
      isSponsored: formData.get('is_sponsored') === 'on',
      isFeatured: formData.get('is_featured') === 'on',
      active: true,
      createdAt: new Date().toISOString(),
    };
    addStore(store as any);
    setOpen(false);
    toast.success('Loja criada');
    form.reset();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Lojas</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild><Button><Plus className="h-4 w-4 mr-1" /> Nova Loja</Button></DialogTrigger>
          <DialogContent className="max-w-lg">
            <DialogHeader><DialogTitle>Nova Loja</DialogTitle></DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div><Label>Nome</Label><Input name="name" required /></div>
              <div><Label>Descrição</Label><Textarea name="description" /></div>
              <div><Label>Categoria</Label><Input name="category" /></div>
              <div><Label>Endereço</Label><Input name="address" /></div>
              <div><Label>Telefone</Label><Input name="phone" /></div>
              <div><Label>WhatsApp</Label><Input name="whatsapp" /></div>
              <div className="flex gap-4">
                <div className="flex items-center gap-2"><input type="checkbox" name="is_sponsored" id="ss" /><Label htmlFor="ss">Patrocinado</Label></div>
                <div className="flex items-center gap-2"><input type="checkbox" name="is_featured" id="sf" /><Label htmlFor="sf">Destaque</Label></div>
              </div>
              <Button type="submit" className="w-full">Salvar</Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>
      <Table>
        <TableHeader>
          <TableRow><TableHead>Nome</TableHead><TableHead>Categoria</TableHead><TableHead>Status</TableHead><TableHead>Ações</TableHead></TableRow>
        </TableHeader>
        <TableBody>
          {stores.map(s => (
            <TableRow key={s.id}>
              <TableCell>{s.name} {s.isSponsored && <Badge className="ml-2 bg-amber-500 text-slate-900">Patrocinado</Badge>}</TableCell>
              <TableCell>{s.category}</TableCell>
              <TableCell>{s.active ? <span className="text-green-600">Ativo</span> : <span className="text-slate-400">Inativo</span>}</TableCell>
              <TableCell>
                <Button size="icon" variant="ghost" className="text-red-600" onClick={() => handleDelete(s.id)}><Trash2 className="h-4 w-4" /></Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function BannersTab() {
  const { banners, deleteBanner, addBanner } = useAppStore();
  const [open, setOpen] = useState(false);

  const handleDelete = (id: number) => {
    if (confirm('Tem certeza?')) { deleteBanner(id); toast.success('Banner excluído'); }
  };

  const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const banner = {
      id: Date.now(),
      name: formData.get('name') as string,
      position: formData.get('position') as any,
      imageKey: '',
      linkUrl: formData.get('link_url') as string,
      startDate: formData.get('start_date') as string,
      endDate: formData.get('end_date') as string,
      active: true,
      priority: parseInt(formData.get('priority') as string) || 0,
    };
    addBanner(banner as any);
    setOpen(false);
    toast.success('Banner criado');
    form.reset();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Banners</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild><Button><Plus className="h-4 w-4 mr-1" /> Novo Banner</Button></DialogTrigger>
          <DialogContent className="max-w-lg">
            <DialogHeader><DialogTitle>Novo Banner</DialogTitle></DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div><Label>Nome</Label><Input name="name" required /></div>
              <div><Label>Posição</Label>
                <Select name="position" defaultValue="hero">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="hero">Hero</SelectItem>
                    <SelectItem value="sidebar_top">Sidebar Topo</SelectItem>
                    <SelectItem value="in_feed">In-Feed</SelectItem>
                    <SelectItem value="sticky_footer">Sticky Footer</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div><Label>Link URL</Label><Input name="link_url" required /></div>
              <div className="grid grid-cols-2 gap-4">
                <div><Label>Data Início</Label><Input type="date" name="start_date" required /></div>
                <div><Label>Data Fim</Label><Input type="date" name="end_date" required /></div>
              </div>
              <div><Label>Prioridade</Label><Input type="number" name="priority" defaultValue="0" /></div>
              <Button type="submit" className="w-full">Salvar</Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>
      <Table>
        <TableHeader>
          <TableRow><TableHead>Nome</TableHead><TableHead>Posição</TableHead><TableHead>Período</TableHead><TableHead>Status</TableHead><TableHead>Ações</TableHead></TableRow>
        </TableHeader>
        <TableBody>
          {banners.map(b => (
            <TableRow key={b.id}>
              <TableCell>{b.name}</TableCell>
              <TableCell>{b.position}</TableCell>
              <TableCell>{b.startDate} - {b.endDate}</TableCell>
              <TableCell>{b.active ? <span className="text-green-600">Ativo</span> : <span className="text-slate-400">Inativo</span>}</TableCell>
              <TableCell>
                <Button size="icon" variant="ghost" className="text-red-600" onClick={() => handleDelete(b.id)}><Trash2 className="h-4 w-4" /></Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function PromotionsTab() {
  const { promotions, stores, deletePromotion, addPromotion } = useAppStore();
  const [open, setOpen] = useState(false);

  const handleDelete = (id: number) => {
    if (confirm('Tem certeza?')) { deletePromotion(id); toast.success('Promoção excluída'); }
  };

  const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const storeId = parseInt(formData.get('store_id') as string);
    const store = stores.find(s => s.id === storeId);
    const promo = {
      id: Date.now(),
      storeId,
      storeName: store?.name,
      storeSlug: store?.slug,
      title: formData.get('title') as string,
      slug: (formData.get('title') as string).toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, ''),
      description: formData.get('description') as string,
      priceDisplay: formData.get('price_display') as string,
      startDate: formData.get('start_date') as string,
      endDate: formData.get('end_date') as string,
      status: 'active',
      isSponsored: formData.get('is_sponsored') === 'on',
    };
    addPromotion(promo as any);
    setOpen(false);
    toast.success('Promoção criada');
    form.reset();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Promoções</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild><Button><Plus className="h-4 w-4 mr-1" /> Nova Promoção</Button></DialogTrigger>
          <DialogContent className="max-w-lg">
            <DialogHeader><DialogTitle>Nova Promoção</DialogTitle></DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div><Label>Loja</Label>
                <Select name="store_id" required>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {stores.map((s: Store) => <SelectItem key={s.id} value={s.id.toString()}>{s.name}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
              <div><Label>Título</Label><Input name="title" required /></div>
              <div><Label>Descrição</Label><Textarea name="description" /></div>
              <div><Label>Preço / Oferta</Label><Input name="price_display" placeholder="Ex: R$ 29,90" /></div>
              <div className="grid grid-cols-2 gap-4">
                <div><Label>Data Início</Label><Input type="date" name="start_date" required /></div>
                <div><Label>Data Fim</Label><Input type="date" name="end_date" required /></div>
              </div>
              <div className="flex items-center gap-2"><input type="checkbox" name="is_sponsored" id="ps" /><Label htmlFor="ps">Patrocinado</Label></div>
              <Button type="submit" className="w-full">Salvar</Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>
      <Table>
        <TableHeader>
          <TableRow><TableHead>Título</TableHead><TableHead>Loja</TableHead><TableHead>Preço</TableHead><TableHead>Validade</TableHead><TableHead>Ações</TableHead></TableRow>
        </TableHeader>
        <TableBody>
          {promotions.map(p => (
            <TableRow key={p.id}>
              <TableCell>{p.title} {p.isSponsored && <Badge className="ml-2 bg-amber-500 text-slate-900">Patrocinado</Badge>}</TableCell>
              <TableCell>{p.storeName}</TableCell>
              <TableCell>{p.priceDisplay}</TableCell>
              <TableCell>{p.endDate}</TableCell>
              <TableCell>
                <Button size="icon" variant="ghost" className="text-red-600" onClick={() => handleDelete(p.id)}><Trash2 className="h-4 w-4" /></Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function NeighborhoodsTab() {
  const { neighborhoods, deleteNeighborhood, addNeighborhood } = useAppStore();
  const [name, setName] = useState('');

  const handleCreate = () => {
    if (!name.trim()) return;
    addNeighborhood({
      id: Date.now(),
      name,
      slug: name.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, ''),
      description: '',
    } as any);
    setName('');
    toast.success('Bairro criado');
  };

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold text-slate-900">Bairros</h1>

      <Card>
        <CardContent className="p-4 flex gap-2">
          <Input placeholder="Nome do bairro" value={name} onChange={e => setName(e.target.value)} />
          <Button onClick={handleCreate}><Plus className="h-4 w-4 mr-1" /> Adicionar</Button>
        </CardContent>
      </Card>

      <Table>
        <TableHeader>
          <TableRow><TableHead>Nome</TableHead><TableHead>Slug</TableHead><TableHead>Ações</TableHead></TableRow>
        </TableHeader>
        <TableBody>
          {neighborhoods.map(n => (
            <TableRow key={n.id}>
              <TableCell>{n.name}</TableCell>
              <TableCell>{n.slug}</TableCell>
              <TableCell>
                <Button size="icon" variant="ghost" className="text-red-600" onClick={() => { deleteNeighborhood(n.id); toast.success('Bairro excluído'); }}><Trash2 className="h-4 w-4" /></Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function UsersTab() {
  const { users, addUser } = useAppStore();
  const [open, setOpen] = useState(false);

  const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const user = {
      id: Date.now(),
      name: formData.get('name') as string,
      email: formData.get('email') as string,
      role: formData.get('role') as 'admin' | 'editor' | 'comercial',
      active: true,
    };
    addUser(user as any);
    setOpen(false);
    toast.success('Usuário criado');
    form.reset();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Usuários</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild><Button><Plus className="h-4 w-4 mr-1" /> Novo Usuário</Button></DialogTrigger>
          <DialogContent className="max-w-lg">
            <DialogHeader><DialogTitle>Novo Usuário</DialogTitle></DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div><Label>Nome</Label><Input name="name" required /></div>
              <div><Label>Email</Label><Input name="email" type="email" required /></div>
              <div><Label>Perfil</Label>
                <Select name="role" defaultValue="editor">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="editor">Editor</SelectItem>
                    <SelectItem value="comercial">Comercial</SelectItem>
                    <SelectItem value="admin">Administrador</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Button type="submit" className="w-full">Salvar</Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>
      <Table>
        <TableHeader>
          <TableRow><TableHead>Nome</TableHead><TableHead>Email</TableHead><TableHead>Perfil</TableHead><TableHead>Status</TableHead></TableRow>
        </TableHeader>
        <TableBody>
          {users.map(u => (
            <TableRow key={u.id}>
              <TableCell>{u.name}</TableCell>
              <TableCell>{u.email}</TableCell>
              <TableCell className="capitalize">{u.role}</TableCell>
              <TableCell>{u.active ? <span className="text-green-600">Ativo</span> : <span className="text-slate-400">Inativo</span>}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function MetricsTab() {
  const { posts, stores, promotions } = useAppStore();
  const publishedPosts = posts.filter(p => p.status === 'published');
  const activeStores = stores.filter(s => s.active);
  const activePromos = promotions.filter(p => p.status === 'active');

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold text-slate-900">Métricas</h1>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardContent className="p-6">
            <h3 className="font-bold mb-4">Posts Publicados</h3>
            <p className="text-3xl font-bold">{publishedPosts.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <h3 className="font-bold mb-4">Lojas Ativas</h3>
            <p className="text-3xl font-bold">{activeStores.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <h3 className="font-bold mb-4">Promoções Ativas</h3>
            <p className="text-3xl font-bold">{activePromos.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <h3 className="font-bold mb-4">Taxa de Patrocínio</h3>
            <p className="text-3xl font-bold">
              {publishedPosts.length > 0 ? Math.round((publishedPosts.filter(p => p.isSponsored).length / publishedPosts.length) * 100) : 0}%
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
