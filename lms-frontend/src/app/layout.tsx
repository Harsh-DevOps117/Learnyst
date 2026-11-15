import type { Metadata } from "next";
import { Toaster } from "react-hot-toast";
import { SessionProvider } from "@/components/providers/SessionProvider";
import "./globals.css";
export const metadata: Metadata = {
  title: {
    default: "Learnyst — Make room for your next chapter",
    template: "%s · Learnyst",
  },
  description:
    "A thoughtful workspace for teaching, learning and making progress.",
};
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" data-scroll-behavior="smooth">
      <body>
        <SessionProvider>{children}</SessionProvider>
        <Toaster
          position="bottom-right"
          toastOptions={{
            style: {
              background: "#2a2a2a",
              color: "#fff",
              border: "1px solid #ffffff26",
              borderRadius: "4px",
              fontFamily: "inherit",
              fontSize: "14px",
            },
          }}
        />
      </body>
    </html>
  );
}
