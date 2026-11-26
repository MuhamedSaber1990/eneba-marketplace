import { readFileSync } from 'node:fs'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// In development the Go API runs on port 8080 (go run ./cmd/server).
// Forwarding /api to it means the frontend uses the same relative URLs
// locally and on Vercel, where both live on one domain.
const apiProxy = { '/api': 'http://localhost:8080' }

// The security headers are defined once, in vercel.json, because Vercel is
// what sends them in production. They are read here so that `npm run preview`
// (which serves the built site locally) sends the same ones. That makes it
// possible to see what the Content-Security-Policy blocks before deploying.
type VercelConfig = { headers: { headers: { key: string; value: string }[] }[] }
const vercelConfig: VercelConfig = JSON.parse(readFileSync(new URL('../vercel.json', import.meta.url), 'utf8'))
const securityHeaders = Object.fromEntries(vercelConfig.headers[0].headers.map((header) => [header.key, header.value]))

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: apiProxy,
  },
  preview: {
    proxy: apiProxy,
    headers: securityHeaders,
  },
})
