/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import React, { useState, useRef, useEffect } from "react";
import { Link } from "react-router-dom";
import { AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Trash2, Edit, ShoppingCart, Plus, X, Crop, Copy, CopyPlus, Check } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { motion } from "framer-motion";
import { useAuth } from "@/hooks/useAuth";
import type { Part } from "@/features/parts/types";
import { API_BASE_URL } from "@/lib/api";
import { copyPhotoToClipboard } from "@/lib/photoUtils";
import { copyToClipboard, formatPartCopyText } from "@/lib/clipboard";

export interface PartBlockProps {
    part: Part;
    isLoading?: boolean;
    isSelectionMode?: boolean;
    isSelected?: boolean;
    activeOrderId?: number;
    onLongPress?: () => void;
    onSelect?: (isSelected: boolean) => void;
    onOpenPhotoViewer?: (part: Part, index: number) => void;
    onOpenOrderDialog?: (part: Part) => void;
    onOpenEditDialog?: (part: Part) => void;
    onOpenDeleteDialog?: (part: Part) => void;
    onOpenCropper?: (part: Part, src: string, photoPathToReplace?: string) => void;
    onAddPhoto?: (part: Part, file: File) => Promise<void>;
    onDeletePhoto?: (part: Part, photoPath: string) => Promise<void>;
    onDuplicatePart?: (part: Part) => Promise<void>;
}

