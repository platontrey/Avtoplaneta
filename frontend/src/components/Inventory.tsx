/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState, useCallback, useEffect } from 'react';
import PartsSearch from './PartsSearch';
import PartsList from './PartsList';
import { useParts } from '@/hooks/useParts';
import type { Part } from '@/lib/types';

function Inventory() {
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
                 // Если пагинация, добавляем к существующим
                 if (currentPage === 1) {
                     setAllParts(partsData);
                 } else {
                     setAllParts(prev => [...prev, ...partsData]);
                 }
                 setHasMore(partsData.length === 20);
             }
             setIsLoadingMore(false);
         }
     }, [partsData, displayLimit, currentPage]);

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
     <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4" style={{ minHeight: '100vh' }}>
       <div className="mb-6">
         <h1 className="text-3xl font-bold mb-2">Доступные запчасти</h1>
         <p className="text-gray-600 mb-8">Список всех запасных частей</p>

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