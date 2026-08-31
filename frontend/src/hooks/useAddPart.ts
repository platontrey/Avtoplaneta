/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCreatePart, useUploadPartPhoto } from './useParts';
import type { Part } from '@/features/parts/types';

interface UseAddPartReturn {
  formData: {
    brand: string;
    name: string;
    model: string;
    quantity: string;
    description: string;
    category: string;
    location: string;
    address: string;
    price: string;
    salesman: string;
  };
  photoFile: File | null;
  photoPreview: string | null;
  isLoading: boolean;
  setBrand: (brand: string) => void;
  updateFormField: (field: string, value: string) => void;
  handlePhotoChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  handleSubmit: (e: React.FormEvent<HTMLFormElement>) => Promise<void>;
}

/**
 * Custom hook for add part form logic
 * Handles form state, validation, and submission
 */
export function useAddPart(): UseAddPartReturn {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    brand: '',
    name: '',
    model: '',
    quantity: '0',
    description: '',
    category: '',
    location: '',
    address: '',
    price: '',
    salesman: '',
  });

  const [photoFile, setPhotoFile] = useState<File | null>(null);
  const [photoPreview, setPhotoPreview] = useState<string | null>(null);

  const createPartMutation = useCreatePart();
  const uploadPhotoMutation = useUploadPartPhoto();

  /**
   * Update brand field
   */
  const setBrand = (brand: string) => {
    setFormData(prev => ({ ...prev, brand }));
  };

  /**
   * Update form field
   */
  const updateFormField = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  /**
   * Handle photo file selection
   */
  const handlePhotoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('image/')) {
      alert('Please select a valid image file.');
      return;
    }

    // Validate file size (max 5MB)
    const maxSize = 5 * 1024 * 1024;
    if (file.size > maxSize) {
      alert('File size must be less than 5MB.');
      return;
    }

    setPhotoFile(file);
    const reader = new FileReader();
    reader.onload = (e) => {
      setPhotoPreview(e.target?.result as string);
    };
    reader.readAsDataURL(file);
  };

  /**
   * Handle form submission
   */
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    // Validation
    if (!formData.brand) {
      alert('Please select a brand.');
      return;
    }

    if (!formData.name.trim()) {
      alert('Please enter a part name.');
      return;
    }

    if (!formData.model.trim()) {
      alert('Please enter a model.');
      return;
    }


    const quantity = parseInt(formData.quantity);
    if (isNaN(quantity) || quantity < 0) {
      alert('Please enter a valid quantity.');
      return;
    }

    try {
      // Create part data
      const newPart: Omit<Part, 'id'> = {
        brand: formData.brand,
        name: formData.name,
        model: formData.model,
        quantity,
        description: formData.description,
        category: formData.category,
        location: formData.location || undefined,
        address: formData.address || undefined,
        price: formData.price ? parseFloat(formData.price) : undefined,
        salesman: formData.salesman || undefined,
      };

      // Create the part
      const result = await createPartMutation.mutateAsync(newPart);

      // Upload photo if selected
      if (photoFile && result.id) {
        await uploadPhotoMutation.mutateAsync({
          id: result.id,
          file: photoFile,
        });
      }

      alert('Part added successfully!');
      navigate('/inventory');
    } catch (error) {
      console.error('Failed to add part:', error);
      alert('Failed to add part. Please try again.');
    }
  };

  return {
    formData,
    photoFile,
    photoPreview,
    isLoading: createPartMutation.isPending || uploadPhotoMutation.isPending,
    setBrand,
    updateFormField,
    handlePhotoChange,
    handleSubmit,
  };
}
