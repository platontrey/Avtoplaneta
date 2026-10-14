/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Accordion } from "@/components/ui/accordion";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Loader2, Check, X, Edit, Trash2 } from "lucide-react";
import PartBlock from './PartBlock';
import BulkEditDialog from './BulkEditDialog';
import BulkDeleteDialog from './BulkDeleteDialog';
import BulkOrderDialog from './BulkOrderDialog';
import { useSwipeGesture } from '@/hooks/useSwipeGesture';
import { useOrders } from '@/hooks/useOrders';
import type { Part } from '@/lib/types';
import type { PartFilters } from '@/hooks/useParts';

interface PartsListProps {
    parts: Part[];
    isLoading: boolean;
    error: Error | null;
    onLoadMore?: () => void;
    hasMore?: boolean;
    isInfiniteScroll?: boolean;
    defaultOpenPartId?: number | null;
    activeFilters?: PartFilters;
}

function PartsList({ parts, isLoading, error, onLoadMore, hasMore, isInfiniteScroll = false, defaultOpenPartId = null, activeFilters }: PartsListProps) {
    const loadMoreRef = useRef<HTMLDivElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [isSelectionMode, setIsSelectionMode] = useState(false);
    const [selectedParts, setSelectedParts] = useState<Set<number>>(new Set());
    const [selectedPartsCatalog, setSelectedPartsCatalog] = useState<Map<number, Part>>(new Map());
    const [isBulkEditOpen, setIsBulkEditOpen] = useState(false);
    const [isBulkDeleteOpen, setIsBulkDeleteOpen] = useState(false);
    const [isBulkOrderOpen, setIsBulkOrderOpen] = useState(false);
    const [isAllMatchingFilterSelected, setIsAllMatchingFilterSelected] = useState(false);
    const [openAccordionValue, setOpenAccordionValue] = useState<string>(
        defaultOpenPartId ? `part-${defaultOpenPartId}` : ''
    );

    const cleanedFilters = useMemo(() => {
        if (!activeFilters) return undefined;
        const result: Record<string, string> = {};
        for (const [key, value] of Object.entries(activeFilters)) {
            if (value !== undefined && value !== null && String(value).trim() !== '' && key !== 'limit' && key !== 'page') {
                result[key] = String(value).trim();
            }
        }
        return Object.keys(result).length > 0 ? result : undefined;
    }, [activeFilters]);

    const { data: activeOrders } = useOrders();

    // Карта part_id -> order.id для индикации деталей, находящихся в активных заказах
    const partToOrderMap = useMemo(() => {
        const map = new Map<number, number>();
        if (!activeOrders) return map;
        for (const order of activeOrders) {
            if (order.part_id && order.part_id > 0) {
                map.set(order.part_id, order.id);
            }
            if (order.items) {
                for (const item of order.items) {
                    if (item.part_id && item.part_id > 0) {
                        map.set(item.part_id, order.id);
                    }
                }
            }
        }
        return map;
    }, [activeOrders]);

    useEffect(() => {
        setOpenAccordionValue(defaultOpenPartId ? `part-${defaultOpenPartId}` : '');
    }, [defaultOpenPartId]);

    useEffect(() => {
        setSelectedPartsCatalog((prev) => {
            const next = new Map(prev);
            for (const p of parts) {
                if (selectedParts.has(p.id)) {
                    next.set(p.id, p);
                }
            }
            return next;
        });
    }, [parts, selectedParts]);

    const { bindSwipeEvents } = useSwipeGesture({
        onSwipeLeft: () => {
            if (onLoadMore && hasMore && !isLoading) {
                onLoadMore();
            }
        },
        onSwipeRight: () => {},
        threshold: 75,
        preventDefault: false
    });

    const handleLongPress = useCallback((partId: number) => {
        setIsSelectionMode(true);
        setSelectedParts(new Set([partId]));
    }, []);

    const handlePartSelect = (partId: number, isSelected: boolean) => {
        setSelectedParts((prev) => {
            const newSet = new Set(prev);
            if (isSelected) {
                newSet.add(partId);
            } else {
                newSet.delete(partId);
            }
            return newSet;
        });
    };

    const exitSelectionMode = () => {
        setIsSelectionMode(false);
        setIsAllMatchingFilterSelected(false);
        setSelectedParts(new Set());
        setSelectedPartsCatalog(new Map());
    };

    const selectAll = () => {
        setSelectedParts(new Set(parts.map((p) => p.id)));
    };

    const deselectAll = () => {
        setSelectedParts(new Set());
        setIsAllMatchingFilterSelected(false);
    };

    useEffect(() => {
        if (!isInfiniteScroll || !onLoadMore || !hasMore || isLoading) return;

        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0]?.isIntersecting) {
                    onLoadMore();
                }
            },
            { rootMargin: '400px' }
        );

        const currentTarget = loadMoreRef.current;
        if (currentTarget) {
            observer.observe(currentTarget);
        }

        return () => {
            if (currentTarget) {
                observer.unobserve(currentTarget);
            }
        };
    }, [isInfiniteScroll, onLoadMore, hasMore, isLoading]);

    useEffect(() => {
        return bindSwipeEvents(containerRef.current);
    }, [bindSwipeEvents]);

    if (error) {
        return (
            <div className="text-center py-12">
                <p className="text-red-500 text-lg">Ошибка загрузки: {error.message}</p>
            </div>
        );
    }

    if (isLoading && parts.length === 0) {
        return (
            <div className="space-y-4">
                {Array.from({ length: 6 }).map((_, index) => (
                    <div key={index} className="border rounded-lg mb-4 p-4 relative group hover:shadow-md transition-shadow bg-card">
                        <div className="flex items-center space-x-2 sm:space-x-3 flex-1">
                            <Skeleton className="w-12 h-12 sm:w-16 sm:h-16 rounded-lg" />
                            <div className="flex flex-col min-w-0 flex-1">
                                <Skeleton className="h-4 w-32 sm:w-48 mb-1" />
                                <Skeleton className="h-3 w-20 sm:w-32 mb-2" />
                                <div className="flex gap-2 mt-1">
                                    <Skeleton className="h-5 w-16 rounded-md" />
                                    <Skeleton className="h-5 w-12 rounded-md" />
                                    <Skeleton className="h-5 w-20 rounded-md" />
                                </div>
                            </div>
                            <div className="flex space-x-1 shrink-0">
                                <Skeleton className="h-8 w-8 rounded" />
                                <Skeleton className="h-8 w-8 rounded" />
                                <Skeleton className="h-8 w-8 rounded" />
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        );
    }

    if (parts.length === 0) {
        return (
            <div className="text-center py-12">
                <p className="text-gray-500 text-lg">Запчасти не найдены</p>
            </div>
        );
    }

    const selectedPartObjects = Array.from(selectedParts)
        .map((id) => selectedPartsCatalog.get(id) || parts.find((p) => p.id === id))
        .filter((p): p is Part => Boolean(p));

    const hasOutOfStockSelected = selectedPartObjects.some((p) => (p.quantity ?? 0) <= 0);

    return (
        <div ref={containerRef} className="space-y-4">
            {/* Панель действий для режима выбора */}
            {isSelectionMode && (
                <div className="bg-blue-50 dark:bg-blue-950/40 border border-blue-200 dark:border-blue-800 rounded-lg p-3 sm:p-4 mb-4 sticky top-2 z-20 shadow-sm">
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                        <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
                            <span className="font-medium text-blue-900 dark:text-blue-200 text-sm sm:text-base">
                                В корзине выбора: {selectedParts.size} запч.
                            </span>
                            <div className="flex gap-2 flex-wrap items-center">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={selectedParts.size === parts.length && !isAllMatchingFilterSelected ? deselectAll : selectAll}
                                    className="text-xs sm:text-sm"
                                >
                                    {selectedParts.size === parts.length && !isAllMatchingFilterSelected ? 'Снять все' : 'Выбрать на странице'}
                                </Button>
                                <Button
                                    variant={isAllMatchingFilterSelected ? "default" : "outline"}
                                    size="sm"
                                    onClick={() => {
                                        if (isAllMatchingFilterSelected) {
                                            setIsAllMatchingFilterSelected(false);
                                        } else {
                                            selectAll();
                                            setIsAllMatchingFilterSelected(true);
                                        }
                                    }}
                                    className="text-xs sm:text-sm"
                                >
                                    {isAllMatchingFilterSelected ? 'Снять выбор "Все по запросу"' : 'Выбрать все запчасти в поиске'}
                                </Button>
                            </div>
                        </div>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={exitSelectionMode}
                            className="text-blue-600 dark:text-blue-300 hover:text-blue-800 self-start sm:self-auto"
                        >
                            <X className="h-4 w-4 mr-1" />
                            Отмена
                        </Button>
                    </div>

                    {isAllMatchingFilterSelected && (
                        <div className="mt-2.5 py-2 px-3 bg-blue-100/90 dark:bg-blue-900/40 border border-blue-200 dark:border-blue-800 rounded-md text-xs text-blue-900 dark:text-blue-100 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1.5">
                            <span>
                                🎯 <strong>Выбраны ВСЕ запчасти по текущему запросу.</strong> Сервер применит изменения ко всем подходящим деталям без необходимости их загрузки на устройство.
                            </span>
                            <button
                                type="button"
                                className="text-xs text-blue-700 dark:text-blue-300 font-semibold underline hover:text-blue-900 dark:hover:text-white cursor-pointer self-start sm:self-auto"
                                onClick={() => setIsAllMatchingFilterSelected(false)}
                            >
                                Оставить только запчасти на странице
                            </button>
                        </div>
                    )}

                    {(selectedParts.size > 0 || isAllMatchingFilterSelected) && (
                        <div className="mt-3 flex flex-col sm:flex-row gap-2 flex-wrap">
                            {!isAllMatchingFilterSelected && (
                                <Button
                                    size="sm"
                                    className="bg-blue-600 hover:bg-blue-700 w-full sm:w-auto text-xs sm:text-sm"
                                    onClick={() => setIsBulkOrderOpen(true)}
                                    title={hasOutOfStockSelected ? "Внимание: среди выбранных есть детали с количеством <= 0" : undefined}
                                >
                                    <Check className="h-4 w-4 mr-1" />
                                    Оформить заказ ({selectedParts.size})
                                    {hasOutOfStockSelected && <span className="ml-1 text-xs">⚠️</span>}
                                </Button>
                            )}
                            <Button
                                size="sm"
                                variant="outline"
                                className="border-orange-300 text-orange-700 hover:bg-orange-50 w-full sm:w-auto text-xs sm:text-sm"
                                onClick={() => setIsBulkEditOpen(true)}
                            >
                                <Edit className="h-4 w-4 mr-1" />
                                {isAllMatchingFilterSelected ? 'Изменить все по запросу' : `Изменить выбранные (${selectedParts.size})`}
                            </Button>
                            <Button
                                size="sm"
                                variant="outline"
                                className="border-red-300 text-red-700 hover:bg-red-50 w-full sm:w-auto text-xs sm:text-sm"
                                onClick={() => setIsBulkDeleteOpen(true)}
                            >
                                <Trash2 className="h-4 w-4 mr-1" />
                                {isAllMatchingFilterSelected ? 'Удалить все по запросу' : `Удалить выбранные (${selectedParts.size})`}
                            </Button>
                        </div>
                    )}
                </div>
            )}

            {/* Кнопка входа в режим выбора */}
            {!isSelectionMode && parts.length > 0 && (
                <div className="flex justify-end items-center gap-2 mb-1">
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setIsSelectionMode(true)}
                        className="text-xs text-muted-foreground hover:text-foreground h-8"
                    >
                        <Check className="h-3.5 w-3.5 mr-1.5" />
                        Выбрать запчасти
                    </Button>
                </div>
            )}
            {/* Список запчастей */}
            <Accordion
                type="single"
                collapsible
                value={openAccordionValue}
                onValueChange={setOpenAccordionValue}
                className="w-full"
            >
                {parts.map((part) => (
                    <PartBlock
                        key={part.id}
                        part={part}
                        isLoading={false}
                        isSelectionMode={isSelectionMode}
                        isSelected={selectedParts.has(part.id)}
                        activeOrderId={partToOrderMap.get(part.id)}
                        onLongPress={() => handleLongPress(part.id)}
                        onSelect={(isSelected) => handlePartSelect(part.id, isSelected)}
                    />
                ))}
            </Accordion>

            {/* Триггер бесконечной прокрутки */}
            {isInfiniteScroll && (
                <div ref={loadMoreRef} className="flex justify-center py-4">
                    {isLoading ? (
                        <div className="flex items-center gap-2">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            <span>Загрузка...</span>
                        </div>
                    ) : hasMore ? (
                        <div className="h-4 w-full text-center text-xs text-transparent select-none">Load More Trigger</div>
                    ) : null}
                </div>
            )}

            {/* Кнопка ручной загрузки */}
            {!isInfiniteScroll && onLoadMore && hasMore && (
                <div className="flex justify-center py-4">
                    <Button
                        onClick={onLoadMore}
                        disabled={isLoading}
                        variant="outline"
                        className="gap-2"
                    >
                        {isLoading ? (
                            <>
                                <Loader2 className="h-4 w-4 animate-spin" />
                                Загрузка...
                            </>
                        ) : (
                            'Загрузить еще'
                        )}
                    </Button>
                </div>
            )}

            {/* Диалоги массовых операций */}
            <BulkEditDialog
                isOpen={isBulkEditOpen}
                onClose={() => setIsBulkEditOpen(false)}
                selectedPartIds={Array.from(selectedParts)}
                selectedParts={selectedPartObjects}
                activeFilters={cleanedFilters}
                initialScopeType={isAllMatchingFilterSelected ? 'filter' : 'ids'}
                onSuccess={() => {
                    setIsBulkEditOpen(false);
                    exitSelectionMode();
                }}
            />

            <BulkDeleteDialog
                isOpen={isBulkDeleteOpen}
                onClose={() => setIsBulkDeleteOpen(false)}
                selectedPartIds={Array.from(selectedParts)}
                scopeType={isAllMatchingFilterSelected ? 'filter' : 'ids'}
                activeFilters={cleanedFilters}
                onSuccess={() => {
                    setIsBulkDeleteOpen(false);
                    exitSelectionMode();
                }}
            />

            <BulkOrderDialog
                isOpen={isBulkOrderOpen}
                onClose={() => setIsBulkOrderOpen(false)}
                selectedParts={selectedPartObjects}
                onSuccess={() => {
                    setIsBulkOrderOpen(false);
                    exitSelectionMode();
                }}
            />
        </div>
    );
}

export default PartsList;