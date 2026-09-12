/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import { partsApi } from '@/features/parts/api/partsApi';
import type { Part } from '@/features/parts/types';

/**
 * Ключи запросов для работы с запчастями
 */
export const partsKeys = {
  /** Базовый ключ для всех запросов запчастей */
  all: ['parts'] as const,
  /** Ключ для списков запчастей */
  lists: () => [...partsKeys.all, 'list'] as const,
  /** Ключ для списка с фильтрами */
  list: (filters: Record<string, unknown>) => [...partsKeys.lists(), filters] as const,
  /** Ключ для деталей запчастей */
  details: () => [...partsKeys.all, 'detail'] as const,
  /** Ключ для конкретной запчасти */
  detail: (id: number) => [...partsKeys.details(), id] as const,
};

function replacePartInCache(queryClient: QueryClient, updatedPart: Partial<Part> & { id: number }) {
  queryClient.setQueryData(partsKeys.detail(updatedPart.id), (oldData: Part | undefined) =>
    oldData ? { ...oldData, ...updatedPart } : updatedPart,
  );
  queryClient.setQueriesData({ queryKey: partsKeys.lists() }, (oldData: unknown) => {
    if (Array.isArray(oldData)) {
      return oldData.map((part) =>
        typeof part === 'object' && part !== null && 'id' in part && part.id === updatedPart.id
          ? { ...part, ...updatedPart }
          : part,
      );
    }

    if (
      oldData &&
      typeof oldData === 'object' &&
      'pages' in oldData &&
      Array.isArray((oldData as { pages: unknown[] }).pages)
    ) {
      const infiniteData = oldData as { pages: unknown[] };
      return {
        ...oldData,
        pages: infiniteData.pages.map((page) =>
          Array.isArray(page)
            ? page.map((part) =>
                typeof part === 'object' && part !== null && 'id' in part && part.id === updatedPart.id
                  ? { ...part, ...updatedPart }
                  : part,
              )
            : page,
        ),
      };
    }

    return oldData;
  });
}

function isPartResponse(value: unknown): value is Part {
  return (
    typeof value === 'object' &&
    value !== null &&
    'id' in value &&
    typeof value.id === 'number'
  );
}

/**
 * Хук для получения списка запчастей с фильтрами
 * @param filters - Объект с фильтрами для поиска запчастей
 * @returns Объект с данными, состоянием загрузки и ошибками
 */
export function useParts(filters?: {
  /** Строка поиска */
  search?: string;
  /** Категория запчасти */
  category?: string;
  /** Бренд автомобиля */
  brand?: string;
  /** Модель автомобиля */
  model?: string;
  /** Местоположение */
  location?: string;
  /** Адрес склада */
  address?: string;
  /** Продавец */
  salesman?: string;
  /** Статус запчасти */
  status?: string;
  /** Фильтр по наличию фото */
  hasPhoto?: string;
  /** Лимит результатов */
  limit?: number;
  /** Страница для пагинации */
  page?: number;
}) {
  console.log('useParts called with filters:', filters);
  return useQuery({
    queryKey: partsKeys.list(filters || {}),
    queryFn: () => {
      console.log('useParts queryFn executing with filters:', filters);
      return partsApi.getAll(filters || {});
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}

/**
 * Хук для бесконечной прокрутки запчастей
 * @param filters - Объект с фильтрами для поиска запчастей
 * @returns Объект с данными, функциями для загрузки следующей страницы
 */
export function useInfiniteParts(filters?: {
  /** Строка поиска */
  search?: string;
  /** Категория запчасти */
  category?: string;
  /** Бренд автомобиля */
  brand?: string;
  /** Модель автомобиля */
  model?: string;
  /** Местоположение */
  location?: string;
  /** Адрес склада */
  address?: string;
  /** Продавец */
  salesman?: string;
  /** Статус запчасти */
  status?: string;
  /** Фильтр по наличию фото */
  hasPhoto?: string;
}) {
  console.log('useInfiniteParts called with filters:', filters);
  return useInfiniteQuery({
    queryKey: [...partsKeys.lists(), 'infinite', filters || {}],
    queryFn: ({ pageParam = 1 }) => {
      console.log('useInfiniteParts queryFn executing with filters:', filters, 'page:', pageParam);
      return partsApi.getAll({
        ...filters,
        page: pageParam,
        limit: 20, // Фиксированный лимит для бесконечной прокрутки
      });
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      // Если последняя страница содержит меньше 20 элементов, значит это последняя
      return lastPage.length === 20 ? allPages.length + 1 : undefined;
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}

/**
 * Хук для создания новой запчасти
 * @returns Мутационный объект для создания запчасти
 */
export function useCreatePart() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: partsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
    },
  });
}

/**
 * Хук для обновления существующей запчасти
 * @returns Мутационный объект для обновления запчасти
 */
export function useUpdatePart() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Part> }) =>
      partsApi.update(id, data),
    onSuccess: (updatedPart, { id, data }) => {
      replacePartInCache(queryClient, isPartResponse(updatedPart) ? updatedPart : { ...data, id } as Part);
      void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
      // Also refetch immediately to ensure UI updates
      void queryClient.refetchQueries({ queryKey: partsKeys.lists() });
    },
  });
}

/**
 * Хук для удаления запчасти
 * @returns Мутационный объект для удаления запчасти
 */
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

/**
 * Хук для загрузки фото запчасти
 * @returns Мутационный объект для загрузки фото
 */
export function useUploadPartPhoto() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, file }: { id: number; file: File }) =>
      partsApi.uploadPhoto(id, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
        // Также немедленное повторение загрузки, что бы обеспечить обновление UI
      queryClient.refetchQueries({ queryKey: partsKeys.lists() });
    },
  });
}
