import { Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Search, Menu, X, Shield } from 'lucide-react';
import { useState } from 'react';

export default function Header() {
  const { currentUser, logout } = useAppStore();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      window.location.href = `/busca?q=${encodeURIComponent(searchQuery)}`;
    }
  };

  return (
    <header className="bg-slate-900 text-white sticky top-0 z-50 shadow-md">
      <div className="max-w-7xl mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <Link to="/" className="text-xl font-bold tracking-tight">
            Inhumas <span className="text-amber-400">em Foco</span>
          </Link>

          <nav className="hidden md:flex items-center gap-6">
            <Link to="/" className="text-sm font-medium hover:text-amber-400 transition">Home</Link>
            <Link to="/lojas" className="text-sm font-medium hover:text-amber-400 transition">Lojas</Link>
            <Link to="/promocoes" className="text-sm font-medium hover:text-amber-400 transition">Promoções</Link>
            <Link to="/busca" className="text-sm font-medium hover:text-amber-400 transition">Busca</Link>
          </nav>

          <div className="hidden md:flex items-center gap-3">
            <form onSubmit={handleSearch} className="flex items-center">
              <Input
                type="text"
                placeholder="Buscar..."
                className="w-40 h-8 bg-slate-800 border-slate-700 text-white placeholder:text-slate-400"
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
              />
              <Button type="submit" size="icon" variant="ghost" className="h-8 w-8 text-slate-400 hover:text-white">
                <Search className="h-4 w-4" />
              </Button>
            </form>
            {currentUser ? (
              <div className="flex items-center gap-2">
                <Link to="/painel/7x9k2m">
                  <Button size="sm" variant="outline" className="border-amber-400 text-amber-400 hover:bg-amber-400 hover:text-slate-900">
                    <Shield className="h-3 w-3 mr-1" /> Painel
                  </Button>
                </Link>
                <Button size="sm" variant="ghost" onClick={logout} className="text-slate-400 hover:text-white">Sair</Button>
              </div>
            ) : (
              <Link to="/login">
                <Button size="sm" variant="ghost" className="text-slate-300 hover:text-white">Entrar</Button>
              </Link>
            )}
          </div>

          <button className="md:hidden text-white" onClick={() => setMobileOpen(!mobileOpen)}>
            {mobileOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </button>
        </div>
      </div>

      {mobileOpen && (
        <div className="md:hidden bg-slate-800 border-t border-slate-700 px-4 py-4 space-y-3">
          <Link to="/" className="block text-sm font-medium py-2" onClick={() => setMobileOpen(false)}>Home</Link>
          <Link to="/lojas" className="block text-sm font-medium py-2" onClick={() => setMobileOpen(false)}>Lojas</Link>
          <Link to="/promocoes" className="block text-sm font-medium py-2" onClick={() => setMobileOpen(false)}>Promoções</Link>
          <Link to="/busca" className="block text-sm font-medium py-2" onClick={() => setMobileOpen(false)}>Busca</Link>
          {currentUser && (
            <Link to="/painel/7x9k2m" className="block text-sm font-medium py-2 text-amber-400" onClick={() => setMobileOpen(false)}>Painel</Link>
          )}
          <form onSubmit={handleSearch} className="flex gap-2">
            <Input
              type="text"
              placeholder="Buscar..."
              className="bg-slate-700 border-slate-600 text-white"
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
            />
            <Button type="submit" size="sm">Buscar</Button>
          </form>
        </div>
      )}
    </header>
  );
}
