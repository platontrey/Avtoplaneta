/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { partsApi } from '@/features/parts/api/partsApi';
import type { Shelf } from '@/features/parts/types';

export const shelvesKeys = {
  all: ['shelves'] as const,
  list: () => [...shelvesKeys.all, 'list'] as const,
};

export function useShelves() {
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: shelvesKeys.list(),
    queryFn: () => partsApi.getShelves(),
    staleTime: 5 * 60 * 1000,
  });

  const createMutation = useMutation({
    mutationFn: (name: string) => partsApi.createShelf(name),
    onSuccess: (newShelf) => {
      queryClient.setQueryData<Shelf[]>(shelvesKeys.list(), (old) => {
        if (!old) return [newShelf];
        const exists = old.some((s) => s.id === newShelf.id || s.name.toLowerCase() === newShelf.name.toLowerCase());
        if (exists) return old;
        return [...old, newShelf].sort((a, b) => a.name.localeCompare(b.name));
      });
      queryClient.invalidateQueries({ queryKey: shelvesKeys.list() });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => partsApi.deleteShelf(id),
    onSuccess: (_, deletedId) => {
      queryClient.setQueryData<Shelf[]>(shelvesKeys.list(), (old) => {
        if (!old) return [];
        return old.filter((s) => s.id !== deletedId);
      });
      queryClient.invalidateQueries({ queryKey: shelvesKeys.list() });
    },
  });

  return {
    shelves: query.data || [],
    isLoading: query.isLoading,
    error: query.error,
    createShelf: createMutation.mutateAsync,
    isCreating: createMutation.isPending,
    deleteShelf: deleteMutation.mutateAsync,
    isDeleting: deleteMutation.isPending,
    refetch: query.refetch,
  };
}
