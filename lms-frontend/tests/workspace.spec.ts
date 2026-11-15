import {
  test,
  expect,
  type Page,
  type APIRequestContext,
} from "@playwright/test";
import path from "node:path";
import { cleanTestUsers } from "../../tests/fixtures.mjs";
const suffix = Date.now();
const teacherEmail = `test-browser-teacher-${suffix}@learnyst.local`;
const studentEmail = `test-browser-student-${suffix}@learnyst.local`;
const password = "Browser-password123!";
const title = `Browser course ${suffix}`;
const assignmentTitle = `Practical challenge ${suffix}`;
const ids: number[] = [];
const assets: string[] = [];
let courseID = 0;
let invitation = "";
test.describe.configure({ mode: "serial" });
async function signIn(page: Page, email: string, pass = password) {
  await page.goto("/login");
  await page.locator("input[name=email]").fill(email);
  await page.locator("input[name=password]").fill(pass);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await page.waitForURL("**/dashboard");
  await expect(page.getByRole("heading", { name: /Hello,/ })).toBeVisible();
}
async function register(
  request: APIRequestContext,
  email: string,
  role: number,
) {
  const response = await request.post("/api/auth/register", {
    data: {
      first_name: "Browser",
      last_name: role === 2 ? "Teacher" : "Student",
      email,
      password,
      role_id: role,
    },
  });
  expect(response.status()).toBe(201);
  ids.push((await response.json()).user.id);
  await request.post("/api/auth/logout");
}
async function mailToken(email: string, subject: string) {
  const messages = await (
    await fetch("http://127.0.0.1:8025/api/v1/messages")
  ).json();
  const message = messages.messages.find(
    (m: { To: { Address: string }[]; Subject: string }) =>
      m.To.some((to) => to.Address === email) && m.Subject.includes(subject),
  );
  expect(message).toBeTruthy();
  const content = await (
    await fetch(`http://127.0.0.1:8025/api/v1/message/${message.ID}`)
  ).json();
  return content.Text.match(/token=([a-f0-9]{64})/)[1] as string;
}
test.beforeAll(async ({ request }) => {
  await register(request, teacherEmail, 2);
  await register(request, studentEmail, 3);
});
test.afterAll(async () => {
  await cleanTestUsers(ids, assets);
});
test("signup validates passwords, creates a teacher workspace, and signs out", async ({
  page,
}) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto("/signup");
  await page.getByRole("button", { name: /I’m here to teach/ }).click();
  await page.getByLabel("First name").fill("Signup");
  await page.getByLabel("Last name").fill("Teacher");
  await page
    .getByLabel("Email address")
    .fill(`test-browser-signup-${suffix}@learnyst.local`);
  await page.locator("input[name=password]").fill(password);
  await page.getByLabel("Confirm password").fill("Does-not-match123!");
  await page.getByRole("button", { name: "Create my account" }).click();
  await expect(page.locator(".notice.error")).toContainText("don’t match");
  await page.getByLabel("Confirm password").fill(password);
  await page.getByRole("button", { name: "Create my account" }).click();
  await page.waitForURL("**/dashboard");
  const response = await page.request.get("/api/auth/session");
  ids.push((await response.json()).user.id);
  await expect(page.locator(".header-role")).toContainText("Teacher");
  await page.reload();
  await expect(
    page.getByRole("heading", { name: "Hello, Signup" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Sign out", exact: true }).click();
  await page.waitForURL("**/login");
  expect(errors).toEqual([]);
});
test("teacher creates and edits a course, uploads a playable lesson, publishes and invites", async ({
  page,
}) => {
  await signIn(page, teacherEmail);
  await page.getByRole("link", { name: "My courses", exact: true }).click();
  await page
    .getByRole("button", { name: "Create course", exact: true })
    .first()
    .click();
  const dialog = page.getByRole("dialog");
  await dialog.getByLabel("Course title").fill(title);
  await dialog
    .getByLabel("A little about the course")
    .fill("A real course built through the browser.");
  await dialog
    .getByRole("button", { name: "Create course", exact: true })
    .click();
  await page.waitForURL(/\/dashboard\/courses\/\d+/);
  courseID = Number(page.url().split("/").pop());
  await expect(
    page.getByRole("heading", { name: title, exact: true, level: 1 }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Publish course", exact: true })
    .click();
  await expect(page.locator(".notice.error")).toContainText(
    "Add at least one lesson",
  );
  await page
    .getByLabel("What will your students learn?")
    .fill(
      "A real course with a locally stored lesson, practical work and thoughtful feedback.",
    );
  await page.getByRole("button", { name: "Save details" }).click();
  await page.getByRole("button", { name: "Add lesson", exact: true }).click();
  await dialog.getByLabel("Lesson title").fill("A working local lesson");
  await dialog
    .locator("input[name=video]")
    .setInputFiles(path.resolve("../.local/uploads/sample.mp4"));
  await dialog.getByRole("button", { name: "Add lesson", exact: true }).click();
  await expect(dialog).not.toBeVisible();
  await expect(page.locator(".editor-lesson")).toContainText(
    "A working local lesson",
  );
  const details = await (
    await page.request.get(`/api/core/courses/${courseID}`)
  ).json();
  assets.push(details.videos[0].video_url);
  await page
    .getByRole("button", { name: "Publish course", exact: true })
    .click();
  await expect(page.locator(".heading-badges")).toContainText("Published");
  await page
    .getByRole("button", { name: "Invite a student", exact: true })
    .click();
  await dialog.getByLabel("Student email").fill(studentEmail);
  await dialog.getByRole("button", { name: "Send invitation" }).click();
  await expect(dialog).toContainText("A new beginning, delivered.");
  invitation = await mailToken(studentEmail, "invitation");
  await dialog.getByRole("button", { name: "Close dialog" }).click();
  await page.getByRole("link", { name: "Assignments", exact: true }).click();
  await page
    .getByRole("button", { name: "Create assignment", exact: true })
    .click();
  await dialog
    .getByLabel("Course", { exact: true })
    .selectOption(String(courseID));
  await dialog.getByLabel("Assignment title").fill(assignmentTitle);
  await dialog
    .getByLabel("Instructions")
    .fill("Share one thing you learned from the lesson.");
  await dialog
    .getByRole("button", { name: "Create assignment", exact: true })
    .click();
  await expect(dialog).not.toBeVisible();
  await expect(page.locator(".assignment-summary")).toContainText(
    assignmentTitle,
  );
});
test("student accepts an invitation, watches a video, saves progress and submits work", async ({
  page,
}) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await signIn(page, studentEmail);
  await page.goto(`/invite/accept?token=${invitation}`);
  await page.getByRole("button", { name: "Accept invitation" }).click();
  await page.waitForURL(`**/dashboard/courses/${courseID}`);
  const video = page.locator("video");
  await expect(video).toBeVisible();
  await expect
    .poll(async () => video.evaluate((v) => (v as HTMLVideoElement).readyState))
    .toBeGreaterThan(0);
  await video.evaluate((v) => (v as HTMLVideoElement).play());
  await expect
    .poll(async () =>
      video.evaluate((v) => (v as HTMLVideoElement).currentTime),
    )
    .toBeGreaterThan(0);
  await video.evaluate((v) => (v as HTMLVideoElement).pause());
  await page.getByRole("button", { name: "Mark as complete" }).click();
  await expect(
    page.getByRole("button", { name: "Lesson complete" }),
  ).toBeVisible();
  await page.reload();
  await expect(
    page.getByRole("button", { name: "Lesson complete" }),
  ).toBeVisible();
  await expect(page.locator(".curriculum-progress-label")).toContainText(
    "100%",
  );
  await page.getByRole("link", { name: "Assignments", exact: true }).click();
  const assignment = page
    .locator(".assignment-item")
    .filter({ hasText: assignmentTitle });
  await assignment.locator("summary").click();
  await assignment
    .getByLabel("Your response")
    .fill(
      "I can learn with video files stored locally, and my progress survives a refresh.",
    );
  await assignment.getByRole("button", { name: "Submit assignment" }).click();
  await expect(assignment.locator("summary")).toContainText("Submitted");
  expect(errors).toEqual([]);
});
test("teacher reviews work and the student roster reflects enrollment", async ({
  page,
}) => {
  await signIn(page, teacherEmail);
  await page.getByRole("link", { name: "Students", exact: true }).click();
  await expect(page.locator(".students-table")).toContainText(studentEmail);
  await page.getByRole("link", { name: "Assignments", exact: true }).click();
  const assignment = page
    .locator(".assignment-item")
    .filter({ hasText: assignmentTitle });
  await assignment.locator("summary").click();
  await assignment.getByLabel("Marks out of 10").fill("8");
  await assignment
    .getByLabel("Feedback")
    .fill("Clear, practical and well explained.");
  await assignment.getByRole("button", { name: "Save feedback" }).click();
  await expect(assignment).toContainText("Reviewed");
});
test("mobile student navigation, search, grades and logout work without overflow", async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await signIn(page, studentEmail);
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth,
    ),
  ).toBe(false);
  await page.screenshot({
    path: "../.local/verified-student-mobile.png",
    fullPage: true,
  });
  await page.getByRole("button", { name: "Open navigation" }).click();
  await page.getByRole("link", { name: "My learning", exact: true }).click();
  await page.getByRole("textbox", { name: "Search courses" }).fill(title);
  await expect(page.locator(".course-card")).toHaveCount(1);
  await page.getByRole("button", { name: "Open navigation" }).click();
  await page.getByRole("link", { name: "Assignments", exact: true }).click();
  const assignment = page
    .locator(".assignment-item")
    .filter({ hasText: assignmentTitle });
  await expect(assignment.locator("summary")).toContainText("8/10 marks");
  await assignment.locator("summary").click();
  await expect(assignment).toContainText(
    "Clear, practical and well explained.",
  );
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth,
    ),
  ).toBe(false);
  await page.getByRole("button", { name: "Open navigation" }).click();
  await page.getByRole("button", { name: "Sign out", exact: true }).click();
  await page.waitForURL("**/login");
});
test("password recovery email opens a working browser reset form", async ({
  page,
}) => {
  await page.goto("/forgot-password");
  await page.getByLabel("Email address").fill(studentEmail);
  await page.getByRole("button", { name: "Send reset link" }).click();
  await expect(
    page.getByRole("heading", { name: "Check your inbox." }),
  ).toBeVisible();
  const token = await mailToken(studentEmail, "Reset");
  await page.goto(`/reset-password?token=${token}`);
  await page
    .locator("input[name=password]")
    .fill("Changed-browser-password123!");
  await page
    .getByLabel("Confirm password")
    .fill("Changed-browser-password123!");
  await page.getByRole("button", { name: "Update password" }).click();
  await expect(
    page.getByRole("heading", { name: "You’re all set." }),
  ).toBeVisible();
  await signIn(page, studentEmail, "Changed-browser-password123!");
});
