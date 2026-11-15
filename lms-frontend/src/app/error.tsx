"use client";
import { ErrorState } from "@/components/workspace/Shared";
export default function ErrorPage({ reset }: { reset: () => void }) {
  return (
    <main style={{ padding: 30 }}>
      <ErrorState
        message="The page couldn’t finish loading. Please try again."
        retry={reset}
      />
    </main>
  );
}
