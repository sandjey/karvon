import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: process.env.VITE_BASE_PATH || '/admin/',
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'https://fan.sarbon.me',
        changeOrigin: true,
        secure: true,
      }
    }
  }
})
