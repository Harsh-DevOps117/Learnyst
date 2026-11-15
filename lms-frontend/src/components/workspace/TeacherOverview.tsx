import Link from "next/link";
import { ArrowRight, ArrowUpRight, BookOpen, Sparkles } from "lucide-react";
import type { Course } from "@/lib/types";
export function TeacherOverview({
  firstCourse,
  onCreate,
}: {
  firstCourse?: Course;
  onCreate: () => void;
}) {
  return (
    <div className="overview-feature-grid">
      <section className="welcome-hero">
        <div className="welcome-hero-content">
          <span className="eyebrow">
            <span className="hero-tiny-dot" />
            TURN KNOWLEDGE INTO POSSIBILITY
          </span>
          <h2>
            You know something
            <br />
            worth <em>sharing.</em>
          </h2>
          <p>
            Give your ideas a home. Create a course that makes someone’s next
            step a little easier.
          </p>
          <button onClick={onCreate} className="btn btn-lime">
            Create a course
            <ArrowUpRight size={17} />
          </button>
        </div>
        <div className="hero-illustration" aria-hidden="true">
          <span className="hero-orbit" />
          <span className="hero-orbit orbit-offset" />
          <div className="hero-book">
            <span />
            <BookOpen size={62} strokeWidth={1.1} />
            <i />
          </div>
          <span className="hero-spark spark-one">✳</span>
          <span className="hero-spark spark-two">✦</span>
          <span className="hero-small-label">GROW AT YOUR OWN PACE</span>
        </div>
      </section>
      <section className="rhythm-card">
        <span className="eyebrow">THE CREATIVE PROCESS</span>
        <span className="rhythm-icon">
          <Sparkles size={25} strokeWidth={1.4} />
        </span>
        <h3>A good idea deserves a beginning.</h3>
        <p>Your draft doesn’t have to be perfect. It just has to be a start.</p>
        <div className="rhythm-bottom">
          <Link
            href={
              firstCourse
                ? `/dashboard/courses/${firstCourse.id}`
                : "/dashboard?tab=courses&action=create_course"
            }
          >
            {firstCourse ? "Shape your next course" : "Make your first course"}
            <ArrowRight size={16} />
          </Link>
        </div>
      </section>
    </div>
  );
}
