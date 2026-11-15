"use client";
import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import {
  ArrowUpRight,
  BookOpen,
  Compass,
  GraduationCap,
  LayoutDashboard,
  LogOut,
  Menu,
  PanelLeftClose,
  Settings2,
  Shapes,
  Sparkles,
  Users,
  X,
} from "lucide-react";
import { useAuthStore } from "@/store/authStore";
import { authApi, errorMessage } from "@/api/client";
import { Brand } from "@/components/workspace/Brand";
import { Loading } from "@/components/workspace/Shared";
import toast from "react-hot-toast";
import { WorkspacePicker } from "@/components/workspace/TenantSpace";
function Shell({ children }: { children: React.ReactNode }) {
  const { user, ready, setUser } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();
  const params = useSearchParams();
  const [mobile, setMobile] = useState(false);
  const [signingOut, setSigningOut] = useState(false);
  const owner = user?.role_id === 1;
  const teacher = owner || user?.role_id === 2;
  const roleLabel = owner
    ? "Tenant / academy owner"
    : teacher
      ? "Teacher"
      : "Student";
  const tab = params.get("tab") || "overview";
  const nav = owner
    ? [
        { id: "overview", title: "Overview", icon: LayoutDashboard },
        { id: "academy", title: "Academy settings", icon: Settings2 },
        { id: "team", title: "Teaching team", icon: GraduationCap },
        { id: "courses", title: "Academy courses", icon: BookOpen },
        { id: "students", title: "Students", icon: Users },
      ]
    : teacher
      ? [
          { id: "overview", title: "Overview", icon: LayoutDashboard },
          { id: "courses", title: "My courses", icon: BookOpen },
          { id: "students", title: "Students", icon: Users },
          { id: "assignments", title: "Assignments", icon: Shapes },
        ]
      : [
          { id: "overview", title: "Overview", icon: LayoutDashboard },
          { id: "courses", title: "My learning", icon: BookOpen },
          { id: "explore", title: "Explore courses", icon: Compass },
          { id: "assignments", title: "Assignments", icon: Shapes },
        ];
  useEffect(() => {
    if (ready && !user && !signingOut)
      router.replace(
        `/login?next=${encodeURIComponent(pathname + (params.toString() ? "?" + params.toString() : ""))}`,
      );
  }, [ready, user, signingOut, router, pathname, params]);
  async function logout() {
    setSigningOut(true);
    try {
      await authApi.post("/logout");
      setUser(null);
      router.replace("/login");
    } catch (error) {
      setSigningOut(false);
      toast.error(errorMessage(error));
    }
  }
  if (!ready || !user)
    return (
      <div className="full-loading">
        <Brand />
        <Loading label="Opening your learning space…" />
      </div>
    );
  const title = pathname.includes("/courses/")
    ? "Course workspace"
    : tab === "profile"
      ? "Your profile"
      : nav.find((item) => item.id === tab)?.title || "Overview";
  return (
    <div className="workspace">
      <a className="skip-link" href="#workspace-main">
        Skip to content
      </a>
      {mobile && (
        <button
          className="sidebar-backdrop"
          aria-label="Close navigation"
          onClick={() => setMobile(false)}
        />
      )}
      <aside className={`sidebar ${mobile ? "sidebar-open" : ""}`}>
        <div className="sidebar-brand">
          <Brand light href="/dashboard" />
          <button
            className="icon-button mobile-only"
            onClick={() => setMobile(false)}
            aria-label="Close navigation"
          >
            <X size={20} />
          </button>
        </div>
        <div className="workspace-selector">
          <span className="workspace-selector-icon">
            <GraduationCap size={19} />
          </span>
          <div>
            {teacher ? (
              <WorkspacePicker />
            ) : (
              <strong>Your learning space</strong>
            )}
            <span>{roleLabel + " workspace"}</span>
          </div>
          <span className="selector-dot" />
        </div>
        <span className="nav-caption">WORKSPACE</span>
        <nav aria-label="Workspace navigation">
          {nav.map(({ id, title, icon: Icon }) => (
            <Link
              key={id}
              href={`/dashboard?tab=${id}`}
              onClick={() => setMobile(false)}
              className={`nav-item ${tab === id && !pathname.includes("/courses/") ? "nav-active" : ""}`}
              aria-current={
                tab === id && !pathname.includes("/courses/")
                  ? "page"
                  : undefined
              }
            >
              <Icon size={18} strokeWidth={1.7} />
              {title}
              {tab === id && <span className="nav-active-dot" />}
            </Link>
          ))}
        </nav>
        <div className="sidebar-note">
          <span className="sidebar-note-icon">
            <Sparkles size={21} />
          </span>
          <strong>
            {teacher
              ? "Your knowledge. Their next step."
              : "A little progress, every day."}
          </strong>
          <p>
            {teacher
              ? "Great teaching starts with a little inspiration."
              : "Make time for one good idea today."}
          </p>
          <Link
            href={
              teacher
                ? "/dashboard?tab=courses&action=create_course"
                : "/dashboard?tab=explore"
            }
            onClick={() => setMobile(false)}
          >
            {teacher ? "Create something new" : "Find your next course"}
            <ArrowUpRight size={16} />
          </Link>
        </div>
        <div className="sidebar-bottom">
          <Link
            href="/dashboard?tab=profile"
            onClick={() => setMobile(false)}
            className={`nav-item ${tab === "profile" ? "nav-active" : ""}`}
          >
            <Settings2 size={18} /> Profile & settings
          </Link>
          <div className="sidebar-user">
            <Link
              href="/dashboard?tab=profile"
              className="user-identity"
              onClick={() => setMobile(false)}
            >
              <span className="avatar">
                {user.first_name[0]}
                {user.last_name[0]}
              </span>
              <div>
                <strong>
                  {user.first_name} {user.last_name}
                </strong>
                <span>{roleLabel}</span>
              </div>
            </Link>
            <button
              className="icon-button"
              onClick={() => void logout()}
              disabled={signingOut}
              aria-label="Sign out"
              title="Sign out"
            >
              <LogOut size={18} />
            </button>
          </div>
        </div>
      </aside>
      <div className="workspace-body">
        <header className="workspace-header">
          <div className="header-breadcrumb">
            <button
              className="icon-button mobile-only"
              onClick={() => setMobile(true)}
              aria-label="Open navigation"
            >
              <Menu size={21} />
            </button>
            <PanelLeftClose size={18} className="desktop-only" />
            <span className="desktop-only">Workspace</span>
            <span className="breadcrumb-slash desktop-only">/</span>
            <strong>{title}</strong>
          </div>
          <div className="header-right">
            <span className="header-role">
              <span />
              {roleLabel} workspace
            </span>
            <Link
              className="avatar avatar-small"
              href="/dashboard?tab=profile"
              aria-label="Open your profile"
            >
              {user.first_name[0]}
              {user.last_name[0]}
            </Link>
          </div>
        </header>
        <main id="workspace-main" className="workspace-content">
          {children}
        </main>
        <footer className="workspace-footer">
          <span>Little by little, a lot becomes possible.</span>
          <span>
            Made for curious minds <span className="footer-spark">✳</span>
          </span>
        </footer>
      </div>
    </div>
  );
}
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <Suspense fallback={<Loading />}>
      <Shell>{children}</Shell>
    </Suspense>
  );
}
