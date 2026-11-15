"use client";
import { useState } from "react";
import {
  Check,
  ChevronDown,
  ClipboardList,
  Clock3,
  Plus,
  Send,
} from "lucide-react";
import toast from "react-hot-toast";
import { coreApi, errorMessage } from "@/api/client";
import { useRemote } from "@/lib/useRemote";
import type { Assignment, Course, Submission } from "@/lib/types";
import { EmptyState, ErrorState, Loading, Modal, formatDate } from "./Shared";
function GradeForm({
  submission,
  max,
  reload,
}: {
  submission: Submission;
  max: number;
  reload: () => void;
}) {
  const [busy, setBusy] = useState(false);
  async function grade(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setBusy(true);
    try {
      await coreApi.put(`/submissions/${submission.id}/grade`, {
        marks: Number(form.get("marks")),
        feedback: form.get("feedback"),
      });
      toast.success("Feedback shared. A little encouragement goes a long way.");
      reload();
    } catch (error) {
      toast.error(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="submission-review">
      <div className="submission-review-head">
        <strong>{submission.student_name}</strong>
        <span className="badge badge-green">
          {submission.status === "graded" ? "Reviewed" : "Ready to review"}
        </span>
      </div>
      <p className="preserve-lines">{submission.text}</p>
      <form onSubmit={grade} className="grade-form">
        <label>
          Marks out of {max}
          <input
            type="number"
            name="marks"
            min={0}
            max={max}
            required
            defaultValue={submission.marks ?? ""}
          />
        </label>
        <label>
          Feedback
          <textarea
            name="feedback"
            rows={2}
            placeholder="What went well? What could be better?"
            defaultValue={submission.feedback}
            maxLength={10000}
          />
        </label>
        <button className="btn btn-primary" disabled={busy}>
          {busy ? "Saving…" : "Save feedback"}
          <Check size={15} />
        </button>
      </form>
    </div>
  );
}
function AssignmentItem({
  assignment: a,
  teacher,
  reload,
}: {
  assignment: Assignment;
  teacher: boolean;
  reload: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [now] = useState(() => Date.now());
  const submission = a.submissions[0];
  const overdue = a.due_date && Date.parse(a.due_date) < now;
  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setBusy(true);
    setError("");
    try {
      await coreApi.post(`/assignments/${a.id}/submit`, {
        text: form.get("text"),
      });
      toast.success("Your assignment is in. Nicely done.");
      reload();
    } catch (error) {
      setError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  return (
    <details className="assignment-item">
      <summary>
        <span className="assignment-icon">
          <ClipboardList size={21} />
        </span>
        <div className="assignment-summary">
          <span className="eyebrow">{a.course_title}</span>
          <h3>{a.title}</h3>
          <span className="muted assignment-due">
            <Clock3 size={13} />
            {a.due_date ? `Due ${formatDate(a.due_date)}` : "No deadline"}
            <span>·</span>
            {a.max_marks} marks
          </span>
        </div>
        <span
          className={`badge ${teacher ? "badge-neutral" : submission?.status === "graded" ? "badge-green" : submission ? "badge-blue" : overdue ? "badge-amber" : "badge-neutral"}`}
        >
          {teacher
            ? `${a.submissions.length} submitted`
            : submission?.status === "graded"
              ? `${submission.marks}/${a.max_marks} marks`
              : submission
                ? "Submitted"
                : overdue
                  ? "Deadline passed"
                  : "To do"}
        </span>
        <ChevronDown size={19} className="details-chevron" />
      </summary>
      <div className="assignment-content">
        <p className="preserve-lines">{a.description}</p>
        {teacher ? (
          a.submissions.length ? (
            a.submissions.map((s) => (
              <GradeForm
                key={s.id}
                submission={s}
                max={a.max_marks}
                reload={reload}
              />
            ))
          ) : (
            <p className="inline-note">
              No submissions yet. Your students’ work will appear here.
            </p>
          )
        ) : submission ? (
          <div className="submitted-response">
            <span className="eyebrow">YOUR RESPONSE</span>
            <p className="preserve-lines">{submission.text}</p>
            {submission.feedback && (
              <div className="teacher-feedback">
                <strong>Feedback from your teacher</strong>
                <p>{submission.feedback}</p>
              </div>
            )}
          </div>
        ) : overdue ? (
          <p className="inline-note">
            The deadline for this assignment has passed.
          </p>
        ) : (
          <form onSubmit={submit}>
            {error && (
              <div className="notice error" role="alert">
                {error}
              </div>
            )}
            <label>
              Your response
              <textarea
                name="text"
                placeholder="Share your thoughts, approach or a link to your work…"
                required
                rows={5}
                maxLength={20000}
              />
            </label>
            <button className="btn btn-primary" disabled={busy}>
              {busy ? "Submitting…" : "Submit assignment"}
              <Send size={15} />
            </button>
          </form>
        )}
      </div>
    </details>
  );
}
function CreateAssignment({
  courses,
  onClose,
  reload,
}: {
  courses: Course[];
  onClose: () => void;
  reload: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  async function create(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setBusy(true);
    setError("");
    try {
      await coreApi.post(`/courses/${form.get("course")}/assignments`, {
        title: form.get("title"),
        description: form.get("description"),
        max_marks: Number(form.get("max_marks")),
        due_date: form.get("due_date")
          ? new Date(`${form.get("due_date")}T23:59:59`).toISOString()
          : null,
      });
      toast.success("Assignment created.");
      reload();
      onClose();
    } catch (error) {
      setError(errorMessage(error));
    } finally {
      setBusy(false);
    }
  }
  return (
    <Modal
      title="Learning by doing."
      description="Give your students a chance to put their new skills into practice."
      onClose={onClose}
    >
      <form onSubmit={create} className="stack-form">
        {error && (
          <div role="alert" className="notice error">
            {error}
          </div>
        )}
        <label>
          Course
          <select name="course" aria-label="Course" required>
            {courses.map((c) => (
              <option key={c.id} value={c.id}>
                {c.title}
              </option>
            ))}
          </select>
        </label>
        <label>
          Assignment title
          <input
            name="title"
            placeholder="Give this challenge a name"
            maxLength={255}
            required
          />
        </label>
        <label>
          Instructions
          <textarea
            name="description"
            rows={4}
            placeholder="What should your students work on?"
            required
            maxLength={10000}
          />
        </label>
        <div className="form-row">
          <label>
            Maximum marks
            <input
              name="max_marks"
              type="number"
              min={1}
              max={1000}
              defaultValue={10}
              required
            />
          </label>
          <label>
            Due date <span className="optional">(optional)</span>
            <input
              name="due_date"
              type="date"
              min={new Date().toLocaleDateString("en-CA")}
            />
          </label>
        </div>
        <button className="btn btn-primary full" disabled={busy}>
          {busy ? "Creating…" : "Create assignment"}
          <Plus size={16} />
        </button>
      </form>
    </Modal>
  );
}
export function Assignments({
  teacher,
  courses,
}: {
  teacher: boolean;
  courses: Course[];
}) {
  const { data, error, loading, reload } =
    useRemote<Assignment[]>("/assignments");
  const [create, setCreate] = useState(false);
  return (
    <>
      <div className="page-heading">
        <div>
          <span className="eyebrow">PUT IT INTO PRACTICE</span>
          <h1>Learn it. Try it. Make it yours.</h1>
          <p>
            {teacher
              ? "Meaningful practice, thoughtful feedback, real progress."
              : "A chance to turn what you’ve learned into something real."}
          </p>
        </div>
        {teacher && courses.length > 0 && (
          <button className="btn btn-primary" onClick={() => setCreate(true)}>
            <Plus size={17} />
            Create assignment
          </button>
        )}
      </div>
      {loading ? (
        <Loading />
      ) : error ? (
        <ErrorState message={error} retry={reload} />
      ) : !data?.length ? (
        <EmptyState
          title="A little practice is on the way."
          text={
            teacher
              ? "Create a course, then add your first assignment."
              : "Assignments from your enrolled courses will appear here."
          }
        />
      ) : (
        <div className="assignment-list">
          {data.map((a) => (
            <AssignmentItem
              key={a.id}
              assignment={a}
              teacher={teacher}
              reload={reload}
            />
          ))}
        </div>
      )}
      {create && (
        <CreateAssignment
          courses={courses}
          onClose={() => setCreate(false)}
          reload={reload}
        />
      )}
    </>
  );
}
