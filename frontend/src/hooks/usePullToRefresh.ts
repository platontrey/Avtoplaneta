import { useRef, useCallback, useEffect } from 'react';

interface PullToRefreshConfig {
  onRefresh: () => Promise<void> | void;
  threshold?: number;
  disabled?: boolean;
}

export const usePullToRefresh = (config: PullToRefreshConfig) => {
  const { onRefresh, threshold = 80, disabled = false } = config;

  const containerRef = useRef<HTMLElement>(null);
  const startYRef = useRef<number>(0);
  const isPullingRef = useRef<boolean>(false);
  const isRefreshingRef = useRef<boolean>(false);

  const handleTouchStart = useCallback((e: TouchEvent) => {
    if (disabled || isRefreshingRef.current) return;

    const scrollTop = containerRef.current?.scrollTop || 0;
    if (scrollTop > 0) return; // Only allow pull when at top

    startYRef.current = e.touches[0].clientY;
    isPullingRef.current = true;
  }, [disabled]);

  const handleTouchMove = useCallback((e: TouchEvent) => {
    if (disabled || !isPullingRef.current || isRefreshingRef.current) return;

    const currentY = e.touches[0].clientY;
    const pullDistance = currentY - startYRef.current;

    if (pullDistance > 0) {
      e.preventDefault();

      // Add visual feedback
      const progress = Math.min(pullDistance / threshold, 1);
      const opacity = progress * 0.8;

      // Create or update refresh indicator
      let indicator = document.getElementById('pull-refresh-indicator');
      if (!indicator) {
        indicator = document.createElement('div');
        indicator.id = 'pull-refresh-indicator';
        indicator.style.cssText = `
          position: fixed;
          top: 0;
          left: 50%;
          transform: translateX(-50%);
          background: rgba(0, 123, 255, ${opacity});
          color: white;
          padding: 8px 16px;
          border-radius: 20px;
          font-size: 14px;
          z-index: 1000;
          pointer-events: none;
          transition: opacity 0.2s;
          display: flex;
          align-items: center;
          gap: 8px;
        `;
        indicator.innerHTML = `
          <div style="
            width: 16px;
            height: 16px;
            border: 2px solid rgba(255,255,255,0.3);
            border-top: 2px solid white;
            border-radius: 50%;
            animation: spin 1s linear infinite;
          "></div>
          <span>Потяните для обновления</span>
        `;
        document.body.appendChild(indicator);
      } else {
        indicator.style.background = `rgba(0, 123, 255, ${opacity})`;
        const text = indicator.querySelector('span');
        if (text) {
          text.textContent = progress >= 1 ? 'Отпустите для обновления' : 'Потяните для обновления';
        }
      }
    }
  }, [disabled, threshold]);

  const handleTouchEnd = useCallback(async (e: TouchEvent) => {
    if (disabled || !isPullingRef.current || isRefreshingRef.current) return;

    const currentY = e.changedTouches[0].clientY;
    const pullDistance = currentY - startYRef.current;

    // Remove indicator
    const indicator = document.getElementById('pull-refresh-indicator');
    if (indicator) {
      indicator.remove();
    }

    if (pullDistance >= threshold) {
      isRefreshingRef.current = true;

      // Show refreshing indicator
      const refreshIndicator = document.createElement('div');
      refreshIndicator.id = 'refreshing-indicator';
      refreshIndicator.style.cssText = `
        position: fixed;
        top: 20px;
        left: 50%;
        transform: translateX(-50%);
        background: rgba(0, 123, 255, 0.9);
        color: white;
        padding: 8px 16px;
        border-radius: 20px;
        font-size: 14px;
        z-index: 1000;
        pointer-events: none;
        display: flex;
        align-items: center;
        gap: 8px;
      `;
      refreshIndicator.innerHTML = `
        <div style="
          width: 16px;
          height: 16px;
          border: 2px solid rgba(255,255,255,0.3);
          border-top: 2px solid white;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        "></div>
        <span>Обновление...</span>
      `;
      document.body.appendChild(refreshIndicator);

      try {
        await onRefresh();
      } finally {
        // Remove refreshing indicator
        const existingIndicator = document.getElementById('refreshing-indicator');
        if (existingIndicator) {
          existingIndicator.remove();
        }
        isRefreshingRef.current = false;
      }
    }

    isPullingRef.current = false;
  }, [disabled, threshold, onRefresh]);

  const bindPullToRefresh = useCallback((element: HTMLElement | null) => {
    if (!element || disabled) return;

    containerRef.current = element;

    element.addEventListener('touchstart', handleTouchStart, { passive: false });
    element.addEventListener('touchmove', handleTouchMove, { passive: false });
    element.addEventListener('touchend', handleTouchEnd, { passive: false });

    return () => {
      element.removeEventListener('touchstart', handleTouchStart);
      element.removeEventListener('touchmove', handleTouchMove);
      element.removeEventListener('touchend', handleTouchEnd);
    };
  }, [disabled, handleTouchStart, handleTouchMove, handleTouchEnd]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      const indicator = document.getElementById('pull-refresh-indicator');
      if (indicator) indicator.remove();

      const refreshIndicator = document.getElementById('refreshing-indicator');
      if (refreshIndicator) refreshIndicator.remove();
    };
  }, []);

  return { bindPullToRefresh };
};