import { Link } from 'react-router';

export default function Footer() {
  return (
    <footer className="bg-slate-900 text-slate-400 py-10 mt-auto">
      <div className="max-w-7xl mx-auto px-4 grid grid-cols-1 md:grid-cols-3 gap-8">
        <div>
          <h4 className="text-white font-bold text-lg mb-3">Inhumas em Foco</h4>
          <p className="text-sm">Portal de notícias e comércio local de Inhumas, Goiás. Mantendo a comunidade informada e conectada.</p>
        </div>
        <div>
          <h4 className="text-white font-bold text-lg mb-3">Navegação</h4>
          <div className="flex flex-col gap-2 text-sm">
            <Link to="/" className="hover:text-amber-400 transition">Home</Link>
            <Link to="/lojas" className="hover:text-amber-400 transition">Lojas</Link>
            <Link to="/promocoes" className="hover:text-amber-400 transition">Promoções</Link>
            <Link to="/busca" className="hover:text-amber-400 transition">Busca</Link>
            <Link to="/sobre" className="hover:text-amber-400 transition">Sobre</Link>
            <Link to="/contato" className="hover:text-amber-400 transition">Contato</Link>
          </div>
        </div>
        <div>
          <h4 className="text-white font-bold text-lg mb-3">Legal</h4>
          <p className="text-sm">&copy; 2026 Inhumas em Foco. Todos os direitos reservados.</p>
          <p className="text-sm mt-2">contato@inhumasemfoco.com.br</p>
        </div>
      </div>
    </footer>
  );
}
