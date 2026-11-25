import path from 'node:path'
import {fileURLToPath} from 'node:url'
import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

const root = path.dirname(fileURLToPath(import.meta.url))

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
    strictPort: true,
    // Wails webview loads via wails.localhost during `wails3 task dev`
    allowedHosts: ['.localhost', 'wails.localhost'],
  },
})
