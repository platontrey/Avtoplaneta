/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from './lib/queryClient'
import { pushManager } from './lib/pushNotifications'
import { ThemeProvider } from './contexts/ThemeContext'
import './index.css'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ThemeProvider defaultTheme="system" storageKey="avtoplaneta-theme">
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </ThemeProvider>
    </QueryClientProvider>
  </StrictMode>,
)

const isIpHostname = /^(?:\d{1,3}\.){3}\d{1,3}$/.test(window.location.hostname)
  || window.location.hostname.includes(':')

const clearServiceWorkerState = async () => {
  if ('serviceWorker' in navigator) {
    const registrations = await navigator.serviceWorker.getRegistrations()
    await Promise.all(registrations.map((registration) => registration.unregister()))
  }

  if ('caches' in window) {
    const names = await caches.keys()
    await Promise.all(names.map((name) => caches.delete(name)))
  }
}

const initializeServiceWorker = async () => {
  if (!('serviceWorker' in navigator)) return

  // Public certificates generally do not cover private LAN IP addresses.
  // Avoid a noisy registration failure while preserving PWA support by domain.
  if (isIpHostname) {
    await clearServiceWorkerState()
    console.info('Service Worker disabled for direct IP access')
    return
  }

  try {
    const registration = await navigator.serviceWorker.register('/sw.js', {
      updateViaCache: 'none',
    })
    await registration.update()
    const isInitialized = await pushManager.init()
    console.info(
      isInitialized
        ? 'Push notifications initialized successfully'
        : 'Push notifications not available or already initialized',
    )
  } catch (error) {
    console.error('Failed to initialize Service Worker:', error)
  }
}

window.addEventListener('load', () => {
  void initializeServiceWorker()
})
