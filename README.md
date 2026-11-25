# WreckScope

Live recovery dashboard for Wreckfest 2. Watches the race feed, flags stuck cars, and shows the grid on a map — on PC and on a phone over LAN.

Built with [Wails v3](https://wails.io) (Go + Svelte). Windows NSIS installer; settings and caches under `%AppData%\WreckScope\`.

## Requirements

- Go 1.27+
- Node 24 and pnpm 11
- [Wails CLI v3](https://v3.wails.io/getting-started/installation/) (`wails3`)
- [Task](https://taskfile.dev) (used by `wails3 task`)
- WebView2 (Windows)
- [NSIS](https://nsis.sourceforge.io) (`makensis`) for the installer

## Develop

Preferred (hot reload — Vite + app, waits until Vite accepts connections):

```powershell
pnpm --dir ui install
wails3 task dev
```

Alternate (no watcher): start Vite in one terminal, then build/run in another:

```powershell
pnpm --dir ui exec vite --host 127.0.0.1 --port 9245 --strictPort
# other terminal:
$env:FRONTEND_DEVSERVER_URL = "http://127.0.0.1:9245"
wails3 task build DEV=true
wails3 task run
```

(`pnpm run dev -- --port 9245` can hand Vite a literal `--`, so it stays on :5173 — use `pnpm exec vite` or rely on `vite.config.ts` / `WAILS_VITE_PORT`.)


`ui/bindings/` are the generated Wails v3 TypeScript bindings — keep and regenerate with `wails3 generate bindings` after Go API changes (old `wailsjs` is gone).

Demo mode is available in non-production (debug) builds, or with `-tags dev`.

`uname` / `tail` messages from Task on Windows are harmless noise.

## Build

NSIS installer:

```powershell
.\scripts\package-windows.ps1
```

Outputs under `dist\windows\`:

- `WreckScope-<ver>-windows-amd64-setup.exe` — NSIS installer
- `SHA256SUMS` — checksums for the GitHub updater

Publish both on a GitHub Release. In-app updates download the setup package to `%AppData%\WreckScope\updates\`, open it, then quit. Staged installers are removed after launch when possible, and any leftovers are cleared on the next app start.

Binary-only (no installer):

```powershell
wails3 task build VERSION=0.1.0
```

Artifacts land in `bin\`. Version is baked via ldflags (`internal/version`) and stamped into Windows `info.json` / NSIS `DisplayVersion` by `scripts/stamp-windows-version.ps1` (single source: `build/config.yml` `info.version` unless `-Version` is passed).

## Data locations

Install dir (Program Files) holds binaries only.

Everything else under `%AppData%\WreckScope\`:

- `settings.json` — app settings (survives upgrade; uninstall removes this folder)
- `webview\` — WebView2 profile
- `updates\` — temporary staging for in-app update downloads (cleared after launch / on next start)

First-run with no `settings.json` seeds theme `wreckscope` and mode `system`. Upgrading over an existing install keeps the previous settings.

Game telemetry `config.json` stays under `My Games\Wreckfest 2\...\savegame\telemetry\`. WreckScope can enable UDP on `127.0.0.1:23123` and keeps a sibling backup before writing.

LAN dashboard listens on TCP `8080` by default. Use Device view for the phone QR (same Wi‑Fi).

## Layout

```
main.go          Wails v3 entry (embeds ui/dist)
internal/        app, settings, setup, update, version, racefeed, recovery, wf2
ui/              Svelte UI + generated bindings/
build/           Taskfiles, config.yml, icons, NSIS
scripts/         Windows package script
bin/             build output
```
