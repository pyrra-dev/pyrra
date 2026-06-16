import path from 'path'
import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  base: './',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/objectives.v1alpha1.ObjectiveService': {
        target: 'https://demo.pyrra.dev',
        changeOrigin: true,
        secure: true,
      },
      '/prometheus.v1.PrometheusService': {
        target: 'https://demo.pyrra.dev',
        changeOrigin: true,
        secure: true,
      },
    },
  },
  build: {
    outDir: 'build',
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/setupTests.ts',
  },
})
