import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { readFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const envPath = resolve(__dirname, '../.env')
const env: Record<string, string> = {}
for (const line of readFileSync(envPath, 'utf-8').split('\n')) {
  const t = line.trim()
  if (!t || t.startsWith('#')) continue
  const i = t.indexOf('=')
  if (i > 0) env[t.slice(0, i).trim()] = t.slice(i + 1).trim()
}

const backendPort = env.OTHELLO_BACKEND_PORT || '8088'
const frontendPort = Number(env.OTHELLO_FRONTEND_PORT) || 5173

export default defineConfig({
  plugins: [vue()],
  define: {
    'import.meta.env.VITE_BACKEND_PORT': JSON.stringify(backendPort),
    'import.meta.env.VITE_FRONTEND_PORT': JSON.stringify(String(frontendPort)),
  },
  server: {
    port: frontendPort,
    proxy: {
      '/ws': {
        target: `ws://localhost:${backendPort}`,
        ws: true,
      },
    },
  },
})
