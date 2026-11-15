"use client";
import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  ArrowRight,
  ArrowUpRight,
  BookOpen,
  Building2,
  Check,
  GraduationCap,
  Mail,
  Plus,
  Users,
} from "lucide-react";
import toast from "react-hot-toast";
import { coreApi, errorMessage } from "@/api/client";
import { useRemote } from "@/lib/useRemote";
import { useAuthStore } from "@/store/authStore";
import { EmptyState, ErrorState, Loading, Modal, formatDate } from "./Shared";
interface AcademyData {
  academy: { id: number; name: string; slug: string; description: string };
  teachers: {
    id: number;
    user_id: number;
    name: string;
    email: string;
    active: boolean;
    course_count: number;
  }[];
  invitations: {
    id: number;
    email: string;
    accepted: boolean;
    expires_at: string;
  }[];
  stats: {
    courses: number;
    published: number;
    teachers: number;
    students: number;
  };
}
export function TenantSpace({
  view,
}: {
  view: "overview" | "academy" | "team";
}) {
  const { user } = useAuthStore();
  const { data, loading, error, reload } = useRemote<AcademyData>("/academy");
  const [invite, setInvite] = useState(false);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState("");
  const [query, setQuery] = useState("");
  const [now] = useState(() => Date.now());
  async function save(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const values = Object.fromEntries(new FormData(event.currentTarget));
    setBusy(true);
    setFormError("");
    try {
      await coreApi.put("/academy", values);
      toast.success("Academy details saved.");
      reload();
    } catch (error) {
      setFormError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  async function send(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const values = Object.fromEntries(new FormData(event.currentTarget));
    setBusy(true);
    setFormError("");
    try {
      await coreApi.post("/academy/invitations", values);
      toast.success("Teacher invitation sent to the local inbox.");
      setInvite(false);
      reload();
    } catch (error) {
      setFormError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  async function access(id: number, active: boolean) {
    setBusy(true);
    try {
      await coreApi.put(`/academy/teachers/${id}`, { active });
      toast.success(
        active
          ? "Teacher access restored."
          : "Teacher access removed. Their course records are preserved.",
      );
      reload();
    } catch (error) {
      toast.error(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  if (loading && !data) return <Loading label="Opening your academy…" />;
  if (error || !data)
    return (
      <ErrorState message={error || "Academy could not load."} retry={reload} />
    );
  const filtered = data.teachers.filter((t) =>
    `${t.name} ${t.email}`.toLowerCase().includes(query.toLowerCase()),
  );
  return (
    <>
      {view === "overview" ? (
        <>
          <div className="welcome-row">
            <div>
              <span className="eyebrow">ACADEMY OWNER WORKSPACE</span>
              <h1>
                Hello, {user?.first_name}
                <span className="welcome-star">✳</span>
              </h1>
              <p>
                Manage your academy, support your teachers, and see the whole
                picture.
              </p>
            </div>
            <Link className="btn btn-primary" href="/dashboard?tab=academy">
              <Building2 size={16} />
              Academy settings
            </Link>
          </div>
          <section className="tenant-hero">
            <div>
              <span className="eyebrow">YOUR ACADEMY, YOUR DIRECTION</span>
              <h2>{data.academy.name}</h2>
              <p>
                {data.academy.description ||
                  "Give your academy a clear identity, invite your teaching team, and help their ideas reach your students."}
              </p>
              <Link className="btn btn-lime" href="/dashboard?tab=team">
                Manage teaching team
                <ArrowUpRight size={16} />
              </Link>
            </div>
            <Building2 size={98} strokeWidth={1} aria-hidden="true" />
          </section>
          <div className="stats-grid">
            {[
              {
                label: "Academy courses",
                value: data.stats.courses,
                icon: BookOpen,
              },
              {
                label: "Published courses",
                value: data.stats.published,
                icon: Check,
              },
              {
                label: "Active teachers",
                value: data.stats.teachers,
                icon: GraduationCap,
              },
              {
                label: "Enrolled students",
                value: data.stats.students,
                icon: Users,
              },
            ].map(({ label, value, icon: Icon }) => (
              <section className="stat-card" key={label}>
                <div>
                  <span>{label}</span>
                  <span className="stat-icon">
                    <Icon size={18} />
                  </span>
                </div>
                <strong>{value}</strong>
                <p>Across your academy</p>
              </section>
            ))}
          </div>
          <div className="profile-grid">
            <section className="panel">
              <div className="panel-heading">
                <Building2 size={20} />
                <h2>Academy identity</h2>
              </div>
              <h3>{data.academy.name}</h3>
              <p className="muted" style={{ marginTop: 12 }}>
                {data.academy.description ||
                  "Add a description in academy settings."}
              </p>
              <Link
                href={`/${data.academy.slug}`}
                className="text-action"
                style={{ marginTop: 22 }}
              >
                Open public academy
                <ArrowUpRight size={16} />
              </Link>
            </section>
            <section className="panel">
              <div className="panel-heading">
                <GraduationCap size={20} />
                <h2>Your teaching team</h2>
              </div>
              <p className="muted">
                Owners manage academy settings and team access. Teachers build
                their own courses and review their students’ work.
              </p>
              <Link
                href="/dashboard?tab=team"
                className="text-action"
                style={{ marginTop: 22 }}
              >
                {data.stats.teachers
                  ? "View teachers and invitations"
                  : "Invite your first teacher"}
                <ArrowRight size={16} />
              </Link>
            </section>
          </div>
        </>
      ) : (
        <div className="page-heading">
          <div>
            <span className="eyebrow">ACADEMY OWNER</span>
            <h1>
              {view === "academy"
                ? "Make this academy your own."
                : "A team that helps people grow."}
            </h1>
            <p>
              {view === "academy"
                ? "Set the name, public address and introduction students will see."
                : "Invite teachers and manage their access to your academy."}
            </p>
          </div>
          {view === "team" && (
            <button
              className="btn btn-primary"
              onClick={() => {
                setFormError("");
                setInvite(true);
              }}
            >
              <Plus size={16} />
              Invite teacher
            </button>
          )}
        </div>
      )}
      {view === "academy" && (
        <section className="panel" style={{ maxWidth: 780 }}>
          <div className="panel-heading">
            <Building2 size={20} />
            <h2>Academy settings</h2>
          </div>
          <form className="stack-form" onSubmit={save}>
            {formError && (
              <div className="notice error" role="alert">
                {formError}
              </div>
            )}
            <label>
              Academy name
              <input
                name="name"
                defaultValue={data.academy.name}
                required
                maxLength={255}
              />
            </label>
            <label>
              Public academy address
              <input
                name="slug"
                defaultValue={data.academy.slug}
                required
                maxLength={100}
                pattern="[a-z0-9]+(-[a-z0-9]+)*"
              />
              <span className="field-help">
                Use lowercase letters, numbers and hyphens. Your public page
                opens at /{data.academy.slug}.
              </span>
            </label>
            <label>
              Academy description
              <textarea
                name="description"
                rows={5}
                defaultValue={data.academy.description}
                maxLength={5000}
                placeholder="What does your academy help people learn?"
              />
            </label>
            <div className="lesson-actions">
              <button className="btn btn-primary" disabled={busy}>
                {busy ? "Saving…" : "Save academy"}
                <Check size={16} />
              </button>
              <Link className="text-action" href={`/${data.academy.slug}`}>
                View public academy
                <ArrowUpRight size={16} />
              </Link>
            </div>
          </form>
        </section>
      )}
      {view === "team" && (
        <>
          <section className="panel student-panel">
            <div className="section-heading">
              <h2>Teaching team</h2>
              <label className="search-box">
                <Users size={16} />
                <input
                  aria-label="Search teachers"
                  placeholder="Search teachers…"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                />
              </label>
            </div>
            {!filtered.length ? (
              <EmptyState
                title={
                  query
                    ? "No teachers match this search."
                    : "Your team starts with an invitation."
                }
                text="Invited teachers join with their own account. Their access stays scoped to your academy."
              />
            ) : (
              <div className="table-scroll">
                <table className="students-table">
                  <thead>
                    <tr>
                      <th>Teacher</th>
                      <th>Courses</th>
                      <th>Access</th>
                      <th>Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((t) => (
                      <tr key={t.id}>
                        <td>
                          <strong>{t.name}</strong>
                          <p className="muted">{t.email}</p>
                        </td>
                        <td>{t.course_count}</td>
                        <td>
                          <span
                            className={`badge ${t.active ? "badge-green" : "badge-amber"}`}
                          >
                            {t.active ? "Active" : "Inactive"}
                          </span>
                        </td>
                        <td>
                          <button
                            className="text-action"
                            disabled={busy}
                            onClick={() => void access(t.id, !t.active)}
                          >
                            {t.active ? "Remove access" : "Restore access"}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
          <section className="panel" style={{ marginTop: 24 }}>
            <div className="panel-heading">
              <Mail size={19} />
              <h2>Teacher invitations</h2>
            </div>
            {!data.invitations.length ? (
              <p className="muted">
                Sent invitations appear here. Emails are delivered to your local
                inbox.
              </p>
            ) : (
              <div className="table-scroll">
                <table className="students-table">
                  <thead>
                    <tr>
                      <th>Email</th>
                      <th>Status</th>
                      <th>Expires</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.invitations.map((i) => (
                      <tr key={i.id}>
                        <td>{i.email}</td>
                        <td>
                          <span
                            className={`badge ${i.accepted ? "badge-green" : "badge-neutral"}`}
                          >
                            {i.accepted
                              ? "Accepted"
                              : Date.parse(i.expires_at) < now
                                ? "Expired"
                                : "Pending"}
                          </span>
                        </td>
                        <td>{formatDate(i.expires_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            <a
              className="text-action"
              href="http://localhost:8025"
              target="_blank"
              rel="noreferrer"
              style={{ marginTop: 20 }}
            >
              Open local email inbox
              <ArrowUpRight size={16} />
            </a>
          </section>
        </>
      )}
      {invite && (
        <Modal
          title="Invite a teacher."
          description="They’ll join your academy with their own teacher account. Owner controls stay with you."
          onClose={() => setInvite(false)}
        >
          <form onSubmit={send} className="stack-form">
            {formError && (
              <div className="notice error" role="alert">
                {formError}
              </div>
            )}
            <label>
              Teacher email
              <input
                name="email"
                type="email"
                required
                maxLength={255}
                placeholder="teacher@example.com"
              />
            </label>
            <button className="btn btn-primary full" disabled={busy}>
              {busy ? "Sending…" : "Send teacher invitation"}
              <Mail size={16} />
            </button>
          </form>
        </Modal>
      )}
    </>
  );
}
export function WorkspacePicker() {
  const router = useRouter();
  const refreshWorkspace = useAuthStore((state) => state.refreshWorkspace);
  const { data, error } =
    useRemote<{ id: number; name: string; current: boolean }[]>("/workspaces");
  const [busy, setBusy] = useState(false);
  async function select(value: string) {
    setBusy(true);
    try {
      await coreApi.put("/workspaces/current", { tenant_id: Number(value) });
      refreshWorkspace();
      router.replace("/dashboard");
      setBusy(false);
    } catch (error) {
      toast.error(errorMessage(error));
      setBusy(false);
    }
  }
  if (error) return <strong>Your academy workspace</strong>;
  if (!data?.length) return <strong>Loading academy…</strong>;
  if (data.length === 1)
    return <strong title={data[0].name}>{data[0].name}</strong>;
  return (
    <select
      className="academy-picker"
      aria-label="Active academy"
      disabled={busy}
      value={data.find((w) => w.current)?.id ?? data[0].id}
      onChange={(e) => void select(e.target.value)}
    >
      {data.map((w) => (
        <option key={w.id} value={w.id}>
          {w.name}
        </option>
      ))}
    </select>
  );
}
