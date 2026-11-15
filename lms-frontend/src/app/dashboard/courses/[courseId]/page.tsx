"use client";
import { useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  ArrowRight,
  ArrowUpRight,
  BookOpen,
  Check,
  CheckCircle2,
  ChevronRight,
  Clock3,
  Download,
  FileVideo2,
  GraduationCap,
  Loader2,
  Plus,
  Save,
  Trash2,
  Upload,
  Video,
} from "lucide-react";
import toast from "react-hot-toast";
import { coreApi, errorMessage } from "@/api/client";
import { useRemote } from "@/lib/useRemote";
import type { CourseDetail, Lesson, Progress } from "@/lib/types";
import {
  Artwork,
  EmptyState,
  ErrorState,
  Loading,
  Modal,
} from "@/components/workspace/Shared";
function durationLabel(seconds: number) {
  return `${Math.floor(seconds / 60)}:${String(Math.floor(seconds % 60)).padStart(2, "0")}`;
}
async function videoDuration(file: File): Promise<number> {
  return new Promise((resolve, reject) => {
    const video = document.createElement("video");
    const url = URL.createObjectURL(file);
    const cleanup = () => {
      URL.revokeObjectURL(url);
      video.removeAttribute("src");
      video.load();
    };
    const timeout = setTimeout(() => {
      cleanup();
      reject(
        new Error(
          "Could not read this video. Please choose an MP4 or WebM file.",
        ),
      );
    }, 10000);
    video.onloadedmetadata = () => {
      clearTimeout(timeout);
      const duration = video.duration;
      cleanup();
      if (Number.isFinite(duration)) resolve(Math.ceil(duration));
      else reject(new Error("This video has an invalid duration."));
    };
    video.onerror = () => {
      clearTimeout(timeout);
      cleanup();
      reject(
        new Error(
          "Your browser could not read this video. Try an H.264 MP4 or WebM file.",
        ),
      );
    };
    video.preload = "metadata";
    video.src = url;
  });
}
function AddLesson({
  courseId,
  close,
  reload,
}: {
  courseId: number;
  close: () => void;
  reload: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const file = data.get("video") as File;
    const notes = data.get("notes") as File;
    if (!file?.size || file.size > 256 * 1024 * 1024) {
      setError("Choose a video smaller than 256 MB.");
      return;
    }
    setBusy(true);
    setError("");
    setStatus("Reading your video…");
    try {
      const duration = await videoDuration(file);
      const upload = new FormData();
      upload.append("file", file);
      upload.append("course_id", String(courseId));
      const { data: uploaded } = await coreApi.post<{ url: string }>(
        "/upload",
        upload,
        {
          timeout: 0,
          onUploadProgress: (progress) =>
            setStatus(
              `Uploading video · ${Math.round((progress.loaded / (progress.total || file.size)) * 100)}%`,
            ),
        },
      );
      let noteUrl = "";
      if (notes?.size) {
        const noteForm = new FormData();
        noteForm.append("file", notes);
        noteForm.append("course_id", String(courseId));
        setStatus("Adding your lesson notes…");
        const { data: note } = await coreApi.post<{ url: string }>(
          "/upload",
          noteForm,
          { timeout: 0 },
        );
        noteUrl = note.url;
      }
      setStatus("Saving your lesson…");
      await coreApi.post(`/courses/${courseId}/videos`, {
        title: data.get("title"),
        description: data.get("description"),
        video_url: uploaded.url,
        note_url: noteUrl,
        duration_seconds: duration,
      });
      toast.success("One more lesson. One more possibility.");
      reload();
      close();
    } catch (err) {
      setError(
        err instanceof Error && !("isAxiosError" in err)
          ? err.message
          : errorMessage(err),
      );
    } finally {
      setBusy(false);
    }
  }
  return (
    <Modal
      title="Give your ideas a voice."
      description="A clear explanation can make all the difference. Add your next lesson."
      onClose={() => {
        if (!busy) close();
      }}
    >
      <form onSubmit={submit} className="stack-form" aria-busy={busy}>
        {error && (
          <div className="notice error" role="alert">
            {error}
          </div>
        )}
        <fieldset disabled={busy}>
          <label>
            Lesson title
            <input
              name="title"
              placeholder="What’s this lesson about?"
              required
              maxLength={255}
            />
          </label>
          <label>
            Description <span className="optional">(optional)</span>
            <textarea
              name="description"
              rows={3}
              placeholder="A little context for your students"
              maxLength={10000}
            />
          </label>
          <label className="upload-label">
            <span>Lesson video</span>
            <span className="file-drop">
              <Upload size={23} />
              <strong>Choose an MP4 or WebM video</strong>
              <span>Up to 256 MB · H.264 MP4 recommended</span>
              <input
                name="video"
                type="file"
                accept="video/mp4,video/webm,.mp4,.webm"
                required
              />
            </span>
          </label>
          <label>
            Lesson notes <span className="optional">(optional PDF)</span>
            <input name="notes" type="file" accept="application/pdf,.pdf" />
          </label>
          <button className="btn btn-primary full" disabled={busy}>
            {busy ? (
              <>
                <Loader2 size={16} className="spin" />
                {status}
              </>
            ) : (
              <>
                Add lesson
                <Plus size={16} />
              </>
            )}
          </button>
        </fieldset>
      </form>
    </Modal>
  );
}
function InviteStudent({
  courseId,
  close,
}: {
  courseId: number;
  close: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [sent, setSent] = useState(false);
  async function invite(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = Object.fromEntries(new FormData(event.currentTarget));
    setBusy(true);
    setError("");
    try {
      await coreApi.post(`/courses/${courseId}/invitations`, data);
      setSent(true);
    } catch (error) {
      setError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  return (
    <Modal
      title={sent ? "A new beginning, delivered." : "Invite someone to learn."}
      description={
        sent
          ? "The invitation is ready in your local email inbox. The student can sign in with their invited email and join this course."
          : "A little encouragement can open a new door. Send a student a course invitation."
      }
      onClose={close}
    >
      {sent ? (
        <a
          className="btn btn-primary full"
          style={{ marginTop: 25 }}
          href="http://localhost:8025"
          target="_blank"
          rel="noreferrer"
        >
          Open local inbox
          <ArrowUpRight size={16} />
        </a>
      ) : (
        <form onSubmit={invite} className="stack-form">
          {error && (
            <div className="notice error" role="alert">
              {error}
            </div>
          )}
          <label>
            Student email
            <input
              name="email"
              type="email"
              required
              maxLength={255}
              placeholder="student@example.com"
            />
          </label>
          <button className="btn btn-primary full" disabled={busy}>
            {busy ? "Sending…" : "Send invitation"}
            <ArrowRight size={16} />
          </button>
        </form>
      )}
    </Modal>
  );
}
function TeacherCourse({
  detail,
  reload,
}: {
  detail: CourseDetail;
  reload: () => void;
}) {
  const { course, videos } = detail;
  const router = useRouter();
  const [adding, setAdding] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [inviting, setInviting] = useState(false);
  const [preview, setPreview] = useState<Lesson | null>(null);
  async function toggle() {
    setBusy(true);
    setError("");
    try {
      await coreApi.put(`/courses/${course.id}`, {
        is_published: !course.is_published,
      });
      toast.success(
        course.is_published
          ? "Course unpublished."
          : "Your course is ready to inspire.",
      );
      reload();
    } catch (error) {
      setError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  async function save(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      await coreApi.put(`/courses/${course.id}`, Object.fromEntries(form));
      toast.success("Course details saved.");
      reload();
    } catch (error) {
      setError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  async function remove() {
    setBusy(true);
    setError("");
    try {
      await coreApi.delete(`/courses/${course.id}`);
      toast.success("Course deleted.");
      router.replace("/dashboard?tab=courses");
    } catch (error) {
      setError(errorMessage(error));
      setDeleting(false);
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      <Link className="back-link" href="/dashboard?tab=courses">
        <ArrowLeft size={16} />
        Back to your courses
      </Link>
      <div className="page-heading course-page-heading">
        <div>
          <div className="heading-badges">
            <span className="eyebrow">YOUR COURSE WORKSPACE</span>
            <span
              className={`badge ${course.is_published ? "badge-green" : "badge-amber"}`}
            >
              <i />
              {course.is_published ? "Published" : "Draft"}
            </span>
          </div>
          <h1>{course.title}</h1>
          <p>
            A little structure. A little inspiration. A course that’s all yours.
          </p>
        </div>
        <button
          className={`btn ${course.is_published ? "btn-secondary" : "btn-primary"}`}
          disabled={busy}
          onClick={() => void toggle()}
        >
          {course.is_published ? "Unpublish course" : "Publish course"}
          <ArrowUpRight size={17} />
        </button>
      </div>
      {error && (
        <div className="notice error" role="alert">
          {error}
        </div>
      )}
      <div className="editor-grid">
        <div className="editor-main">
          <section className="panel">
            <div className="panel-heading">
              <BookOpen size={19} />
              <h2>Course details</h2>
            </div>
            <form onSubmit={save} className="stack-form">
              <label>
                Course title
                <input
                  name="title"
                  required
                  maxLength={255}
                  defaultValue={course.title}
                />
              </label>
              <label>
                What will your students learn?
                <textarea
                  name="description"
                  rows={4}
                  maxLength={10000}
                  defaultValue={course.description}
                />
              </label>
              <button className="btn btn-secondary" disabled={busy}>
                <Save size={16} />
                {busy ? "Saving…" : "Save details"}
              </button>
            </form>
          </section>
          <section className="panel curriculum-panel">
            <div className="section-heading">
              <div>
                <h2>Your curriculum</h2>
                <p>One lesson at a time, an idea takes shape.</p>
              </div>
              <button
                className="btn btn-primary btn-small"
                onClick={() => setAdding(true)}
              >
                <Plus size={16} />
                Add lesson
              </button>
            </div>
            {!videos.length ? (
              <EmptyState
                title="Start with your first lesson."
                text="Upload a video to give your students somewhere to begin."
              />
            ) : (
              <div className="editor-lessons">
                {videos.map((lesson, index) => (
                  <div className="editor-lesson" key={lesson.id}>
                    <span className="lesson-number">
                      {String(index + 1).padStart(2, "0")}
                    </span>
                    <span className="lesson-type-icon">
                      <FileVideo2 size={20} />
                    </span>
                    <div>
                      <h3>{lesson.title}</h3>
                      <span>
                        <Video size={12} />
                        Video lesson<span>·</span>
                        {durationLabel(lesson.duration_seconds)}
                      </span>
                    </div>
                    <button
                      className="icon-button"
                      aria-label={`Preview ${lesson.title}`}
                      onClick={() => setPreview(lesson)}
                    >
                      <ArrowUpRight size={19} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>
        <aside className="course-editor-aside">
          <div className="course-preview-card">
            <Artwork category={course.category} large />
            <div>
              <span className="eyebrow">A LITTLE AT A GLANCE</span>
              <h3>{course.title}</h3>
              <dl>
                <div>
                  <dt>Category</dt>
                  <dd>{course.category}</dd>
                </div>
                <div>
                  <dt>Level</dt>
                  <dd>{course.level}</dd>
                </div>
                <div>
                  <dt>Lessons</dt>
                  <dd>{videos.length}</dd>
                </div>
                <div>
                  <dt>Enrolled students</dt>
                  <dd>{course.student_count}</dd>
                </div>
                <div>
                  <dt>Course access</dt>
                  <dd>{course.price ? `₹${course.price}` : "Free"}</dd>
                </div>
              </dl>
            </div>
          </div>
          <div className="editor-tip">
            <GraduationCap size={23} />
            <h3>Clarity is a superpower.</h3>
            <p>
              Short, focused lessons help students feel a little more capable
              with every step.
            </p>
          </div>
          {course.is_published && (
            <button
              className="btn btn-secondary full"
              style={{ marginTop: 18 }}
              onClick={() => setInviting(true)}
            >
              Invite a student
              <ArrowUpRight size={16} />
            </button>
          )}
          <button className="delete-course" onClick={() => setDeleting(true)}>
            <Trash2 size={15} />
            Delete this course
          </button>
        </aside>
      </div>
      {inviting && (
        <InviteStudent courseId={course.id} close={() => setInviting(false)} />
      )}
      {adding && (
        <AddLesson
          courseId={course.id}
          close={() => setAdding(false)}
          reload={reload}
        />
      )}
      {preview && (
        <Modal title={preview.title} onClose={() => setPreview(null)}>
          <video
            className="preview-video"
            src={preview.video_url}
            controls
            playsInline
            preload="metadata"
          />
          <p className="muted">{preview.description}</p>
        </Modal>
      )}
      {deleting && (
        <Modal
          title="Delete this course?"
          description="This removes the course, its lessons and uploaded files. Courses with enrolled students are protected; you can unpublish them instead."
          onClose={() => setDeleting(false)}
        >
          <div className="modal-actions">
            <button
              className="btn btn-secondary"
              onClick={() => setDeleting(false)}
            >
              Keep course
            </button>
            <button
              className="btn btn-danger"
              disabled={busy}
              onClick={() => void remove()}
            >
              {busy ? "Deleting…" : "Delete course"}
            </button>
          </div>
        </Modal>
      )}
    </>
  );
}
function StudentCourse({ detail }: { detail: CourseDetail }) {
  const [selected, setSelected] = useState(
    () =>
      detail.videos.find(
        (v) =>
          (detail.progress.find((p) => p.video_id === v.id)
            ?.completion_percent || 0) < 90,
      )?.id ??
      detail.videos[0]?.id ??
      0,
  );
  const [localProgress, setLocalProgress] = useState<Record<number, Progress>>(
    {},
  );
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [videoError, setVideoError] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const videoRef = useRef<HTMLVideoElement>(null);
  const lastSave = useRef(0);
  const progressFor = (id: number) =>
    localProgress[id] || detail.progress.find((p) => p.video_id === id);
  const lesson =
    detail.videos.find((v) => v.id === selected) || detail.videos[0];
  const currentIndex = detail.videos.findIndex((v) => v.id === lesson?.id);
  const completed = detail.videos.filter(
    (v) => (progressFor(v.id)?.completion_percent || 0) >= 90,
  ).length;
  async function saveProgress(
    video: HTMLVideoElement | null,
    complete = false,
  ) {
    if (
      !lesson ||
      (!complete && (!video?.duration || !Number.isFinite(video.duration)))
    )
      return;
    const p: Progress = {
      video_id: lesson.id,
      watch_time_seconds: Math.floor(
        video?.currentTime || lesson.duration_seconds,
      ),
      completion_percent: complete
        ? 100
        : Math.min(
            100,
            Math.floor((video!.currentTime / video!.duration) * 100),
          ),
    };
    if (
      p.completion_percent <= (progressFor(lesson.id)?.completion_percent || 0)
    )
      return;
    setSaving(true);
    try {
      await coreApi.post("/progress/video", p);
      setLocalProgress((previous) => {
        const existing = previous[p.video_id];
        return {
          ...previous,
          [p.video_id]: {
            ...p,
            completion_percent: Math.max(
              existing?.completion_percent || 0,
              p.completion_percent,
            ),
          },
        };
      });
      setSaveError("");
    } catch (error) {
      setSaveError(
        errorMessage(error, "Could not save progress. Please try again."),
      );
    } finally {
      setSaving(false);
    }
  }
  function choose(id: number) {
    void saveProgress(videoRef.current);
    setSelected(id);
    setLoaded(false);
    setVideoError(false);
    lastSave.current = 0;
  }
  return (
    <>
      <Link className="back-link" href="/dashboard?tab=courses">
        <ArrowLeft size={16} />
        Back to your learning
      </Link>
      <div className="page-heading course-page-heading">
        <div>
          <span className="eyebrow">
            {detail.course.category} · {detail.course.level}
          </span>
          <h1>{detail.course.title}</h1>
          <p>
            With {detail.course.instructor}
            <span className="inline-dot">·</span>
            {completed} of {detail.videos.length} lessons completed
          </p>
        </div>
        <span className="count-pill">
          <GraduationCap size={19} />
          At your own pace
        </span>
      </div>
      <div className="learning-grid">
        <div>
          <div className="video-container">
            {lesson ? (
              <>
                <video
                  key={lesson.id}
                  ref={videoRef}
                  src={lesson.video_url}
                  controls
                  playsInline
                  preload="metadata"
                  onLoadedMetadata={(event) => {
                    const p = progressFor(lesson.id);
                    if (p && p.completion_percent < 90)
                      event.currentTarget.currentTime = Math.min(
                        p.watch_time_seconds,
                        Math.max(0, event.currentTarget.duration - 0.5),
                      );
                    setLoaded(true);
                  }}
                  onError={() => {
                    setVideoError(true);
                    setLoaded(true);
                  }}
                  onTimeUpdate={(event) => {
                    if (Date.now() - lastSave.current > 8000) {
                      lastSave.current = Date.now();
                      void saveProgress(event.currentTarget);
                    }
                  }}
                  onPause={(event) => void saveProgress(event.currentTarget)}
                  onEnded={(event) =>
                    void saveProgress(event.currentTarget, true)
                  }
                />
                {!loaded && (
                  <div className="video-status">
                    <Loader2 size={27} className="spin" />
                    <span>Getting your lesson ready…</span>
                  </div>
                )}
                {videoError && (
                  <div className="video-status">
                    <FileVideo2 size={30} />
                    <p>
                      This video couldn’t be played. Refresh the page or try
                      another lesson.
                    </p>
                    <button
                      className="btn btn-lime"
                      onClick={() => {
                        setVideoError(false);
                        setLoaded(false);
                        videoRef.current?.load();
                      }}
                    >
                      Try again
                    </button>
                  </div>
                )}
              </>
            ) : (
              <EmptyState
                title="Lessons are on their way."
                text="Your teacher is preparing something good."
              />
            )}
          </div>
          {lesson && (
            <section className="lesson-detail panel">
              <div className="section-heading">
                <div>
                  <span className="eyebrow">
                    LESSON {String(currentIndex + 1).padStart(2, "0")}
                  </span>
                  <h2>{lesson.title}</h2>
                </div>
                <span className="lesson-duration">
                  <Clock3 size={14} />
                  {durationLabel(lesson.duration_seconds)}
                </span>
              </div>
              <p className="preserve-lines">
                {lesson.description || detail.course.description}
              </p>
              {saveError && (
                <div className="notice error" role="alert">
                  {saveError}
                  <button
                    className="text-action"
                    onClick={() => void saveProgress(videoRef.current)}
                  >
                    Retry saving
                  </button>
                </div>
              )}
              <div className="lesson-actions">
                <button
                  className={`btn ${(progressFor(lesson.id)?.completion_percent || 0) >= 90 ? "btn-secondary" : "btn-primary"}`}
                  disabled={
                    saving ||
                    (progressFor(lesson.id)?.completion_percent || 0) >= 90
                  }
                  onClick={() => void saveProgress(videoRef.current, true)}
                >
                  {saving ? (
                    <Loader2 className="spin" size={16} />
                  ) : (
                    <CheckCircle2 size={16} />
                  )}
                  {(progressFor(lesson.id)?.completion_percent || 0) >= 90
                    ? "Lesson complete"
                    : "Mark as complete"}
                </button>
                {detail.videos[currentIndex + 1] && (
                  <button
                    className="text-action"
                    onClick={() => choose(detail.videos[currentIndex + 1].id)}
                  >
                    Next lesson
                    <ArrowRight size={16} />
                  </button>
                )}
                {lesson.note_url && (
                  <a href={lesson.note_url} className="text-action" download>
                    <Download size={16} />
                    Lesson notes
                  </a>
                )}
              </div>
            </section>
          )}
          <div className="learning-reminder">
            <span>✳</span>
            <p>
              You don’t have to learn it all today.
              <br />
              <strong>A little more than yesterday is enough.</strong>
            </p>
          </div>
        </div>
        <aside className="curriculum-sidebar">
          <div className="curriculum-header">
            <span className="eyebrow">A LITTLE ROADMAP</span>
            <h2>Your course journey</h2>
            <div className="curriculum-progress-label">
              <span>
                {completed} of {detail.videos.length} complete
              </span>
              <strong>
                {detail.videos.length
                  ? Math.round((completed / detail.videos.length) * 100)
                  : 0}
                %
              </strong>
            </div>
            <div className="progress-track">
              <span
                style={{
                  width: `${detail.videos.length ? (completed / detail.videos.length) * 100 : 0}%`,
                }}
              />
            </div>
          </div>
          <nav aria-label="Course lessons">
            {detail.videos.map((v, index) => {
              const done = (progressFor(v.id)?.completion_percent || 0) >= 90;
              return (
                <button
                  key={v.id}
                  onClick={() => choose(v.id)}
                  className={`curriculum-lesson ${v.id === lesson?.id ? "curriculum-active" : ""}`}
                  aria-current={v.id === lesson?.id ? "step" : undefined}
                >
                  <span className={done ? "lesson-done" : ""}>
                    {done ? (
                      <Check size={15} />
                    ) : (
                      String(index + 1).padStart(2, "0")
                    )}
                  </span>
                  <div>
                    <strong>{v.title}</strong>
                    <small>
                      <Video size={11} />
                      {durationLabel(v.duration_seconds)}
                    </small>
                  </div>
                  <ChevronRight size={16} />
                </button>
              );
            })}
          </nav>
          <Link
            href="/dashboard?tab=assignments"
            className="curriculum-assignment"
          >
            <BookOpen size={19} />
            <div>
              <strong>Put it into practice</strong>
              <span>Explore your assignments</span>
            </div>
            <ArrowUpRight size={16} />
          </Link>
        </aside>
      </div>
    </>
  );
}
export default function Page() {
  const { courseId } = useParams<{ courseId: string }>();
  const { data, loading, error, reload } = useRemote<CourseDetail>(
    `/courses/${courseId}`,
  );
  if (loading && !data) return <Loading label="Opening your course…" />;
  if (error || !data)
    return (
      <>
        <Link className="back-link" href="/dashboard?tab=explore">
          <ArrowLeft size={16} />
          Explore courses
        </Link>
        <ErrorState
          message={error || "This course could not be found."}
          retry={reload}
        />
      </>
    );
  return data.can_manage ? (
    <TeacherCourse detail={data} reload={reload} />
  ) : (
    <StudentCourse detail={data} />
  );
}
