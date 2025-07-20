import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  server: {
    host: '0.0.0.0',
    port: 8081,
    hmr: {
      clientPort: 443,
      protocol: 'wss',
    },
    watch: {
      usePolling: true,
    }
  },
  resolve: {
    alias: {
      '@': '/app/src',
    },
  },
})