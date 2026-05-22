// Custom service worker for push notifications
importScripts('https://storage.googleapis.com/workbox-cdn/releases/6.5.4/workbox-sw.js');

if (workbox) {
  // Set up workbox
  workbox.setConfig({
    debug: false
  });

  // Cache strategies
  workbox.routing.registerRoute(
    /^https:\/\/api\./i,
    new workbox.strategies.NetworkFirst({
      cacheName: 'api-cache',
      plugins: [
        new workbox.expiration.ExpirationPlugin({
          maxEntries: 100,
          maxAgeSeconds: 24 * 60 * 60 // 24 hours
        })
      ]
    })
  );

  // Handle push notifications
  self.addEventListener('push', function(event) {
    console.log('Push message received:', event);

    let data = {};
    if (event.data) {
      data = event.data.json();
    }

    const options = {
      body: data.body || 'Новое уведомление',
      icon: data.icon || '/pwa-192x192.png',
      badge: data.badge || '/pwa-192x192.png',
      data: data.data || {},
      actions: data.actions || [],
      requireInteraction: true,
      silent: false
    };

    event.waitUntil(
      self.registration.showNotification(data.title || 'Автопланета', options)
    );
  });

  // Handle notification click
  self.addEventListener('notificationclick', function(event) {
    console.log('Notification click received:', event);

    event.notification.close();

    // Handle action clicks
    if (event.action) {
      // Handle specific actions
      switch (event.action) {
        case 'view':
          event.waitUntil(
            clients.openWindow(event.notification.data.url || '/')
          );
          break;
        case 'dismiss':
          // Just close the notification
          break;
        default:
          event.waitUntil(
            clients.openWindow('/')
          );
      }
    } else {
      // Default click behavior
      event.waitUntil(
        clients.openWindow(event.notification.data.url || '/')
      );
    }
  });

  // Handle background sync (if needed)
  self.addEventListener('sync', function(event) {
    console.log('Background sync triggered:', event.tag);

    if (event.tag === 'background-sync') {
      event.waitUntil(doBackgroundSync());
    }
  });

  // Background sync function
  async function doBackgroundSync() {
    try {
      // Implement background sync logic here
      console.log('Performing background sync...');
      // For example, retry failed API calls
    } catch (error) {
      console.error('Background sync failed:', error);
    }
  }

  // Handle messages from main thread
  self.addEventListener('message', function(event) {
    if (event.data && event.data.type === 'SKIP_WAITING') {
      self.skipWaiting();
    }
  });
} else {
  console.log('Workbox could not be loaded. No offline functionality available.');
}