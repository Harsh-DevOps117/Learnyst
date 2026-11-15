import Link from "next/link";
export function Brand({
  light = false,
  href = "/",
}: {
  light?: boolean;
  href?: string;
}) {
  return (
    <Link
      href={href}
      className={`brand ${light ? "brand-light" : ""}`}
      aria-label="Learnyst home"
    >
      <span className="brand-symbol" aria-hidden="true">
        <i />
        <i />
        <i />
      </span>
      learnyst<span className="brand-dot">.</span>
    </Link>
  );
}
