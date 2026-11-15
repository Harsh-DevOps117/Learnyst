"use client";
import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { ArrowRight, MailOpen } from "lucide-react";
import { coreApi, errorMessage } from "@/api/client";
import { useAuthStore } from "@/store/authStore";
import { Brand } from "@/components/workspace/Brand";
import { Loading } from "@/components/workspace/Shared";
function Invitation() {
  const params = useSearchParams();
  const token = params.get("token");
  const router = useRouter();
  const { user, ready } = useAuthStore();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function accept() {
    setBusy(true);
    setError("");
    try {
      const { data } = await coreApi.post<{
        course_id?: number;
        redirect?: string;
      }>("/invitations/accept", { token });
      router.replace(data.redirect || `/dashboard/courses/${data.course_id}`);
    } catch (error) {
      setError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  const next = `/invite/accept?token=${encodeURIComponent(token || "")}`;
  return (
    <main className="full-loading">
      <Brand />
      <section
        className="panel"
        style={{ width: "min(440px, calc(100vw - 40px))", marginTop: 35 }}
      >
        <MailOpen size={30} color="currentColor" />
        <h1 style={{ fontSize: 29, margin: "20px 0 12px" }}>
          You’re invited to join.
        </h1>
        <p className="muted">
          Sign in with the email and account type your invitation was sent to,
          then join your course or teaching team.
        </p>
        {error && (
          <div className="notice error" role="alert">
            {error}
          </div>
        )}
        {!token || !/^[a-f0-9]{64}$/.test(token) ? (
          <div className="notice error">
            This invitation link is incomplete. Ask the sender for a new one.
          </div>
        ) : !ready ? (
          <Loading />
        ) : user ? (
          <>
            <p className="muted" style={{ marginTop: 20 }}>
              Signed in as {user.email}
            </p>
            <button
              className="btn btn-primary full"
              style={{ marginTop: 20 }}
              disabled={busy}
              onClick={() => void accept()}
            >
              {busy ? "Joining…" : "Accept invitation"}
              <ArrowRight size={16} />
            </button>
          </>
        ) : (
          <Link
            className="btn btn-primary full"
            style={{ marginTop: 25 }}
            href={`/login?next=${encodeURIComponent(next)}`}
          >
            Sign in to join
            <ArrowRight size={16} />
          </Link>
        )}
      </section>
    </main>
  );
}
export default function Page() {
  return (
    <Suspense fallback={<Loading />}>
      <Invitation />
    </Suspense>
  );
}
