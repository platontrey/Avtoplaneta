/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import { API_BASE_URL } from './api';

export interface ClientErrorPayload {
  message: string;
  stack?: string;
  url: string;
  timestamp: string;
  userAgent: string;
  userId?: string | null;
  context?: Record<string, unknown>;
}

/**
 * Отправка критических клиентских ошибок на бэкенд для аудита и мониторинга (Observability).
 */
export function logClientError(error: unknown, context?: Record<string, unknown>): void {
  const err = error instanceof Error ? error : new Error(String(error));
  const userId = typeof localStorage !== 'undefined' ? localStorage.getItem('userId') : null;

  const payload: ClientErrorPayload = {
    message: err.message,
    stack: err.stack,
    url: typeof window !== 'undefined' ? window.location.href : '',
    timestamp: new Date().toISOString(),
    userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : '',
    userId,
    context,
  };

  // Не спамим сетевыми логами при локальной разработке, если бэкенд не слушает телеметрию
  if (import.meta.env.DEV) {
    console.debug('[Telemetry] Captured error:', payload);
    return;
  }

  try {
    const endpoint = `${API_BASE_URL}/api/v1/telemetry/errors`;
    const body = JSON.stringify(payload);

    if (typeof navigator !== 'undefined' && navigator.sendBeacon) {
      navigator.sendBeacon(endpoint, body);
    } else {
      fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body,
        keepalive: true,
      }).catch(() => {
        // Ошибка отправки телеметрии не должна прерывать работу UI
      });
    }
  } catch {
    // Безопасный перехват
  }
}

/**
 * Глобальная инициализация перехватчиков критических ошибок приложения.
 */
export function initGlobalErrorHandlers(): void {
  if (typeof window === 'undefined') return;

  // Перехват unhandled promise rejections (необработанных асинхронных сбоев)
  window.addEventListener('unhandledrejection', (event) => {
    // Игнорируем штатные отмены запросов (AbortController / DOMException)
    if (event.reason instanceof DOMException && event.reason.name === 'AbortError') {
      return;
    }

    console.error('Unhandled Promise Rejection:', event.reason);
    logClientError(event.reason, { type: 'unhandledrejection' });
  });

  // Перехват глобальных синтаксических или рантайм ошибок
  window.addEventListener('error', (event) => {
    // Фильтруем внешние ошибки скриптов расширений браузера
    if (event.filename && !event.filename.includes(window.location.origin)) {
      return;
    }

    logClientError(event.error || event.message, {
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      type: 'global_error',
    });
  });
}
