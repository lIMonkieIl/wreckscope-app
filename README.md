# WreckScope

Live recovery dashboard for Wreckfest 2. Watches the race feed, flags stuck cars, and shows the grid on a map — on PC and on a phone over LAN.

Built with [Wails](https://wails.io) (Go + Svelte). Windows installer (default) or portable zip.

## Requirements

- Go 1.23+
- Node 24 and pnpm 11
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation)
- WebView2 (Windows)
- [NSIS](https://nsis.sourceforge.io) (`makensis`) for the installer

## Develop

```powershell
pnpm --dir ui install
wails dev -tags dev
```

Demo mode is available in this build only.

## Build

Installer + portable zip:

```powershell
.\scripts\package-windows.ps1
```

Outputs under `dist\windows\`:

- `WreckScope-<ver>-windows-amd64-setup.exe` — NSIS installer
- `WreckScope-<ver>-windows-amd64-portable.zip` — exe + `portable` marker

Binary-only (no installer):

```powershell
wails build
```

## Settings

- **Installed** (no `portable` file beside the exe): `%AppData%\WreckScope\settings.json`
- **Portable** (zip includes an empty `portable` file): `settings.json` next to the exe

Game telemetry `config.json` stays under `My Games\Wreckfest 2\...\savegame\telemetry\`. WreckScope can enable UDP on `127.0.0.1:23123` and keeps a sibling backup before writing.

LAN dashboard listens on TCP `8080` by default. Use Device view for the phone QR (same Wi‑Fi).

## Layout

```
main.go          Wails entry (embeds ui/dist)
internal/        app, settings, setup, racefeed, recovery, wf2
ui/              Svelte UI
build/           icons and packaging
scripts/         Windows package script
```
