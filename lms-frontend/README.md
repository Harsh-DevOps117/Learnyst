# Learnyst frontend

Next.js 16, React 19 and TypeScript. Authentication and the Tenant, Teacher and Student workspaces use the public landing page's charcoal palette, Space Grotesk headings, Inter body text, Space Mono labels and white pill buttons. The responsive shell has role-specific navigation and real API-backed actions. Fonts and artwork are local assets.

Start the full app from the **repository root** with `npm run dev`. Open http://localhost:3005. See [the root README](../README.md) for demo accounts, local email, configuration and verification.

Browser API requests use `/api/auth`, `/api/core` and `/uploads`, proxied by `next.config.ts`. Session credentials live in HttpOnly cookies; the Zustand store holds public profile data only. Fonts and course artwork are served locally.

```bash
npm ci
npm run lint
npx tsc --noEmit
npm run build
npx playwright install chromium
npm run test:e2e   # Full local app must already be running.
```

Use `AUTH_SERVICE_URL` and `CORE_SERVICE_URL` to override backend proxy targets. The root launcher supplies these values when custom ports are configured. For a standalone production preview, `npm run start -- --port 3005 --hostname 127.0.0.1` serves the compiled frontend while the Go services remain running separately.
