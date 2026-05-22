/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState, useEffect } from 'react';

export function useDebouncedSearch(initialValue: string = '', delay: number = 500) {
  const [debouncedValue, setDebouncedValue] = useState(initialValue);
  const [immediateValue, setImmediateValue] = useState(initialValue);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(immediateValue);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [immediateValue, delay]);

  return {
    debouncedValue,
    immediateValue,
    setImmediateValue
  };
}