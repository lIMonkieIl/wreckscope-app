# WreckScope

Live recovery dashboard for Wreckfest 2. Watches the race feed, flags stuck cars, and shows the grid on a map — on PC and on a phone over LAN.

Built with [Wails](https://wails.io) (Go + Svelte). Portable Windows exe; no installer required.

## Requirements

- Go 1.23+
- Node 24 and pnpm 11
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation)
- WebView2 (Windows)

## Develop

```powershell
pnpm --dir ui install
wails dev -tags dev
```

Demo mode is available in this build only.

## Build

```powershell
wails build
```

Binary: `build\bin\WreckScope.exe`

## Config

Point Settings → Config at the game’s telemetry `config.json` (under `My Games\Wreckfest 2\...\savegame\telemetry\`). WreckScope can enable UDP on `127.0.0.1:23123` and keeps a sibling backup before writing.

LAN dashboard listens on TCP `8080` by default. Use Device view for the phone QR (same Wi‑Fi).

## Layout

```
main.go          Wails entry (embeds ui/dist)
internal/        app, settings, setup, racefeed, recovery, wf2
ui/              Svelte UI
build/           icons and packaging
```
