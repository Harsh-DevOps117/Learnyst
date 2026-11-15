import { Brand } from "@/components/workspace/Brand";
import { ArrowUpRight, Check, Sparkles } from "lucide-react";
export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="auth-page">
      <aside className="auth-story">
        <Brand light />
        <div className="auth-story-main">
          <span className="eyebrow">
            <Sparkles size={15} /> THE NEXT CHAPTER IS YOURS
          </span>
          <h1>
            Small steps.
            <br />
            Extraordinary
            <br />
            <em>possibilities.</em>
          </h1>
          <p>
            A space to share what you know.
            <br />
            And discover what you’re capable of.
          </p>
          <div className="auth-preview">
            <span className="auth-preview-check">
              <Check size={19} />
            </span>
            <div>
              <strong>One step closer.</strong>
              <p>Every lesson is a new beginning.</p>
            </div>
            <ArrowUpRight size={22} />
            <div className="auth-preview-progress">
              <span />
            </div>
          </div>
        </div>
        <div className="auth-story-footer">
          <span className="tiny-orbit" /> MADE FOR CURIOUS MINDS.
        </div>
        <div className="auth-decor decor-one" />
        <div className="auth-decor decor-two" />
      </aside>
      <main className="auth-main">
        <div className="auth-mobile-brand">
          <Brand />
        </div>
        <div className="auth-form-wrap">{children}</div>
        <footer className="auth-footnote">
          A little curiosity can change a lot.
        </footer>
      </main>
    </div>
  );
}
