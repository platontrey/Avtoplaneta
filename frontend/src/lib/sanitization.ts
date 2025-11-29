/**
 * Утилиты для санитизации и валидации данных
 */

/**
 * Санитизирует строку, удаляя потенциально опасные символы
 * @param input - Входная строка
 * @returns Санитизированная строка
 */
export function sanitizeString(input: string): string {
  if (typeof input !== 'string') return '';
  return input.replace(/[<>]/g, '').trim();
}

/**
 * Санитизирует HTML, удаляя скрипты и потенциально опасные теги
 * @param html - HTML строка
 * @returns Санитизированный HTML
 */
export function sanitizeHtml(html: string): string {
  if (typeof html !== 'string') return '';
  // Простая санитизация - удаляем script и style теги
  return html.replace(/<script[^>]*>.*?<\/script>/gi, '').replace(/<style[^>]*>.*?<\/style>/gi, '');
}

/**
 * Санитизирует email адрес
 * @param email - Email строка
 * @returns Санитизированный email
 */
export function sanitizeEmail(email: string): string {
  if (typeof email !== 'string') return '';
  return email.toLowerCase().trim();
}

/**
 * Санитизирует числовое значение
 * @param value - Значение для санитизации
 * @returns Число или null если невалидно
 */
export function sanitizeNumber(value: unknown): number | null {
  const num = Number(value);
  return isNaN(num) ? null : num;
}

/**
 * Санитизирует массив строк
 * @param arr - Массив строк
 * @returns Массив санитизированных строк
 */
export function sanitizeStringArray(arr: unknown[]): string[] {
  if (!Array.isArray(arr)) return [];
  return arr.filter((item): item is string => typeof item === 'string').map(sanitizeString);
}

/**
 * Проверяет, является ли строка безопасной для использования в URL
 * @param str - Строка для проверки
 * @returns true если безопасна
 */
export function isSafeUrl(str: string): boolean {
  if (typeof str !== 'string') return false;
  // Простая проверка на наличие потенциально опасных протоколов
  return !str.match(/^(javascript|data|vbscript):/i);
}