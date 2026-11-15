import { test, expect } from "@playwright/test";
import { cleanTestUsers } from "../../tests/fixtures.mjs";
test("tenant configures academy and invites a teacher; teacher works in that academy with separate permissions", async ({
  browser,
}) => {
  const tag = Date.now(),
    password = "Academy-browser123!";
  const ids: number[] = [];
  const owner = await browser.newContext(),
    teacher = await browser.newContext(),
    learner = await browser.newContext();
  const p = await owner.newPage(),
    t = await teacher.newPage();
  const teacherEmail = `test-owner-ui-teacher-${tag}@learnyst.local`,
    ownerEmail = `test-owner-ui-${tag}@learnyst.local`;
  try {
    await p.goto("/signup?role=tenant");
    await expect(
      p.getByRole("button", { name: /I run an academy/ }),
    ).toHaveAttribute("aria-pressed", "true");
    await p.getByLabel("First name").fill("Academy");
    await p.getByLabel("Last name").fill("Owner");
    await p.getByLabel("Email address").fill(ownerEmail);
    await p.locator("input[name=password]").fill(password);
    await p.getByLabel("Confirm password").fill(password);
    await p.getByRole("button", { name: "Create my account" }).click();
    await p.waitForURL("**/dashboard");
    ids.push((await (await p.request.get("/api/auth/session")).json()).user.id);
    await expect(p.locator(".header-role")).toContainText("Tenant");
    await p
      .locator("aside")
      .getByRole("link", { name: "Academy settings", exact: true })
      .click();
    await p.getByLabel("Academy name").fill("Owners Studio");
    await p.getByLabel("Public academy address").fill(`owners-studio-${tag}`);
    await p
      .getByLabel("Academy description")
      .fill("A workspace shared safely with invited teachers.");
    await p.getByRole("button", { name: "Save academy" }).click();
    await p
      .locator("aside")
      .getByRole("link", { name: "Teaching team", exact: true })
      .click();
    await p
      .getByRole("button", { name: "Invite teacher", exact: true })
      .click();
    await p.getByRole("dialog").getByLabel("Teacher email").fill(teacherEmail);
    await p.getByRole("button", { name: "Send teacher invitation" }).click();
    await expect(p.getByRole("dialog")).not.toBeVisible();
    await expect(p.locator("table")).toContainText(teacherEmail);
    const messages = await (
      await fetch("http://localhost:8025/api/v1/messages")
    ).json();
    const m = messages.messages.find((m: { To: { Address: string }[] }) =>
      m.To.some((x) => x.Address === teacherEmail),
    );
    const email = await (
      await fetch(`http://localhost:8025/api/v1/message/${m.ID}`)
    ).json();
    const token = email.Text.match(/token=([a-f0-9]{64})/)[1];
    await t.goto("/signup?role=teacher");
    await t.getByLabel("First name").fill("Invited");
    await t.getByLabel("Last name").fill("Teacher");
    await t.getByLabel("Email address").fill(teacherEmail);
    await t.locator("input[name=password]").fill(password);
    await t.getByLabel("Confirm password").fill(password);
    await t.getByRole("button", { name: "Create my account" }).click();
    await t.waitForURL("**/dashboard");
    ids.push((await (await t.request.get("/api/auth/session")).json()).user.id);
    await t.goto(`/invite/accept?token=${token}`);
    await t.getByRole("button", { name: "Accept invitation" }).click();
    await t.waitForURL("**/dashboard");
    await expect(t.locator(".header-role")).toContainText("Teacher");
    await expect(
      t.getByRole("combobox", { name: "Active academy" }),
    ).toContainText("Owners Studio");
    await expect(
      t
        .locator("aside")
        .getByRole("link", { name: "Academy settings", exact: true }),
    ).toHaveCount(0);
    expect((await t.request.get("/api/core/academy")).status()).toBe(403);
    await t
      .locator("aside")
      .getByRole("link", { name: "My courses", exact: true })
      .click();
    await t
      .getByRole("button", { name: "Create course", exact: true })
      .first()
      .click();
    await t
      .getByRole("dialog")
      .getByLabel("Course title")
      .fill("Teacher course in shared academy");
    await t
      .getByRole("dialog")
      .getByRole("button", { name: "Create course", exact: true })
      .click();
    await t.waitForURL(/\/dashboard\/courses\/\d+/);
    await p.goto("/dashboard?tab=courses");
    await expect(p.locator(".course-card")).toContainText(
      "Teacher course in shared academy",
    );
    await p.goto("/dashboard?tab=team");
    await expect(p.locator("table").first()).toContainText("Invited Teacher");
    await expect(p.locator("table").last()).toContainText("Accepted");
    const response = await learner.request.post("/api/auth/register", {
      data: {
        first_name: "Role",
        last_name: "Student",
        email: `test-owner-ui-student-${tag}@learnyst.local`,
        password,
        role_id: 3,
      },
    });
    expect(response.status()).toBe(201);
    ids.push((await response.json()).user.id);
    expect((await learner.request.get("/api/core/academy")).status()).toBe(403);
  } finally {
    await owner.close();
    await teacher.close();
    await learner.close();
    await cleanTestUsers(ids);
  }
});
