/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from './lib/queryClient'
import { pushManager } from './lib/pushNotifications'
import './index.css'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
)

// Unregister service worker and clear cache to force updates
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.getRegistrations().then((registrations) => {
    for (const registration of registrations) {
      registration.unregister().then((success) => {
        if (success) console.log('Service Worker unregistered successfully');
      });
    }
  });
}

if ('caches' in window) {
  caches.keys().then((names) => {
    for (const name of names) {
      caches.delete(name).then(() => {
        console.log('Cache cleared:', name);
      });
    }
  });
}

// Initialize push notifications
pushManager.init().then((isInitialized) => {
  if (isInitialized) {
    console.log('Push notifications initialized successfully');
  } else {
    console.log('Push notifications not available or already initialized');
  }
}).catch((error) => {
  console.error('Failed to initialize push notifications:', error);
});