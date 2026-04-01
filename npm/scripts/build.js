const cp = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

function arg(name) {
  const idx = process.argv.indexOf(name);
  if (idx === -1) return null;
  return process.argv[idx + 1] || null;
}

function must(name) {
  const v = arg(name);
  if (!v) throw new Error(`Missing ${name}`);
  return v;
}

function run(cmd, args, opts) {
  const res = cp.spawnSync(cmd, args, { stdio: "inherit", ...opts });
  if (res.error) throw res.error;
  if (res.status !== 0) process.exit(res.status ?? 1);
}

function main() {
  const goos = must("--os");
  const goarch = must("--arch");
  const out = must("--out");
  const moduleDir = arg("--module") || "../blender-mcp-go";

  const outPath = path.resolve(process.cwd(), out);
  fs.mkdirSync(path.dirname(outPath), { recursive: true });

  const cwd = path.resolve(process.cwd(), moduleDir);
  if (!fs.existsSync(path.join(cwd, "go.mod"))) {
    throw new Error(`go.mod not found in ${cwd}`);
  }

  const env = {
    ...process.env,
    CGO_ENABLED: "0",
    GOOS: goos,
    GOARCH: goarch
  };

  run("go", ["build", "-trimpath", "-o", outPath, "."], { env, cwd });
}

main();
