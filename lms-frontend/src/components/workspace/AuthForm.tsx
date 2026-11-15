"use client";
import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  ArrowRight,
  BookOpen,
  CheckCircle2,
  Eye,
  EyeOff,
  GraduationCap,
  Loader2,
  Mail,
} from "lucide-react";
import { authApi, errorMessage } from "@/api/client";
import { useAuthStore } from "@/store/authStore";
import { Loading } from "./Shared";
import type { User } from "@/lib/types";
type Mode = "login" | "signup" | "forgot" | "reset";
function safeNext(value: string | null) {
  return value &&
    /^\/(?:dashboard(?:[/?]|$)|invite\/accept(?:\?|$))/.test(value) &&
    !value.includes("\\")
    ? value
    : "/dashboard";
}
function Form({ mode }: { mode: Mode }) {
  const router = useRouter();
  const params = useSearchParams();
  const { user, ready, setUser } = useAuthStore();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [show, setShow] = useState(false);
  const [role, setRole] = useState(
    params.get("role") === "tenant"
      ? 1
      : params.get("role") === "teacher"
        ? 2
        : 3,
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [sent, setSent] = useState(false);
  const [demo, setDemo] = useState(false);
  useEffect(() => {
    if (mode === "login")
      authApi
        .get("/config")
        .then(({ data }) => setDemo(data.demo_enabled))
        .catch(() => {});
  }, [mode]);
  useEffect(() => {
    if (ready && user && (mode === "login" || mode === "signup"))
      router.replace(safeNext(params.get("next")));
  }, [ready, user, mode, router, params]);
  async function signIn(loginEmail: string, loginPassword: string) {
    setBusy(true);
    setError("");
    try {
      const { data } = await authApi.post<{ user: User }>("/login", {
        email: loginEmail.trim().toLowerCase(),
        password: loginPassword,
      });
      setUser(data.user);
      router.replace(safeNext(params.get("next")));
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }
  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    const form = new FormData(event.currentTarget);
    if (mode === "login") {
      await signIn(email, password);
      return;
    }
    if ((mode === "signup" || mode === "reset") && password !== confirm) {
      setError("Your passwords don’t match. Please check and try again.");
      return;
    }
    if (
      (mode === "signup" || mode === "reset") &&
      new TextEncoder().encode(password).length > 72
    ) {
      setError("Please use a password of 72 bytes or fewer.");
      return;
    }
    setBusy(true);
    try {
      if (mode === "signup") {
        const { data } = await authApi.post<{ user: User }>("/register", {
          first_name: form.get("first_name"),
          last_name: form.get("last_name"),
          email: email.trim().toLowerCase(),
          password,
          role_id: role,
        });
        setUser(data.user);
        router.replace(safeNext(params.get("next")));
      } else if (mode === "forgot") {
        await authApi.post("/reset-password", {
          email: email.trim().toLowerCase(),
        });
        setSent(true);
      } else {
        await authApi.post("/reset-password/complete", {
          token: params.get("token"),
          new_password: password,
        });
        setUser(null);
        setSent(true);
      }
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }
  const copy = {
    login: [
      "GOOD TO HAVE YOU BACK",
      "Pick up where you left off.",
      "Sign in to your learning space. Your next idea is waiting.",
    ],
    signup: [
      "SOMETHING GOOD STARTS HERE",
      "Make room for what’s next.",
      "A new skill. A new perspective. A whole new possibility.",
    ],
    forgot: [
      "LET’S GET YOU BACK IN",
      "Forgot your password?",
      "It happens. Enter your email and we’ll send you a reset link.",
    ],
    reset: [
      "A FRESH START",
      "Choose a new password.",
      "Make it something secure, with at least 8 characters.",
    ],
  }[mode];
  if (sent)
    return (
      <div className="auth-success">
        <span className="success-icon">
          {mode === "forgot" ? <Mail size={28} /> : <CheckCircle2 size={28} />}
        </span>
        <h1>{mode === "forgot" ? "Check your inbox." : "You’re all set."}</h1>
        <p>
          {mode === "forgot"
            ? "If this email has an account, you’ll receive a link to reset your password. It’s valid for 30 minutes."
            : "Your password has been updated and previous sessions have been signed out. Sign in with your new password."}
        </p>
        {mode === "forgot" && (
          <a
            className="btn btn-primary full"
            href="http://localhost:8025"
            target="_blank"
            rel="noreferrer"
          >
            Open local inbox <ArrowUpRightIcon />
          </a>
        )}
        <Link href="/login" className="text-action">
          <ArrowLeft size={16} /> Back to sign in
        </Link>
      </div>
    );
  return (
    <>
      <span className="eyebrow">{copy[0]}</span>
      <h1 className="auth-title">{copy[1]}</h1>
      <p className="auth-description">{copy[2]}</p>
      {mode === "login" && params.get("verified") && (
        <div className="notice success" role="status">
          Your account is ready. Sign in to continue.
        </div>
      )}
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      {mode === "reset" && !params.get("token") ? (
        <div className="notice error">
          This link is incomplete.{" "}
          <Link href="/forgot-password">Request a new reset link.</Link>
        </div>
      ) : (
        <form onSubmit={submit} className="auth-form" aria-busy={busy}>
          <fieldset disabled={busy}>
            {mode === "signup" && (
              <>
                <div
                  className="role-choice"
                  role="group"
                  aria-label="Account type"
                >
                  <button
                    type="button"
                    className={role === 3 ? "selected" : ""}
                    onClick={() => setRole(3)}
                    aria-pressed={role === 3}
                  >
                    <BookOpen size={20} />
                    <span>
                      I’m here to learn<small>Explore something new</small>
                    </span>
                  </button>
                  <button
                    type="button"
                    className={role === 2 ? "selected" : ""}
                    onClick={() => setRole(2)}
                    aria-pressed={role === 2}
                  >
                    <GraduationCap size={20} />
                    <span>
                      I’m here to teach<small>Share what you know</small>
                    </span>
                  </button>
                  <button
                    type="button"
                    className={role === 1 ? "selected" : ""}
                    onClick={() => setRole(1)}
                    aria-pressed={role === 1}
                  >
                    <GraduationCap size={20} />
                    <span>
                      I run an academy
                      <small>Manage your academy and teaching team</small>
                    </span>
                  </button>
                </div>
                <div className="form-row">
                  <label>
                    First name
                    <input
                      name="first_name"
                      autoComplete="given-name"
                      placeholder="Jamie"
                      required
                      maxLength={100}
                    />
                  </label>
                  <label>
                    Last name
                    <input
                      name="last_name"
                      autoComplete="family-name"
                      placeholder="Singh"
                      required
                      maxLength={100}
                    />
                  </label>
                </div>
              </>
            )}
            {mode !== "reset" && (
              <label>
                Email address
                <input
                  name="email"
                  type="email"
                  autoComplete="email"
                  placeholder="you@example.com"
                  required
                  maxLength={255}
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                />
              </label>
            )}
            {mode !== "forgot" && (
              <>
                <label>
                  <span className="label-row">
                    {mode === "reset" ? "New password" : "Password"}
                    {mode === "login" && (
                      <Link href="/forgot-password">Forgot password?</Link>
                    )}
                  </span>
                  <div className="password-input">
                    <input
                      name="password"
                      type={show ? "text" : "password"}
                      autoComplete={
                        mode === "login" ? "current-password" : "new-password"
                      }
                      placeholder={
                        mode === "login"
                          ? "Enter your password"
                          : "At least 8 characters"
                      }
                      required
                      minLength={mode === "login" ? undefined : 8}
                      maxLength={72}
                      value={password}
                      onChange={(event) => setPassword(event.target.value)}
                    />
                    <button
                      type="button"
                      onClick={() => setShow(!show)}
                      aria-label={show ? "Hide password" : "Show password"}
                    >
                      {show ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </div>
                </label>
                {mode !== "login" && (
                  <label>
                    Confirm password
                    <input
                      name="confirm_password"
                      type={show ? "text" : "password"}
                      autoComplete="new-password"
                      placeholder="Enter it once more"
                      required
                      minLength={8}
                      maxLength={72}
                      value={confirm}
                      onChange={(event) => setConfirm(event.target.value)}
                    />
                  </label>
                )}
              </>
            )}
            <button
              className="btn btn-primary full auth-submit"
              type="submit"
              disabled={busy}
            >
              {busy ? (
                <>
                  <Loader2 size={17} className="spin" /> Just a moment…
                </>
              ) : (
                <>
                  {mode === "login"
                    ? "Sign in"
                    : mode === "signup"
                      ? "Create my account"
                      : mode === "forgot"
                        ? "Send reset link"
                        : "Update password"}
                  <ArrowRight size={17} />
                </>
              )}
            </button>
          </fieldset>
        </form>
      )}
      {mode === "login" && demo && (
        <div className="demo-options">
          <div className="divider-label">
            <span />
            TAKE A LOOK AROUND
            <span />
          </div>
          <div className="demo-buttons">
            <button
              disabled={busy}
              onClick={() =>
                void signIn("student@learnyst.local", "LocalStudent123!")
              }
            >
              <BookOpen size={17} /> Student demo <ArrowRight size={15} />
            </button>
            <button
              disabled={busy}
              onClick={() =>
                void signIn("teacher@learnyst.local", "LocalTeacher123!")
              }
            >
              <GraduationCap size={17} /> Teacher demo <ArrowRight size={15} />
            </button>
          </div>
          <p>Demo accounts are part of your local workspace.</p>
        </div>
      )}
      <p className="auth-switch">
        {mode === "login" ? (
          <>
            New here?{" "}
            <Link
              href={
                params.get("next")
                  ? `/signup?next=${encodeURIComponent(safeNext(params.get("next")))}`
                  : "/signup"
              }
            >
              Create an account <ArrowUpRightIcon />
            </Link>
          </>
        ) : mode === "signup" ? (
          <>
            Already have an account?{" "}
            <Link
              href={
                params.get("next")
                  ? `/login?next=${encodeURIComponent(safeNext(params.get("next")))}`
                  : "/login"
              }
            >
              Sign in <ArrowUpRightIcon />
            </Link>
          </>
        ) : (
          <Link href="/login">
            <ArrowLeft size={15} /> Back to sign in
          </Link>
        )}
      </p>
    </>
  );
}
function ArrowUpRightIcon() {
  return <ArrowRight size={15} style={{ transform: "rotate(-45deg)" }} />;
}
export function AuthForm({ mode }: { mode: Mode }) {
  return (
    <Suspense fallback={<Loading label="Getting things ready…" />}>
      <Form mode={mode} />
    </Suspense>
  );
}