function PartBlock({
    part,
    isLoading = false,
    isSelectionMode = false,
    isSelected = false,
    activeOrderId,
    onLongPress,
    onSelect,
    onOpenPhotoViewer,
    onOpenOrderDialog,
    onOpenEditDialog,
    onOpenDeleteDialog,
    onOpenCropper,
    onAddPhoto,
    onDeletePhoto,
    onDuplicatePart,
}: PartBlockProps) {
    const { user } = useAuth();
    const accordionRef = useRef<HTMLDivElement>(null);
    const longPressTimerRef = useRef<NodeJS.Timeout | null>(null);
    const [, setIsPressed] = useState(false);

    const [copiedPhotoPath, setCopiedPhotoPath] = useState<string | null>(null);
    const [isPartCopied, setIsPartCopied] = useState(false);
    const [isPartDuplicated, setIsPartDuplicated] = useState(false);
    const [isDuplicating, setIsDuplicating] = useState(false);

    const handleCopyPhoto = async (photoPath: string, e: React.MouseEvent) => {
        e.stopPropagation();
        const photoUrl = photoPath.startsWith('http') || photoPath.startsWith('data:')
            ? photoPath
            : `${API_BASE_URL}${photoPath.startsWith('/') ? photoPath : `/${photoPath}`}`;
        try {
            await copyPhotoToClipboard(photoUrl);
            setCopiedPhotoPath(photoPath);
            setTimeout(() => setCopiedPhotoPath(null), 2000);
        } catch (err) {
            console.error('Ошибка копирования фото:', err);
        }
    };

    const handleCopyPartInfo = async (e: React.MouseEvent) => {
        e.stopPropagation();
        const text = formatPartCopyText(part);
        const success = await copyToClipboard(text);
        if (success) {
            setIsPartCopied(true);
            setTimeout(() => setIsPartCopied(false), 2000);
        }
    };

    const handleDuplicateClick = async (e: React.MouseEvent) => {
        e.stopPropagation();
        if (isDuplicating || !onDuplicatePart) return;
        setIsDuplicating(true);
        try {
            await onDuplicatePart(part);
            setIsPartDuplicated(true);
            setTimeout(() => setIsPartDuplicated(false), 2000);
        } catch (err) {
            console.error('Ошибка дублирования:', err);
        } finally {
            setIsDuplicating(false);
        }
    };

    const handlePointerDown = (e: React.PointerEvent) => {
        if (e.pointerType === 'mouse') {
            return;
        }
        const target = e.target as HTMLElement;
        if (target.closest('input, a, [role="button"], button')) {
            return;
        }

        setIsPressed(true);
        longPressTimerRef.current = setTimeout(() => {
            if (onLongPress) {
                onLongPress();
            }
        }, 300);
    };

    const handlePointerUp = () => {
        setIsPressed(false);
        if (longPressTimerRef.current) {
            clearTimeout(longPressTimerRef.current);
            longPressTimerRef.current = null;
        }
    };

    const handlePointerLeave = () => {
        setIsPressed(false);
        if (longPressTimerRef.current) {
            clearTimeout(longPressTimerRef.current);
            longPressTimerRef.current = null;
        }
    };

    useEffect(() => {
        return () => {
            if (longPressTimerRef.current) {
                clearTimeout(longPressTimerRef.current);
            }
        };
    }, []);

    if (!part) {
        return null;
    }

    if (isLoading) {
        return (
            <div className="border rounded-lg mb-2 p-4">
                <div className="flex items-center space-x-4">
                    <Skeleton className="h-4 w-48" />
                    <Skeleton className="h-8 w-8 rounded" />
                </div>
                <div className="mt-4 space-y-2">
                    {[32, 40, 36, 28, 30, 34, 26].map((width, index) => (
                        <Skeleton key={index} className={`h-4 w-${width}`} />
                    ))}
                </div>
            </div>
        );
    }

    const mainPhotoUrl = (part.photos && part.photos.length > 0)
        ? `${API_BASE_URL}${part.photos[0]}${part.updated_at ? `?v=${encodeURIComponent(part.updated_at)}` : ''}`
        : '/placeholder-part.svg';

    return (
        <motion.div
            className="bg-card border border-border rounded-lg mb-4 relative group hover:shadow-md transition-shadow p-4"
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.2, ease: "easeOut" }}
            onClick={(e) => e.stopPropagation()}
            onPointerDown={handlePointerDown}
            onPointerUp={handlePointerUp}
            onPointerLeave={handlePointerLeave}
        >
            {/* Чекбокс для режима выбора */}
            {isSelectionMode && (
                <div className="absolute top-2 left-2 sm:top-11 sm:left-[-3rem] z-10" onClick={(e) => e.stopPropagation()}>
                    <Checkbox
                        className="scale-125 sm:scale-150"
                        checked={isSelected}
                        onCheckedChange={(checked) => {
                            if (onSelect) {
                                onSelect(checked as boolean);
                            }
                        }}
                    />
                </div>
            )}

            {/* Элемент аккордеона */}
            <AccordionItem
                value={`part-${part.id}`}
                data-testid="part-accordion"
                ref={accordionRef}
                className="border-b last:border-b-0"
            >
                <AccordionTrigger
                    className="flex items-center justify-between w-full hover:bg-accent hover:text-accent-foreground pr-4 pointer-events-auto"
                    onClick={(e) => {
                        const selection = window.getSelection()?.toString();
                        if (selection && selection.trim().length > 0) {
                            e.preventDefault();
                            e.stopPropagation();
                        }
                    }}
                >
                    <div className="flex items-center space-x-2 sm:space-x-3 flex-1 min-w-0">
                        <img
                            src={mainPhotoUrl}
                            alt={part.name || 'Изображение детали'}
                            onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                e.currentTarget.onerror = null;
                                e.currentTarget.src = '/placeholder-part.svg';
                            }}
                            loading="lazy"
                            decoding="async"
                            className="w-12 h-12 sm:w-16 sm:h-16 object-cover rounded-lg border cursor-pointer ml-3 mr-3 hover:ring-2 hover:ring-primary/50 transition-all shrink-0"
                            onClick={(e) => {
                                e.stopPropagation();
                                onOpenPhotoViewer?.(part, 0);
                            }}
                            title="Нажмите для увеличения фото"
                        />
                        <div
                            className="flex flex-col min-w-0 flex-1 select-text cursor-text"
                            onPointerDown={(e) => e.stopPropagation()}
                        >
                            {/* Строка 1: Название детали */}
                            <span className="truncate text-sm sm:text-base font-semibold select-text cursor-text">
                                {part.name || 'Unnamed Part'}
                            </span>

                            {/* Строка 2: Бренд и модель */}
                            {(part.brand || part.model) && (
                                <span className="text-xs sm:text-sm text-muted-foreground truncate select-text cursor-text">
                                    {part.brand && part.model ? `${part.brand} ${part.model}` : part.brand || part.model}
                                </span>
                            )}

                            {/* Строка 3: Код производителя, категория, количество */}
                            <div className="flex flex-wrap items-center gap-2 mt-1 select-text cursor-text">
                                {part.manufacturer_code && (
                                    <span
                                        className="text-sm bg-secondary text-secondary-foreground border border-border px-3 py-1 rounded-md font-medium select-text cursor-text"
                                        title={`Код производителя: ${part.manufacturer_code}`}
                                    >
                                        Код: {part.manufacturer_code}
                                    </span>
                                )}
                                {part.category && (
                                    <span className="text-sm bg-secondary text-secondary-foreground border border-border px-3 py-1 rounded-md font-medium select-text cursor-text">
                                        {part.category}
                                    </span>
                                )}
                                <span className="text-sm bg-secondary text-secondary-foreground border border-border px-3 py-1 rounded-md font-medium select-text cursor-text">
                                    Кол: {part.quantity ?? 0}
                                </span>
                            </div>

                            {/* Строка 4: Цена, номер детали и активный заказ */}
                            <div className="flex flex-wrap items-center gap-2 mt-1 select-text cursor-text">
                                <span className="text-sm bg-primary text-primary-foreground dark:bg-emerald-500/20 dark:text-emerald-300 dark:border-emerald-500/30 border border-primary/20 px-3 py-1 rounded-md font-semibold select-text cursor-text">
                                    Цена: {part.price && Number(part.price) > 0 ? `₽${part.price}` : 'отсутствует'}
                                </span>
                                {part.number && (
                                    <span
                                        className="text-sm bg-secondary text-secondary-foreground border border-border px-3 py-1 rounded-md font-medium select-text cursor-text"
                                        title={`Номер детали: ${part.number}`}
                                    >
                                        № {part.number}
                                    </span>
                                )}
                                {activeOrderId && (
                                    <Link
                                        to="/orders"
                                        onClick={(e) => e.stopPropagation()}
                                        className="text-xs sm:text-sm bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/40 px-2.5 py-1 rounded-md font-semibold hover:bg-amber-500/25 transition-colors select-text cursor-text"
                                        title={`Эта деталь находится в активном заказе #${activeOrderId}`}
                                    >
                                        📦 В заказе #{activeOrderId}
                                    </Link>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Кнопки действий */}
                    <div className="flex space-x-1 shrink-0" onClick={(e) => e.stopPropagation()}>
                        {/* Быстрое копирование */}
                        <div
                            className={`w-8 h-8 flex items-center justify-center rounded hover:bg-accent cursor-pointer transition-colors ${
                                isPartCopied ? 'text-emerald-500 bg-emerald-500/10' : 'text-muted-foreground hover:text-foreground'
                            }`}
                            onClick={handleCopyPartInfo}
                            title={isPartCopied ? "Скопировано в буфер обмена!" : "Скопировать информацию о детали"}
                        >
                            {isPartCopied ? <Check className="h-4 w-4 text-emerald-500" /> : <Copy className="h-4 w-4" />}
                        </div>

                        {/* Дублирование */}
                        <div
                            className={`w-8 h-8 flex items-center justify-center rounded hover:bg-accent cursor-pointer transition-colors ${
                                isPartDuplicated
                                    ? 'text-emerald-500 bg-emerald-500/10'
                                    : isDuplicating
                                    ? 'text-muted-foreground animate-pulse'
                                    : 'text-muted-foreground hover:text-foreground'
                            }`}
                            onClick={handleDuplicateClick}
                            title={isPartDuplicated ? "Запчасть успешно продублирована!" : isDuplicating ? "Создание дубликата без фото..." : "Дублировать запчасть (без фото)"}
                        >
                            {isPartDuplicated ? <Check className="h-4 w-4 text-emerald-500" /> : <CopyPlus className="h-4 w-4" />}
                        </div>

                        {/* Добавление в заказ */}
                        {(part.quantity ?? 0) > 0 ? (
                            <div
                                className="w-8 h-8 flex items-center justify-center rounded hover:bg-accent cursor-pointer"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    onOpenOrderDialog?.(part);
                                }}
                                title="Оформить заказ / Продать"
                            >
                                <ShoppingCart className="h-4 w-4" />
                            </div>
                        ) : (
                            <div
                                className="opacity-30 w-8 h-8 flex items-center justify-center rounded cursor-not-allowed"
                                title="Нет в наличии (нельзя оформить в заказ)"
                                onClick={(e) => e.stopPropagation()}
                            >
                                <ShoppingCart className="h-4 w-4 text-muted-foreground" />
                            </div>
                        )}

                        {/* Редактирование */}
                        {user?.role === 'admin' && (
                            <div
                                className="w-8 h-8 flex items-center justify-center rounded hover:bg-accent cursor-pointer"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    onOpenEditDialog?.(part);
                                }}
                                title="Редактировать деталь"
                            >
                                <Edit className="h-4 w-4" />
                            </div>
                        )}

                        {/* Удаление */}
                        {user?.role === 'admin' && (
                            <div
                                className="w-8 h-8 flex items-center justify-center rounded hover:bg-accent cursor-pointer"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    onOpenDeleteDialog?.(part);
                                }}
                                title="Удалить деталь"
                            >
                                <Trash2 className="h-4 w-4 text-destructive hover:text-destructive/90" />
                            </div>
                        )}
                    </div>
                </AccordionTrigger>

                {/* Развернутый контент */}
                <AccordionContent>
                    <div className="hover:bg-accent/50 border-t border-border pt-3 px-4">
                        <div className="space-y-3">
                            {/* Галерея фотографий */}
                            {(part.photos && part.photos.length > 0) && (
                                <div>
                                    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                                        {part.photos.map((photoPath, index) => (
                                            <div key={index} className="relative group">
                                                <button
                                                    type="button"
                                                    onClick={() => onOpenPhotoViewer?.(part, index)}
                                                    className="bg-transparent border-none p-0 pointer-events-auto w-full group/img focus:outline-none"
                                                    title="Нажмите для полноэкранного просмотра"
                                                >
                                                    <img
                                                        src={`${API_BASE_URL}${photoPath}${part.updated_at ? `?v=${encodeURIComponent(part.updated_at)}` : ''}`}
                                                        alt={`${part.name} - фото ${index + 1}`}
                                                        onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                                            e.currentTarget.onerror = null;
                                                            e.currentTarget.src = '/placeholder-part.svg';
                                                        }}
                                                        loading="lazy"
                                                        decoding="async"
                                                        className="w-full h-24 object-cover rounded-lg border cursor-pointer group-hover/img:ring-2 group-hover/img:ring-primary/50 group-hover/img:brightness-105 transition-all"
                                                    />
                                                </button>
                                                {user?.role === 'admin' && (
                                                    <div className="absolute top-1 right-1 flex space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                                        <Button
                                                            size="icon"
                                                            variant="secondary"
                                                            className="h-6 w-6 rounded-full shadow-sm bg-background/80 backdrop-blur-sm hover:bg-background"
                                                            onClick={(e) => {
                                                                e.stopPropagation();
                                                                onOpenCropper?.(part, `${API_BASE_URL}${photoPath}`, photoPath);
                                                            }}
                                                            title="Обрезать / Редактировать фото"
                                                        >
                                                            <Crop className="h-3 w-3" />
                                                        </Button>
                                                        <Button
                                                            size="icon"
                                                            variant="secondary"
                                                            className={`h-6 w-6 rounded-full shadow-sm backdrop-blur-sm transition-colors ${
                                                                copiedPhotoPath === photoPath
                                                                    ? 'bg-emerald-500 text-white hover:bg-emerald-600'
                                                                    : 'bg-background/80 hover:bg-background'
                                                            }`}
                                                            onClick={(e) => handleCopyPhoto(photoPath, e)}
                                                            title={copiedPhotoPath === photoPath ? "Скопировано в буфер!" : "Скопировать фото в буфер обмена"}
                                                        >
                                                            {copiedPhotoPath === photoPath ? (
                                                                <Check className="h-3 w-3 text-white" />
                                                            ) : (
                                                                <Copy className="h-3 w-3" />
                                                            )}
                                                        </Button>
                                                        <Button
                                                            size="icon"
                                                            variant="destructive"
                                                            className="h-6 w-6 rounded-full shadow-sm"
                                                            onClick={(e) => {
                                                                e.stopPropagation();
                                                                onDeletePhoto?.(part, photoPath);
                                                            }}
                                                            title="Удалить это фото"
                                                        >
                                                            <X className="h-3 w-3" />
                                                        </Button>
                                                    </div>
                                                )}
                                                {index === 0 && (
                                                    <span className="absolute bottom-1 left-1 bg-black/60 text-white text-[10px] px-1 rounded backdrop-blur-sm pointer-events-none">
                                                        Главное
                                                    </span>
                                                )}
                                            </div>
                                        ))}

                                        {/* Кнопка добавления фото */}
                                        {user?.role === 'admin' && (
                                            <div className="relative group border-2 border-dashed border-border rounded-lg h-24 flex items-center justify-center hover:border-primary transition-colors">
                                                <input
                                                    type="file"
                                                    accept="image/*"
                                                    multiple
                                                    className="hidden"
                                                    id={`add-photo-${part.id}`}
                                                    onChange={async (e) => {
                                                        const files = e.target.files;
                                                        if (files && files.length > 0 && onAddPhoto) {
                                                            for (let i = 0; i < files.length; i++) {
                                                                await onAddPhoto(part, files[i]);
                                                            }
                                                            e.target.value = '';
                                                        }
                                                    }}
                                                />
                                                <label htmlFor={`add-photo-${part.id}`} className="inline-block">
                                                    <Button
                                                        size="sm"
                                                        variant="ghost"
                                                        type="button"
                                                        className="cursor-pointer"
                                                        tabIndex={-1}
                                                    >
                                                        <Plus className="h-4 w-4 mr-1" />
                                                        Добавить
                                                    </Button>
                                                </label>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            )}

                            {/* Загрузка первого фото, если фоток нет */}
                            {(!part.photos || part.photos.length === 0) && user?.role === 'admin' && (
                                <div>
                                    <input
                                        type="file"
                                        accept="image/*"
                                        multiple
                                        className="hidden"
                                        id={`add-first-photo-${part.id}`}
                                        onChange={async (e) => {
                                            const files = e.target.files;
                                            if (files && files.length > 0 && onAddPhoto) {
                                                for (let i = 0; i < files.length; i++) {
                                                    await onAddPhoto(part, files[i]);
                                                }
                                                e.target.value = '';
                                            }
                                        }}
                                    />
                                    <label htmlFor={`add-first-photo-${part.id}`} className="inline-block">
                                        <Button
                                            size="sm"
                                            variant="outline"
                                            type="button"
                                            className="cursor-pointer"
                                            tabIndex={-1}
                                        >
                                            <Plus className="h-4 w-4 mr-1" />
                                            Добавить фотографии
                                        </Button>
                                    </label>
                                </div>
                            )}

                            {/* Описание */}
                            {part.description && (
                                <div className="bg-muted/50 rounded-lg p-3 text-sm select-text cursor-text">
                                    <div className="font-semibold text-muted-foreground mb-1 select-text cursor-text">
                                        Описание:
                                    </div>
                                    <div className="text-foreground whitespace-pre-line leading-relaxed select-text cursor-text">
                                        {part.description}
                                    </div>
                                </div>
                            )}

                            {/* Дефект */}
                            {(part.category !== 'Автохимия и масла' && part.category !== 'Аксессуары и тюннинг') && part.defect && (
                                <div className="bg-amber-500/10 border border-amber-500/25 text-amber-950 dark:text-amber-200 rounded-xl p-3 text-sm flex items-start gap-2.5 select-text cursor-text">
                                    <div className="flex-1 select-text cursor-text">
                                        <div className="text-[11px] font-bold uppercase tracking-wider text-amber-700 dark:text-amber-400 mb-0.5 select-text cursor-text">
                                            Дефект
                                        </div>
                                        <div className="font-medium select-text cursor-text">{part.defect}</div>
                                    </div>
                                </div>
                            )}

                            {/* Характеристики */}
                            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-2.5">
                                {(() => {
                                    const formatPosValue = (code: string, val?: string) => {
                                        if (!val) return null;
                                        if (code === 'front_rear') {
                                            if (val === 'F') return 'F (Перед)';
                                            if (val === 'R') return 'R (Зад)';
                                        }
                                        if (code === 'left_right') {
                                            if (val === 'L') return 'L (Лево)';
                                            if (val === 'R') return 'R (Право)';
                                        }
                                        return val;
                                    };

                                    const allSpecs = [
                                        { code: 'body_brand', label: 'Марка кузова', value: part.body_brand },
                                        { code: 'engine_brand', label: 'Марка двигателя', value: part.engine_brand },
                                        { code: 'car_release_date', label: 'Год выпуска', value: part.car_release_date },
                                        { code: 'vin', label: 'VIN / Номер кузова (Н.К.)', value: part.vin },
                                        { code: 'car_release_period', label: 'Период выпуска', value: part.car_release_period },
                                        { code: 'drive', label: 'Привод', value: part.drive },
                                        { code: 'color', label: 'Цвет кузовных деталей', value: part.color },
                                        { code: 'front_rear', label: 'Перед/зад', value: formatPosValue('front_rear', part.front_rear) },
                                        { code: 'left_right', label: 'Право/лево', value: formatPosValue('left_right', part.left_right) },
                                        { code: 'top_bottom', label: 'Верх/низ', value: part.top_bottom },
                                        { code: 'number', label: 'Номер детали', value: part.number },
                                        { code: 'manufacturer', label: 'Производитель', value: part.manufacturer },
                                        { code: 'manufacturer_code', label: 'Код производителя', value: part.manufacturer_code },
                                        { code: 'oem_code', label: 'OEM код', value: part.oem_code },
                                        { code: 'condition', label: 'Состояние', value: part.condition },
                                        { code: 'supplier_code', label: 'Код поставки', value: part.supplier_code },
                                        { code: 'transmission', label: 'Трансмиссия', value: part.transmission },
                                        { code: 'transmission_model', label: 'Модель трансмиссии', value: part.transmission_model },
                                        { code: 'wear_percentage', label: 'Процент износа', value: part.wear_percentage ? `${part.wear_percentage}%` : null },
                                        { code: 'season', label: 'Сезон', value: part.season },
                                        { code: 'diameter', label: 'Диаметр', value: part.diameter },
                                        { code: 'width', label: 'Ширина', value: part.width },
                                        { code: 'profile', label: 'Профиль', value: part.profile },
                                        { code: 'tire_quantity', label: 'Количество шин', value: part.tire_quantity },
                                        { code: 'drilling', label: 'Сверловка', value: part.drilling },
                                        { code: 'offset', label: 'Вылет', value: part.offset },
                                        { code: 'center_hole_diameter', label: 'Диаметр ЦО', value: part.center_hole_diameter },
                                        { code: 'tire_model', label: 'Модель шины', value: part.tire_model },
                                        { code: 'location', label: 'Местоположение', value: part.location },
                                        { code: 'address', label: 'Адрес склада', value: part.address },
                                        { code: 'salesman', label: 'Продавец', value: part.salesman },
                                    ];

                                    return allSpecs
                                        .filter(item => Boolean(item.value))
                                        .map((item, idx) => (
                                            <div
                                                key={idx}
                                                className="bg-muted/40 hover:bg-muted/70 border border-border/60 rounded-xl p-2.5 flex flex-col justify-between transition-colors min-h-[58px] select-text cursor-text"
                                            >
                                                <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground truncate select-text cursor-text" title={item.label}>
                                                    {item.label}
                                                </span>
                                                <span className="text-sm font-semibold text-foreground break-words mt-0.5 line-clamp-2 select-text cursor-text" title={String(item.value)}>
                                                    {item.value}
                                                </span>
                                            </div>
                                        ));
                                })()}
                            </div>
                        </div>
                    </div>
                </AccordionContent>
            </AccordionItem>
        </motion.div>
    );
}

export default React.memo(PartBlock);
