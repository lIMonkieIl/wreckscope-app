import path from 'node:path'
import {fileURLToPath} from 'node:url'
import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

const root = path.dirname(fileURLToPath(import.meta.url))

// Wails sets WAILS_VITE_PORT (default 9245). Keep config in sync so a bare
// `pnpm dev` still matches the AssetServer even if CLI flags are dropped.
const vitePort = Number(process.env.WAILS_VITE_PORT || process.env.VITE_PORT || 9245)

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [tailwindcss(), svelte()],
  publicDir: 'static',
  resolve: {
    alias: {
      '#lib': path.resolve(root, 'src/lib'),
    },
  },
  clearScreen: false,
  server: {
    // IPv4 loopback: Wails AssetServer dials tcp4 127.0.0.1 (Windows).
    host: '127.0.0.1',
    port: Number.isFinite(vitePort) && vitePort > 0 ? vitePort : 9245,
    strictPort: true,
    // Wails webview loads via wails.localhost during `wails3 task dev`
    allowedHosts: ['.localhost', 'wails.localhost'],
  },
})
