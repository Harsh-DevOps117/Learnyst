import { test, after } from "node:test";
import assert from "node:assert/strict";
import { cleanTestUsers } from "./fixtures.mjs";
const url = "http://localhost:3005";
const ids = [];
const stamp = Date.now();
class Account {
  cookie = "";
  async call(route, method = "GET", data) {
    const response = await fetch(url + route, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...(this.cookie ? { Cookie: this.cookie } : {}),
      },
      body: data === undefined ? undefined : JSON.stringify(data),
    });
    const cookie = response.headers.getSetCookie()[0];
    if (cookie) this.cookie = cookie.split(";")[0];
    return { status: response.status, body: await response.json() };
  }
  async register(role, name) {
    this.email = `test-tenant-${name}-${stamp}@learnyst.local`;
    const r = await this.call("/api/auth/register", "POST", {
      first_name: "Role",
      last_name: name,
      email: this.email,
      password: "Roles-password123!",
      role_id: role,
    });
    assert.equal(r.status, 201, JSON.stringify(r.body));
    ids.push(r.body.user.id);
    return this;
  }
}
after(async () => cleanTestUsers(ids));
test("owner settings, teacher invitations, academy isolation and access revocation work end to end", async () => {
  const owner = await new Account().register(1, "owner");
  const initialViews = await Promise.all(
    Array.from({ length: 12 }, (_, i) =>
      owner.call(i % 2 ? "/api/core/academy" : "/api/core/workspaces"),
    ),
  );
  for (const view of initialViews)
    assert.equal(view.status, 200, JSON.stringify(view.body));
  const ownerWorkspaces = await owner.call("/api/core/workspaces");
  assert.equal(
    ownerWorkspaces.body.length,
    1,
    "Concurrent initial views create a single academy",
  );
  const otherOwner = await new Account().register(1, "other-owner");
  const teacher = await new Account().register(2, "teacher");
  const wrongTeacher = await new Account().register(2, "wrong-teacher");
  const student = await new Account().register(3, "student");
  let a = await owner.call("/api/core/academy");
  assert.equal(a.status, 200);
  const academyID = a.body.academy.id;
  assert.equal((await teacher.call("/api/core/academy")).status, 403);
  assert.equal((await student.call("/api/core/academy")).status, 403);
  const settings = {
    name: "Role Test Academy",
    slug: `role-test-${stamp}`,
    description: "A real owner workspace with invited teaching staff.",
  };
  assert.equal(
    (await owner.call("/api/core/academy", "PUT", settings)).status,
    200,
  );
  assert.equal(
    (
      await owner.call("/api/core/academy", "PUT", {
        ...settings,
        slug: "dashboard",
      })
    ).status,
    400,
  );
  assert.equal(
    (await otherOwner.call("/api/core/academy", "PUT", settings)).status,
    409,
  );
  const publicPage = await owner.call(
    `/api/core/tenants/slug/${settings.slug}`,
  );
  assert.equal(publicPage.body.tenant.academy_name, settings.name);
  assert.equal(
    (
      await teacher.call("/api/core/academy/invitations", "POST", {
        email: wrongTeacher.email,
      })
    ).status,
    403,
  );
  assert.equal(
    (
      await owner.call("/api/core/academy/invitations", "POST", {
        email: student.email,
      })
    ).status,
    400,
  );
  assert.equal(
    (
      await owner.call("/api/core/academy/invitations", "POST", {
        email: teacher.email,
      })
    ).status,
    201,
  );
  const inbox = await (
    await fetch("http://localhost:8025/api/v1/messages")
  ).json();
  const message = inbox.messages.find(
    (m) =>
      m.To.some((t) => t.Address === teacher.email) &&
      m.Subject.includes("academy teacher"),
  );
  assert.ok(message);
  const mail = await (
    await fetch(`http://localhost:8025/api/v1/message/${message.ID}`)
  ).json();
  const token = mail.Text.match(/token=([a-f0-9]{64})/)[1];
  assert.equal(
    (await wrongTeacher.call("/api/core/invitations/accept", "POST", { token }))
      .status,
    400,
  );
  assert.equal(
    (await teacher.call("/api/core/invitations/accept", "POST", { token }))
      .status,
    200,
  );
  assert.equal(
    (await teacher.call("/api/core/invitations/accept", "POST", { token }))
      .status,
    400,
  );
  const workspaces = await teacher.call("/api/core/workspaces");
  assert.ok(workspaces.body.some((w) => w.id === academyID && w.current));
  const course = await teacher.call("/api/core/courses", "POST", {
    title: "Invited teacher course",
    level: "Beginner",
    category: "Development",
    tenant_id: 9999,
  });
  assert.equal(course.status, 201);
  assert.equal(course.body.tenant_id, academyID);
  const ownerCourses = await owner.call("/api/core/courses");
  assert.ok(ownerCourses.body.some((c) => c.id === course.body.id));
  const privateCourse = await owner.call("/api/core/courses", "POST", {
    title: "Owner-only draft",
    level: "Beginner",
  });
  assert.equal(privateCourse.status, 201);
  assert.equal(
    (await teacher.call(`/api/core/courses/${privateCourse.body.id}`)).status,
    403,
  );
  assert.equal(
    (
      await otherOwner.call(`/api/core/courses/${course.body.id}`, "PUT", {
        title: "Stolen",
      })
    ).status,
    403,
  );
  a = await owner.call("/api/core/academy");
  const member = a.body.teachers.find((t) => t.email === teacher.email);
  assert.ok(member);
  assert.equal(member.course_count, 1);
  assert.equal(a.body.stats.teachers, 1);
  assert.ok(a.body.invitations.some((i) => i.accepted));
  assert.equal(
    (
      await otherOwner.call(`/api/core/academy/teachers/${member.id}`, "PUT", {
        active: false,
      })
    ).status,
    404,
  );
  assert.equal(
    (
      await owner.call(`/api/core/academy/teachers/${member.id}`, "PUT", {
        active: false,
      })
    ).status,
    200,
  );
  assert.equal(
    (await teacher.call(`/api/core/courses/${course.body.id}`)).status,
    403,
  );
  assert.equal(
    (
      await teacher.call("/api/core/workspaces/current", "PUT", {
        tenant_id: academyID,
      })
    ).status,
    403,
  );
  assert.equal(
    (
      await owner.call(`/api/core/academy/teachers/${member.id}`, "PUT", {
        active: true,
      })
    ).status,
    200,
  );
  assert.equal(
    (
      await teacher.call("/api/core/workspaces/current", "PUT", {
        tenant_id: academyID,
      })
    ).status,
    200,
  );
  assert.equal(
    (await teacher.call(`/api/core/courses/${course.body.id}`)).status,
    200,
  );
  assert.equal(
    (await teacher.call("/api/core/academy", "PUT", settings)).status,
    403,
  );
});
