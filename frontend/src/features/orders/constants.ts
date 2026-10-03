/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

export const SOURCE_LABELS: Record<string, string> = {
  drom: 'Дром',
  avito: 'Авито',
  messenger: 'Мессенджер',
  pickup: 'Самовывоз',
};

export const PAYMENT_OPTIONS = [
  { value: 'unpaid', label: '💳 Не оплачен', badgeClass: 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30' },
  { value: 'prepaid', label: '💵 Предоплата', badgeClass: 'bg-blue-500/15 text-blue-700 dark:text-blue-300 border-blue-500/30' },
  { value: 'paid', label: '✅ Оплачен', badgeClass: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/30' },
] as const;

export const WAREHOUSE_OPTIONS = [
  { value: 'inspecting', label: '📷 Проверка / Фото' },
  { value: 'transfer', label: '🔄 Перемещение' },
  { value: 'ready', label: '📦 Готов к выдаче' },
] as const;

export const DELIVERY_OPTIONS = [
  { value: 'pickup', label: '🏃 Самовывоз' },
  { value: 'tk', label: '🚚 Отправка ТК' },
  { value: 'city', label: '🚕 По городу' },
] as const;
