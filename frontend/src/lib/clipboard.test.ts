/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { copyToClipboard, formatPartCopyText } from './clipboard';
import type { Part } from '@/features/parts/types';

describe('clipboard utils', () => {
  describe('formatPartCopyText', () => {
    it('should format all provided part properties cleanly', () => {
      const part: Part = {
        id: 1,
        name: 'Фара передняя левая',
        brand: 'Toyota',
        model: 'Camry',
        price: 15000,
        quantity: 2,
        location: 'Склад А',
        address: 'ул. Ленина 10',
        oem_code: '81150-33620',
        number: '12345',
        condition: 'Б/У',
        description: 'В отличном состоянии, все крепления целые.',
      };

      const text = formatPartCopyText(part);
      expect(text).toContain('Фара передняя левая');
      expect(text).toContain('Автомобиль: Toyota Camry');
      expect(text).toMatch(/Цена:\s*₽15[\s\u00A0]000/);
      expect(text).toContain('В наличии: 2 шт.');
      expect(text).toContain('Склад: Склад А (ул. Ленина 10)');
      expect(text).toContain('OEM: 81150-33620');
      expect(text).toContain('Номер: 12345');
      expect(text).toContain('Состояние: Б/У');
      expect(text).toContain('В отличном состоянии, все крепления целые.');
    });

    it('should handle partial or empty part properties without errors', () => {
      const part: Part = {
        id: 2,
        name: 'Колодки',
        quantity: 0,
      };

      const text = formatPartCopyText(part);
      expect(text).toContain('Колодки');
      expect(text).toContain('В наличии: 0 шт.');
      expect(text).not.toContain('Автомобиль:');
      expect(text).not.toContain('Цена:');
      expect(text).not.toContain('OEM:');
    });
  });

  describe('copyToClipboard', () => {
    beforeEach(() => {
      vi.restoreAllMocks();
      (globalThis as any).window = { isSecureContext: true };
      (globalThis as any).navigator = {};
      (globalThis as any).document = {
        createElement: vi.fn(() => ({
          style: {},
          setAttribute: vi.fn(),
          focus: vi.fn(),
          select: vi.fn(),
          setSelectionRange: vi.fn(),
        })),
        body: {
          appendChild: vi.fn(),
          removeChild: vi.fn(),
        },
        execCommand: vi.fn().mockReturnValue(true),
      };
    });

    it('should return false when text is empty', async () => {
      const res = await copyToClipboard('');
      expect(res).toBe(false);
    });

    it('should use navigator.clipboard if available and secure', async () => {
      const writeTextMock = vi.fn().mockResolvedValue(undefined);
      (globalThis as any).navigator = {
        clipboard: { writeText: writeTextMock },
      };
      (globalThis as any).window = { isSecureContext: true };

      const res = await copyToClipboard('test text');
      expect(res).toBe(true);
      expect(writeTextMock).toHaveBeenCalledWith('test text');
    });

    it('should fallback to execCommand if navigator.clipboard fails', async () => {
      const writeTextMock = vi.fn().mockRejectedValue(new Error('Permission denied'));
      (globalThis as any).navigator = {
        clipboard: { writeText: writeTextMock },
      };
      const execCommandMock = vi.fn().mockReturnValue(true);
      (globalThis as any).document.execCommand = execCommandMock;

      const res = await copyToClipboard('fallback text');
      expect(res).toBe(true);
      expect(execCommandMock).toHaveBeenCalledWith('copy');
    });
  });
});
