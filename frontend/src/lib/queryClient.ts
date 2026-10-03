/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { QueryClient } from '@tanstack/react-query';

/**
 * Эшелонированные политики свежести данных (Enterprise Tiered Caching).
 */
export const CACHE_TIERS = {
  /**
   * Справочники и стабильные сущности: бренды авто, категории, склады, роли сотрудников.
   * Данные меняются крайне редко — держим в памяти и не дергаем сеть.
   */
  STATIC: {
    staleTime: 10 * 60 * 1000, // 10 минут
    gcTime: 60 * 60 * 1000, // 1 час
    refetchOnWindowFocus: false,
  },

  /**
   * Каталог запчастей и инвентарь: сбалансированная свежесть с тихой фоновой ревалидацией.
   */
  CATALOG: {
    staleTime: 60 * 1000, // 1 минута
    gcTime: 15 * 60 * 1000, // 15 минут
    refetchOnWindowFocus: false,
  },

  /**
   * Оперативные бизнес-данные: заказы, сделки, сообщения, аудит.
   * Требуют своевременной синхронизации при активности оператора.
   */
  REALTIME: {
    staleTime: 15 * 1000, // 15 секунд
    gcTime: 5 * 60 * 1000, // 5 минут
    refetchOnWindowFocus: true,
  },
} as const;

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1 минута по умолчанию
      gcTime: 1000 * 60 * 15, // 15 минут
      refetchOnWindowFocus: false, // Защита от шквала запросов при переключении между окнами
      refetchOnReconnect: true, // Авторевалидация при восстановлении интернет-связи
      retry: (failureCount, error) => {
        // Не повторяем запросы при фатальных 4xx статусах
        if (
          error instanceof Error &&
          (error.message.includes('401') ||
            error.message.includes('403') ||
            error.message.includes('404'))
        ) {
          return false;
        }
        return failureCount < 2;
      },
      // Экспоненциальный бэкофф с защитой от перегрузки бэкенда
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    },
    mutations: {
      retry: false,
    },
  },
});