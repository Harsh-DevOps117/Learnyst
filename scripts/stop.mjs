import { readFile, readlink } from "node:fs/promises";
import { spawnSync } from "node:child_process";
try {
  const state = JSON.parse(await readFile(".local/run.json", "utf8"));
  const cmd = await readFile(`/proc/${state.pid}/cmdline`, "utf8");
  const cwd = await readlink(`/proc/${state.pid}/cwd`);
  if (cwd === state.root && cmd.includes("scripts/dev.mjs"))
    process.kill(state.pid, "SIGTERM");
} catch {}
spawnSync("docker", ["compose", "stop"], { stdio: "inherit" });
