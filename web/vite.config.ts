import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      // Proxy API requests to the backend server
      '/api': {
        target: 'http://localhost:8080', // Default Go server port
        changeOrigin: true,
        secure: false,
        // Don't rewrite paths - the server expects /api prefix
      },
      // Proxy WebSocket connections for terminal
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
        secure: false,
      },
    },
  },
})
