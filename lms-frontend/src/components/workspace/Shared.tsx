"use client";
import { useEffect, useRef } from "react";
import {
  AlertCircle,
  ArrowUpRight,
  BookOpen,
  Code2,
  Layers3,
  Loader2,
  Terminal,
  X,
} from "lucide-react";
import Link from "next/link";
import type { Course } from "@/lib/types";
export function Loading({
  label = "Loading your workspace…",
}: {
  label?: string;
}) {
  return (
    <div className="loading-state" role="status">
      <Loader2 className="spin" size={26} />
      <span>{label}</span>
    </div>
  );
}
export function ErrorState({
  message,
  retry,
}: {
  message: string;
  retry?: () => void;
}) {
  return (
    <div className="error-state" role="alert">
      <AlertCircle size={23} />
      <div>
        <strong>We couldn’t load this just yet.</strong>
        <p>{message}</p>
      </div>
      {retry && (
        <button className="btn btn-secondary" onClick={retry}>
          Try again
        </button>
      )}
    </div>
  );
}
export function EmptyState({
  title,
  text,
  children,
}: {
  title: string;
  text: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="empty-state">
      <span className="empty-icon">
        <BookOpen size={26} />
      </span>
      <h3>{title}</h3>
      <p>{text}</p>
      {children}
    </div>
  );
}
export function Modal({
  title,
  description,
  onClose,
  children,
}: {
  title: string;
  description?: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const dialog = ref.current;
    dialog?.showModal();
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previous;
      dialog?.close();
    };
  }, []);
  return (
    <dialog
      ref={ref}
      className="modal"
      onCancel={(event) => {
        event.preventDefault();
        onClose();
      }}
      onClick={(event) => {
        if (event.target === event.currentTarget) onClose();
      }}
    >
      <div className="modal-content">
        <button
          className="icon-button modal-close"
          onClick={onClose}
          aria-label="Close dialog"
        >
          <X size={20} />
        </button>
        <span className="eyebrow">A LITTLE PROGRESS STARTS HERE</span>
        <h2>{title}</h2>
        {description && <p className="muted">{description}</p>}
        {children}
      </div>
    </dialog>
  );
}
export function Artwork({
  category,
  large = false,
}: {
  category: string;
  large?: boolean;
}) {
  const design = category === "Design";
  const devops = category === "DevOps";
  const Icon = design ? Layers3 : devops ? Terminal : Code2;
  return (
    <div
      className={`course-art ${design ? "art-design" : devops ? "art-devops" : "art-code"} ${large ? "art-large" : ""}`}
      aria-hidden="true"
    >
      <div className="art-grid" />
      <span className="art-orbit orbit-one" />
      <span className="art-orbit orbit-two" />
      <span className="art-orbit orbit-three" />
      <div className="art-window">
        <div className="art-window-top">
          <i />
          <i />
          <i />
          <span>
            {design
              ? "creative.studio"
              : devops
                ? "~/your-next-project"
                : "hello.world"}
          </span>
        </div>
        <div className="art-window-body">
          <Icon size={large ? 64 : 40} strokeWidth={1.4} />
          <span>
            {design
              ? "Make it meaningful."
              : devops
                ? "Build. Ship. Repeat."
                : "Ideas into interfaces."}
          </span>
        </div>
      </div>
      <span className="art-caption">
        {design
          ? "DESIGN & CREATE"
          : devops
            ? "BUILD & DEPLOY"
            : "LEARN & BUILD"}{" "}
        <ArrowUpRight size={14} />
      </span>
    </div>
  );
}
export function CourseCard({
  course,
  teacher,
  onEnroll,
  enrolling,
}: {
  course: Course;
  teacher: boolean;
  onEnroll?: () => void;
  enrolling?: boolean;
}) {
  const href = `/dashboard/courses/${course.id}`;
  return (
    <article className={`course-card ${teacher ? "" : "student-course-card"}`}>
      {teacher || course.enrolled ? (
        <Link href={href} aria-label={`Open ${course.title}`}>
          <Artwork category={course.category} />
        </Link>
      ) : (
        <Artwork category={course.category} />
      )}
      <div className="course-card-body">
        <div className="course-card-meta">
          <span className="category-label">{course.category || "General"}</span>
          {teacher ? (
            <span
              className={`badge ${course.is_published ? "badge-green" : "badge-amber"}`}
            >
              <i />
              {course.is_published ? "Published" : "Draft"}
            </span>
          ) : course.enrolled ? (
            <span
              className={`badge ${course.progress >= 90 ? "badge-green" : course.progress ? "badge-blue" : "badge-neutral"}`}
            >
              <i />
              {course.progress >= 90
                ? "Completed"
                : course.progress
                  ? "In progress"
                  : "Not started"}
            </span>
          ) : (
            <span className="level-label">{course.level}</span>
          )}
        </div>
        <h3>
          {teacher || course.enrolled ? (
            <Link href={href}>{course.title}</Link>
          ) : (
            course.title
          )}
        </h3>
        <p className="course-description">
          {course.description || "A new idea, ready to take shape."}
        </p>
        <div className="course-small-meta">
          <span>
            <BookOpen size={14} />
            {course.lesson_count}{" "}
            {course.lesson_count === 1 ? "lesson" : "lessons"}
          </span>
          <span>
            {teacher
              ? `${course.student_count} ${course.student_count === 1 ? "student" : "students"}`
              : course.level}
          </span>
        </div>
        {!teacher && course.enrolled && (
          <div className="course-progress">
            <div>
              <span>
                {course.progress >= 90
                  ? "Looking good!"
                  : course.progress
                    ? "Keep the momentum"
                    : "Ready when you are"}
              </span>
              <strong>{course.progress}%</strong>
            </div>
            <div className="progress-track">
              <span style={{ width: `${course.progress}%` }} />
            </div>
          </div>
        )}
        <div className="course-card-bottom">
          {teacher ? (
            <>
              <span className="muted">{course.level}</span>
              <Link href={href} className="text-action">
                Manage course <ArrowUpRight size={16} />
              </Link>
            </>
          ) : course.enrolled ? (
            <>
              <span className="student-instructor">
                <span className="avatar">
                  {course.instructor
                    .split(" ")
                    .map((part) => part[0])
                    .slice(0, 2)
                    .join("")}
                </span>
                {course.instructor}
              </span>
              <Link href={href} className="btn btn-primary">
                {course.progress >= 90
                  ? "Revisit course"
                  : course.progress
                    ? "Continue learning"
                    : "Start learning"}{" "}
                <ArrowUpRight size={16} />
              </Link>
            </>
          ) : (
            <>
              <strong>
                {course.price ? `₹${course.price}` : "Free to explore"}
              </strong>
              <button
                className="btn btn-primary"
                onClick={onEnroll}
                disabled={enrolling}
              >
                {enrolling ? "Enrolling…" : "Enroll now"}{" "}
                <ArrowUpRight size={16} />
              </button>
            </>
          )}
        </div>
      </div>
    </article>
  );
}
export function formatDate(value: string) {
  return new Date(value).toLocaleDateString("en-IN", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}
