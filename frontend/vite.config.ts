import { readFileSync } from 'node:fs'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// In development the Go API runs on port 8080 (go run ./cmd/server).
// Forwarding /api to it means the frontend uses the same relative URLs
// locally and on Vercel, where both live on one domain. Set API_URL to
// point at an API on another port, for example when 8080 is taken.
//
// changeOrigin is false so the API receives the same Host header the browser
// sent (localhost:5173). The API compares that with the request's Origin to
// refuse requests forged by other sites, and on Vercel the two match because
// frontend and API share a domain. Vite's short form ('/api': 'http://...')
// would rewrite the Host to the API's own address and every sign-in would be
// refused as cross-site.
const apiProxy = {
  '/api': { target: process.env.API_URL ?? 'http://localhost:8080', changeOrigin: false },
}

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
