# WreckScope

Live recovery dashboard for Wreckfest 2. Watches the race feed, flags stuck cars, and shows the grid on a map — on PC and on a phone over LAN.

Built with [Wails v3](https://wails.io) (Go + Svelte). Windows installer (default) or portable zip.

## Requirements

- Go 1.27+
- Node 24 and pnpm 11
- [Wails CLI v3](https://v3.wails.io/getting-started/installation/) (`wails3`)
- [Task](https://taskfile.dev) (used by `wails3 task`)
- WebView2 (Windows)
- [NSIS](https://nsis.sourceforge.io) (`makensis`) for the installer

## Develop

```powershell
pnpm --dir ui install
wails3 task build DEV=true
wails3 task run
# or: wails3 task dev
```

Demo mode is available in non-production (debug) builds, or with `-tags dev`.

## Build

Installer + portable zip:

```powershell
.\scripts\package-windows.ps1
```

Outputs under `dist\windows\`:

- `WreckScope-<ver>-windows-amd64-setup.exe` — NSIS installer
- `WreckScope-<ver>-windows-amd64-portable.zip` — exe + `settings.json` (`portable: true`)

Binary-only (no installer):

```powershell
wails3 task build
```

Artifacts land in `bin\`.

## Settings

- **Installed** (no portable settings beside the exe): `%AppData%\WreckScope\settings.json`
- **Portable** (zip ships `settings.json` with `"portable": true`): settings stay next to the exe

Game telemetry `config.json` stays under `My Games\Wreckfest 2\...\savegame\telemetry\`. WreckScope can enable UDP on `127.0.0.1:23123` and keeps a sibling backup before writing.

LAN dashboard listens on TCP `8080` by default. Use Device view for the phone QR (same Wi‑Fi).

## Layout

```
main.go          Wails v3 entry (embeds ui/dist)
internal/        app, settings, setup, racefeed, recovery, wf2
ui/              Svelte UI + generated bindings/
build/           Taskfiles, config.yml, icons, NSIS
scripts/         Windows package script
bin/             build output
```
