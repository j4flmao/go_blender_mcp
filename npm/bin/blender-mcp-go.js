#!/usr/bin/env node
const path = require('path');
const fs = require('fs');
const cp = require('child_process');

function resolveBin() {
  const p = process.platform;
  const a = process.arch;
  if (p === 'win32' && a === 'x64') return path.join(__dirname, '..', 'dist', 'blender-mcp-go-win-x64.exe');
  if (p === 'linux' && a === 'x64') return path.join(__dirname, '..', 'dist', 'blender-mcp-go-linux-x64');
  if (p === 'darwin' && a === 'arm64') return path.join(__dirname, '..', 'dist', 'blender-mcp-go-darwin-arm64');
  return null;
}

const bin = resolveBin();
if (!bin || !fs.existsSync(bin)) {
  console.error('No prebuilt binary available for this platform/arch. Build with "npm run build:all" or provide a compatible binary under dist/.');
  process.exit(1);
}

const child = cp.spawn(bin, [], { stdio: 'inherit', env: process.env });
child.on('exit', (code) => process.exit(code));
