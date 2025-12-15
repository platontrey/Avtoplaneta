/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import {useCallback, useEffect, useRef, useState} from "react";
import { Accordion } from "@/components/ui/accordion";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Loader2, Check, X, Edit, Trash2 } from "lucide-react";
import PartBlock from './PartBlock';
import BulkEditDialog from './BulkEditDialog';
import BulkDeleteDialog from './BulkDeleteDialog';
import BulkOrderDialog from './BulkOrderDialog';
import type { Part } from '@/lib/types';

interface PartsListProps {
    parts: Part[];
    isLoading: boolean;
    error: Error | null;
    onLoadMore?: () => void;
    hasMore?: boolean;
    isInfiniteScroll?: boolean;
}

function PartsList({ parts, isLoading, error, onLoadMore, hasMore, isInfiniteScroll = false }: PartsListProps) {
    const loadMoreRef = useRef<HTMLDivElement>(null);
    const [isSelectionMode, setIsSelectionMode] = useState(false);
    const [selectedParts, setSelectedParts] = useState<Set<number>>(new Set());
    const [isBulkEditOpen, setIsBulkEditOpen] = useState(false);
    const [isBulkDeleteOpen, setIsBulkDeleteOpen] = useState(false);
    const [isBulkOrderOpen, setIsBulkOrderOpen] = useState(false);

    // Обработчик долгого нажатия для активации режима выбора
    const handleLongPress = useCallback((partId: number) => {
        setIsSelectionMode(true);
        setSelectedParts(new Set([partId]));
    }, []);

    // Обработчик выбора/снятия выбора запчасти
    const handlePartSelect = (partId: number, isSelected: boolean) => {
        setSelectedParts(prev => {
            const newSet = new Set(prev);
            if (isSelected) {
                newSet.add(partId);
            } else {
                newSet.delete(partId);
            }
            return newSet;
        });
    };

    // Выход из режима выбора
    const exitSelectionMode = () => {
        setIsSelectionMode(false);
        setSelectedParts(new Set());
    };

    // Выбрать все
    const selectAll = () => {
        setSelectedParts(new Set(parts.map(p => p.id)));
    };

    // Снять выбор со всех
    const deselectAll = () => {
        setSelectedParts(new Set());
    };

    // Intersection Observer для бесконечной прокрутки
    useEffect(() => {
        if (!isInfiniteScroll || !onLoadMore || !hasMore) return;

        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting && !isLoading) {
                    onLoadMore();
                }
            },
            { threshold: 0.1 }
        );

        if (loadMoreRef.current) {
            observer.observe(loadMoreRef.current);
        }

        return () => observer.disconnect();
    }, [isInfiniteScroll, onLoadMore, hasMore, isLoading]);

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

    return (
        <div className="space-y-4">
            {/* Панель действий для режима выбора */}
            {isSelectionMode && (
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 sm:p-4 mb-4">
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                        <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
              <span className="font-medium text-blue-900 text-sm sm:text-base">
                Выбрано: {selectedParts.size} из {parts.length}
              </span>
                            <div className="flex gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={selectedParts.size === parts.length ? deselectAll : selectAll}
                                    className="text-xs sm:text-sm"
                                >
                                    {selectedParts.size === parts.length ? 'Снять все' : 'Выбрать все'}
                                </Button>
                            </div>
                        </div>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={exitSelectionMode}
                            className="text-blue-600 hover:text-blue-800 self-start sm:self-auto"
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
                                Заказать выбранные ({selectedParts.size})
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

            {/* Список запчастей */}
            <Accordion type="single" collapsible className="w-full">
                {parts.map((part) => (
                    <PartBlock
                        key={part.id}
                        part={part}
                        isLoading={false}
                        isSelectionMode={isSelectionMode}
                        isSelected={selectedParts.has(part.id)}
                        onLongPress={() => handleLongPress(part.id)}
                        onSelect={(isSelected) => handlePartSelect(part.id, isSelected)}
                    />
                ))}
            </Accordion>

            {/* Триггер бесконечной прокрутки */}
            {isInfiniteScroll && hasMore && (
                <div ref={loadMoreRef} className="flex justify-center py-4">
                    {isLoading ? (
                        <div className="flex items-center gap-2">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            <span>Загрузка...</span>
                        </div>
                    ) : (
                        <div className="h-4" /> // Невидимый элемент триггера
                    )}
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
                selectedParts={parts.filter(part => selectedParts.has(part.id))}
                onSuccess={() => {
                    setIsBulkOrderOpen(false);
                    exitSelectionMode();
                }}
            />
        </div>
    );
}

export default PartsList;