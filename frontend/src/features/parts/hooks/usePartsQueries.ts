/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { Part } from '../types';
import { partsApi } from '../api/partsApi';

// Query keys
const PARTS_QUERY_KEYS = {
  all: ['parts'] as const,
} as const;

export const partsKeys = {
  all: PARTS_QUERY_KEYS.all,
  lists: () => [...PARTS_QUERY_KEYS.all, 'list'] as const,
  list: (filters: Record<string, unknown>) => [...PARTS_QUERY_KEYS.all, 'list', filters] as const,
  details: () => [...PARTS_QUERY_KEYS.all, 'detail'] as const,
  detail: (id: number) => [...PARTS_QUERY_KEYS.all, 'detail', id] as const,
};

// Get all parts with optional filters
export function useParts(filters?: {
  search?: string;
  category?: string;
  brand?: string;
  model?: string;
  location?: string;
  salesman?: string;
  status?: string;
}) {
  return useQuery({
    queryKey: partsKeys.list(filters || {}),
    queryFn: () => partsApi.getAll(filters || {}),
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}

// Create part mutation
export function useCreatePart() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: partsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
    },
  });
}

// Update part mutation
export function useUpdatePart() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Part> }) =>
      partsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
      // Also refetch immediately to ensure UI updates
      queryClient.refetchQueries({ queryKey: partsKeys.lists() });
    },
  });
}

// Delete part mutation
export function useDeletePart() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: partsApi.delete,
    onSuccess: () => {
      console.log('useDeletePart: onSuccess called, invalidating queries');
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
    },
    onError: (error) => {
      console.error('useDeletePart: onError called:', error);
    },
  });
}

// Upload part photo mutation
export function useUploadPartPhoto() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, file }: { id: number; file: File }) =>
      partsApi.uploadPhoto(id, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
    },
  });
}