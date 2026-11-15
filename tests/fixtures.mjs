import { execFileSync } from "node:child_process";
import { unlink } from "node:fs/promises";
import { existsSync } from "node:fs";
import { resolve } from "node:path";
const root = resolve(existsSync("compose.yaml") ? "." : "..") + "/";
export async function cleanTestUsers(ids, assetURLs = []) {
  if (!ids.length || !ids.every((id) => Number.isSafeInteger(id) && id > 0))
    return;
  // The email predicate additionally protects real and demo users from fixture cleanup.
  const list = `SELECT id FROM users WHERE id IN (${ids.join(",")}) AND email LIKE 'test-%@learnyst.local'`;
  const tenants = `SELECT id FROM tenants WHERE owner_user_id IN (${list})`;
  const courses = `SELECT id FROM courses WHERE tenant_id IN (${tenants})`;
  const sql = `BEGIN;
    DELETE FROM video_progresses WHERE student_id IN (${list}) OR video_id IN (SELECT id FROM course_videos WHERE course_id IN (${courses}));
    DELETE FROM assignment_submissions WHERE student_id IN (${list}) OR assignment_id IN (SELECT id FROM assignments WHERE course_id IN (${courses}));
    DELETE FROM enrollments WHERE student_id IN (${list}) OR batch_id IN (SELECT id FROM batches WHERE course_id IN (${courses}));
    DELETE FROM invitations WHERE tenant_id IN (${tenants});
    DELETE FROM batches WHERE tenant_id IN (${tenants});
    DELETE FROM course_videos WHERE tenant_id IN (${tenants});
    DELETE FROM assignments WHERE tenant_id IN (${tenants});
    DELETE FROM local_assets WHERE user_id IN (${list});
    DELETE FROM courses WHERE tenant_id IN (${tenants});
    DELETE FROM tenant_staffs WHERE user_id IN (${list});
    DELETE FROM tenants WHERE owner_user_id IN (${list});
    DELETE FROM workspace_preferences WHERE user_id IN (${list});
    DELETE FROM auth_sessions WHERE user_id IN (${list});
    DELETE FROM password_resets WHERE user_id IN (${list});
    DELETE FROM users WHERE id IN (${list}); COMMIT;`;
  execFileSync(
    "docker",
    [
      "compose",
      "exec",
      "-T",
      "db",
      "psql",
      "-U",
      "learnyst",
      "-d",
      "learnyst",
      "-v",
      "ON_ERROR_STOP=1",
      "-c",
      sql,
    ],
    { cwd: root, stdio: "pipe" },
  );
  for (const url of assetURLs)
    if (/^\/uploads\/[a-f0-9]{32}\.[a-z0-9]+$/.test(url))
      await unlink(root + ".local" + url).catch(() => {});
}
