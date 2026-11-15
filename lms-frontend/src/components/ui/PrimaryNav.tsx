'use client';
import Link from 'next/link';
import { Brand } from '@/components/workspace/Brand';
import { useAuthStore } from '@/store/authStore';
export function PrimaryNav() {
  const { user } = useAuthStore();
  return <header className="landing-nav"><Brand /><div>{user ? <Link className="btn btn-primary" href="/dashboard">Your workspace</Link> : <><Link href="/login">Sign in</Link><Link className="btn btn-primary" href="/signup">Get started</Link></>}</div></header>;
}
