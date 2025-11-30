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