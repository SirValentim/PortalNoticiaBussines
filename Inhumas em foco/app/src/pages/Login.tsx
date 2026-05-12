import { useState } from 'react';
import { useNavigate, Link } from 'react-router';
import { useAppStore } from '@/hooks/useAppStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Shield, AlertCircle } from 'lucide-react';
import { toast } from 'sonner';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { login } = useAppStore();
  const navigate = useNavigate();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    const user = login(email, password);
    if (user) {
      toast.success(`Bem-vindo, ${user.name}!`);
      navigate('/painel/7x9k2m');
    } else {
      setError('Credenciais inválidas');
    }
  };

  return (
    <div className="min-h-[80vh] flex items-center justify-center px-4">
      <div className="w-full max-w-md bg-white rounded-xl shadow-lg border p-8">
        <div className="text-center mb-6">
          <Shield className="h-10 w-10 text-blue-600 mx-auto mb-2" />
          <h1 className="text-2xl font-bold text-slate-900">Painel Administrativo</h1>
          <p className="text-sm text-slate-500">Acesso restrito a colaboradores</p>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4 flex items-center gap-2 text-sm text-red-700">
            <AlertCircle className="h-4 w-4" /> {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
              placeholder="admin@inhumasemfoco.com.br"
            />
          </div>
          <div>
            <Label htmlFor="password">Senha</Label>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              placeholder="••••••••"
            />
          </div>
          <Button type="submit" className="w-full">Entrar</Button>
        </form>

        <div className="mt-6 text-center text-xs text-slate-500">
          <p>Demo: admin@inhumasemfoco.com.br / admin123</p>
          <Link to="/" className="text-blue-600 hover:underline mt-2 block">Voltar para o site</Link>
        </div>
      </div>
    </div>
  );
}
