import { test, before, after } from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { cleanTestUsers } from "./fixtures.mjs";
const auth = process.env.TEST_AUTH_URL || "http://127.0.0.1:3010/v1/auth";
const core = process.env.TEST_CORE_URL || "http://127.0.0.1:3011/api/v1";
const web = process.env.TEST_WEB_URL || "http://localhost:3005";
const stamp = `${Date.now()}`;
const ids = [];
let teacher,
  otherTeacher,
  student,
  otherStudent,
  course,
  lesson,
  assignment,
  assetURL,
  resetToken,
  invitationToken;
class Client {
  cookie = "";
  async request(base, path, method = "GET", data, extra = {}) {
    const headers = {
      ...(this.cookie ? { Cookie: this.cookie } : {}),
      ...(data instanceof FormData
        ? {}
        : { "Content-Type": "application/json" }),
      ...extra,
    };
    const response = await fetch(base + path, {
      method,
      headers,
      body:
        data === undefined
          ? undefined
          : data instanceof FormData
            ? data
            : JSON.stringify(data),
    });
    const cookies = response.headers.getSetCookie();
    if (cookies.length) this.cookie = cookies[0].split(";")[0];
    let body;
    try {
      body = await response.json();
    } catch {
      body = null;
    }
    return { status: response.status, body, headers: response.headers };
  }
}
async function register(role, label) {
  const client = new Client();
  client.email = `test-${label}-${stamp}@learnyst.local`;
  client.password = "A-local-test-password!";
  const result = await client.request(auth, "/register", "POST", {
    first_name: "Test",
    last_name: label,
    email: client.email,
    password: client.password,
    role_id: role,
  });
  assert.equal(result.status, 201, JSON.stringify(result.body));
  ids.push(result.body.user.id);
  client.user = result.body.user;
  assert.match(result.headers.get("set-cookie"), /HttpOnly/i);
  assert.match(result.headers.get("set-cookie"), /SameSite=Lax/i);
  assert.equal(result.body.user.password_hash, undefined);
  assert.equal(result.body.access_token, undefined);
  return client;
}
async function inboxToken(email, marker) {
  const response = await fetch("http://127.0.0.1:8025/api/v1/messages");
  const data = await response.json();
  const message = data.messages.find(
    (m) =>
      m.To.some((to) => to.Address === email) && m.Subject.includes(marker),
  );
  assert.ok(message, "Local inbox contains the expected email");
  const full = await (
    await fetch(`http://127.0.0.1:8025/api/v1/message/${message.ID}`)
  ).json();
  const token = full.Text.match(/token=([a-f0-9]{64})/)?.[1];
  assert.ok(token, "Email contains a working frontend link");
  return token;
}
before(async () => {
  for (const url of [
    auth.replace("/v1/auth", "/health"),
    core.replace("/api/v1", "/health"),
  ])
    assert.equal((await fetch(url)).status, 200);
  teacher = await register(2, "teacher");
  otherTeacher = await register(2, "other-teacher");
  student = await register(3, "student");
  otherStudent = await register(3, "other-student");
});
after(async () => {
  await cleanTestUsers(ids, assetURL ? [assetURL] : []);
});
test("auth rejects forged roles and credentials; same-origin proxy restores the session", async () => {
  const anonymous = new Client();
  assert.equal((await anonymous.request(core, "/courses")).status, 401);
  assert.equal(
    (
      await anonymous.request(auth, "/register", "POST", {
        first_name: "Bad",
        last_name: "Role",
        email: `bad-${stamp}@learnyst.local`,
        password: "Password123!",
        role_id: 4,
      })
    ).status,
    400,
  );
  assert.equal(
    (
      await anonymous.request(auth, "/login", "POST", {
        email: student.email,
        password: "wrong",
      })
    ).status,
    401,
  );
  const session = await student.request(web, "/api/auth/session");
  assert.equal(session.status, 200);
  assert.equal(session.body.user.role_id, 3);
  assert.equal(
    (
      await student.request(core, "/courses", "POST", {
        title: "Unauthorized",
        level: "Beginner",
      })
    ).status,
    403,
  );
  assert.equal(
    (
      await student.request(
        auth,
        "/logout",
        "POST",
        {},
        { Origin: "https://untrusted.example" },
      )
    ).status,
    403,
  );
  assert.equal(
    (
      await teacher.request(
        core,
        "/courses",
        "POST",
        { title: "Bad origin", level: "Beginner" },
        { Origin: "https://untrusted.example" },
      )
    ).status,
    403,
  );
});
test("teachers get isolated workspaces and drafts stay protected", async () => {
  const profile = await teacher.request(core, "/me");
  assert.equal(profile.status, 200);
  assert.ok(profile.body.tenant_id);
  const result = await teacher.request(core, "/courses", "POST", {
    title: `Integration course ${stamp}`,
    description: "A course created by the real API test.",
    category: "Development",
    level: "Beginner",
    created_by: otherTeacher.user.id,
    tenant_id: 99999,
  });
  assert.equal(result.status, 201, JSON.stringify(result.body));
  course = result.body;
  assert.equal(course.tenant_id, profile.body.tenant_id);
  assert.equal(
    (
      await otherTeacher.request(core, `/courses/${course.id}`, "PUT", {
        title: "Stolen",
      })
    ).status,
    403,
  );
  assert.equal(
    (await student.request(core, `/courses/${course.id}`)).status,
    403,
  );
  const catalog = await student.request(core, "/courses?scope=catalog");
  assert.ok(!catalog.body.some((c) => c.id === course.id));
  assert.equal(
    (
      await teacher.request(core, `/courses/${course.id}`, "PUT", {
        is_published: true,
      })
    ).status,
    400,
  );
});
test("file uploads validate content, stay local, and require course access", async () => {
  const bad = new FormData();
  bad.append("course_id", String(course.id));
  bad.append("file", new Blob(["not an MP4"]), "disguised.mp4");
  assert.equal(
    (await teacher.request(core, "/upload", "POST", bad)).status,
    400,
  );
  const form = new FormData();
  form.append("course_id", String(course.id));
  form.append(
    "file",
    new Blob([await readFile(".local/uploads/sample.mp4")], {
      type: "video/mp4",
    }),
    "lesson.mp4",
  );
  const result = await teacher.request(core, "/upload", "POST", form);
  assert.equal(result.status, 201, JSON.stringify(result.body));
  assetURL = result.body.url;
  assert.match(assetURL, /^\/uploads\/[a-f0-9]{32}\.mp4$/);
  const response = await teacher.request(
    core,
    `/courses/${course.id}/videos`,
    "POST",
    { title: "A local video", video_url: assetURL, duration_seconds: 18 },
  );
  assert.equal(response.status, 201, JSON.stringify(response.body));
  lesson = response.body;
  assert.equal(
    (
      await teacher.request(core, `/courses/${course.id}`, "PUT", {
        is_published: true,
      })
    ).status,
    200,
  );
  assert.equal((await fetch(web + assetURL)).status, 401);
  assert.equal(
    (await fetch(web + assetURL, { headers: { Cookie: otherStudent.cookie } }))
      .status,
    403,
  );
});
test("enrollment is idempotent, video seeking works, and progress belongs to the session", async () => {
  for (let i = 0; i < 2; i++)
    assert.equal(
      (await student.request(core, `/courses/${course.id}/enroll`, "POST", {}))
        .status,
      200,
    );
  const detail = await student.request(core, `/courses/${course.id}`);
  assert.equal(detail.status, 200);
  assert.equal(detail.body.course.student_count, 1);
  const range = await fetch(web + assetURL, {
    headers: { Cookie: student.cookie, Range: "bytes=0-31" },
  });
  assert.equal(range.status, 206);
  assert.equal((await range.arrayBuffer()).byteLength, 32);
  assert.equal(
    (
      await student.request(core, "/progress/video", "POST", {
        video_id: lesson.id,
        watch_time_seconds: 9,
        completion_percent: 50,
        student_id: otherStudent.user.id,
      })
    ).status,
    200,
  );
  assert.equal(
    (
      await student.request(core, "/progress/video", "POST", {
        video_id: lesson.id,
        completion_percent: 20,
      })
    ).status,
    200,
  );
  const saved = await student.request(core, `/courses/${course.id}`);
  assert.equal(saved.body.progress[0].completion_percent, 50);
  assert.equal(saved.body.course.progress, 50);
  assert.equal(
    (
      await otherStudent.request(core, "/progress/video", "POST", {
        video_id: lesson.id,
        completion_percent: 100,
      })
    ).status,
    403,
  );
  assert.equal(
    (await teacher.request(core, `/courses/${course.id}`, "DELETE")).status,
    409,
  );
});
test("students submit once and only their teacher can grade", async () => {
  const result = await teacher.request(
    core,
    `/courses/${course.id}/assignments`,
    "POST",
    {
      title: "Explain the idea",
      description: "Write one thing you learned.",
      max_marks: 10,
    },
  );
  assert.equal(result.status, 201);
  assignment = result.body;
  assert.equal(
    (
      await student.request(
        core,
        `/assignments/${assignment.id}/submit`,
        "POST",
        {
          text: "I learned how to keep local development independent from cloud credentials.",
        },
      )
    ).status,
    201,
  );
  assert.equal(
    (
      await student.request(
        core,
        `/assignments/${assignment.id}/submit`,
        "POST",
        { text: "Duplicate" },
      )
    ).status,
    409,
  );
  const resultList = await teacher.request(core, "/assignments");
  const submission = resultList.body.find((a) => a.id === assignment.id)
    .submissions[0];
  assert.equal(
    (
      await otherTeacher.request(
        core,
        `/submissions/${submission.id}/grade`,
        "PUT",
        { marks: 9 },
      )
    ).status,
    403,
  );
  assert.equal(
    (
      await teacher.request(
        core,
        `/submissions/${submission.id}/grade`,
        "PUT",
        { marks: 11 },
      )
    ).status,
    400,
  );
  assert.equal(
    (
      await teacher.request(
        core,
        `/submissions/${submission.id}/grade`,
        "PUT",
        { marks: 9, feedback: "Clear and practical." },
      )
    ).status,
    200,
  );
  const studentView = await student.request(core, "/assignments");
  assert.equal(
    studentView.body.find((a) => a.id === assignment.id).submissions[0].marks,
    9,
  );
});
test("course invitations arrive locally and only the invited email can accept", async () => {
  assert.equal(
    (
      await teacher.request(core, `/courses/${course.id}/invitations`, "POST", {
        email: otherStudent.email,
      })
    ).status,
    201,
  );
  invitationToken = await inboxToken(otherStudent.email, "invitation");
  assert.equal(
    (
      await student.request(core, "/invitations/accept", "POST", {
        token: invitationToken,
      })
    ).status,
    400,
  );
  assert.equal(
    (
      await otherStudent.request(core, "/invitations/accept", "POST", {
        token: invitationToken,
      })
    ).status,
    200,
  );
  assert.equal(
    (
      await otherStudent.request(core, "/invitations/accept", "POST", {
        token: invitationToken,
      })
    ).status,
    400,
  );
});
test("password reset is single-use and invalidates existing sessions immediately", async () => {
  assert.equal(
    (
      await otherStudent.request(auth, "/reset-password", "POST", {
        email: otherStudent.email,
      })
    ).status,
    200,
  );
  resetToken = await inboxToken(otherStudent.email, "Reset");
  const anonymous = new Client();
  assert.equal(
    (
      await anonymous.request(auth, "/reset-password/complete", "POST", {
        token: resetToken,
        new_password: "A-new-local-password!",
      })
    ).status,
    200,
  );
  assert.equal((await otherStudent.request(core, "/me")).status, 401);
  assert.equal(
    (
      await anonymous.request(auth, "/reset-password/complete", "POST", {
        token: resetToken,
        new_password: "Another-password!",
      })
    ).status,
    400,
  );
  assert.equal(
    (
      await anonymous.request(auth, "/login", "POST", {
        email: otherStudent.email,
        password: otherStudent.password,
      })
    ).status,
    401,
  );
  assert.equal(
    (
      await anonymous.request(auth, "/login", "POST", {
        email: otherStudent.email,
        password: "A-new-local-password!",
      })
    ).status,
    200,
  );
});
test("logout and password change revoke database sessions, including copied cookies", async () => {
  const copied = new Client();
  copied.cookie = student.cookie;
  assert.equal(
    (await student.request(auth, "/logout", "POST", {})).status,
    200,
  );
  assert.equal((await copied.request(core, "/me")).status, 401);
  assert.equal(
    (
      await teacher.request(auth, "/change-password", "POST", {
        old_password: "wrong",
        new_password: "New-password123!",
      })
    ).status,
    400,
  );
  const copiedTeacher = new Client();
  copiedTeacher.cookie = teacher.cookie;
  assert.equal(
    (
      await teacher.request(auth, "/change-password", "POST", {
        old_password: teacher.password,
        new_password: "New-password123!",
      })
    ).status,
    200,
  );
  assert.equal((await copiedTeacher.request(core, "/courses")).status, 401);
});
