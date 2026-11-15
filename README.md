# Learnyst — local teaching & learning workspace

Run the Next.js frontend and Go auth/core services locally, with PostgreSQL for data and files on your own disk. **No Azure account, cloud keys, SMS service, or paid API is needed.**

## Start

Prerequisites: Node.js 22+, Go 1.26.3+, Docker with Compose, and FFmpeg with either the `libx264` or `libopenh264` encoder (for the sample video).

```bash
git clone https://github.com/Harsh-DevOps117/Learnyst.git
cd Learnyst
npm run dev
```

The launcher starts the database and Mailpit, installs frontend dependencies if needed, generates an 18-second preview video, builds the Go services, runs migrations and seeds local sample data. Initial setup needs internet access for dependencies and container images. Subsequent runs reuse the database and local uploads.

- **App:** http://localhost:3005
- **Local email inbox:** http://localhost:8025
- Auth: `127.0.0.1:3010`; core: `127.0.0.1:3011`
- PostgreSQL: `127.0.0.1:55432`, database/user `learnyst`, local-only password `learnyst-local-only`

The default ports avoid the old project's `3000`/`3001` ports. Browser requests go through the frontend's same-origin API proxy.

```bash
# Stop the app and its two containers. Saved data is retained.
npm run stop

# Or stop app processes with Ctrl+C in the launcher terminal.
# Containers can then be stopped with: docker compose stop
```

Do not run `docker compose down -v` unless you intend to erase this project's local database. Uploaded files live in `.local/uploads/`, independently of the database volume.

## Try each role

The login page includes **Student demo** and **Teacher demo** buttons. Signup also offers **I run an academy**, **Teacher** and **Student** accounts; local signup immediately signs you in, with no external OTP dependency.

| Account | Email                    | Development password |
| ------- | ------------------------ | -------------------- |
| Teacher | `teacher@learnyst.local` | `LocalTeacher123!`   |
| Student | `student@learnyst.local` | `LocalStudent123!`   |

Demo passwords are only used when first creating the accounts. Changing a password persists across restarts. Demo seeding is idempotent and only runs with both `ENV=development` and `SEED_DEMO=true`. It never creates a default superadmin. Sample videos are explicitly labelled previews, ready to replace with your own content.

**Tenant / academy owner:** configure the academy name, public address and description, invite teachers through local email, remove or restore their academy access, and monitor academy courses and enrollment totals. Owner settings and team controls are unavailable to teachers and students.

**Teacher:** create a draft course, edit its details, upload an MP4/WebM video and optional PDF notes, publish, send student invitations, create assignments and review submissions. A teacher can accept an owner invitation and select the shared academy, or work in an automatically created private academy. Courses with enrolled students can be unpublished but cannot be deleted, preserving their work.

**Student:** browse published courses, enroll in free courses, watch and seek local videos, resume saved progress, mark lessons complete, submit assignments and read teacher feedback. Profile changes and password updates are available in both workspaces. The sample academy is at http://localhost:3005/learning-studio.

**Password recovery and invitations:** emails are delivered to Mailpit instead of the internet. Open the local inbox, select the message and follow its frontend link. Reset links expire in 30 minutes and are single-use. Course invitations expire in seven days and must be accepted by the invited student's email.

## Authentication and access

- Random 256-bit sessions in HttpOnly, SameSite=Lax cookies. Only hashed session tokens are stored in PostgreSQL; no credentials are stored in localStorage.
- Login restores after a page reload. Logout, password reset and password change revoke sessions in the database immediately.
- Server-side user, role, academy, enrollment and course-ownership checks. Client-supplied user/teacher IDs cannot impersonate another account.
- Signup cannot request the superadmin role. Passwords are bcrypt hashed with a 72-byte limit. Auth routes are rate limited, and mutation requests check the browser origin.
- Uploads validate file type, have a 256 MB limit and use random filenames. Local media requires course access, supports byte-range requests and does not expose the upload directory.

