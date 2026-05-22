/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState } from 'react';
import { useParts, useCreatePart, useDeletePart } from './useParts';
import type { Part } from '@/lib/types';

interface UseInventoryReturn {
  parts: Part[];
  isLoading: boolean;
  error: Error | null;
  addPart: (partData: Omit<Part, 'id'>) => Promise<void>;
  deletePart: (id: number) => Promise<void>;
  isAddingPart: boolean;
  isDeletingPart: boolean;
}

/**
 * Custom hook for inventory management logic
 * Handles parts data fetching, adding, and deleting
 */
export function useInventory(): UseInventoryReturn {
  const [isDeletingPart, setIsDeletingPart] = useState(false);

  // React Query hooks
  const {
    data: parts = [],
    isLoading,
    error,
  } = useParts();

  const createPartMutation = useCreatePart();
  const deletePartMutation = useDeletePart();

  /**
   * Add a new part to inventory
   */
  const addPart = async (partData: Omit<Part, 'id'>) => {
    try {
      await createPartMutation.mutateAsync(partData);
    } catch (error) {
      console.error('Failed to add part:', error);
      throw error;
    }
  };

  /**
   * Delete a part from inventory
   */
  const deletePart = async (id: number) => {
    if (isDeletingPart) return;

    setIsDeletingPart(true);
    try {
      await deletePartMutation.mutateAsync(id);
    } catch (error) {
      console.error('Failed to delete part:', error);
      throw error;
    } finally {
      setIsDeletingPart(false);
    }
  };

  return {
    parts,
    isLoading,
    error: error as Error | null,
    addPart,
    deletePart,
    isAddingPart: createPartMutation.isPending,
    isDeletingPart: deletePartMutation.isPending || isDeletingPart,
  };
}
