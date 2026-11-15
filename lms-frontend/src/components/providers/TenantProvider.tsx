'use client';
import { createContext, useContext } from 'react';
import { useParams } from 'next/navigation';
import { useRemote } from '@/lib/useRemote';
interface TenantConfig { academy_name: string; academy_slug: string; description: string }
interface TenantContextType { tenant: TenantConfig | null; loading: boolean; error: string | null }
const TenantContext = createContext<TenantContextType>({ tenant: null, loading: true, error: null });
export const useTenant = () => useContext(TenantContext);
export function TenantProvider({ children }: { children: React.ReactNode }) {
  const { slug } = useParams<{ slug?: string }>();
  const { data, loading, error } = useRemote<{ tenant: TenantConfig }>(slug ? `/tenants/slug/${encodeURIComponent(slug)}` : null);
  return <TenantContext.Provider value={{ tenant: data?.tenant || null, loading: Boolean(slug) && loading, error: error || null }}>{children}</TenantContext.Provider>;
}
