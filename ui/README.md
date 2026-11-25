# WreckScope UI

Svelte 5 + Vite + Skeleton (`#lib` → `src/lib`).

```
src/
  App.svelte / main.ts / app.css
  lib/
    components/     # brand, dialogs, picker, field help, maturity
    features/
      recovery/     # dashboard, map, telemetry client + setup
      appearance/   # themes + ThemePicker
      settings/     # NumberSetting, SettingSwitch
```

```powershell
pnpm --dir ui install
pnpm --dir ui run check
pnpm --dir ui run build
```

Usual path: `wails dev` from the repo root.
