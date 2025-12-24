import React, { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { haptic } from '@/lib/haptic';

const PWAInstallPrompt: React.FC = () => {
  const [deferredPrompt, setDeferredPrompt] = useState<any>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    // Проверяем, установлено ли уже PWA
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches ||
                        (window.navigator as any).standalone === true;

    if (isStandalone) {
      return; // Уже установлено, не показываем кнопку
    }

    const handleBeforeInstallPrompt = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e);
      setIsVisible(true);
    };

    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt);

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
    };
  }, []);

  const handleInstallClick = async () => {
    if (!deferredPrompt) return;

    haptic.medium();
    deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;

    if (outcome === 'accepted') {
      console.log('PWA установлено');
      haptic.success();
    } else {
      console.log('PWA установка отклонена');
      haptic.error();
    }

    setDeferredPrompt(null);
    setIsVisible(false);
  };

  if (!isVisible) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <Button onClick={handleInstallClick} variant="default" size="touch">
        Установить приложение
      </Button>
    </div>
  );
};

export default PWAInstallPrompt;