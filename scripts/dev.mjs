import { spawn } from "node:child_process";
import { mkdir, readFile, writeFile, rm } from "node:fs/promises";
import { existsSync } from "node:fs";
import { createServer } from "node:net";
import { fileURLToPath } from "node:url";
import path from "node:path";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
process.chdir(root);
if (existsSync(".env")) process.loadEnvFile(".env");
const ports = {
  web: process.env.LEARNYST_WEB_PORT || "3005",
  auth: process.env.LEARNYST_AUTH_PORT || "3010",
  core: process.env.LEARNYST_CORE_PORT || "3011",
  db: process.env.LEARNYST_DB_PORT || "55432",
};
const env = {
  ...process.env,
  ENV: "development",
  SEED_DEMO: process.env.SEED_DEMO || "true",
  FRONTEND_URL: `http://localhost:${ports.web}`,
  DB_URI: `postgresql://learnyst:learnyst-local-only@127.0.0.1:${ports.db}/learnyst?sslmode=disable`,
  STORAGE_DIR: path.join(root, ".local/uploads"),
  SMTP_ADDR: "127.0.0.1:1025",
  AUTH_SERVICE_URL: `http://127.0.0.1:${ports.auth}`,
  CORE_SERVICE_URL: `http://127.0.0.1:${ports.core}`,
  COOKIE_SECURE: "false",
  GOMAXPROCS: "4",
};
const children = [];
let stopping = false;
function run(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: root,
      env,
      stdio: "inherit",
      ...options,
    });
    child.once("error", reject);
    child.once("exit", (code) =>
      code === 0
        ? resolve()
        : reject(new Error(`${command} failed (${code}).`)),
    );
  });
}
async function available(port) {
  await new Promise((resolve, reject) => {
    const server = createServer();
    server.once("error", () =>
      reject(
        new Error(
          `Port ${port} is already in use. Stop the other Learnyst process or set LEARNYST_*_PORT in .env.`,
        ),
      ),
    );
    server.listen(Number(port), "127.0.0.1", () => server.close(resolve));
  });
}
function start(command, args, options = {}) {
  const child = spawn(command, args, {
    cwd: root,
    env,
    stdio: "inherit",
    detached: true,
    ...options,
  });
  children.push(child);
  child.once("error", (error) => {
    console.error(error.message);
    void shutdown(1);
  });
  child.once("exit", (code) => {
    if (!stopping) {
      console.error(`A Learnyst service stopped (${code}).`);
      void shutdown(code || 1);
    }
  });
  return child;
}
async function health(url) {
  for (let i = 0; i < 120; i++) {
    try {
      const response = await fetch(url, { signal: AbortSignal.timeout(1000) });
      if (response.ok) return;
    } catch {}
    if (stopping) throw new Error("Startup stopped.");
    await new Promise((resolve) => setTimeout(resolve, 500));
  }
  throw new Error(`Service did not become healthy: ${url}`);
}
async function shutdown(code = 0) {
  if (stopping) return;
  stopping = true;
  for (const child of children) {
    try {
      process.kill(-child.pid, "SIGTERM");
    } catch {}
  }
  await rm(".local/run.json", { force: true });
  setTimeout(() => process.exit(code), 500);
}
process.once("SIGINT", () => void shutdown());
process.once("SIGTERM", () => void shutdown());
try {
  await mkdir(".local/uploads", { recursive: true });
  await Promise.all(Object.values(ports).slice(0, 3).map(available));
  await run("docker", ["compose", "up", "-d", "--wait"]);
  if (!existsSync("lms-frontend/node_modules/next/package.json"))
    await run("npm", ["ci"], { cwd: path.join(root, "lms-frontend") });
  if (
    env.SEED_DEMO === "true" &&
    (!existsSync(".local/uploads/sample.mp4") ||
      (await readFile(".local/uploads/sample.mp4")).length === 0)
  ) {
    console.log("Creating a small local preview video…");
    const { execFileSync } = await import("node:child_process");
    const encoders = execFileSync("ffmpeg", ["-hide_banner", "-encoders"], {
      encoding: "utf8",
    });
    const encoder = encoders.includes("libx264") ? "libx264" : "libopenh264";
    await run("ffmpeg", [
      "-hide_banner",
      "-loglevel",
      "error",
      "-y",
      "-f",
      "lavfi",
      "-i",
      "color=c=0x143c34:s=960x540:r=24:d=18",
      "-f",
      "lavfi",
      "-i",
      "anullsrc=r=44100:cl=stereo",
      "-vf",
      "drawtext=text='Welcome to Learnyst':fontcolor=0xdbf5a3:fontsize=48:x=(w-text_w)/2:y=210,drawtext=text='Your next chapter starts here.':fontcolor=white:fontsize=24:x=(w-text_w)/2:y=285",
      "-c:v",
      encoder,
      "-threads",
      "2",
      "-pix_fmt",
      "yuv420p",
      "-c:a",
      "aac",
      "-shortest",
      "-movflags",
      "+faststart",
      ".local/uploads/sample.mp4",
    ]);
  }
  console.log("Building Go services…");
  await Promise.all([
    run(
      "go",
      ["build", "-p", "2", "-o", path.join(root, ".local/core"), "./cmd/api"],
      { cwd: path.join(root, "lms-core-service") },
    ),
    run(
      "go",
      ["build", "-p", "2", "-o", path.join(root, ".local/auth"), "./cmd/api"],
      { cwd: path.join(root, "lms-api-gateway") },
    ),
  ]);
  start(path.join(root, ".local/core"), [], {
    env: { ...env, PORT: ports.core },
  });
  await health(`${env.CORE_SERVICE_URL}/health`);
  start(path.join(root, ".local/auth"), [], {
    env: { ...env, PORT: ports.auth },
  });
  await health(`${env.AUTH_SERVICE_URL}/health`);
  start(
    "node",
    [
      path.join(root, "lms-frontend/node_modules/next/dist/bin/next"),
      "dev",
      "--port",
      ports.web,
      "--hostname",
      "127.0.0.1",
    ],
    { cwd: path.join(root, "lms-frontend"), env },
  );
  await writeFile(
    ".local/run.json",
    JSON.stringify({ pid: process.pid, root, ports }),
  );
  await health(env.FRONTEND_URL);
  console.log(
    `\nLearnyst is ready: ${env.FRONTEND_URL}\nLocal email inbox: http://localhost:8025\nStop with Ctrl+C or npm run stop. Data is saved in Docker and .local/uploads.\n`,
  );
} catch (error) {
  console.error(
    `\nLearnyst startup failed: ${error.message}\nRequires Docker, Go 1.26.3+, Node 22+, and FFmpeg for demo videos.\n`,
  );
  await shutdown(1);
}
