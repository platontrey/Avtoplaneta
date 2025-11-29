/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useCallback } from 'react';
import type { Part, PartFormData } from '../types';

interface UsePartFormOptions {
  initialPart?: Part;
}

interface UsePartFormReturn {
  formData: PartFormData;
  updateField: (field: keyof PartFormData, value: string | boolean) => void;
  resetForm: () => void;
  getFormData: () => Omit<Part, 'id'>;
}

const createInitialFormData = (part?: Part): PartFormData => ({
  name: part?.name || '',
  quantity: part?.quantity?.toString() || '',
  description: part?.description || '',
  category: part?.category || '',
  price: part?.price?.toString() || '',
  salesman: part?.salesman || '',
  location: part?.location || '',
  status: part?.status ?? true,
});

export function usePartForm(options: UsePartFormOptions = {}): UsePartFormReturn {
  const [formData, setFormData] = useState<PartFormData>(() =>
    createInitialFormData(options.initialPart)
  );

  const updateField = useCallback((field: keyof PartFormData, value: string | boolean) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  }, []);

  const resetForm = useCallback(() => {
    setFormData(createInitialFormData(options.initialPart));
  }, [options.initialPart]);

  const getFormData = useCallback((): Omit<Part, 'id'> => ({
    name: formData.name,
    quantity: parseInt(formData.quantity) || 0,
    description: formData.description || undefined,
    category: formData.category || undefined,
    price: formData.price ? parseFloat(formData.price) : undefined,
    salesman: formData.salesman || undefined,
    location: formData.location || undefined,
    status: formData.status,
  }), [formData]);

  return {
    formData,
    updateField,
    resetForm,
    getFormData,
  };
}