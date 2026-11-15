"use client";
import { Suspense, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  ArrowDownWideNarrow,
  ArrowRight,
  ArrowUpRight,
  BookOpen,
  CheckCheck,
  ClipboardCheck,
  GraduationCap,
  Layers3,
  Plus,
  Search,
  Sparkles,
  Target,
  Users,
} from "lucide-react";
import toast from "react-hot-toast";
import { coreApi, errorMessage } from "@/api/client";
import { useAuthStore } from "@/store/authStore";
import { useRemote } from "@/lib/useRemote";
import type { Course, Student } from "@/lib/types";
import {
  CourseCard,
  EmptyState,
  ErrorState,
  Loading,
  Modal,
  formatDate,
} from "@/components/workspace/Shared";
import { Assignments } from "@/components/workspace/Assignments";
import { Profile } from "@/components/workspace/Profile";
import { TenantSpace } from "@/components/workspace/TenantSpace";
import { StudentOverview } from "@/components/workspace/StudentOverview";
import { TeacherOverview } from "@/components/workspace/TeacherOverview";
function CreateCourse({ onClose }: { onClose: () => void }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function create(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = Object.fromEntries(new FormData(event.currentTarget));
    setBusy(true);
    setError("");
    try {
      const { data: course } = await coreApi.post<Course>("/courses", data);
      toast.success("A new course, a new possibility.");
      router.push(`/dashboard/courses/${course.id}`);
      onClose();
    } catch (error) {
      setError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  return (
    <Modal
      title="What will you teach?"
      description="Start with an idea. You can shape the rest as you go."
      onClose={onClose}
    >
      <form onSubmit={create} className="stack-form">
        {error && (
          <div className="notice error" role="alert">
            {error}
          </div>
        )}
        <label>
          Course title
          <input
            name="title"
            placeholder="e.g. A practical guide to TypeScript"
            required
            maxLength={255}
            autoFocus
          />
        </label>
        <label>
          A little about the course
          <textarea
            name="description"
            rows={4}
            placeholder="What will your students learn?"
            maxLength={10000}
          />
        </label>
        <div className="form-row">
          <label>
            Category
            <select name="category">
              <option>Development</option>
              <option>Design</option>
              <option>DevOps</option>
              <option>Business</option>
              <option>Other</option>
            </select>
          </label>
          <label>
            Level
            <select name="level">
              <option>Beginner</option>
              <option>Intermediate</option>
              <option>Advanced</option>
            </select>
          </label>
        </div>
        <p className="field-help">
          Your course starts as a draft. Add your lessons before publishing.
        </p>
        <button className="btn btn-primary full" disabled={busy}>
          {busy ? "Creating…" : "Create course"}
          <ArrowRight size={16} />
        </button>
      </form>
    </Modal>
  );
}
function StudentsView({
  students,
  loading,
  error,
  retry,
}: {
  students: Student[];
  loading: boolean;
  error: string;
  retry: () => void;
}) {
  const [query, setQuery] = useState("");
  const filtered = students.filter((s) =>
    `${s.first_name} ${s.last_name} ${s.email}`
      .toLowerCase()
      .includes(query.toLowerCase()),
  );
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">THE PEOPLE YOU INSPIRE</span>
          <h1>A space to grow, together.</h1>
          <p>Your students and the learning journeys you’re part of.</p>
        </div>
        <span className="count-pill">
          <Users size={17} />
          {students.length} {students.length === 1 ? "student" : "students"}
        </span>
      </div>
      <div className="panel student-panel">
        <div className="section-heading">
          <h2>Your students</h2>
          <label className="search-box">
            <Search size={17} />
            <input
              aria-label="Search students"
              placeholder="Search students…"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </label>
        </div>
        {loading ? (
          <Loading />
        ) : error ? (
          <ErrorState message={error} retry={retry} />
        ) : !filtered.length ? (
          <EmptyState
            title={
              query
                ? "No matches just yet."
                : "Your first student is on the way."
            }
            text={
              query
                ? "Try another name or email."
                : "Publish a course so students can discover and enroll in it."
            }
          />
        ) : (
          <div className="table-scroll">
            <table className="students-table">
              <thead>
                <tr>
                  <th>Student</th>
                  <th>Enrolled courses</th>
                  <th>Joined your courses</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((s) => (
                  <tr key={s.id}>
                    <td>
                      <div className="table-user">
                        <span className="avatar">
                          {s.first_name[0]}
                          {s.last_name[0]}
                        </span>
                        <div>
                          <strong>
                            {s.first_name} {s.last_name}
                          </strong>
                          <span>{s.email}</span>
                        </div>
                      </div>
                    </td>
                    <td>
                      {s.course_count}{" "}
                      {s.course_count === 1 ? "course" : "courses"}
                    </td>
                    <td>{formatDate(s.joined_at)}</td>
                    <td>
                      <span className="badge badge-green">
                        <i />
                        Enrolled
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </>
  );
}
function Dashboard() {
  const { user } = useAuthStore();
  const params = useSearchParams();
  const router = useRouter();
  const teacher = user?.role_id === 1 || user?.role_id === 2;
  const rawTab = params.get("tab") || "overview";
  const tab = [
    "overview",
    "academy",
    "team",
    "courses",
    "explore",
    "assignments",
    "students",
    "profile",
  ].includes(rawTab)
    ? rawTab
    : "overview";
  const { data, loading, error, reload } = useRemote<Course[]>(
    user ? `/courses${tab === "explore" ? "?scope=catalog" : ""}` : null,
  );
  const students = useRemote<Student[]>(teacher ? "/students" : null);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState("all");
  const [sort, setSort] = useState("recent");
  const [creating, setCreating] = useState(false);
  const [enrolling, setEnrolling] = useState<number | null>(null);
  const courses = data || [];
  const createOpen =
    teacher && (creating || params.get("action") === "create_course");
  function closeCreate() {
    setCreating(false);
    if (params.get("action")) router.replace("/dashboard?tab=courses");
  }
  async function enroll(course: Course) {
    setEnrolling(course.id);
    try {
      await coreApi.post(`/courses/${course.id}/enroll`);
      toast.success("You’re in. Your next chapter starts now.");
      router.push(`/dashboard/courses/${course.id}`);
    } catch (error) {
      toast.error(errorMessage(error));
    } finally {
      setEnrolling(null);
    }
  }
  if (
    user?.role_id === 1 &&
    (tab === "overview" || tab === "academy" || tab === "team")
  )
    return <TenantSpace view={tab} />;
  if (tab === "profile") return <Profile />;
  if (tab === "assignments")
    return <Assignments teacher={teacher} courses={courses} />;
  if (tab === "students" && teacher)
    return (
      <StudentsView
        students={students.data || []}
        loading={students.loading}
        error={students.error}
        retry={students.reload}
      />
    );
  const published = courses.filter((c) => c.is_published).length;
  const complete = courses.filter((c) => c.progress >= 90).length;
  const lessons = courses.reduce((sum, c) => sum + c.lesson_count, 0);
  const finished = courses.reduce((sum, c) => sum + c.completed_lessons, 0);
  const stats = teacher
    ? [
        {
          label: "Your courses",
          value: courses.length,
          note: "Ideas taking shape",
          icon: BookOpen,
        },
        {
          label: "Published",
          value: published,
          note: "Ready to inspire",
          icon: Layers3,
        },
        {
          label: "Your students",
          value: students.data?.length ?? 0,
          note: "Learning with you",
          icon: Users,
        },
        {
          label: "In the making",
          value: courses.length - published,
          note: "Drafts with potential",
          icon: Sparkles,
        },
      ]
    : [
        {
          label: "Enrolled courses",
          value: courses.length,
          note: "Possibilities to explore",
          icon: BookOpen,
        },
        {
          label: "Lessons completed",
          value: finished,
          note: "Every step counts",
          icon: CheckCheck,
        },
        {
          label: "Courses completed",
          value: complete,
          note: "Something to be proud of",
          icon: GraduationCap,
        },
        {
          label: "Still to discover",
          value: Math.max(0, lessons - finished),
          note: "Your next little step",
          icon: Target,
        },
      ];
  let filtered = courses.filter((c) =>
    `${c.title} ${c.category} ${c.instructor}`
      .toLowerCase()
      .includes(query.toLowerCase()),
  );
  if (tab !== "overview" && tab !== "explore" && filter !== "all")
    filtered = filtered.filter((c) =>
      teacher
        ? filter === "published"
          ? c.is_published
          : !c.is_published
        : filter === "completed"
          ? c.progress >= 90
          : c.progress < 90,
    );
  if (sort === "title")
    filtered = [...filtered].sort((a, b) => a.title.localeCompare(b.title));
  if (tab === "overview") filtered = filtered.slice(0, 3);
  const firstCourse = courses.find((c) =>
    teacher ? !c.is_published : c.progress < 90,
  );
  return (
    <>
      {tab === "overview" ? (
        <>
          <div className="welcome-row">
            <div>
              <span className="eyebrow">
                YOUR SPACE TO {teacher ? "INSPIRE" : "GROW"}
              </span>
              <h1>
                Hello, {user?.first_name}
                <span className="welcome-star">✳</span>
              </h1>
              <p>
                {teacher
                  ? "Good things happen when you share what you know."
                  : "A little focus today. A little more possibility tomorrow."}
              </p>
            </div>
            <span className="date-label">
              {new Date().toLocaleDateString("en-IN", {
                weekday: "long",
                day: "numeric",
                month: "long",
              })}
            </span>
          </div>
          {teacher ? (
            <TeacherOverview
              firstCourse={firstCourse}
              onCreate={() => setCreating(true)}
            />
          ) : (
            <StudentOverview courses={courses} loading={loading} />
          )}
          <div className="stats-grid">
            {stats.map(({ label, value, note, icon: Icon }) => (
              <section className="stat-card" key={label}>
                <div>
                  <span>{label}</span>
                  <span className="stat-icon">
                    <Icon size={18} strokeWidth={1.7} />
                  </span>
                </div>
                <strong>{loading ? "—" : value}</strong>
                <p>{note}</p>
              </section>
            ))}
          </div>
        </>
      ) : (
        <div className="page-heading">
          <div>
            <span className="eyebrow">
              {tab === "explore"
                ? "FOLLOW YOUR CURIOSITY"
                : teacher
                  ? "MADE BY YOU, FOR WHAT’S NEXT"
                  : "LITTLE STEPS, BIG POSSIBILITIES"}
            </span>
            <h1>
              {tab === "explore"
                ? "There’s more to discover."
                : teacher
                  ? "Your ideas, taking shape."
                  : "Keep your curiosity going."}
            </h1>
            <p>
              {tab === "explore"
                ? "Find something that makes you think, build or see things a little differently."
                : teacher
                  ? "A home for everything you’re creating and sharing."
                  : "Everything you’re learning, all in one thoughtful little space."}
            </p>
          </div>
          {teacher && (
            <button
              className="btn btn-primary"
              onClick={() => setCreating(true)}
            >
              <Plus size={17} />
              Create course
            </button>
          )}
        </div>
      )}
      <section className="courses-section">
        <div className="section-heading">
          <div>
            <h2>
              {tab === "overview"
                ? teacher
                  ? "Your courses"
                  : "Continue your journey"
                : tab === "explore"
                  ? "A world of possibilities"
                  : teacher
                    ? "Course library"
                    : "Your learning library"}
            </h2>
            {tab === "overview" && (
              <p>
                {teacher
                  ? "A little of what you’re putting into the world."
                  : "Pick up a little momentum, right where you left off."}
              </p>
            )}
          </div>
          {tab === "overview" ? (
            <Link href="/dashboard?tab=courses" className="text-action">
              View all courses
              <ArrowUpRight size={16} />
            </Link>
          ) : (
            <span className="subtle-count">
              {courses.length} {courses.length === 1 ? "course" : "courses"}
            </span>
          )}
        </div>
        {tab !== "overview" && (
          <div className="course-toolbar">
            {tab !== "explore" && (
              <div className="filter-tabs" aria-label="Filter courses">
                {[
                  { value: "all", label: "All courses" },
                  {
                    value: teacher ? "published" : "progress",
                    label: teacher ? "Published" : "In progress",
                  },
                  {
                    value: teacher ? "draft" : "completed",
                    label: teacher ? "Drafts" : "Completed",
                  },
                ].map((f) => (
                  <button
                    key={f.value}
                    className={filter === f.value ? "selected" : ""}
                    onClick={() => setFilter(f.value)}
                    aria-pressed={filter === f.value}
                  >
                    {f.label}
                  </button>
                ))}
              </div>
            )}
            <div className="toolbar-search">
              <label className="search-box">
                <Search size={17} />
                <input
                  aria-label="Search courses"
                  placeholder="Find a course…"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                />
              </label>
              <label className="sort-select">
                <ArrowDownWideNarrow size={16} />
                <select
                  aria-label="Sort courses"
                  value={sort}
                  onChange={(e) => setSort(e.target.value)}
                >
                  <option value="recent">Recent</option>
                  <option value="title">A to Z</option>
                </select>
              </label>
            </div>
          </div>
        )}
        {loading ? (
          <div className="course-grid skeleton-grid">
            {[1, 2, 3].map((i) => (
              <div className="skeleton-card" key={i}>
                <div />
                <span />
                <span />
                <span />
              </div>
            ))}
          </div>
        ) : error ? (
          <ErrorState message={error} retry={reload} />
        ) : !filtered.length ? (
          <EmptyState
            title={
              query
                ? "A little change of search?"
                : teacher
                  ? "Your next great course starts here."
                  : "Your next chapter is waiting."
            }
            text={
              query
                ? "No courses match this search. Try another title or category."
                : teacher
                  ? "Turn something you know into something worth sharing."
                  : "Explore the library and find a course that sparks your curiosity."
            }
          >
            {teacher ? (
              <button
                className="btn btn-primary"
                onClick={() => setCreating(true)}
              >
                <Plus size={16} />
                Create a course
              </button>
            ) : (
              <Link href="/dashboard?tab=explore" className="btn btn-primary">
                Explore courses
                <ArrowRight size={16} />
              </Link>
            )}
          </EmptyState>
        ) : (
          <div
            className={`course-grid ${teacher ? "" : "student-course-grid"}`}
          >
            {filtered.map((course) => (
              <CourseCard
                key={course.id}
                course={course}
                teacher={teacher}
                onEnroll={() => void enroll(course)}
                enrolling={enrolling === course.id}
              />
            ))}
          </div>
        )}
      </section>
      {tab === "overview" && (
        <div className="bottom-note">
          <span>
            <ClipboardCheck size={20} />
          </span>
          <div>
            <strong>
              {teacher
                ? "A little practice makes a lot of difference."
                : "Put a little of that learning into practice."}
            </strong>
            <p>
              {teacher
                ? "Create meaningful assignments and help your students find their own way forward."
                : "Your assignments are a chance to turn new ideas into something you can call your own."}
            </p>
          </div>
          <Link href="/dashboard?tab=assignments" className="text-action">
            See assignments
            <ArrowUpRight size={17} />
          </Link>
        </div>
      )}
      {createOpen && <CreateCourse onClose={closeCreate} />}
    </>
  );
}
export default function Page() {
  return (
    <Suspense fallback={<Loading />}>
      <Dashboard />
    </Suspense>
  );
}
