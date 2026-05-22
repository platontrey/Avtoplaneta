/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import type { OrderForm } from './useOrderDialog';

export const INPUT_LIMITS = {
  TEXT_MAX_LENGTH: 100,
  CUSTOMER_ID_MAX_LENGTH: 10,
  QUANTITY_MAX: 9999,
  QUANTITY_MIN: 1,
} as const;

/**
 * Sanitizes text input by removing HTML tags and trimming whitespace
 */
export const sanitizeText = (text: string, maxLength: number = INPUT_LIMITS.TEXT_MAX_LENGTH): string => {
  return text.trim().replace(/[<>]/g, '').substring(0, maxLength);
};

/**
 * Sanitizes number input within specified range
 */
export const sanitizeNumber = (value: string, min: number = INPUT_LIMITS.QUANTITY_MIN, max: number = INPUT_LIMITS.QUANTITY_MAX): number => {
  const num = parseInt(value, 10);
  return Math.max(min, Math.min(max, isNaN(num) ? min : num));
};

/**
 * Sanitizes customer ID (only digits, max length)
 */
export const sanitizeCustomerId = (value: string): string => {
  return value.replace(/[^0-9]/g, '').substring(0, INPUT_LIMITS.CUSTOMER_ID_MAX_LENGTH);
};

/**
 * Validates order form data
 */
export const validateOrderForm = (form: OrderForm): { isValid: boolean; errors: string[] } => {
  const errors: string[] = [];

  if (!form.customer_id.trim()) {
    errors.push('ID клиента обязателен');
  } else if (!/^\d+$/.test(form.customer_id)) {
    errors.push('ID клиента должен содержать только цифры');
  }

  if (!form.order_number.trim()) {
    errors.push('Номер заказа обязателен');
  }

  if (!form.buyer_number.trim()) {
    errors.push('Номер покупателя обязателен');
  }

  if (form.quantity < INPUT_LIMITS.QUANTITY_MIN || form.quantity > INPUT_LIMITS.QUANTITY_MAX) {
    errors.push(`Количество должно быть от ${INPUT_LIMITS.QUANTITY_MIN} до ${INPUT_LIMITS.QUANTITY_MAX}`);
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};