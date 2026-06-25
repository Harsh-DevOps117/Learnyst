'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { authApi, coreApi } from '@/api/client';
import { useAuthStore } from '@/store/authStore';
import Link from 'next/link';

export default function TenantLoginPage() {
  const router = useRouter();
  const { slug } = useParams();

  const [tenant, setTenant] = useState<any>(null);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const setAuth = useAuthStore((state) => state.setAuth);

  useEffect(() => {
    // Fetch tenant details for branding
    coreApi.get(`/tenants/slug/${slug}`).then((res) => {
      setTenant(res.data.tenant);
    }).catch(err => {
      console.error("Failed to load tenant details", err);
    });
  }, [slug]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await authApi.post('/login', {
        email,
        password,
        device_name: "Web Browser"
      });

      const { access_token, user_id, role_id } = response.data;
      setAuth(access_token, role_id || 1, user_id);
      
      // Redirect to the tenant's student dashboard instead of the main dashboard
      router.push(`/${slug}/dashboard`);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Invalid credentials.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#fafafa] flex flex-col justify-center py-12 sm:px-6 lg:px-8 font-ui">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <div className="flex justify-center mb-6">
          {tenant?.logo_url ? (
            <img src={tenant.logo_url} alt="Logo" className="h-12 w-auto object-contain" />
          ) : (
            <span className="font-display text-[28px] font-bold text-gray-900 tracking-tight">
              {tenant?.academy_name || 'Loading...'}
            </span>
          )}
        </div>
        <h2 className="text-center text-3xl font-display font-extrabold text-gray-900">
          Sign in to your account
        </h2>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10 border border-gray-200">
          <form className="space-y-6" onSubmit={handleSubmit}>
            {error && (
              <div className="bg-red-50 border-l-4 border-red-400 p-4">
                <p className="text-sm text-red-700">{error}</p>
              </div>
            )}
            
            <div>
              <label className="block text-sm font-medium text-gray-700">Email address</label>
              <div className="mt-1">
                <Input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="w-full"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Password</label>
              <div className="mt-1">
                <Input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  className="w-full"
                />
              </div>
            </div>

            <Button type="submit" className="w-full bg-black text-white hover:bg-gray-800" disabled={loading}>
              {loading ? 'Signing in...' : 'Sign in'}
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              Don't have an account?{' '}
              <Link href={`/${slug}/signup`} className="font-medium text-black hover:underline">
                Sign up
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
