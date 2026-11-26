import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // In development the Go API runs on port 8080 (go run ./cmd/server).
    // Forwarding /api to it means the frontend uses the same relative URLs
    // locally and on Vercel, where both live on one domain.
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
