import Link from "next/link";
import { Brand } from "@/components/workspace/Brand";
export default function NotFound() {
  return (
    <main className="full-loading">
      <Brand />
      <div className="empty-state" style={{ marginTop: 35 }}>
        <span className="eyebrow">A LITTLE DETOUR</span>
        <h1 style={{ marginTop: 15 }}>This page took a different path.</h1>
        <p>Let’s get you back to somewhere familiar.</p>
        <Link href="/dashboard" className="btn btn-primary">
          Back to your workspace
        </Link>
      </div>
    </main>
  );
}
