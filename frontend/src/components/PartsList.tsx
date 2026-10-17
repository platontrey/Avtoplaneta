/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
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
import PartOrderDialog from './PartOrderDialog';
import EditPartDialog from './EditPartDialog';
import PartPhotoViewer from './PartPhotoViewer';
import ImageEditor from './ImageEditor';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useSwipeGesture } from '@/hooks/useSwipeGesture';
import { useOrders } from '@/hooks/useOrders';
import { useDeletePart, useDuplicatePart, partsKeys } from '@/hooks/useParts';
import { usePartEdit } from '@/hooks/usePartEdit';
import { partsApi } from '@/features/parts/api/partsApi';
import { useQueryClient } from '@tanstack/react-query';
import { getAuthHeaders } from '@/lib/csrf';
import { API_BASE_URL } from '@/lib/api';
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

interface EditPartModalProps {
    part: Part;
    onClose: () => void;
    onCrop: (src: string | File | undefined, photoPathToReplace?: string) => void;
}

function EditPartModal({ part, onClose, onCrop }: EditPartModalProps) {
    const queryClient = useQueryClient();
    const partEdit = usePartEdit({ initialPart: part });
    const [originalFile, setOriginalFile] = useState<File | null>(null);

    useEffect(() => {
        partEdit.setIsEditing(true);
    }, []);

    const handleOpenChange = (open: boolean) => {
        partEdit.setIsEditing(open);
        if (!open) {
            onClose();
        }
    };

    const handlePhotoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            setOriginalFile(file);
            partEdit.photoUpload.handlePhotoChange(e);
        }
    };

    const handleDeletePhoto = async (photoPath?: string) => {
        if (partEdit.photoUpload.photoFile) {
            partEdit.photoUpload.resetPhoto();
            partEdit.updateFormField('photo', '');
            setOriginalFile(null);
            return;
        }

        const targetPhoto = photoPath || (part.photos && part.photos.length > 0 ? part.photos[0] : '');
        if (!targetPhoto) return;

        if (!confirm('Вы уверены, что хотите удалить это фото?')) return;

        try {
            const deleteUrl = `${API_BASE_URL}/api/v1/parts/${part.id}/photo?photo_url=${encodeURIComponent(targetPhoto)}`;
            const response = await fetch(deleteUrl, {
                method: 'DELETE',
                headers: getAuthHeaders(),
                credentials: 'include',
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            partEdit.updateFormField('photo', '');
            void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
        } catch (error) {
            console.error('Ошибка при удалении фото:', error);
            alert('Не удалось удалить фото. Попробуйте еще раз.');
        }
    };

    return (
        <EditPartDialog
            partEdit={{
                ...partEdit,
                setIsEditing: handleOpenChange,
            }}
            part={part}
            onPhotoChange={handlePhotoChange}
            onCrop={onCrop}
            onDeletePhoto={handleDeletePhoto}
            originalFile={originalFile}
        />
    );
}

function PartsList({ parts, isLoading, error, onLoadMore, hasMore, isInfiniteScroll = false, defaultOpenPartId = null, activeFilters }: PartsListProps) {
    const loadMoreRef = useRef<HTMLDivElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const queryClient = useQueryClient();
    const deletePartMutation = useDeletePart();
    const duplicatePartMutation = useDuplicatePart();

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

    // Синглтон-состояния модалок
    const [editingPart, setEditingPart] = useState<Part | null>(null);
    const [orderingPart, setOrderingPart] = useState<Part | null>(null);
    const [deletingPart, setDeletingPart] = useState<Part | null>(null);
    const [photoViewerState, setPhotoViewerState] = useState<{
        photos: string[];
        index: number;
        name: string;
        subtitle?: string;
        timestamp?: number;
    } | null>(null);
    const [cropperState, setCropperState] = useState<{
        src: string | File;
        part: Part;
        photoToReplace?: string | null;
    } | null>(null);

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
        if (isAllMatchingFilterSelected) {
            setIsAllMatchingFilterSelected(false);
            const newSet = new Set(parts.map((p) => p.id));
            if (!isSelected) {
                newSet.delete(partId);
            } else {
                newSet.add(partId);
            }
            setSelectedParts(newSet);
            return;
        }

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

    // При активном режиме "все по запросу" автоматически добавляем подгруженные скроллом запчасти в selectedParts
    useEffect(() => {
        if (isAllMatchingFilterSelected) {
            setSelectedParts(new Set(parts.map((p) => p.id)));
        }
    }, [isAllMatchingFilterSelected, parts]);

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

    const handleDuplicatePart = async (part: Part) => {
        try {
            const duplicated = await duplicatePartMutation.mutateAsync(part.id);
            alert(`Запчасть успешно продублирована!\nСоздан дубликат без фото с ID: ${duplicated.id}.\nЕсли включен фильтр "Только с фото", переключите его на "Все", чтобы увидеть новую деталь.`);
        } catch (err) {
            console.error('Ошибка дублирования детали:', err);
            alert('Не удалось дублировать деталь');
            throw err;
        }
    };

    const handleAddPhoto = async (part: Part, file: File) => {
        if (!file.type.startsWith('image/')) {
            alert('Пожалуйста, выберите файл изображения');
            return;
        }
        if (file.size > 30 * 1024 * 1024) {
            alert('Размер файла должен быть менее 30МБ.');
            return;
        }
        try {
            await partsApi.uploadPhoto(part.id, file);
            void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
        } catch (error) {
            console.error('Ошибка при загрузке фото:', error);
            alert('Не удалось загрузить фото. Попробуйте еще раз.');
        }
    };

    const handleDeletePhoto = async (part: Part, photoPath: string) => {
        if (!confirm('Вы уверены, что хотите удалить это фото?')) return;
        try {
            const response = await fetch(`${API_BASE_URL}/api/v1/parts/${part.id}/photo?photo_url=${encodeURIComponent(photoPath)}`, {
                method: 'DELETE',
                headers: getAuthHeaders(),
                credentials: 'include',
            });
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
        } catch (error) {
            console.error('Ошибка удаления фото:', error);
            alert('Не удалось удалить фото.');
        }
    };

    // Двойной механизм бесконечной подгрузки: Scroll Listener с упреждением + IntersectionObserver
    useEffect(() => {
        if (!isInfiniteScroll || !onLoadMore || !hasMore) return;

        let ticking = false;

        const checkAndLoad = () => {
            if (isLoading || !hasMore) return;

            const scrollHeight = document.documentElement.scrollHeight;
            const scrollTop = window.scrollY || document.documentElement.scrollTop;
            const clientHeight = window.innerHeight;

            // Срабатывает с упреждением 600px до конца страницы для бесшовной подгрузки
            if (scrollTop + clientHeight >= scrollHeight - 600) {
                onLoadMore();
            }
        };

        const handleScroll = () => {
            if (!ticking) {
                window.requestAnimationFrame(() => {
                    checkAndLoad();
                    ticking = false;
                });
                ticking = true;
            }
        };

        window.addEventListener('scroll', handleScroll, { passive: true });
        window.addEventListener('resize', handleScroll, { passive: true });

        // IntersectionObserver в качестве второго параллельного триггера
        let observer: IntersectionObserver | null = null;
        const currentTarget = loadMoreRef.current;

        if (typeof IntersectionObserver !== 'undefined' && currentTarget) {
            observer = new IntersectionObserver(
                (entries) => {
                    if (entries[0]?.isIntersecting && !isLoading && hasMore) {
                        onLoadMore();
                    }
                },
                { rootMargin: '600px' }
            );
            observer.observe(currentTarget);
        }

        // Проверяем сразу (например, если контента мало для появления полосы прокрутки)
        checkAndLoad();

        return () => {
            window.removeEventListener('scroll', handleScroll);
            window.removeEventListener('resize', handleScroll);
            if (observer && currentTarget) {
                observer.unobserve(currentTarget);
                observer.disconnect();
            }
        };
    }, [isInfiniteScroll, onLoadMore, hasMore, isLoading, parts.length]);

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
                                {isAllMatchingFilterSelected ? '🎯 Выбраны ВСЕ детали по запросу' : `В корзине выбора: ${selectedParts.size} запч.`}
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
                        isSelected={isAllMatchingFilterSelected || selectedParts.has(part.id)}
                        activeOrderId={partToOrderMap.get(part.id)}
                        onLongPress={() => handleLongPress(part.id)}
                        onSelect={(isSelected) => handlePartSelect(part.id, isSelected)}
                        onOpenPhotoViewer={(p, index) => {
                            setPhotoViewerState({
                                photos: p.photos || [],
                                index,
                                name: p.name,
                                subtitle: p.brand && p.model ? `${p.brand} ${p.model}` : p.brand || p.model,
                                timestamp: Date.now(),
                            });
                        }}
                        onOpenOrderDialog={(p) => setOrderingPart(p)}
                        onOpenEditDialog={(p) => setEditingPart(p)}
                        onOpenDeleteDialog={(p) => setDeletingPart(p)}
                        onOpenCropper={(p, src, photoPathToReplace) => {
                            setCropperState({ src, part: p, photoToReplace: photoPathToReplace || null });
                        }}
                        onAddPhoto={handleAddPhoto}
                        onDeletePhoto={handleDeletePhoto}
                        onDuplicatePart={handleDuplicatePart}
                    />
                ))}
            </Accordion>

            {/* Триггер бесконечной прокрутки */}
            {isInfiniteScroll && hasMore && (
                <div ref={loadMoreRef} className="flex justify-center items-center py-6 min-h-[48px]">
                    {isLoading ? (
                        <div className="flex items-center gap-2 text-muted-foreground text-sm">
                            <Loader2 className="h-4 w-4 animate-spin text-primary" />
                            <span>Загрузка запчастей...</span>
                        </div>
                    ) : (
                        <div className="h-6 w-full text-center text-xs text-transparent select-none">
                            Load More Trigger
                        </div>
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

            {/* Диалог редактирования детали (монтируется только для активной детали) */}
            {editingPart && (
                <EditPartModal
                    part={editingPart}
                    onClose={() => setEditingPart(null)}
                    onCrop={(src, photoPathToReplace) => {
                        if (src) {
                            setCropperState({ src, part: editingPart, photoToReplace: photoPathToReplace || null });
                        }
                    }}
                />
            )}

            {/* Диалог заказа */}
            {orderingPart && (
                <PartOrderDialog
                    part={orderingPart}
                    isOpen={Boolean(orderingPart)}
                    onOpenChange={(open) => {
                        if (!open) setOrderingPart(null);
                    }}
                />
            )}

            {/* Редактор / обрезка фото */}
            {cropperState && (
                <ImageEditor
                    src={cropperState.src}
                    aspect={null}
                    onEditComplete={async (croppedImageBlob: Blob) => {
                        if (croppedImageBlob.size === 0) return;
                        const croppedFile = new File([croppedImageBlob], 'cropped-image.jpg', { type: 'image/jpeg' });
                        try {
                            await partsApi.uploadPhoto(cropperState.part.id, croppedFile);
                            if (cropperState.photoToReplace) {
                                try {
                                    await fetch(`${API_BASE_URL}/api/v1/parts/${cropperState.part.id}/photo?photo_url=${encodeURIComponent(cropperState.photoToReplace)}`, {
                                        method: 'DELETE',
                                        headers: getAuthHeaders(),
                                        credentials: 'include',
                                    });
                                } catch (delErr) {
                                    console.warn('Не удалось удалить старое фото:', delErr);
                                }
                            }
                            void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
                        } catch (err) {
                            console.error('Ошибка сохранения фото:', err);
                            alert('Не удалось сохранить отредактированное фото');
                        } finally {
                            setCropperState(null);
                        }
                    }}
                    onCancel={() => setCropperState(null)}
                />
            )}

            {/* Просмотрщик фото */}
            {photoViewerState && (
                <PartPhotoViewer
                    isOpen={Boolean(photoViewerState)}
                    onClose={() => setPhotoViewerState(null)}
                    photos={photoViewerState.photos}
                    initialIndex={photoViewerState.index}
                    partName={photoViewerState.name}
                    partSubtitle={photoViewerState.subtitle}
                    uploadTimestamp={photoViewerState.timestamp}
                />
            )}

            {/* Подтверждение удаления детали */}
            {deletingPart && (
                <AlertDialog open={Boolean(deletingPart)} onOpenChange={(open) => !open && setDeletingPart(null)}>
                    <AlertDialogContent>
                        <AlertDialogHeader>
                            <AlertDialogTitle>Вы абсолютно уверены?</AlertDialogTitle>
                            <AlertDialogDescription>
                                Это действие нельзя отменить. Деталь «{deletingPart.name}» (ID: {deletingPart.id}) будет удалена из базы склада.
                            </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                            <AlertDialogCancel onClick={() => setDeletingPart(null)}>Отмена</AlertDialogCancel>
                            <AlertDialogAction
                                className="bg-destructive hover:bg-destructive/90 text-destructive-foreground"
                                onClick={async () => {
                                    try {
                                        await deletePartMutation.mutateAsync(deletingPart.id);
                                    } catch (err) {
                                        console.error('Ошибка удаления:', err);
                                        alert('Не удалось удалить деталь');
                                    } finally {
                                        setDeletingPart(null);
                                    }
                                }}
                            >
                                Удалить
                            </AlertDialogAction>
                        </AlertDialogFooter>
                    </AlertDialogContent>
                </AlertDialog>
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