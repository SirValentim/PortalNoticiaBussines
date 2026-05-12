import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { toast } from 'sonner';

export default function Contact() {
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
    toast.success('Mensagem enviada com sucesso!');
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-12">
      <h1 className="text-3xl font-bold text-slate-900 mb-2">Contato</h1>
      <p className="text-slate-600 mb-8">
        Entre em contato conosco para anunciar, enviar sugestões de pauta ou para qualquer outra informação.
      </p>

      {submitted ? (
        <div className="bg-green-50 border border-green-200 rounded-xl p-8 text-center">
          <h2 className="text-xl font-bold text-green-800 mb-2">Mensagem enviada!</h2>
          <p className="text-green-700">Agradecemos o contato. Responderemos em breve.</p>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border p-8 space-y-6">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Nome</label>
            <Input required />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Email</label>
            <Input type="email" required />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Assunto</label>
            <Select defaultValue="anuncio">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="anuncio">Quero anunciar</SelectItem>
                <SelectItem value="pauta">Sugestão de pauta</SelectItem>
                <SelectItem value="denuncia">Denúncia</SelectItem>
                <SelectItem value="outro">Outro</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Mensagem</label>
            <Textarea rows={5} required />
          </div>
          <Button type="submit" className="w-full">Enviar mensagem</Button>
        </form>
      )}

      <div className="mt-8 text-center text-sm text-slate-500">
        <p>Email: contato@inhumasemfoco.com.br</p>
        <p className="mt-1">WhatsApp: (62) 9XXXX-XXXX</p>
      </div>
    </div>
  );
}
