/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState, useCallback, useEffect, useRef } from 'react';
import PartsSearch from './PartsSearch';
import PartsList from './PartsList';
import { useParts } from '@/hooks/useParts';
import { usePullToRefresh } from '@/hooks/usePullToRefresh';
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/hooks/useParts';
import type { Part } from '@/lib/types';

function Inventory() {
    const containerRef = useRef<HTMLDivElement>(null);
    const queryClient = useQueryClient();

    const [filters, setFilters] = useState<{
          search: string;
          category: string;
          brand: string;
          model: string;
          location: string;
          status: string;
          hasPhoto?: string;
      }>({
          search: '',
          category: '',
          brand: '',
          model: '',
          location: '',
          status: ''
      });

      const [displayLimit, setDisplayLimit] = useState<number | undefined>(20);
      const [currentPage, setCurrentPage] = useState(1);
      const [allParts, setAllParts] = useState<Part[]>([]);
      const [hasMore, setHasMore] = useState(true);
      const [isLoadingMore, setIsLoadingMore] = useState(false);
  
      // Pull to refresh functionality
      const { bindPullToRefresh } = usePullToRefresh({
          onRefresh: async () => {
              await queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
          },
          threshold: 80
      });
  
       // Используем React Query хук для загрузки данных
     const { data: partsData, isLoading, error } = useParts({
         ...filters,
         limit: displayLimit === undefined ? 10000 : (displayLimit || 20),
         page: displayLimit !== undefined ? 1 : currentPage
     });

     // Обновляем allParts при получении данных
     useEffect(() => {
         if (partsData) {
             if (displayLimit !== undefined) {
                 // Если limit установлен, просто используем данные
                 setAllParts(partsData);
                 setHasMore(false); // Все данные загружены сразу
             } else {
                 // Если пагинация, добавляем к существующим (оптимизированная конкатенация)
                 setAllParts(prev => {
                     if (currentPage === 1) {
                         return partsData;
                     } else {
                         // Используем более эффективную конкатенацию для больших массивов
                         const newArray = new Array(prev.length + partsData.length);
                         for (let i = 0; i < prev.length; i++) {
                             newArray[i] = prev[i];
                         }
                         for (let i = 0; i < partsData.length; i++) {
                             newArray[prev.length + i] = partsData[i];
                         }
                         return newArray;
                     }
                 });
                 setHasMore(partsData.length === 20);
             }
             setIsLoadingMore(false);
         }
     }, [partsData, displayLimit, currentPage]);

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
         status: string;
         hasPhoto?: string;
     }) => {
         setFilters(newFilters);
         setCurrentPage(1);
         setAllParts([]);
         setHasMore(true);
     }, []);

     // Обработчик изменения лимита отображения
     const handleDisplayLimitChange = useCallback((newLimit: number | undefined) => {
         setDisplayLimit(newLimit);
         setCurrentPage(1);
         setAllParts([]);
         setHasMore(true);
     }, []);

     // Функция загрузки дополнительных данных
     const loadMore = useCallback(() => {
         if (!isLoadingMore && hasMore) {
             setIsLoadingMore(true);
             setCurrentPage(prev => prev + 1);
         }
     }, [isLoadingMore, hasMore]);


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
         isLoading={isLoading || isLoadingMore}
         error={error}
         onLoadMore={displayLimit ? undefined : loadMore}
         hasMore={hasMore}
         isInfiniteScroll={displayLimit === undefined}
       />
     </div>
   );
 }

export default Inventory;