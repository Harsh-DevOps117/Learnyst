"use client";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Brand } from "@/components/workspace/Brand";
import {
  CourseCard,
  EmptyState,
  ErrorState,
  Loading,
} from "@/components/workspace/Shared";
import { useRemote } from "@/lib/useRemote";
import { useAuthStore } from "@/store/authStore";
import { coreApi, errorMessage } from "@/api/client";
import type { Course } from "@/lib/types";
import toast from "react-hot-toast";
import { useState } from "react";
export default function Academy() {
  const { slug } = useParams<{ slug: string }>();
  const router = useRouter();
  const { user } = useAuthStore();
  const [enrolling, setEnrolling] = useState<number | null>(null);
  const { data, error, loading, reload } = useRemote<{
    tenant: { academy_name: string; description: string };
    courses: Course[];
  }>(`/tenants/slug/${encodeURIComponent(slug)}`);
  async function enroll(course: Course) {
    if (!user) {
      router.push(
        `/login?next=${encodeURIComponent("/dashboard?tab=explore")}`,
      );
      return;
    }
    if (user.role_id !== 3) {
      router.push("/dashboard");
      return;
    }
    setEnrolling(course.id);
    try {
      await coreApi.post(`/courses/${course.id}/enroll`);
      router.push(`/dashboard/courses/${course.id}`);
    } catch (error) {
      toast.error(errorMessage(error));
    } finally {
      setEnrolling(null);
    }
  }
  return (
    <div className="landing">
      <header className="landing-nav">
        <Brand />
        <Link className="btn btn-primary" href={user ? "/dashboard" : "/login"}>
          {user ? "Your workspace" : "Sign in"}
        </Link>
      </header>
      <main className="public-academy">
        {loading ? (
          <Loading />
        ) : error || !data ? (
          <ErrorState message={error || "Academy not found."} retry={reload} />
        ) : (
          <>
            <div className="page-heading">
              <div>
                <span className="eyebrow">A SPACE TO GROW</span>
                <h1>{data.tenant.academy_name}</h1>
                <p>{data.tenant.description}</p>
              </div>
            </div>
            {data.courses.length ? (
              <div className="course-grid">
                {data.courses.map((course) => (
                  <CourseCard
                    key={course.id}
                    course={course}
                    teacher={false}
                    onEnroll={() => void enroll(course)}
                    enrolling={enrolling === course.id}
                  />
                ))}
              </div>
            ) : (
              <EmptyState
                title="Something good is on the way."
                text="This academy is preparing its first courses."
              />
            )}
          </>
        )}
      </main>
    </div>
  );
}
