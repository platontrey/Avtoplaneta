
import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Bell, BellOff } from 'lucide-react';
import { pushManager, isPushSupported, getNotificationPermission } from '@/lib/pushNotifications';

export default function PushNotifications() {
  const [isSupported, setIsSupported] = useState(false);
  const [permission, setPermission] = useState<NotificationPermission>('default');
  const [isSubscribed, setIsSubscribed] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    checkSupport();
    checkSubscription();
  }, []);

  const checkSupport = () => {
    const supported = isPushSupported();
    setIsSupported(supported);
    if (supported) {
      setPermission(getNotificationPermission());
    }
  };

  const checkSubscription = async () => {
    try {
      const subscription = await pushManager.getSubscription();
      setIsSubscribed(!!subscription);
    } catch (error) {
      console.error('Error checking subscription:', error);
    }
  };

  const requestPermission = async () => {
    setIsLoading(true);
    try {
      const newPermission = await pushManager.requestPermission();
      setPermission(newPermission);

      if (newPermission === 'granted') {
        // For demo purposes, we'll use a test VAPID key
        // In production, this should come from your backend
        const testVapidKey = 'BKxSosZ2-VLcYRG0FgJ8B6F7Z8gXH0X3K9V5X8X8X8X8X8X8X8X8X8X8X8X8X8X8X8X8X8X8X8X8';
        await pushManager.subscribe(testVapidKey);
        await checkSubscription();
      }
    } catch (error) {
      console.error('Error requesting permission:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const unsubscribe = async () => {
    setIsLoading(true);
    try {
      await pushManager.unsubscribe();
      await checkSubscription();
    } catch (error) {
      console.error('Error unsubscribing:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const testNotification = () => {
    pushManager.showNotification({
      title: 'Тестовое уведомление',
      body: 'Это тестовое push-уведомление от Автопланеты',
      icon: '/pwa-192x192.png',
      data: { url: '/' }
    });
  };

  if (!isSupported) {
    return (
      <div className="text-center py-8 text-gray-500">
        Push-уведомления не поддерживаются в этом браузере
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-medium">Статус уведомлений</h3>
          <p className="text-sm text-gray-600">
            Разрешение: {permission === 'granted' ? 'Разрешено' :
                         permission === 'denied' ? 'Запрещено' : 'Не запрошено'}
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {isSubscribed ? (
            <Bell className="w-5 h-5 text-green-600" />
          ) : (
            <BellOff className="w-5 h-5 text-gray-400" />
          )}
        </div>
      </div>

      <div className="flex gap-2">
        {permission !== 'granted' ? (
          <Button
            onClick={requestPermission}
            disabled={isLoading}
            className="flex-1"
          >
            {isLoading ? 'Запрашиваем...' : 'Разрешить уведомления'}
          </Button>
        ) : (
          <>
            {isSubscribed ? (
              <Button
                variant="outline"
                onClick={unsubscribe}
                disabled={isLoading}
              >
                Отписаться
              </Button>
            ) : (
              <Button
                onClick={requestPermission}
                disabled={isLoading}
              >
                Подписаться
              </Button>
            )}
            <Button
              variant="outline"
              onClick={testNotification}
            >
              Тест
            </Button>
          </>
        )}
      </div>

      <div className="text-xs text-gray-500">
        Push-уведомления позволяют получать важные обновления о запчастях и заказах.
      </div>
    </div>
  );
}