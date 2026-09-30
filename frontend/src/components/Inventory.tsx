/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState, useCallback, useEffect, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { ArrowLeft, X } from 'lucide-react';
import PartsSearch from './PartsSearch';
import PartsList from './PartsList';
import { useParts, usePart, useInfiniteParts, type PartFilters } from '@/hooks/useParts';
import { usePullToRefresh } from '@/hooks/usePullToRefresh';
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/hooks/useParts';
import { Button } from './ui/button';
import type { Part } from '@/lib/types';

const hasSamePartObjects = (currentParts: Part[], nextParts: Part[]) =>
    currentParts.length === nextParts.length &&
    currentParts.every((part, index) => part === nextParts[index]);

function Inventory() {
    const containerRef = useRef<HTMLDivElement>(null);
    const queryClient = useQueryClient();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();

    const rawPartId = searchParams.get('partId');
    const parsedPartId = rawPartId ? Number(rawPartId) : NaN;
    const targetPartId = Number.isFinite(parsedPartId) && parsedPartId > 0 ? parsedPartId : null;
    const isSinglePartMode = targetPartId !== null;

    const [filters, setFilters] = useState<PartFilters>({
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

      const clearSinglePartSelection = useCallback(() => {
          setSearchParams((prev) => {
              const next = new URLSearchParams(prev);
              next.delete('partId');
              return next;
          }, { replace: true });
      }, [setSearchParams]);
  
      // Pull to refresh functionality
      const { bindPullToRefresh } = usePullToRefresh({
          onRefresh: async () => {
              if (isSinglePartMode && targetPartId) {
                  await queryClient.invalidateQueries({ queryKey: partsKeys.detail(targetPartId) });
              } else {
                  await queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
              }
          },
          threshold: 80
      });
  
       // Выбираем хук в зависимости от режима
       const isInfiniteMode = !isSinglePartMode && displayLimit === undefined;

       // Запрос конкретной запчасти при переходе по ссылке из заказов (?partId=...)
       const singlePartQuery = usePart(targetPartId);
 
       // Для бесконечной прокрутки
       const infiniteQuery = useInfiniteParts(isInfiniteMode ? filters : undefined);
 
       // Для фиксированного лимита
       const regularQuery = useParts(!isSinglePartMode && !isInfiniteMode ? {
          ...filters,
          limit: displayLimit || 20,
          page: 1
       } : undefined);
 
       // Выбираем данные в зависимости от режима
       const { data: partsData, isLoading, error, hasNextPage, fetchNextPage, isFetchingNextPage } = isSinglePartMode ? {
          data: singlePartQuery.data ? [singlePartQuery.data] : [],
          isLoading: singlePartQuery.isLoading,
          error: singlePartQuery.error,
          hasNextPage: false,
          fetchNextPage: () => {},
          isFetchingNextPage: false,
       } : isInfiniteMode ? {
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
     const handleFiltersChange = useCallback((newFilters: PartFilters) => {
         setFilters(newFilters);
         setAllParts([]);
         if (isSinglePartMode && Object.values(newFilters).some((v) => v !== undefined && v !== '')) {
             clearSinglePartSelection();
         }
     }, [isSinglePartMode, clearSinglePartSelection]);

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

         {isSinglePartMode && (
           <div className="mt-4 flex flex-wrap items-center justify-between gap-2 rounded-lg border border-primary/30 bg-primary/5 px-4 py-3 text-sm">
             <div className="font-medium">
               Просмотр запчасти <span className="font-semibold">#{targetPartId}</span>
               {singlePartQuery.data?.name ? ` — ${singlePartQuery.data.name}` : ''}
             </div>
             <div className="flex items-center gap-2">
               <Button
                 variant="outline"
                 size="sm"
                 onClick={() => navigate('/orders')}
                 className="gap-1.5"
               >
                 <ArrowLeft className="h-4 w-4" />
                 К заказам
               </Button>
               <Button
                 variant="secondary"
                 size="sm"
                 onClick={clearSinglePartSelection}
                 className="gap-1.5"
               >
                 <X className="h-4 w-4" />
                 Показать все запчасти
               </Button>
             </div>
           </div>
         )}
       </div>

       {/* Компонент списка запчастей */}
       <PartsList
         parts={allParts}
         isLoading={isLoading || isFetchingNextPage}
         error={error}
         onLoadMore={isInfiniteMode ? loadMore : undefined}
         hasMore={hasNextPage}
         isInfiniteScroll={isInfiniteMode}
         defaultOpenPartId={targetPartId}
       />
     </div>
   );
 }

export default Inventory;
