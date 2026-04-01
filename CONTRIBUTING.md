# Contributing

## Overview
- This project provides a Go MCP server for Blender, plus a Blender bridge add-on and an npm wrapper for distribution.
- Contributions should keep binaries lightweight, outputs concise, and integrations optional.

## Prerequisites
- Go ≥ 1.22
- Node.js ≥ 18
- Blender ≥ 3.6 (tested on 5.0.1)

## Local Setup
- Build Go binaries:
  - Windows x64: `npm run build:win-x64` (from `npm/`)
  - Linux x64: `npm run build:linux-x64` (from `npm/`)
  - macOS arm64: `npm run build:darwin-arm64` (from `npm/`)
- Blender bridge:
  - Install `blender_bridge/blender_bridge.py` in Blender Add-ons
  - N-panel → MCP → Start Bridge
- Claude config:
  - Recommended: `npx @j4flmao/go_blender_mcp` in `claude_desktop_config.json`

## Branching Workflow
- Create feature branch from `main`
- Keep PRs focused and small
- Link issues in PR description

## Commit & PR Guidelines
- Clear commit messages: scope + short action + rationale if needed
- Include screenshots for Blender-visible changes when relevant
- Avoid large vendor drops; prefer small, reviewable diffs

## Coding Standards
- Go:
  - `gofmt -w .`
  - Prefer stdlib, avoid heavy dependencies
  - Single-sentence tool outputs to minimize MCP context
- Python (Blender):
  - Execute `bpy` ops on main thread via timer/queue
  - Defensive error handling; cap string outputs to keep context low
- Security:
  - Do not commit secrets or API keys
  - Use environment variables or UI fields for keys

## Testing & QA
- Manual tool calls:
  - `tools/list`, `get_scene_info`, `create_object`, `exec_python`
- MCP Inspector:
  - `npx @modelcontextprotocol/inspector ./npm/dist/blender-mcp-go-{platform}`
- Validate optional integrations with toggles enabled/disabled

## CI/CD
- GitHub Actions builds Windows/Linux/macOS and publishes npm when `NPM_TOKEN` is configured
- Version bump in `npm/package.json` and tag `vX.Y.Z` triggers release
- Artifacts copied into `npm/dist` before publish

## Issue Reporting
- Provide OS, Blender version, steps to reproduce, and relevant logs
- Include Claude config snippet if tool discovery fails

## License
- Contributions are licensed under MIT; see `LICENSE`
