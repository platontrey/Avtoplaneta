import React, { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { X } from 'lucide-react';
import { haptic } from '@/lib/haptic';

const PWAInstallPrompt: React.FC = () => {
  const [deferredPrompt, setDeferredPrompt] = useState<any>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    // Проверяем, установлено ли уже PWA
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches ||
                        (window.navigator as any).standalone === true;

    if (isStandalone) {
      return; // Уже установлено, не показываем баннер
    }

    // Показываем баннер сразу, если не установлено
    setIsVisible(true);

    const handleBeforeInstallPrompt = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e);
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

  const handleDismiss = () => {
    setIsVisible(false);
  };

  if (!isVisible) return null;

  return (
    <div className="fixed top-0 left-0 right-0 z-50 bg-gray-600 text-black p-4 shadow-lg">
      <div className="max-w-7xl mx-auto flex items-center justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium">
            Установите наше приложение для лучшего опыта работы!
          </p>
        </div>
        <div className="flex items-center gap-2 ml-4">
          {deferredPrompt ? (
            <Button onClick={handleInstallClick} variant="secondary" size="sm">
              Установить
            </Button>
          ) : (
            <span className="text-sm">Доступно в поддерживаемых браузерах</span>
          )}
          <Button onClick={handleDismiss} variant="ghost" size="sm" className="text-white hover:bg-blue-700">
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  );
};

export default PWAInstallPrompt;