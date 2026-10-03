/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useWindowVirtualizer } from '@tanstack/react-virtual';
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

interface PartsListProps {
    parts: Part[];
    isLoading: boolean;
    error: Error | null;
    onLoadMore?: () => void;
    hasMore?: boolean;
    isInfiniteScroll?: boolean;
    defaultOpenPartId?: number | null;
}

function PartsList({ parts, isLoading, error, onLoadMore, hasMore, isInfiniteScroll = false, defaultOpenPartId = null }: PartsListProps) {
    const loadMoreRef = useRef<HTMLDivElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const listRef = useRef<HTMLDivElement>(null);

    const virtualizer = useWindowVirtualizer({
        count: parts.length,
        estimateSize: () => 140,
        overscan: 5,
        scrollMargin: listRef.current?.offsetTop ?? 0,
    });
    const [isSelectionMode, setIsSelectionMode] = useState(false);
    const [selectedParts, setSelectedParts] = useState<Set<number>>(new Set());
    const [selectedPartsCatalog, setSelectedPartsCatalog] = useState<Map<number, Part>>(new Map());
    const [isBulkEditOpen, setIsBulkEditOpen] = useState(false);
    const [isBulkDeleteOpen, setIsBulkDeleteOpen] = useState(false);
    const [isBulkOrderOpen, setIsBulkOrderOpen] = useState(false);
    const [openAccordionValue, setOpenAccordionValue] = useState<string>(
        defaultOpenPartId ? `part-${defaultOpenPartId}` : ''
    );

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
        setSelectedParts(new Set());
        setSelectedPartsCatalog(new Map());
    };

    const selectAll = () => {
        setSelectedParts(new Set(parts.map((p) => p.id)));
    };

    const deselectAll = () => {
        setSelectedParts(new Set());
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
                            <div className="flex gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={selectedParts.size === parts.length ? deselectAll : selectAll}
                                    className="text-xs sm:text-sm"
                                >
                                    {selectedParts.size === parts.length ? 'Снять все' : 'Выбрать все на странице'}
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
                    {selectedParts.size > 0 && (
                        <div className="mt-3 flex flex-col sm:flex-row gap-2 flex-wrap">
                            <Button
                                size="sm"
                                className="bg-blue-600 hover:bg-blue-700 w-full sm:w-auto text-xs sm:text-sm"
                                onClick={() => setIsBulkOrderOpen(true)}
                            >
                                <Check className="h-4 w-4 mr-1" />
                                Оформить заказ ({selectedParts.size})
                            </Button>
                            <Button
                                size="sm"
                                variant="outline"
                                className="border-orange-300 text-orange-700 hover:bg-orange-50 w-full sm:w-auto text-xs sm:text-sm"
                                onClick={() => setIsBulkEditOpen(true)}
                            >
                                <Edit className="h-4 w-4 mr-1" />
                                Изменить выбранные ({selectedParts.size})
                            </Button>
                            <Button
                                size="sm"
                                variant="outline"
                                className="border-red-300 text-red-700 hover:bg-red-50 w-full sm:w-auto text-xs sm:text-sm"
                                onClick={() => setIsBulkDeleteOpen(true)}
                            >
                                <Trash2 className="h-4 w-4 mr-1" />
                                Удалить выбранные ({selectedParts.size})
                            </Button>
                        </div>
                    )}
                </div>
            )}

            {/* Список запчастей с виртуализацией */}
            <div ref={listRef} className="w-full">
                <Accordion
                    type="single"
                    collapsible
                    value={openAccordionValue}
                    onValueChange={setOpenAccordionValue}
                    className="w-full relative"
                    style={{
                        height: `${virtualizer.getTotalSize()}px`,
                    }}
                >
                    {virtualizer.getVirtualItems().map((virtualRow) => {
                        const part = parts[virtualRow.index];
                        if (!part) return null;
                        return (
                            <div
                                key={part.id}
                                ref={virtualizer.measureElement}
                                data-index={virtualRow.index}
                                className="w-full"
                                style={{
                                    position: 'absolute',
                                    top: 0,
                                    left: 0,
                                    width: '100%',
                                    transform: `translateY(${virtualRow.start - (virtualizer.options.scrollMargin || 0)}px)`,
                                }}
                            >
                                <PartBlock
                                    part={part}
                                    isLoading={false}
                                    isSelectionMode={isSelectionMode}
                                    isSelected={selectedParts.has(part.id)}
                                    activeOrderId={partToOrderMap.get(part.id)}
                                    onLongPress={() => handleLongPress(part.id)}
                                    onSelect={(isSelected) => handlePartSelect(part.id, isSelected)}
                                />
                            </div>
                        );
                    })}
                </Accordion>
            </div>

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
                onSuccess={() => {
                    setIsBulkEditOpen(false);
                    exitSelectionMode();
                }}
            />

            <BulkDeleteDialog
                isOpen={isBulkDeleteOpen}
                onClose={() => setIsBulkDeleteOpen(false)}
                selectedPartIds={Array.from(selectedParts)}
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