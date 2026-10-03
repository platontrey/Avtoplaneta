/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { describe, it, expect } from 'vitest';
import { cn, formatCarReleasePeriod } from './utils';

describe('formatCarReleasePeriod', () => {
  it('should return empty string for falsy input', () => {
    expect(formatCarReleasePeriod('')).toBe('');
  });

  it('should return 4 digits when exactly 4 digits entered', () => {
    expect(formatCarReleasePeriod('2005')).toBe('2005');
  });

  it('should insert hyphen when more than 4 digits entered', () => {
    expect(formatCarReleasePeriod('20052010')).toBe('2005-2010');
  });

  it('should preserve hyphen if trailing hyphen entered on 4 digits', () => {
    expect(formatCarReleasePeriod('2005-')).toBe('2005-');
  });

  it('should strip non-digits', () => {
    expect(formatCarReleasePeriod('abc2005xyz2010')).toBe('2005-2010');
  });
});

describe('cn utility', () => {
  it('should merge classes correctly', () => {
    expect(cn('p-4', 'bg-red-500')).toBe('p-4 bg-red-500');
  });

  it('should resolve tailwind class conflicts', () => {
    expect(cn('p-4', 'p-8')).toBe('p-8');
  });

  it('should handle conditional classes', () => {
    const isHidden = false;
    const isVisible = true;
    expect(cn('base', isHidden && 'hidden', isVisible && 'block')).toBe('base block');
  });
});
