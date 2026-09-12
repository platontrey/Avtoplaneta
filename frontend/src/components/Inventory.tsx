/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState, useCallback, useEffect, useRef } from 'react';
import PartsSearch from './PartsSearch';
import PartsList from './PartsList';
import { useParts, useInfiniteParts } from '@/hooks/useParts';
import { usePullToRefresh } from '@/hooks/usePullToRefresh';
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/hooks/useParts';
import type { Part } from '@/lib/types';

const hasSamePartObjects = (currentParts: Part[], nextParts: Part[]) =>
    currentParts.length === nextParts.length &&
    currentParts.every((part, index) => part === nextParts[index]);

function Inventory() {
    const containerRef = useRef<HTMLDivElement>(null);
    const queryClient = useQueryClient();

    const [filters, setFilters] = useState<{
          search: string;
          category: string;
          brand: string;
          model: string;
          location: string;
          address: string;
          status: string;
          hasPhoto?: string;
      }>({
          search: '',
          category: '',
          brand: '',
          model: '',
          location: '',
          address: '',
          status: ''
      });

      const [displayLimit, setDisplayLimit] = useState<number | undefined>(undefined);
      const [allParts, setAllParts] = useState<Part[]>([]);
  
      // Pull to refresh functionality
      const { bindPullToRefresh } = usePullToRefresh({
          onRefresh: async () => {
              await queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
          },
          threshold: 80
      });
  
       // Выбираем хук в зависимости от режима
       const isInfiniteMode = displayLimit === undefined;
 
       // Для бесконечной прокрутки
       const infiniteQuery = useInfiniteParts(isInfiniteMode ? filters : undefined);
 
       // Для фиксированного лимита
       const regularQuery = useParts(!isInfiniteMode ? {
          ...filters,
          limit: displayLimit || 20,
          page: 1
       } : undefined);
 
       // Выбираем данные в зависимости от режима
       const { data: partsData, isLoading, error, hasNextPage, fetchNextPage, isFetchingNextPage } = isInfiniteMode ? {
          data: infiniteQuery.data?.pages.flat() || [],
          isLoading: infiniteQuery.isLoading,
          error: infiniteQuery.error,
          hasNextPage: infiniteQuery.hasNextPage,
          fetchNextPage: infiniteQuery.fetchNextPage,
          isFetchingNextPage: infiniteQuery.isFetchingNextPage,
       } : {
          data: regularQuery.data || [],
          isLoading: regularQuery.isLoading,
          error: regularQuery.error,
          hasNextPage: false,
          fetchNextPage: () => {},
          isFetchingNextPage: false,
       };


     // Обновляем allParts при получении данных
     useEffect(() => {
         if (partsData) {
             // Используем функциональное обновление, чтобы избежать лишних ререндеров
             setAllParts(prev => {
                 // Если это первая страница или данные полностью заменились (например, при фильтрации)
                 if (!isInfiniteMode) {
                     if (hasSamePartObjects(prev, partsData)) {
                         return prev;
                     }
                     return partsData;
                 }

                 // Для бесконечной прокрутки:
                 // partsData содержит все загруженные страницы (flat).
                 // Просто обновляем состояние, если оно отличается.
                 if (hasSamePartObjects(prev, partsData)) {
                     return prev;
                 }
                 return partsData;
             });
         }
     }, [partsData, isInfiniteMode]);

     // Bind pull to refresh
     useEffect(() => {
         return bindPullToRefresh(containerRef.current);
     }, [bindPullToRefresh]);

     // Обработчик применения фильтров
     const handleFiltersChange = useCallback((newFilters: {
         search: string;
         category: string;
         brand: string;
         model: string;
         location: string;
         address: string;
         status: string;
         hasPhoto?: string;
     }) => {
         setFilters(newFilters);
         setAllParts([]);
     }, []);

     // Обработчик изменения лимита отображения
     const handleDisplayLimitChange = useCallback((newLimit: number | undefined) => {
         setDisplayLimit(newLimit);
         setAllParts([]);
     }, []);

     // Функция загрузки дополнительных данных
     const loadMore = useCallback(() => {
         if (isInfiniteMode && hasNextPage && !isFetchingNextPage) {
             fetchNextPage();
         }
     }, [isInfiniteMode, hasNextPage, isFetchingNextPage, fetchNextPage]);


   return (
     <div ref={containerRef} className="max-w-7xl mx-auto px-4 sm:px-6 py-4" style={{ minHeight: '100vh' }}>
       <div className="mb-4 sm:mb-6">
         <h1 className="text-2xl sm:text-3xl font-bold mb-2">Доступные запчасти</h1>
         <p className="text-gray-600 mb-6 sm:mb-8 text-sm sm:text-base">Управляйте вашими запчастями в инвентаре</p>

         {/* Компонент поиска */}
         <PartsSearch
           onFiltersChange={handleFiltersChange}
           onDisplayLimitChange={handleDisplayLimitChange}
           currentDisplayLimit={displayLimit}
         />
       </div>

       {/* Компонент списка запчастей */}
       <PartsList
         parts={allParts}
         isLoading={isLoading || isFetchingNextPage}
         error={error}
         onLoadMore={isInfiniteMode ? loadMore : undefined}
         hasMore={hasNextPage}
         isInfiniteScroll={isInfiniteMode}
       />
     </div>
   );
 }

export default Inventory;
