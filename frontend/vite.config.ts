import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'masked-icon.svg'],
      manifest: {
        name: 'Автопланета - Система управления запчастями',
        short_name: 'Автопланета',
        description: 'Система для управления инвентарем автозапчастей',
        theme_color: '#ffffff',
        background_color: '#ffffff',
        display: 'standalone',
        icons: [
          {
            src: 'pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png'
          }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/api\./i,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              expiration: {
                maxEntries: 100,
                maxAgeSeconds: 60 * 60 * 24 // 24 hours
              }
            }
          }
        ]
      }
    })
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
      host: '0.0.0.0',
      allowedHosts: ['spectrologically-seeable-zenobia.ngrok-free.dev'],
      proxy: {
      '/api/inventory': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/api/addpart': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/api/deletepart': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/api/updatepart': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/api/uploadpartphoto': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/api/statistics': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/auth': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('proxyReq', (proxyReq, req) => {
            if (req.headers.cookie) {
              proxyReq.setHeader('Cookie', req.headers.cookie);
            }
          });
        },
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          ui: ['lucide-react', '@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu', '@radix-ui/react-accordion', '@radix-ui/react-select', '@radix-ui/react-tabs'],
          query: ['@tanstack/react-query'],
          forms: ['react-hook-form', 'zod'],
          utils: ['clsx', 'tailwind-merge', 'styled-components'],
          // auth: ['js-cookie'], // Commented out as js-cookie is not used
        },
      },
    },
    chunkSizeWarningLimit: 1000,
  },
})