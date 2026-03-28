import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ShieldCheck } from 'lucide-react';
import { useAuth } from '../auth/AuthProvider';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';

export default function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({ username: 'customer', password: 'Customer1234!', totpCode: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const submit = async (event) => {
    event.preventDefault();
    setLoading(true);
    setError('');
    try {
      await login(form);
      navigate('/overview', { replace: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen grid place-items-center bg-[linear-gradient(130deg,#020617,#0f172a,#134e4a)] p-4">
      <form className="w-full max-w-md space-y-4 rounded-2xl border border-slate-700 bg-slate-950/70 p-6" onSubmit={submit}>
        <div className="flex items-center gap-2">
          <ShieldCheck className="h-5 w-5 text-cyan-300" />
          <h1 className="text-2xl font-semibold">Sign In</h1>
        </div>
        <Input placeholder="Username" value={form.username} onChange={(e) => setForm((prev) => ({ ...prev, username: e.target.value }))} />
        <Input type="password" placeholder="Password" value={form.password} onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))} />
        <Input placeholder="TOTP Code (optional)" value={form.totpCode} onChange={(e) => setForm((prev) => ({ ...prev, totpCode: e.target.value }))} />
        <Button type="submit" className="w-full" disabled={loading}>{loading ? 'Signing in...' : 'Sign In'}</Button>
        {error && <p className="text-sm text-rose-300">{error}</p>}
      </form>
    </div>
  );
}
