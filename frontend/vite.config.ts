import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

const BACKEND = 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Backend internal routes (/_/health, ...)
      '^/_/': { target: BACKEND, changeOrigin: true },
      // Backend API routes
      '/api': { target: BACKEND, changeOrigin: true },
    },
  },
})