This setup binds app services and container ports to loopback and uses HTTP cookies for localhost. It is a development setup; public deployment would need HTTPS, secure cookies and deployment-specific configuration.

## Configuration

All defaults work without creating an environment file. To override ports or turn off demos:

```bash
cp .env.example .env
# Edit LEARNYST_WEB_PORT, LEARNYST_AUTH_PORT, LEARNYST_CORE_PORT,
# LEARNYST_DB_PORT or SEED_DEMO, then restart npm run dev.
```

Changing `SEED_DEMO=false` stops future seeding; it does not delete existing demo data. Mailpit uses ports `1025` and `8025`. If a required port is occupied, the launcher reports the conflict; it does not stop unrelated projects.

The original Kafka video worker, site builder, tenant proxy and AI service remain as source. The default local app uses direct MP4/WebM playback, so Kafka, Cassandra, Redis, HLS transcoding and site generation are not startup dependencies. External AI and paid checkout are not exposed by the local API; it does not simulate payments or AI responses. Azure adapters have been replaced with filesystem adapters, and tracked binaries and hardcoded credentials have been removed.

**Credential cleanup:** removing secrets from the working tree does not remove them from earlier Git commits or existing clones. Revoke/rotate the Azure keys that were previously committed before reusing that account. No remote history has been rewritten.

## Verify

Keep `npm run dev` running in another terminal.

```bash
# Nine backend integration tests against real local services
npm test

# Type/lint checks and an optimized production build
npm --prefix lms-frontend run lint
npm --prefix lms-frontend run typecheck
npm run build

# Seven browser workflows: owner/team permissions, signup, teacher authoring,
# student video/progress, assignment feedback, mobile and password recovery
cd lms-frontend
npx playwright install chromium
npm run test:e2e

# Dependency audit
npm audit
```

If Chromium is already installed elsewhere, set `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` to its executable when running browser tests. Tests create uniquely named local fixtures and clean up only users with test email addresses and their records. They do not alter the demo users' passwords or progress. Screenshots/traces from failed browser tests are saved under `lms-frontend/test-results/`.

For the current 7:30 role-based frontend tour with pages scrolled from top to bottom, run:

```bash
npm run record:full
```

The `record` and `record:roles` commands run the same tour. Install Playwright Chromium first; FFmpeg is required for MP4 export. Set `LEARNYST_DEMO_DIR` to override the default Desktop output directory. Generated videos are local artifacts and are not committed to the repository.

This saves `Learnyst-Landing-Design-Full-Demo.mp4` and its chapter index on the Desktop. The landing page and all authentication, Tenant, Teacher and Student screens share its charcoal palette, Space Grotesk headings, Inter body text, mono labels and white pill buttons. The hero keeps “THE INFRASTRUCTURE” together at desktop and mobile widths. The tour has separate Tenant, Teacher and Student chapters: the owner configures an academy and invites a teacher; the teacher joins that academy, publishes lessons and reviews work; the student enrolls, watches, downloads notes, submits work and reads feedback. Desktop/mobile screens and password recovery are included. Hinglish captions appear below the product viewport. Footage is grouped by role, with the teacher review recorded after the student submission.

The recorder exercises teacher signup, course creation, local video upload, publishing, invitations, assignments, student signup/enrollment, video playback and saved progress, teacher feedback, profile changes, password recovery and mobile navigation. It uses temporary test accounts and removes them after recording.

Go services can also be checked with `go test ./...` inside each service directory. Auth/core run together via the root launcher; running the frontend alone does not start its API services.

## Layout

```text
scripts/dev.mjs            One-command local launcher
scripts/stop.mjs           Stop this project's app and containers
compose.yaml              PostgreSQL + local email inbox
lms-api-gateway/           Cookie sessions, signup/login and password recovery
lms-core-service/          Authorized courses, enrollment, media and assignments
lms-frontend/             Next.js teacher/student workspaces
.local/                   Ignored build binaries, process info and uploaded files
tests/                    Backend integration tests and fixture cleanup
```
