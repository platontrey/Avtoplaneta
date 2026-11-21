/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

export const SOURCE_LABELS: Record<string, string> = {
  drom: 'Дром',
  avito: 'Авито',
  messenger: 'Мессенджер',
  pickup: 'Самовывоз',
};

// 1. Статус сделки (воронка продажи)
export const ORDER_STATUS_OPTIONS = [
  { value: 'new', label: '🆕 Новый заказ', badgeClass: 'bg-blue-500/15 text-blue-700 dark:text-blue-300 border-blue-500/30' },
  { value: 'processing', label: '⚙️ В работе', badgeClass: 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30' },
  { value: 'ready', label: '📦 Готов к выдаче', badgeClass: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/30' },
  { value: 'completed', label: '✅ Выдан / Завершён', badgeClass: 'bg-slate-500/15 text-slate-700 dark:text-slate-300 border-slate-500/30' },
  { value: 'cancelled', label: '❌ Отменён', badgeClass: 'bg-rose-500/15 text-rose-700 dark:text-rose-300 border-rose-500/30' },
] as const;

// 2. Статус оплаты
export const PAYMENT_OPTIONS = [
  { value: 'unpaid', label: '🔴 Не оплачен', badgeClass: 'bg-rose-500/15 text-rose-700 dark:text-rose-300 border-rose-500/30' },
  { value: 'prepaid', label: '🟡 Предоплата', badgeClass: 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30' },
  { value: 'paid', label: '🟢 Оплачен 100%', badgeClass: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/30' },
  { value: 'refunded', label: '🔄 Возврат средств', badgeClass: 'bg-purple-500/15 text-purple-700 dark:text-purple-300 border-purple-500/30' },
] as const;

// 3. Складской статус WMS (строго в хронологическом порядке прохождения заказа)
export const WAREHOUSE_OPTIONS = [
  { value: 'pending', label: '📋 В очереди сборки', badgeClass: 'bg-slate-500/15 text-slate-700 dark:text-slate-300 border-slate-500/30' },
  { value: 'inspecting', label: '🔍 Осмотр / Фото', badgeClass: 'bg-cyan-500/15 text-cyan-700 dark:text-cyan-300 border-cyan-500/30' },
  { value: 'picking', label: '🛒 В сборке', badgeClass: 'bg-indigo-500/15 text-indigo-700 dark:text-indigo-300 border-indigo-500/30' },
  { value: 'packed', label: '🎁 Собран и упакован', badgeClass: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/30' },
  { value: 'shipped', label: '🚚 Отгружен / Выдан', badgeClass: 'bg-blue-500/15 text-blue-700 dark:text-blue-300 border-blue-500/30' },
] as const;

// Хелпер для поиска понятного ярлыка склада даже по устаревшим значениям
export const getWarehouseStatusLabel = (status?: string): string => {
  if (!status) return '📋 В очереди сборки';
  if (status === 'transfer') return '🛒 В сборке';
  if (status === 'ready') return '🎁 Собран и упакован';
  const found = WAREHOUSE_OPTIONS.find((w) => w.value === status);
  return found ? found.label : status;
};

// Нормализация устаревших значений warehouse_status в выпадающих списках
export const normalizeWarehouseStatus = (status?: string): string => {
  if (!status || status === 'pending') return 'pending';
  if (status === 'transfer' || status === 'picking') return 'picking';
  if (status === 'ready' || status === 'packed') return 'packed';
  if (status === 'inspecting') return 'inspecting';
  if (status === 'shipped') return 'shipped';
  return status;
};

// 4. Способ получения
export const DELIVERY_OPTIONS = [
  { value: 'pickup', label: '🏃 Самовывоз' },
  { value: 'tk', label: '🚚 Отправка ТК' },
  { value: 'city', label: '🚕 Доставка по городу' },
] as const;
