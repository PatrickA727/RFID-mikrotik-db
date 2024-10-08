import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {                               // Any request using the /api prefix will be proxied
        target: 'http://localhost:5000',      // Target server for backend API server, this is where the proxied request will be forwarded to 
        changeOrigin: true                    // Modifies the Host header of the proxied request to match the target server's host
      } 
    }
  }
})
