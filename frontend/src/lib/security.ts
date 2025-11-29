/**
 * Утилиты безопасности для frontend
 * Frontend Security Utilities
 */

import DOMPurify from 'dompurify';

// Конфигурация DOMPurify для различных типов контента
export const SANITIZE_CONFIGS = {
  // Для описаний запчастей - разрешить базовое форматирование
  DESCRIPTION: {
    ALLOWED_TAGS: ['b', 'i', 'u', 'br', 'p', 'strong', 'em'],
    ALLOWED_ATTR: [],
    ALLOW_DATA_ATTR: false
  },

  // Для названий - только текст, без HTML
  NAME: {
    ALLOWED_TAGS: [],
    ALLOWED_ATTR: [],
    ALLOW_DATA_ATTR: false
  },

  // Для комментариев - минимальное форматирование
  COMMENT: {
    ALLOWED_TAGS: ['b', 'i', 'u', 'br'],
    ALLOWED_ATTR: [],
    ALLOW_DATA_ATTR: false
  },

  // Для отображения - разрешить больше тегов
  DISPLAY: {
    ALLOWED_TAGS: ['b', 'i', 'u', 'br', 'p', 'strong', 'em', 'span'],
    ALLOWED_ATTR: ['class'],
    ALLOW_DATA_ATTR: false
  }
};

/**
 * Санитизация HTML контента
 * @param {string} html - HTML строка для очистки
 * @param {string} type - Тип контента ('DESCRIPTION', 'NAME', 'COMMENT', 'DISPLAY')
 * @returns {string} Очищенная строка
 */
export function sanitizeHtml(html: string, type: keyof typeof SANITIZE_CONFIGS = 'DESCRIPTION'): string {
  if (!html || typeof html !== 'string') {
    return '';
  }

  const config = SANITIZE_CONFIGS[type] || SANITIZE_CONFIGS.DESCRIPTION;
  return DOMPurify.sanitize(html, config);
}

/**
 * Санитизация объекта с HTML полями
 * @param {Object} obj - Объект для очистки
 * @param {Array} htmlFields - Массив названий полей, содержащих HTML
 * @param {string} type - Тип контента для всех полей
 * @returns {Object} Объект с очищенными полями
 */
export function sanitizeObject<T extends Record<string, unknown>>(
  obj: T,
  htmlFields: (keyof T)[] = [],
  type: keyof typeof SANITIZE_CONFIGS = 'DESCRIPTION'
): T {
  if (!obj || typeof obj !== 'object') {
    return obj;
  }

  const sanitized = { ...obj };

  htmlFields.forEach(field => {
    if (sanitized[field] && typeof sanitized[field] === 'string') {
      (sanitized as Record<string, unknown>)[field as string] = sanitizeHtml(sanitized[field], type);
    }
  });

  return sanitized;
}

/**
 * Санитизация данных запчасти
 * @param {Object} part - Объект запчасти
 * @returns {Object} Запчасть с очищенными полями
 */
export function sanitizePart(part: Record<string, unknown>) {
  if (!part) return part;

  return sanitizeObject(part, ['name', 'description'], 'DESCRIPTION');
}

/**
 * Санитизация данных заказа
 * @param {Object} order - Объект заказа
 * @returns {Object} Заказ с очищенными полями
 */
export function sanitizeOrder(order: Record<string, unknown>) {
  if (!order) return order;

  return sanitizeObject(order, ['part'], 'NAME');
}

/**
 * Санитизация массива объектов
 * @param {Array} items - Массив объектов
 * @param {Function} sanitizer - Функция санитизации для одного объекта
 * @returns {Array} Массив с очищенными объектами
 */
export function sanitizeArray<T>(items: T[], sanitizer: (item: T) => T): T[] {
  if (!Array.isArray(items)) return items;

  return items.map(item => sanitizer(item));
}

/**
 * Безопасное отображение HTML контента в React
 * @param {string} html - HTML строка
 * @param {string} type - Тип контента
 * @returns {Object} Объект для dangerouslySetInnerHTML
 */
export function createSafeHtml(html: string, type: keyof typeof SANITIZE_CONFIGS = 'DISPLAY') {
  const sanitized = sanitizeHtml(html, type);
  return {
    __html: sanitized
  };
}

/**
 * Валидация и санитизация формы
 * @param {Object} formData - Данные формы
 * @param {Object} rules - Правила валидации
 * @returns {Object} Результат валидации и очищенные данные
 */
interface ValidationRule {
  required?: boolean;
  maxLength?: number;
  sanitize?: keyof typeof SANITIZE_CONFIGS;
}

export function validateAndSanitizeForm(formData: Record<string, unknown>, rules: Record<string, ValidationRule> = {}) {
  const errors: string[] = [];
  const sanitized: Record<string, unknown> = {};

  Object.keys(formData).forEach(field => {
    const value = formData[field];
    const rule = rules[field];

    // Санитизация
    if (rule && rule.sanitize && typeof value === 'string') {
      sanitized[field] = sanitizeHtml(value, rule.sanitize);
    } else {
      sanitized[field] = value;
    }

    // Валидация
    if (rule && rule.required && (!sanitized[field] || (typeof sanitized[field] === 'string' && sanitized[field].trim() === ''))) {
      errors.push(`${field} обязательно для заполнения`);
    }

    if (rule && rule.maxLength && typeof sanitized[field] === 'string' && sanitized[field].length > rule.maxLength) {
      errors.push(`${field} слишком длинный (макс. ${rule.maxLength} символов)`);
    }
  });

  return {
    isValid: errors.length === 0,
    errors,
    sanitizedData: sanitized
  };
}

/**
 * Правила валидации для форм
 */
export const VALIDATION_RULES = {
  part: {
    name: {
      required: true,
      maxLength: 255,
      sanitize: 'NAME' as const
    },
    description: {
      maxLength: 1000,
      sanitize: 'DESCRIPTION' as const
    },
    category: {
      sanitize: 'NAME' as const
    },
    brand: {
      sanitize: 'NAME' as const
    },
    model: {
      sanitize: 'NAME' as const
    },
    salesman: {
      sanitize: 'NAME' as const
    },
    location: {
      sanitize: 'NAME' as const
    }
  },

  order: {
    part: {
      required: true,
      sanitize: 'NAME' as const
    },
    order_number: {
      required: true,
      sanitize: 'NAME' as const
    },
    buyer_number: {
      required: true,
      sanitize: 'NAME' as const
    }
  }
};

/**
 * XSS защита для URL параметров
 * @param {string} param - URL параметр
 * @returns {string} Очищенный параметр
 */
export function sanitizeUrlParam(param: string): string {
  // Удалить потенциально опасные символы
  return param.replace(/[<>'"&]/g, '');
}

/**
 * Защита от XSS в JSON данных
 * @param {any} data - Данные для очистки
 * @returns {any} Очищенные данные
 */
export function sanitizeJsonData(data: unknown): unknown {
  if (Array.isArray(data)) {
    return data.map(item => sanitizeJsonData(item));
  }

  if (data && typeof data === 'object') {
    const sanitized: Record<string, unknown> = {};
    Object.keys(data).forEach(key => {
      // Санитизировать только строковые значения
      if (typeof (data as Record<string, unknown>)[key] === 'string') {
        sanitized[key] = sanitizeHtml((data as Record<string, unknown>)[key] as string, 'NAME');
      } else {
        sanitized[key] = sanitizeJsonData((data as Record<string, unknown>)[key]);
      }
    });
    return sanitized;
  }

  return data;
}