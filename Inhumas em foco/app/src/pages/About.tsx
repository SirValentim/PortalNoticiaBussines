import { Link } from 'react-router';
import { Badge } from '@/components/ui/badge';

export default function About() {
  return (
    <div className="max-w-3xl mx-auto px-4 py-12">
      <h1 className="text-3xl font-bold text-slate-900 mb-6">Sobre o Inhumas em Foco</h1>

      <div className="prose prose-slate max-w-none">
        <p className="text-lg text-slate-700 mb-4">
          O <strong>Inhumas em Foco</strong> é o portal de notícias e comércio local de Inhumas, Goiás.
          Nosso objetivo é manter a comunidade informada sobre tudo que acontece na cidade,
          desde política e bastidores até eventos, promoções e novidades do comércio local.
        </p>

        <p className="text-slate-700 mb-4">
          Fundado com a missão de fortalecer a comunicação local e valorizar os empreendedores
          de Inhumas, o portal oferece espaço para notícias, divulgação de lojas, promoções,
          eventos e guia de bairros.
        </p>

        <h2 className="text-2xl font-bold text-slate-900 mt-8 mb-4">Nossos Pilares</h2>
        <ul className="list-disc list-inside text-slate-700 space-y-2 mb-6">
          <li>Jornalismo local independente e responsável</li>
          <li>Valorização do comércio e empreendedorismo local</li>
          <li>Transparência e ética na informação</li>
          <li>Compromisso com a comunidade inhumense</li>
        </ul>

        <h2 className="text-2xl font-bold text-slate-900 mt-8 mb-4">Monetização e Parcerias</h2>
        <p className="text-slate-700 mb-4">
          O portal é sustentado por meio de publicidade local, parcerias comerciais e
          conteúdo patrocinado. Anunciantes locais podem divulgar suas lojas, promoções
          e eventos de forma direcionada ao público inhumense.
        </p>

        <div className="bg-slate-50 rounded-xl p-6 mt-8">
          <h3 className="font-bold text-lg mb-4">Produtos Disponíveis</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[
              { name: 'Banner Hero', price: 'R$ 300-400/mês', desc: 'Topo da home, full-width' },
              { name: 'Vitrine de Lojas', price: 'R$ 80-150/loja/mês', desc: 'Destaque na home' },
              { name: 'Promoção do Dia', price: 'R$ 30-50/promo', desc: 'Bloco com countdown' },
              { name: 'Conteúdo Patrocinado', price: 'R$ 150-300/matéria', desc: 'Layout editorial + selo' },
            ].map(item => (
              <div key={item.name} className="bg-white rounded-lg p-4 border">
                <div className="flex items-center justify-between mb-1">
                  <span className="font-bold">{item.name}</span>
                  <Badge variant="secondary">{item.price}</Badge>
                </div>
                <p className="text-sm text-slate-500">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>

        <h2 className="text-2xl font-bold text-slate-900 mt-8 mb-4">Fale Conosco</h2>
        <p className="text-slate-700">
          Para anunciar, enviar pautas ou entrar em contato, utilize nossa{' '}
          <Link to="/contato" className="text-blue-600 hover:underline">página de contato</Link>.
        </p>
      </div>
    </div>
  );
}
