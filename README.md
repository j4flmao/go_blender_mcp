# Blender MCP Go (npm-wrapped)

![CI](https://github.com/j4flmao/go_blender_mcp/actions/workflows/ci.yml/badge.svg)
[![npm version](https://img.shields.io/npm/v/%40j4flmao/go_blender_mcp.svg)](https://www.npmjs.com/package/@j4flmao/go_blender_mcp)
![license](https://img.shields.io/badge/license-MIT-green.svg)
![Go](https://img.shields.io/badge/Go-1.22-blue?logo=go)
![Node](https://img.shields.io/badge/Node-%3E%3D18-green?logo=node.js)
![Blender](https://img.shields.io/badge/Blender-5.0.1-tested-orange?logo=blender)

Lightweight Model Context Protocol server for Blender written in Go, wrapped for npm distribution. One binary, minimal context, optional integrations (Poly Haven, Hyper3D, Sketchfab, Hunyuan) controlled via UI or env flags.

## Overview
- Go MCP server: `blender-mcp-go`
- Blender bridge add-on: simple TCP server that executes commands in Blender
- npm wrapper: cross-platform binaries in `dist/`, launcher at `npm/bin/blender-mcp-go.js`

Inspired by and referencing BlenderMCP (Python) project — see “References”.

## Prerequisites
- Blender ≥ 3.6 (tested on 5.0.1)
- Go ≥ 1.22 (for building binaries)
- Node.js ≥ 18 (for npm wrapper & publishing)

## Quick Setup
1. Build binaries (or use CI artifacts):
   - Windows: `npm run build:win-x64`
   - Linux: `npm run build:linux-x64`
   - macOS ARM: `npm run build:darwin-arm64`
   - All: `npm run build:all`

2. Install Blender bridge add-on:
   - File: [blender_bridge/blender_bridge.py](blender_bridge/blender_bridge.py) or get it from: https://github.com/j4flmao/go_blender_mcp.git
   - Blender → Edit → Preferences → Add-ons → Install → select the file → enable
   - N-panel → MCP → Start Bridge (port 9876)

3. Add MCP server to Claude Desktop (`claude_desktop_config.json`) — Recommended (npx):
```json
{
  "mcpServers": {
    "blender": {
      "command": "npx",
      "args": ["@j4flmao/go_blender_mcp"],
      "env": { }
    }
  }
}
```
Restart Claude Desktop after editing.

Alternative (absolute path to binary):
```json
{
  "mcpServers": {
    "blender": {
      "command": "D:\\blender_go_mcp\\npm\\dist\\blender-mcp-go-win-x64.exe",
      "args": [],
      "env": { }
    }
  }
}
```

4. Optional integrations (UI-driven):
   - In Blender N-panel, open “Integrations”
   - Tick Sketchfab / Hyper3D / Hunyuan / Poly Haven
   - Enter API keys where applicable
   - Tools will respect the current toggles at call time

5. npm usage (as a CLI):
   - Recommended: `npx @j4flmao/go_blender_mcp`
   - Optional: `npm start` (launches platform-specific binary from `npm/dist/`)

## CI/CD
- GitHub Actions workflow builds Windows/Linux/macOS and publishes npm if `NPM_TOKEN` is set
- See [.github/workflows/ci.yml](.github/workflows/ci.yml)

## Tools (high level)
- Core: scene info, list/get/move/create/delete objects, materials, render, set engine, exec Python
- Optional:
  - Poly Haven: HDRI/texture/model import
  - Sketchfab: model search (downloadable/animated/rigged filters)
  - Hyper3D: job submission (Rodin)
  - Hunyuan: placeholder for image→3D

## Result Preview
![Result](result.png)

## References
- BlenderMCP (Python, official repo): https://github.com/ahujasid/blender-mcp

## Notes
- Bridge now executes ops on Blender main thread via a small timer/queue to avoid crashes.
- Tools always list optionals; call-time checks ensure disabled integrations return a clear message.

## Troubleshooting
- Bridge not running: open Blender N-panel → MCP → Start Bridge; ensure port 9876 is free.
- Sketchfab returns non-character results: use filters (`animated=true`, `rigged=true`) or refine query.
- Render slow: lower samples or use EEVEE for previews.
- Binary missing after install: run `npm run build:all` or ensure CI artifacts are available.
- Claude doesn’t see tools: verify config path uses absolute path; restart Claude after changes.

## Contributing
- Fork and create feature branches.
- Keep Go code formatted (`gofmt -w .`) and avoid adding heavy dependencies.
- Prefer concise one-line tool outputs to minimize MCP context.
- Pull Requests should include a short description and, when relevant, screenshots of Blender results.

## License
MIT License. See [LICENSE](LICENSE) for details.

## Acknowledgements
- Thanks to BlenderMCP community and integrations such as Poly Haven, Hyper3D (Rodin), Sketchfab, and Hunyuan.
- Blender trademark is owned by Blender Foundation.
