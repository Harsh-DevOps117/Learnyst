import { chromium, expect as baseExpect } from "@playwright/test";
import { mkdir, readFile, writeFile, copyFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import path from "node:path";
import { homedir } from "node:os";
import { cleanTestUsers } from "../../tests/fixtures.mjs";
const root = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
process.chdir(root);
const work = path.join(root, ".local/landing-design-demo");
const stateFile = path.join(work, "state.json");
const base = process.env.TEST_WEB_URL || "http://localhost:3005";
await mkdir(path.join(work, "raw"), { recursive: true });
const previous = existsSync(stateFile)
  ? JSON.parse(await readFile(stateFile, "utf8"))
  : null;
const stamp = String(Date.now()).slice(-7);
const S =
  previous && !previous.finished
    ? previous
    : {
        owner: `test-role-owner-${stamp}@learnyst.local`,
        teacher: `test-role-teacher-${stamp}@learnyst.local`,
        student: `test-role-student-${stamp}@learnyst.local`,
        password: "Learnyst-walkthrough123!",
        resetPassword: "My-new-chapter123!",
        finalPassword: "My-secure-workspace123!",
        ids: [],
        assets: [],
        clips: [],
        courses: {},
        cookieFile: path.join(work, "session.json"),
      };
const save = () => writeFile(stateFile, JSON.stringify(S, null, 2));
await save();
const expect = baseExpect.configure({ timeout: 12000 });
const browser = await chromium.launch({
  headless: true,
  executablePath:
    process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH ||
    chromium.executablePath(),
});
const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
let active,
  started = 0,
  cues = [];
async function explain(page, chapter, text, hold = 3300) {
  console.log(`[${active}] ${chapter}: ${text}`);
  cues.push({ start: (Date.now() - started) / 1000, chapter, text });
  await pause(hold);
}
async function top(page) {
  await page.evaluate(() => window.scrollTo({ top: 0, behavior: "smooth" }));
  await pause(450);
}
async function bottom(page) {
  await page.evaluate(() =>
    window.scrollTo({
      top: document.documentElement.scrollHeight,
      behavior: "smooth",
    }),
  );
  await pause(1300);
}
async function scan(
  page,
  chapter,
  text,
  { hold = 3300, dwell = 1500, reset = true } = {},
) {
  await top(page);
  await explain(page, chapter, text, hold);
  const bounds = await page.evaluate(() => ({
    end: Math.max(0, document.documentElement.scrollHeight - innerHeight),
    step: Math.floor(innerHeight * 0.68),
  }));
  for (
    let y = Math.min(bounds.step, bounds.end);
    y > 0;
    y = Math.min(y + bounds.step, bounds.end)
  ) {
    await page.evaluate(
      (y) => window.scrollTo({ top: y, behavior: "smooth" }),
      y,
    );
    await pause(dwell);
    if (y === bounds.end) break;
  }
  await pause(900);
  if (reset) await top(page);
}
async function focus(locator, hold = 1200) {
  await locator.scrollIntoViewIfNeeded();
  await locator.evaluate((el) =>
    el.animate(
      [
        { boxShadow: "0 0 0 3px rgba(255,255,255,.75)" },
        { boxShadow: "0 0 0 3px rgba(255,255,255,0)" },
      ],
      { duration: 1900 },
    ),
  );
  await pause(hold);
}
async function fill(locator, value) {
  await locator.scrollIntoViewIfNeeded();
  await locator.fill(value);
  await pause(350);
}
async function nav(page, name, tab) {
  if (await page.getByRole("button", { name: "Open navigation" }).isVisible()) {
    await page.getByRole("button", { name: "Open navigation" }).click();
    await pause(800);
  }
  await page
    .locator("aside.sidebar")
    .getByRole("link", { name, exact: true })
    .click();
  await page.waitForURL(
    (url) =>
      url.pathname === "/dashboard" && url.searchParams.get("tab") === tab,
  );
  await pause(350);
}
async function logout(page) {
  if (await page.getByRole("button", { name: "Open navigation" }).isVisible()) {
    await page.getByRole("button", { name: "Open navigation" }).click();
    await pause(500);
  }
  await page.getByRole("button", { name: "Sign out", exact: true }).click();
  await page.waitForURL((url) => url.pathname === "/login");
  await pause(800);
}
async function login(page, email, password = S.password) {
  await page.goto(base + "/login");
  await fill(page.locator("input[name=email]"), email);
  await fill(page.locator("input[name=password]"), password);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await page.waitForURL("**/dashboard");
  await expect(page.locator(".welcome-row h1")).toBeVisible();
  await pause(800);
}
async function signup(page, first, last, email, teacher = false) {
  await page.goto(base + "/signup");
  if (teacher)
    await page.getByRole("button", { name: /I’m here to teach/ }).click();
  await fill(page.getByLabel("First name"), first);
  await fill(page.getByLabel("Last name"), last);
  await fill(page.getByLabel("Email address"), email);
  await fill(page.locator("input[name=password]"), S.password);
  await fill(page.getByLabel("Confirm password"), S.password);
  await page.getByRole("button", { name: "Create my account" }).click();
  await page.waitForURL("**/dashboard");
  await expect(page.locator(".welcome-row h1")).toBeVisible();
  const u = await (await page.request.get(base + "/api/auth/session")).json();
  S.ids.push(u.user.id);
  await save();
  await pause(900);
}
async function createCourse(
  page,
  title,
  category = "Development",
  level = "Beginner",
  description = "",
) {
  await page.goto(base + "/dashboard?tab=courses");
  await page
    .getByRole("button", { name: "Create course", exact: true })
    .first()
    .click();
  const d = page.getByRole("dialog");
  await fill(d.getByLabel("Course title"), title);
  await fill(
    d.getByLabel("A little about the course"),
    description ||
      "A practical course with short lessons, useful notes and a chance to put new skills into practice.",
  );
  await d.locator("select[name=category]").selectOption(category);
  await d.locator("select[name=level]").selectOption(level);
  await pause(1000);
  await d.getByRole("button", { name: "Create course", exact: true }).click();
  await page.waitForURL(/\/dashboard\/courses\/\d+/);
  await expect(page.locator(".course-page-heading h1")).toHaveText(title);
  return Number(page.url().split("/").pop());
}
async function addLesson(page, courseID, title, notes = false) {
  await page.getByRole("button", { name: "Add lesson", exact: true }).click();
  const d = page.getByRole("dialog");
  await fill(d.getByLabel("Lesson title"), title);
  await fill(
    d.locator("textarea[name=description]"),
    "This preview demonstrates the learning workflow. Replace this short clip with your own teaching video when creating a real course.",
  );
  await d
    .locator("input[name=video]")
    .setInputFiles(path.join(root, ".local/uploads/sample.mp4"));
  if (notes)
    await d
      .locator("input[name=notes]")
      .setInputFiles(path.join(work, "lesson-notes.pdf"));
  await pause(1500);
  await d.getByRole("button", { name: "Add lesson", exact: true }).click();
  await expect(d).not.toBeVisible();
  await expect(page.locator(".editor-lessons")).toContainText(title);
  const detail = await (
    await page.request.get(base + `/api/core/courses/${courseID}`)
  ).json();
  for (const v of detail.videos)
    for (const u of [v.video_url, v.note_url])
      if (u && !S.assets.includes(u)) S.assets.push(u);
  await save();
}
async function assignment(page, courseID, title, due = false) {
  await page.goto(base + "/dashboard?tab=assignments");
  await page
    .getByRole("button", { name: "Create assignment", exact: true })
    .click();
  const d = page.getByRole("dialog");
  await d.getByLabel("Course", { exact: true }).selectOption(String(courseID));
  await fill(d.getByLabel("Assignment title"), title);
  await fill(
    d.getByLabel("Instructions"),
    "Describe the audience for your page, the main message you want to share, and the first three steps you will take. You can include a link to your work.",
  );
  if (due)
    await d.locator("input[name=due_date]").fill(
      new Date(Date.now() + 7 * 86400000).toLocaleDateString("en-CA", {
        timeZone: "Asia/Kolkata",
      }),
    );
  await pause(1200);
  await d
    .getByRole("button", { name: "Create assignment", exact: true })
    .click();
  await expect(d).not.toBeVisible();
  await expect(
    page.locator(".assignment-summary").filter({ hasText: title }),
  ).toBeVisible();
}
async function mail(email, subject) {
  const list = await (
    await fetch("http://127.0.0.1:8025/api/v1/messages")
  ).json();
  const m = list.messages.find(
    (m) => m.To.some((t) => t.Address === email) && m.Subject.includes(subject),
  );
  if (!m) throw new Error("Local email did not arrive");
  const detail = await (
    await fetch(`http://127.0.0.1:8025/api/v1/message/${m.ID}`)
  ).json();
  return { ...m, token: detail.Text.match(/token=([a-f0-9]{64})/)[1] };
}
async function showInbox(page, email, subject) {
  await page.goto("http://localhost:8025");
  const search = page.getByRole("textbox", { name: "Search", exact: true });
  await search.fill("to:" + email);
  await search.press("Enter");
  await page.getByText(subject, { exact: true }).first().click();
  await pause(1000);
}
async function segment(name, fn, mobile = false, target = 0) {
  if (S.clips.some((c) => c.name === name)) {
    console.log(`Skipping completed segment ${name}`);
    return;
  }
  active = name;
  cues = [];
  const viewport = mobile
    ? { width: 390, height: 844 }
    : { width: 1600, height: 900 };
  const context = await browser.newContext({
    viewport,
    timezoneId: "Asia/Kolkata",
    acceptDownloads: true,
    storageState: existsSync(S.cookieFile) ? S.cookieFile : undefined,
    recordVideo: { dir: path.join(work, "raw"), size: viewport },
  });
  await context.addInitScript(() => {
    document.addEventListener("pointerdown", (e) => {
      const ring = document.createElement("span");
      ring.style.cssText = `pointer-events:none;position:fixed;z-index:2147483647;left:${e.clientX - 18}px;top:${e.clientY - 18}px;width:36px;height:36px;border:2px solid #fff;border-radius:50%;background:#ffffff20;`;
      document.documentElement.appendChild(ring);
      ring.animate(
        [
          { transform: "scale(.7)", opacity: 1 },
          { transform: "scale(1.7)", opacity: 0 },
        ],
        { duration: 700 },
      );
      setTimeout(() => ring.remove(), 750);
    });
  });
  const page = await context.newPage();
  page.setDefaultTimeout(14000);
  started = Date.now();
  try {
    await fn(page);
    const contentDuration = (Date.now() - started) / 1000;
    const remaining = Math.min(1000, target * 1000 - (Date.now() - started));
    if (remaining > 0) await pause(remaining);
    await pause(700);
    await context.storageState({ path: S.cookieFile });
    const video = page.video();
    await context.close();
    const output = path.join(work, "raw", name + ".webm");
    await video.saveAs(output);
    S.clips.push({ name, mobile, file: output, cues, contentDuration });
    await save();
    console.log(`Completed ${name}`);
  } catch (e) {
    await page
      .screenshot({
        path: path.join(work, name + "-failure.png"),
        fullPage: true,
      })
      .catch(() => {});
    await context.close();
    throw e;
  }
}
// A small real PDF to demonstrate secure lesson-note uploads/downloads.
const pdfObjects = [
  "<< /Type /Catalog /Pages 2 0 R >>",
  "<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
  "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
];
const pdfText =
  "BT /F1 23 Tf 50 730 Td (Build a useful web page) Tj /F1 13 Tf 0 -45 Td (Lesson notes - Learnyst local walkthrough) Tj 0 -40 Td (1. Start with the person who will use your page.) Tj 0 -26 Td (2. Give the content a clear, readable structure.) Tj 0 -26 Td (3. Test the page on a small screen.) Tj 0 -40 Td (Practice: describe your audience and your first three steps.) Tj ET";
pdfObjects.push(
  `<< /Length ${Buffer.byteLength(pdfText)} >>\nstream\n${pdfText}\nendstream`,
  "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
);
let pdf = "%PDF-1.4\n";
const offsets = [0];
for (let i = 0; i < pdfObjects.length; i++) {
  offsets.push(Buffer.byteLength(pdf));
  pdf += `${i + 1} 0 obj\n${pdfObjects[i]}\nendobj\n`;
}
const xref = Buffer.byteLength(pdf);
pdf +=
  `xref\n0 6\n0000000000 65535 f \n` +
  offsets
    .slice(1)
    .map((n) => String(n).padStart(10, "0") + " 00000 n \n")
    .join("") +
  `trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`;
await writeFile(path.join(work, "lesson-notes.pdf"), pdf);
const COURSE = "Build your first web experience",
  TASK = "From an idea to a thoughtful page";
async function ownerSignup(page) {
  await page.goto(base + "/signup?role=tenant");
  await expect(
    page.getByRole("button", { name: /I run an academy/ }),
  ).toHaveAttribute("aria-pressed", "true");
  await fill(page.getByLabel("First name"), "Neha");
  await fill(page.getByLabel("Last name"), "Verma");
  await fill(page.getByLabel("Email address"), S.owner);
  await fill(page.locator("input[name=password]"), S.password);
  await fill(page.getByLabel("Confirm password"), S.password);
  await page.getByRole("button", { name: "Create my account" }).click();
  await page.waitForURL("**/dashboard");
  await expect(page.locator(".header-role")).toContainText("Tenant");
  const user = await (
    await page.request.get(base + "/api/auth/session")
  ).json();
  S.ids.push(user.user.id);
  await save();
}
S.slug ||= "northstar-" + stamp;
await save();
try {
  await segment(
    "01-homepage",
    async (page) => {
      await page.goto(base);
      await expect(
        page.getByRole("heading", { name: /THE INFRASTRUCTURE/ }),
      ).toBeVisible();
      await focus(
        page.getByRole("button", { name: "START TEACHING", exact: true }),
      );
      await scan(
        page,
        "ORIGINAL HOMEPAGE",
        "Landing page ka dark design ab poori app mein same hai: charcoal surfaces, white buttons aur Space Grotesk headings. THE heading ke saath ek line mein hai.",
        { dwell: 1200, hold: 3500 },
      );
      await page
        .getByRole("button", { name: "SIGN UP/IN", exact: true })
        .click();
      await page.waitForURL((url) => url.pathname === "/login");
      await explain(
        page,
        "THREE ROLES — TENANT, TEACHER, STUDENT",
        "Tenant academy/team manage karta hai; teacher courses aur reviews karta hai; student enroll, learn aur submit karta hai.",
      );
    },
    false,
    35,
  );
  await segment(
    "02-tenant-setup",
    async (page) => {
      await ownerSignup(page);
      await explain(
        page,
        "TENANT / ACADEMY OWNER — SEPARATE WORKSPACE",
        "Neha academy owner hai. Is role ko academy settings, teaching team, academy courses aur students ke controls milte hain.",
      );
      await nav(page, "Academy settings", "academy");
      await fill(page.getByLabel("Academy name"), "Northstar Academy");
      await fill(page.getByLabel("Public academy address"), S.slug);
      await fill(
        page.getByLabel("Academy description"),
        "Practical web development, thoughtful design and useful feedback. Learn with our teaching team, one clear step at a time.",
      );
      await page.getByRole("button", { name: "Save academy" }).click();
      await pause(900);
      const academy = await (
        await page.request.get(base + "/api/core/academy")
      ).json();
      S.tenantID = academy.academy.id;
      await save();
      await explain(
        page,
        "TENANT — ACADEMY IDENTITY",
        "Owner academy ka naam, public address aur introduction set karta hai. Save ke baad public academy page par wahi details aati hain.",
      );
      await page.goto(base + "/" + S.slug);
      await scan(
        page,
        "TENANT — PUBLIC ACADEMY PAGE",
        "Academy ab apne address par available hai. Teacher ke publish karne ke baad courses yahan students ko dikhne lagenge.",
        { hold: 2200, dwell: 1200 },
      );
      await page.goto(base + "/dashboard?tab=team");
      await page
        .getByRole("button", { name: "Invite teacher", exact: true })
        .click();
      await fill(
        page.getByRole("dialog").getByLabel("Teacher email"),
        S.teacher,
      );
      await explain(
        page,
        "TENANT — INVITE A TEACHER",
        "Owner teacher ko invite karta hai. Teacher apne account se join karega; owner ke settings aur team-management rights usko nahi milte.",
      );
      await page
        .getByRole("button", { name: "Send teacher invitation" })
        .click();
      await expect(page.getByRole("dialog")).not.toBeVisible();
      S.teacherToken = (await mail(S.teacher, "academy teacher")).token;
      await save();
      await expect(page.locator("table")).toContainText("Pending");
      await showInbox(
        page,
        S.teacher,
        "You have an academy teacher invitation",
      );
      await explain(
        page,
        "TENANT — REAL LOCAL INVITATION EMAIL",
        "Invitation local inbox mein deliver hui hai. Teacher ko isi invited email aur teacher account se accept karna hoga.",
      );
      await page.goto(base + "/dashboard");
      await scan(
        page,
        "TENANT — OWNER OVERVIEW",
        "Owner ka overview academy identity aur actual academy totals dikhata hai. Teaching aur student learning apne alag roles mein honge.",
        { hold: 2200, dwell: 1200 },
      );
    },
    false,
    65,
  );
  await segment(
    "03-teacher-authoring",
    async (page) => {
      await page.goto(base + "/dashboard");
      await expect(page.locator(".welcome-row h1")).toBeVisible();
      await logout(page);
      if (S.resumeTeacher) await login(page, S.teacher);
      else await signup(page, "Aarav", "Mehta", S.teacher, true);
      await page.goto(base + "/invite/accept?token=" + S.teacherToken);
      await page.getByRole("button", { name: "Accept invitation" }).click();
      await page.waitForURL("**/dashboard");
      await expect(
        page.getByRole("combobox", { name: "Active academy" }),
      ).toHaveValue(String(S.tenantID));
      await explain(
        page,
        "TEACHER — JOIN THE TENANT’S ACADEMY",
        "Aarav teacher account se Northstar Academy join karta hai. Academy selector aur Teacher workspace label selected academy confirm karte hain.",
      );
      await nav(page, "My courses", "courses");
      S.courses.main = await createCourse(
        page,
        COURSE,
        "Development",
        "Beginner",
        "Understand your audience, give your content a readable structure, and build a practical first web experience. Includes preview lessons and PDF notes.",
      );
      await save();
      await explain(
        page,
        "TEACHER — COURSE DRAFT AND EDITOR",
        "Teacher apna course title, description, category aur level set karta hai. Course pehle draft hai; owner academy aur team control karta rehta hai.",
      );
      await page
        .getByRole("button", { name: "Publish course", exact: true })
        .click();
      await expect(page.locator(".notice.error")).toContainText(
        "Add at least one lesson",
      );
      await pause(1200);
      await addLesson(page, S.courses.main, "Start with your audience", true);
      await addLesson(page, S.courses.main, "Organise your content");
      await bottom(page);
      await explain(
        page,
        "TEACHER — LESSONS, LOCAL VIDEO AND PDF NOTES",
        "Teacher curriculum mein videos aur optional PDF notes add karta hai. Uploads local storage mein save hote hain; Azure credentials ki zaroorat nahi.",
      );
      await top(page);
      await page
        .getByRole("button", { name: "Preview Start with your audience" })
        .click();
      await page
        .getByRole("dialog")
        .locator("video")
        .evaluate((v) => v.play());
      await pause(2200);
      await page
        .getByRole("dialog")
        .locator("video")
        .evaluate((v) => v.pause());
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Close dialog" })
        .click();
      await page
        .getByRole("button", { name: "Publish course", exact: true })
        .click();
      await expect(page.locator(".heading-badges")).toContainText("Published");
      await explain(
        page,
        "TEACHER — PREVIEW AND PUBLISH",
        "Teacher actual lesson preview karke course publish karta hai. Published course tenant ki public academy aur student catalog mein available hota hai.",
      );
      await assignment(page, S.courses.main, TASK, true);
      await page
        .locator(".assignment-item")
        .filter({ hasText: TASK })
        .locator("summary")
        .click();
      await scan(
        page,
        "TEACHER — ASSIGNMENT INSTRUCTIONS",
        "Teacher course ke liye instructions, maximum marks aur deadline set karta hai. Student ka submitted work isi panel mein review hoga.",
        { hold: 2200, dwell: 1200 },
      );
      S.courses.draft = await createCourse(
        page,
        "Next course - planning",
        "Design",
        "Intermediate",
        "A private draft inside Northstar Academy.",
      );
      await save();
      await page.goto(base + "/dashboard?tab=courses");
      await page
        .getByRole("button", { name: "Published", exact: true })
        .click();
      await pause(1000);
      await page.getByRole("button", { name: "Drafts", exact: true }).click();
      await pause(1000);
      await page
        .getByRole("button", { name: "All courses", exact: true })
        .click();
      await explain(
        page,
        "TEACHER — OWN COURSE LIBRARY",
        "Published/draft filters aur search se teacher apne courses manage karta hai. Woh academy owner ki settings ya doosre teacher ke private course ko edit nahi kar sakta.",
      );
    },
    false,
    110,
  );
  await segment(
    "04-teacher-mobile",
    async (page) => {
      await page.goto(base + "/dashboard?tab=courses");
      await scan(
        page,
        "TEACHER — MOBILE WORKSPACE",
        "Teacher ki course library aur create/manage actions mobile par bhi available hain.",
        { hold: 2300, dwell: 1000 },
      );
      await page.goto(base + `/dashboard/courses/${S.courses.main}`);
      await scan(
        page,
        "TEACHER — RESPONSIVE COURSE EDITOR",
        "Mobile editor mein details, curriculum, preview, publish aur resource controls neeche tak accessible hain.",
        { hold: 2200, dwell: 900 },
      );
    },
    true,
    20,
  );
  await segment(
    "05-student-learning",
    async (page) => {
      await page.goto(base + "/dashboard");
      await expect(page.locator(".welcome-row h1")).toBeVisible();
      await logout(page);
      await signup(page, "Riya", "Sharma", S.student);
      await explain(
        page,
        "STUDENT — OWN LEARNING WORKSPACE",
        "Riya student account se sign in karti hai. Is role ko My learning, Explore, assignments aur personal profile milte hain; teacher/owner controls nahi.",
      );
      await page.goto(base + "/" + S.slug);
      await expect(page.locator(".course-card")).toContainText(COURSE);
      await scan(
        page,
        "STUDENT — DISCOVER AND ENROLL",
        "Student tenant ke public academy page par published course dekhti hai aur free course mein enroll karti hai.",
        { hold: 2300, dwell: 1000 },
      );
      await page
        .locator(".course-card")
        .getByRole("button", { name: "Enroll now" })
        .click();
      await page.waitForURL(`**/dashboard/courses/${S.courses.main}`);
      const video = page.locator(".video-container video");
      await expect
        .poll(() => video.evaluate((v) => v.readyState))
        .toBeGreaterThan(0);
      await video.evaluate((v) => v.play());
      await pause(2400);
      await video.evaluate((v) => {
        v.currentTime = 8;
      });
      await pause(700);
      await video.evaluate((v) => v.pause());
      await pause(800);
      await page.reload();
      await expect
        .poll(() =>
          page.locator(".video-container video").evaluate((v) => v.readyState),
        )
        .toBeGreaterThan(0);
      await expect
        .poll(() =>
          page.locator(".video-container video").evaluate((v) => v.currentTime),
        )
        .toBeGreaterThan(6);
      await explain(
        page,
        "STUDENT — VIDEO PLAYBACK, SEEK AND SAVED RESUME",
        "Lesson play/pause aur seeking work karte hain. Refresh ke baad saved watch position restore hui; progress database mein persist hoti hai.",
      );
      await page
        .getByRole("link", { name: "Lesson notes", exact: true })
        .scrollIntoViewIfNeeded();
      const dl = page.waitForEvent("download");
      await page
        .getByRole("link", { name: "Lesson notes", exact: true })
        .click();
      await (await dl).saveAs(path.join(work, "downloaded-notes.pdf"));
      await page.getByRole("button", { name: "Mark as complete" }).click();
      await expect(
        page.getByRole("button", { name: "Lesson complete" }),
      ).toBeVisible();
      await page.locator(".curriculum-lesson").nth(1).click();
      await page.getByRole("button", { name: "Mark as complete" }).click();
      await expect(page.locator(".curriculum-progress-label")).toContainText(
        "100%",
      );
      await scan(
        page,
        "STUDENT — CURRICULUM, PDF NOTES AND COMPLETION",
        "PDF notes download ho gaye. Student lesson list se next video choose karti hai; completed lessons se course progress update hota hai.",
        { hold: 2200, dwell: 1000 },
      );
      await nav(page, "Overview", "overview");
      await scan(
        page,
        "STUDENT — COMPLETE HOME DASHBOARD",
        "Student home mein next course/resume, real progress ring, pending assignments, learning statistics aur course cards hain.",
        { hold: 2300, dwell: 1100 },
      );
      await nav(page, "My learning", "courses");
      await page
        .getByRole("button", { name: "Completed", exact: true })
        .click();
      await pause(1000);
      await page
        .getByRole("button", { name: "All courses", exact: true })
        .click();
      await fill(
        page.getByRole("textbox", { name: "Search courses" }),
        "web experience",
      );
      await explain(
        page,
        "STUDENT — LEARNING LIBRARY",
        "My learning mein completed/in-progress filters, search aur Revisit/Continue controls hain. Yeh teacher ki course-management screen se alag hai.",
      );
      await nav(page, "Assignments", "assignments");
      const a = page.locator(".assignment-item").filter({ hasText: TASK });
      await a.locator("summary").click();
      await fill(
        a.getByLabel("Your response"),
        "My page will help a new creator explain their work.\n\nI will identify the audience, sketch a readable layout, and test it on a small screen. Clear headings will make the next action easy to find.",
      );
      await a.getByRole("button", { name: "Submit assignment" }).click();
      await expect(a.locator("summary")).toContainText("Submitted");
      await explain(
        page,
        "STUDENT — SUBMIT PRACTICAL WORK",
        "Student instructions padhkar apna response submit karti hai. Submitted state aur original response save rehte hain; teacher alag review screen se grade karega.",
      );
    },
    false,
    100,
  );
  await segment(
    "06-teacher-review",
    async (page) => {
      await page.goto(base + "/dashboard");
      await expect(page.locator(".welcome-row h1")).toBeVisible();
      await logout(page);
      await login(page, S.teacher);
      await nav(page, "Students", "students");
      await expect(page.locator(".students-table")).toContainText(
        "Riya Sharma",
      );
      await explain(
        page,
        "TEACHER — STUDENTS AND SUBMITTED WORK",
        "Teacher apne course ke enrolled students dekhta hai. Is example mein Riya ka submitted assignment review ke liye ready hai.",
      );
      await nav(page, "Assignments", "assignments");
      const a = page.locator(".assignment-item").filter({ hasText: TASK });
      await a.locator("summary").click();
      await bottom(page);
      await fill(a.getByLabel("Marks out of 10"), "9");
      await fill(
        a.getByLabel("Feedback"),
        "A clear audience and a practical plan. Keep the headings specific, then ask someone else to test the page and share where they hesitate.",
      );
      await a.getByRole("button", { name: "Save feedback" }).click();
      await expect(a).toContainText("Reviewed");
      await explain(
        page,
        "TEACHER — MARKS AND FEEDBACK",
        "Teacher response review karke 9/10 marks aur feedback save karta hai. Student apne account mein yahi grade aur feedback padhega.",
      );
    },
    false,
    35,
  );
  await segment(
    "07-tenant-monitoring",
    async (page) => {
      await page.goto(base + "/dashboard");
      await expect(page.locator(".welcome-row h1")).toBeVisible();
      await logout(page);
      await login(page, S.owner);
      await nav(page, "Teaching team", "team");
      await expect(page.locator("table").first()).toContainText("Aarav Mehta");
      await expect(page.locator("table").last()).toContainText("Accepted");
      await explain(
        page,
        "TENANT — TEAM ACCESS AND ACADEMY OVERSIGHT",
        "Active academy mein owner accepted teachers, their course counts aur pending invitations monitor karta hai. Team-access control sirf owner ke paas hai.",
      );
      await page
        .getByRole("button", { name: "Remove access", exact: true })
        .click();
      await page
        .getByRole("button", { name: "Restore access", exact: true })
        .click();
      await expect(page.locator("table").first()).toContainText("Active");
      await nav(page, "Overview", "overview");
      await scan(
        page,
        "TENANT — ACTUAL ACADEMY TOTALS",
        "Owner ko academy ke courses, published courses, active teachers aur enrolled students ke real totals dikhte hain. Teacher teaching karta hai; owner academy manage karta hai.",
        { hold: 2200, dwell: 1000 },
      );
    },
    false,
    25,
  );
  await segment(
    "08-student-feedback-security",
    async (page) => {
      await page.goto(base + "/dashboard");
      await expect(page.locator(".welcome-row h1")).toBeVisible();
      await logout(page);
      await login(page, S.student);
      await nav(page, "Assignments", "assignments");
      const a = page.locator(".assignment-item").filter({ hasText: TASK });
      await a.locator("summary").click();
      await expect(a).toContainText("9/10 marks");
      await scan(
        page,
        "STUDENT — READ MARKS AND TEACHER FEEDBACK",
        "Student ko original submission ke saath 9/10 grade aur teacher ka feedback milta hai. Yeh teacher-review aur student-learning loop complete karta hai.",
        { hold: 2300, dwell: 1000 },
      );
      await nav(page, "Profile & settings", "profile");
      await fill(page.getByLabel("Location"), "New Delhi, India");
      await page.getByRole("button", { name: "Save changes" }).click();
      await explain(
        page,
        "STUDENT — PERSONAL PROFILE AND SECURITY",
        "Student apne naam/location aur password ko manage kar sakti hai. Academy settings aur teaching-team rights usko nahi milte.",
      );
      await logout(page);
      await page.goto(base + "/forgot-password");
      await fill(page.getByLabel("Email address"), S.student);
      await page.getByRole("button", { name: "Send reset link" }).click();
      await expect(
        page.getByRole("heading", { name: "Check your inbox." }),
      ).toBeVisible();
      const token = (await mail(S.student, "Reset")).token;
      await page.goto(base + "/reset-password?token=" + token);
      await fill(page.locator("input[name=password]"), S.resetPassword);
      await fill(page.getByLabel("Confirm password"), S.resetPassword);
      await page.getByRole("button", { name: "Update password" }).click();
      await expect(
        page.getByRole("heading", { name: "You’re all set." }),
      ).toBeVisible();
      await explain(
        page,
        "STUDENT — LOCAL PASSWORD RECOVERY",
        "Forgot password ka single-use email link naya password set karta hai. Purane sessions revoke hote hain; ab naya password se sign in hoga.",
      );
      await login(page, S.student, S.resetPassword);
    },
    false,
    30,
  );
  await segment(
    "09-student-mobile",
    async (page) => {
      await page.goto(base + "/dashboard");
      await expect(
        page.getByRole("heading", { name: "On your horizon" }),
      ).toBeVisible();
      if (
        await page.evaluate(
          () => document.documentElement.scrollWidth > innerWidth,
        )
      )
        throw new Error("Student mobile overflow");
      await scan(
        page,
        "STUDENT — MOBILE HOME TOP TO BOTTOM",
        "Same student workspace phone par: resume, progress ring, assignments, stats, course cards aur footer. Horizontal overflow nahi hai.",
        { hold: 2200, dwell: 900 },
      );
      await nav(page, "Assignments", "assignments");
      await page
        .locator(".assignment-item")
        .filter({ hasText: TASK })
        .locator("summary")
        .click();
      await scan(
        page,
        "STUDENT — MOBILE ASSIGNMENTS AND FEEDBACK",
        "Phone par instructions, submitted response, marks aur teacher feedback poora readable hai. Menu se baaki student pages khulte hain.",
        { hold: 2200, dwell: 900 },
      );
      await nav(page, "Profile & settings", "profile");
      await scan(
        page,
        "STUDENT — MOBILE PROFILE / DEMO COMPLETE",
        "Tenant academy/team manage karta hai, teacher courses aur review karta hai, aur student enroll/learn/submit karta hai. Har role ka apna actual workspace hai.",
        { hold: 2200, dwell: 800 },
      );
    },
    true,
    30,
  );
  await segment("10-tenant-library-students", async (page) => {
    await page.goto(base + "/dashboard");
    await logout(page);
    await login(page, S.owner);
    await nav(page, "Academy courses", "courses");
    await scan(
      page,
      "TENANT — ACADEMY COURSE LIBRARY",
      "Owner academy ke published courses aur drafts ko oversee karta hai. Search, status filters aur course details isi academy ke records dikhate hain.",
    );
    await nav(page, "Students", "students");
    await scan(
      page,
      "TENANT — ACADEMY STUDENT ROSTER",
      "Owner academy ke enrolled students aur unke course records monitor karta hai. Teaching aur grading teacher ke workflow mein hoti hai.",
    );
  });
  await segment("11-teacher-overview-profile", async (page) => {
    await page.goto(base + "/dashboard");
    await logout(page);
    await login(page, S.teacher);
    await scan(
      page,
      "TEACHER — HOME / TEACHING OVERVIEW",
      "Teacher overview mein uske courses, student totals aur teaching activity milti hai. Same academy ke andar bhi course ownership alag rehti hai.",
    );
    await nav(page, "Profile & settings", "profile");
    await scan(
      page,
      "TEACHER — PERSONAL PROFILE AND PASSWORD",
      "Teacher apni personal details aur password manage karta hai. Academy ka public naam, address aur staff permissions owner ke controls hain.",
    );
  });
  await segment("12-student-explore", async (page) => {
    await page.goto(base + "/dashboard");
    await logout(page);
    await login(page, S.student, S.resetPassword);
    await nav(page, "Explore courses", "explore");
    await scan(
      page,
      "STUDENT — EXPLORE COURSES",
      "Explore page par published courses, categories aur course search available hain. Student course choose karke enroll ya enrolled course ko continue kar sakti hai.",
    );
  });
} finally {
  await browser.close();
}
const ordered = [
  "01-homepage",
  "02-tenant-setup",
  "07-tenant-monitoring",
  "10-tenant-library-students",
  "03-teacher-authoring",
  "06-teacher-review",
  "04-teacher-mobile",
  "11-teacher-overview-profile",
  "05-student-learning",
  "12-student-explore",
  "08-student-feedback-security",
  "09-student-mobile",
].map((name) => {
  const clip = S.clips.find((c) => c.name === name);
  if (!clip) throw new Error("Missing role section " + name);
  return clip;
});
const probe = (f) =>
  Number(
    execFileSync(
      "ffprobe",
      [
        "-v",
        "error",
        "-show_entries",
        "format=duration",
        "-of",
        "default=noprint_wrappers=1:nokey=1",
        f,
      ],
      { encoding: "utf8" },
    ),
  );
// Keep the tutorial readable without long holds on an unchanged end screen.
const clipDuration = (c) =>
  Math.min(
    probe(c.file),
    c.contentDuration ?? (c.cues.at(-1)?.start ?? 0) + 12,
  );
const factor = 450 / ordered.reduce((n, c) => n + clipDuration(c), 0);
const encoder = execFileSync("ffmpeg", ["-hide_banner", "-encoders"], {
  encoding: "utf8",
}).includes("libx264")
  ? "libx264"
  : "libopenh264";
const assTime = (t) =>
  `${Math.floor(t / 3600)}:${String(Math.floor((t % 3600) / 60)).padStart(2, "0")}:${(t % 60).toFixed(2).padStart(5, "0")}`;
const stampTime = (t) =>
  `${Math.floor(t / 60)}:${String(Math.floor(t % 60)).padStart(2, "0")}`;
const safe = (s) =>
  s.replaceAll("\\", "/").replaceAll("{", "(").replaceAll("}", ")");
const header =
  "[Script Info]\nScriptType: v4.00+\nPlayResX: 1920\nPlayResY: 1080\nWrapStyle: 0\n\n[V4+ Styles]\nFormat: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\nStyle: Caption,DejaVu Sans,23,&H00EEEEEE,&H00FFFFFF,&H001C1C1C,&H00000000,0,0,0,0,100,100,0,0,1,0,0,2,110,110,22,1\n\n[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n";
let offset = 0;
const parts = [],
  chapters = [],
  roleStarts = [];
for (const [i, c] of ordered.entries()) {
  const rawDuration = clipDuration(c),
    duration = rawDuration * factor;
  if (
    [
      "01-homepage",
      "02-tenant-setup",
      "03-teacher-authoring",
      "05-student-learning",
    ].includes(c.name)
  )
    roleStarts.push({
      start: offset,
      title:
        c.name === "01-homepage"
          ? "Original Homepage"
          : c.name === "02-tenant-setup"
            ? "TENANT - Academy Owner"
            : c.name === "03-teacher-authoring"
              ? "TEACHER - Courses, Students and Review"
              : "STUDENT - Enrollment, Learning and Feedback",
    });
  const events = c.cues
    .map((q, j) => {
      chapters.push({ time: offset + q.start * factor, title: q.chapter });
      return `Dialogue: 0,${assTime(q.start * factor)},${assTime((c.cues[j + 1]?.start ?? rawDuration) * factor)},Caption,,0,0,0,,{\\fs15\\c&H00AAAAAA&}${safe(q.chapter)}\\N{\\fs23\\c&H00EEEEEE&}${safe(q.text)}`;
    })
    .join("\n");
  const ass = path.join(work, c.name + ".ass");
  await writeFile(ass, header + events + "\n");
  const rendered = path.join(work, c.name + ".mp4");
  parts.push(rendered);
  console.log(`Exporting ${i + 1}/${ordered.length}: ${c.name}`);
  const filter = `trim=duration=${rawDuration},setpts=${factor}*(PTS-STARTPTS),scale=1728:972:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:0:color=0x141414,drawbox=x=0:y=972:w=1920:h=108:color=0x1c1c1c:t=fill,subtitles=filename=${ass},setsar=1,fps=30`;
  execFileSync(
    "ffmpeg",
    [
      "-hide_banner",
      "-loglevel",
      "error",
      "-y",
      "-i",
      c.file,
      "-vf",
      filter,
      "-an",
      "-c:v",
      encoder,
      "-threads",
      "2",
      "-b:v",
      "3000k",
      "-pix_fmt",
      "yuv420p",
      "-movflags",
      "+faststart",
      rendered,
    ],
    { stdio: ["ignore", "ignore", "pipe"] },
  );
  offset += duration;
}
const desktop =
    process.env.LEARNYST_DEMO_DIR || path.join(homedir(), "Desktop"),
  out = path.join(desktop, "Learnyst-Landing-Design-Full-Demo.mp4"),
  concat = path.join(work, "concat.txt");
await mkdir(desktop, { recursive: true });
await mkdir(path.join(root, "artifacts"), { recursive: true });
await writeFile(concat, parts.map((p) => `file '${p}'`).join("\n") + "\n");
let metadata =
  ";FFMETADATA1\ntitle=Learnyst - Tenant, Teacher and Student Demo\n";
for (let i = 0; i < roleStarts.length; i++)
  metadata += `\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=${Math.round(roleStarts[i].start * 1000)}\nEND=${Math.round((roleStarts[i + 1]?.start ?? 450) * 1000)}\ntitle=${roleStarts[i].title}\n`;
const meta = path.join(work, "roles.ffmetadata");
await writeFile(meta, metadata);
execFileSync("ffmpeg", [
  "-hide_banner",
  "-loglevel",
  "error",
  "-y",
  "-f",
  "concat",
  "-safe",
  "0",
  "-i",
  concat,
  "-i",
  meta,
  "-map",
  "0:v:0",
  "-map_metadata",
  "1",
  "-map_chapters",
  "1",
  "-c",
  "copy",
  "-movflags",
  "+faststart",
  out,
]);
const chapterText =
  "LEARNYST — ROLE-BASED APP DEMO\n\n" +
  chapters.map((c) => `${stampTime(c.time)}  ${c.title}`).join("\n") +
  "\n";
await writeFile(
  path.join(desktop, "Learnyst-Landing-Design-Chapters.txt"),
  chapterText,
);
await copyFile(
  out,
  path.join(root, "artifacts", "learnyst-landing-design-demo.mp4"),
);
await cleanTestUsers(S.ids, S.assets);
S.finished = true;
S.output = out;
S.duration = probe(out);
await save();
console.log(
  JSON.stringify(
    { file: out, duration: S.duration, sections: roleStarts },
    null,
    2,
  ),
);
