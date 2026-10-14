/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { CustomerCategory } from '@/lib/types';

export const CATEGORY_LABELS: Record<CustomerCategory, { label: string; badgeClass: string; icon: string }> = {
  regular: {
    label: 'Обычный',
    badgeClass: 'bg-zinc-500/15 text-zinc-700 dark:text-zinc-300 border-zinc-500/30',
    icon: '👤',
  },
  vip: {
    label: 'VIP / СТО',
    badgeClass: 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30',
    icon: '⭐',
  },
  wholesale: {
    label: 'Оптовик',
    badgeClass: 'bg-blue-500/15 text-blue-700 dark:text-blue-300 border-blue-500/30',
    icon: '🏢',
  },
  blacklist: {
    label: 'Чёрный список',
    badgeClass: 'bg-rose-500/15 text-rose-700 dark:text-rose-300 border-rose-500/30 font-semibold',
    icon: '⛔',
  },
};
