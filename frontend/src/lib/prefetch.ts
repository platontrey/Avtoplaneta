/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

/**
 * Безопасный запуск фоновых задач в свободные от рендера циклы процессора.
 */
export function scheduleIdle(callback: () => void, timeout = 2000): void {
  if (typeof window !== 'undefined' && 'requestIdleCallback' in window) {
    (window as unknown as { requestIdleCallback: (cb: () => void, opts: { timeout: number }) => void }).requestIdleCallback(
      callback,
      { timeout }
    );
  } else {
    setTimeout(callback, Math.min(timeout, 500));
  }
}

// Кэш для предотвращения повторных вызовов предзагрузки
const preloadedModules = new Set<string>();

function preloadModule(key: string, loader: () => Promise<unknown>): void {
  if (preloadedModules.has(key)) return;
  preloadedModules.add(key);
  loader().catch((err) => {
    preloadedModules.delete(key);
    console.debug(`Prefetch failed for ${key}:`, err);
  });
}

// Точечные префетчеры компонентов
export const prefetchInventory = () => preloadModule('inventory', () => import('@/components/Inventory'));
export const prefetchOrders = () => preloadModule('orders', () => import('@/components/Orders'));
export const prefetchStatistics = () => preloadModule('statistics', () => import('@/components/Statistics'));
export const prefetchMessages = () => preloadModule('messages', () => import('@/components/MessagesPage'));
export const prefetchAddCar = () => preloadModule('add-car', () => import('@/components/AddCar'));
export const prefetchAddPart = () => preloadModule('add-part', () => import('@/components/AddPart'));
export const prefetchDefectReport = () => preloadModule('defect-report', () => import('@/components/DefectReport'));
export const prefetchAdmin = () => preloadModule('admin', () => import('@/components/AdminPanel'));

/**
 * Предзагрузка ключевых экранов при нахождении на странице логина
 */
export function preloadCoreRoutes(): void {
  prefetchInventory();
  prefetchOrders();
}

/**
 * Отложенный фоновый прогрев вторичных разделов после монтирования инвентаря
 */
export function preloadSecondaryRoutes(): void {
  scheduleIdle(() => {
    prefetchStatistics();
    prefetchMessages();
    prefetchAddCar();
  }, 1500);
}
