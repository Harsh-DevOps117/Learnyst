"use client";
import { useState } from "react";
import Link from "next/link";
import {
  ArrowRight,
  ArrowUpRight,
  BookOpen,
  CheckCircle2,
  Clock3,
  Play,
  Sparkles,
} from "lucide-react";
import type { Assignment, Course, CourseDetail } from "@/lib/types";
import { useRemote } from "@/lib/useRemote";
import { Artwork } from "./Shared";
export function learningDuration(seconds: number) {
  if (seconds < 60) return `${seconds} sec`;
  if (seconds < 3600) return `${Math.ceil(seconds / 60)} min`;
  return `${Math.floor(seconds / 3600)} hr ${Math.round((seconds % 3600) / 60)} min`;
}
export function StudentOverview({
  courses,
  loading,
}: {
  courses: Course[];
  loading: boolean;
}) {
  const [now] = useState(() => Date.now());
  const unfinished = [...courses]
    .filter((c) => c.progress < 90)
    .sort((a, b) => b.progress - a.progress);
  const next = unfinished[0] || courses[0];
  const { data: detail } = useRemote<CourseDetail>(
    next ? `/courses/${next.id}` : null,
  );
  const assignments = useRemote<Assignment[]>("/assignments");
  const pending = (assignments.data || [])
    .filter(
      (a) =>
        !a.submissions.length && (!a.due_date || Date.parse(a.due_date) >= now),
    )
    .sort(
      (a, b) =>
        (a.due_date ? Date.parse(a.due_date) : Infinity) -
        (b.due_date ? Date.parse(b.due_date) : Infinity),
    );
  const lesson =
    detail?.videos.find(
      (v) =>
        (detail.progress.find((p) => p.video_id === v.id)?.completion_percent ||
          0) < 90,
    ) || detail?.videos[0];
  const total = courses.reduce((sum, c) => sum + c.lesson_count, 0);
  const done = courses.reduce((sum, c) => sum + c.completed_lessons, 0);
  const percent = total ? Math.round((done / total) * 100) : 0;
  return (
    <div className="student-overview-grid">
      <section
        className={`student-resume ${loading ? "student-resume-loading" : ""}`}
      >
        <div className="student-resume-main">
          <div className="student-resume-top">
            <span className="eyebrow">
              <span className="hero-tiny-dot" />
              {next
                ? next.progress
                  ? "PICK UP WHERE YOU LEFT OFF"
                  : "YOUR NEXT SMALL WIN"
                : "GOOD THINGS START WITH CURIOSITY"}
            </span>
            <span className="student-hero-status">
              <BookOpen size={12} />
              {next ? next.category : "Your space to grow"}
            </span>
          </div>
          <h2>
            {next ? (
              next.title
            ) : (
              <>
                A new skill.
                <br />A fresh <em>perspective.</em>
              </>
            )}
          </h2>
          <p>
            {next
              ? next.description
              : "Find a course that sparks something. Your next chapter can start with a single lesson."}
          </p>
          {next && (
            <div className="student-resume-meta">
              <span>
                <BookOpen size={14} />
                {next.lesson_count}{" "}
                {next.lesson_count === 1 ? "lesson" : "lessons"}
              </span>
              <span>
                <Clock3 size={14} />
                {learningDuration(next.duration_seconds)}
              </span>
              <span>{next.level}</span>
            </div>
          )}
          <div className="student-resume-actions">
            <Link
              href={
                next
                  ? `/dashboard/courses/${next.id}`
                  : "/dashboard?tab=explore"
              }
              className="btn btn-lime"
            >
              {next ? (
                <Play size={14} fill="currentColor" />
              ) : (
                <Sparkles size={15} />
              )}
              {next
                ? next.progress >= 90
                  ? "Revisit course"
                  : next.progress
                    ? "Continue learning"
                    : "Start learning"
                : "Explore courses"}
              <ArrowUpRight size={16} />
            </Link>
            {next && lesson && (
              <span className="student-up-next">
                UP NEXT<strong>{lesson.title}</strong>
              </span>
            )}
          </div>
        </div>
        <div className="student-resume-art" aria-hidden="true">
          <Artwork category={next?.category || "Development"} large />
          <span className="student-art-star">✳</span>
        </div>
        <div className="student-resume-bottom">
          <div>
            <span>
              {next
                ? "Your course progress"
                : "A little more capable, every day."}
            </span>
            {next && <strong>{next.progress}% complete</strong>}
          </div>
          {next && (
            <div className="progress-track">
              <span style={{ width: `${next.progress}%` }} />
            </div>
          )}
        </div>
      </section>
      <aside className="student-horizon">
        <div className="student-horizon-heading">
          <span className="eyebrow">MAKE A LITTLE MOMENTUM</span>
          <h2>
            On your horizon<span>✦</span>
          </h2>
        </div>
        <div className="student-progress-summary">
          <div
            className="student-progress-ring"
            style={{
              background: `conic-gradient(var(--ink) ${percent}%, var(--surface-raised) 0)`,
            }}
          >
            <span>
              {percent}
              <small>%</small>
            </span>
          </div>
          <div>
            <strong>
              {done} of {total} lessons
            </strong>
            <p>
              {done ? "Look how far you’ve come." : "Every first step matters."}
            </p>
            <span>YOUR LEARNING JOURNEY</span>
          </div>
        </div>
        <div className="student-horizon-divider">
          <span>YOUR NEXT STEPS</span>
          <span>
            {assignments.loading ? "…" : `${pending.length} to explore`}
          </span>
        </div>
        <div className="student-next-steps">
          {assignments.error ? (
            <p className="student-next-empty">
              Assignments couldn’t load.{" "}
              <button onClick={assignments.reload}>Try again</button>
            </p>
          ) : assignments.loading ? (
            <p className="student-next-empty">Finding your next steps…</p>
          ) : pending.length ? (
            pending.slice(0, 2).map((a) => (
              <Link
                key={a.id}
                href="/dashboard?tab=assignments"
                className="student-next-step"
              >
                <span className="student-task-icon">
                  <BookOpen size={16} />
                </span>
                <div>
                  <strong>{a.course_title}</strong>
                  <p>{a.title}</p>
                  <small>
                    {a.due_date
                      ? `Due ${new Date(a.due_date).toLocaleDateString("en-IN", { day: "numeric", month: "short" })}`
                      : "At your own pace"}
                  </small>
                </div>
                <ArrowUpRight size={15} />
              </Link>
            ))
          ) : (
            <div className="student-next-empty">
              <CheckCircle2 size={21} />
              <strong>
                {courses.length
                  ? "You’re all caught up."
                  : "Something good is waiting."}
              </strong>
              <p>
                {courses.length
                  ? "Make a little room for your next lesson."
                  : "Explore a course and start your own learning journey."}
              </p>
            </div>
          )}
        </div>
        <Link
          href={
            courses.length
              ? "/dashboard?tab=assignments"
              : "/dashboard?tab=explore"
          }
          className="student-horizon-link"
        >
          {courses.length ? "View your assignments" : "Find your first course"}
          <ArrowRight size={15} />
        </Link>
      </aside>
    </div>
  );
}
