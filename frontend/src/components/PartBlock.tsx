/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import React, { useState, useRef, useEffect } from "react";
import { AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Trash2, Edit, ShoppingCart, Plus, X, Crop } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { motion, AnimatePresence } from "framer-motion";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { usePartEdit } from "@/hooks/usePartEdit";
import { useDeletePart, partsKeys } from "@/hooks/useParts";
import { useAuth } from "@/hooks/useAuth";
import { partsApi } from "@/features/parts/api/partsApi";
import PartOrderDialog from "./PartOrderDialog";
import ImageEditor from "./ImageEditor";
import EditPartDialog from "./EditPartDialog";
import PartPhotoViewer from "./PartPhotoViewer";
import type { Part } from "@/features/parts/types";
import { API_BASE_URL } from "@/lib/api";
import { useQueryClient } from '@tanstack/react-query';
import { getAuthHeaders } from '@/lib/csrf';

interface PartBlockProps {
    part: Part;
    isLoading?: boolean;
    isSelectionMode?: boolean;
    isSelected?: boolean;
    onLongPress?: () => void;
    onSelect?: (isSelected: boolean) => void;
}

function PartBlock({
                       part,
                       isLoading = false,
                       isSelectionMode = false,
                       isSelected = false,
                       onLongPress,
                       onSelect
                   }: PartBlockProps) {
    const { user } = useAuth();
    const accordionRef = useRef<HTMLDivElement>(null);
    const longPressTimerRef = useRef<NodeJS.Timeout | null>(null);
    const [, setIsPressed] = useState(false);

    const [isDeleting, setIsDeleting] = useState(false);
    const [isOrderDialogOpen, setIsOrderDialogOpen] = useState(false);
    const [isPhotoViewerOpen, setIsPhotoViewerOpen] = useState(false);
    const [photoViewerIndex, setPhotoViewerIndex] = useState(0);
    const [showCropper, setShowCropper] = useState(false);
    const [tempImageSrc, setTempImageSrc] = useState<string | File>("");
    const [photoToReplace, setPhotoToReplace] = useState<string | null>(null);
    const [originalFile, setOriginalFile] = useState<File | null>(null);
    const queryClient = useQueryClient();

    const handleOpenPhotoViewer = (index: number = 0, e?: React.MouseEvent) => {
        if (e) {
            e.stopPropagation();
        }
        setPhotoViewerIndex(index);
        setIsPhotoViewerOpen(true);
    };

    // Для множественных фото используем partsApi напрямую
    const [photoUploadTimestamp, setPhotoUploadTimestamp] = useState<number>(Date.now());
    const partEdit = usePartEdit({
        initialPart: part,
    });
    const deletePartMutation = useDeletePart();

    // Обработчики долгого нажатия
    const handlePointerDown = (e: React.PointerEvent) => {
        // Игнорируем если клик на интерактивных элементах
        const target = e.target as HTMLElement;
        if (target.closest('input, a, [role="button"]')) {
            return;
        }

        setIsPressed(true);
        longPressTimerRef.current = setTimeout(() => {
            if (onLongPress) {
                onLongPress();
            }
        }, 300); // 0.3 секунды
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

    // Очистка таймера при размонтировании
    useEffect(() => {
        return () => {
            if (longPressTimerRef.current) {
                clearTimeout(longPressTimerRef.current);
            }
        };
    }, []);

    // Закрываем диалог заказа при изменении детали (редактирование/удаление)
    useEffect(() => {
        console.log('PartBlock: part.id changed to', part.id, 'closing order dialog');
        setIsOrderDialogOpen(false);
    }, [part.id]);

    const handlePhotoChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
        console.log('PartBlock handlePhotoChange: START - event target files:', e.target.files);
        const file = e.target.files?.[0];
        console.log('PartBlock handlePhotoChange: file selected:', file?.name, 'size:', file?.size, 'type:', file?.type);
        if (file) {
            console.log('PartBlock handlePhotoChange: file validation passed');
            // Validate file type - only allow images
            if (!file.type.startsWith('image/')) {
                console.error('PartBlock handlePhotoChange: invalid file type:', file.type);
                alert('Please select a valid image file.');
                return;
            }

            // Validate file size (max 5MB to prevent memory issues)
            const maxSize = 5 * 1024 * 1024; // 5MB
            if (file.size > maxSize) {
                console.error('PartBlock handlePhotoChange: file too large:', file.size, 'max:', maxSize);
                alert('File size must be less than 5MB.');
                return;
            }

            console.log('PartBlock handlePhotoChange: current part name:', partEdit.editForm.name);

            setOriginalFile(file);
            console.log('PartBlock handlePhotoChange: calling partEdit.photoUpload.handlePhotoChange');
            partEdit.photoUpload.handlePhotoChange({
                target: { files: [file] }
            } as unknown as React.ChangeEvent<HTMLInputElement>);
        } else {
            console.warn('PartBlock handlePhotoChange: no file selected');
        }
        console.log('PartBlock handlePhotoChange: END');
    };

    const handleCropComplete = async (croppedImageBlob: Blob) => {
        console.log('PartBlock handleCropComplete called with blob size:', croppedImageBlob.size, 'type:', croppedImageBlob.type);
        if (croppedImageBlob.size === 0) {
            alert('Ошибка: обрезанное изображение пустое');
            return;
        }
        const croppedFile = new File([croppedImageBlob], 'cropped-image.jpg', { type: 'image/jpeg' });
        console.log('PartBlock created file:', croppedFile.name, 'size:', croppedFile.size, 'type:', croppedFile.type);

        try {
            // 1. Загружаем новое (отредактированное) фото
            await partsApi.uploadPhoto(part.id, croppedFile);

            // 2. Если мы редактировали существующее фото на сервере, удаляем старую версию
            if (photoToReplace) {
                try {
                    console.log('Удаляем старую версию фото после редактирования:', photoToReplace);
                    await fetch(`${API_BASE_URL}/api/v1/parts/${part.id}/photo?photo_url=${encodeURIComponent(photoToReplace)}`, {
                        method: 'DELETE',
                        headers: getAuthHeaders(),
                        credentials: 'include',
                    });
                } catch (delErr) {
                    console.warn('Не удалось удалить старое фото при замене:', delErr);
                }
            }

            setPhotoUploadTimestamp(Date.now());
            void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
        } catch (error) {
            console.error('PartBlock handleCropComplete: upload failed:', error);
            alert('Не удалось сохранить отредактированное фото. Попробуйте еще раз.');
        } finally {
            setShowCropper(false);
            setTempImageSrc("");
            setPhotoToReplace(null);
            setOriginalFile(null);
        }
    };

    const handleCropCancel = () => {
        setShowCropper(false);
        setTempImageSrc("");
        setPhotoToReplace(null);
        setOriginalFile(null);
        // Reset the inputs
        const input = document.getElementById('photo') as HTMLInputElement;
        if (input) input.value = '';
        const addInput = document.getElementById('add-photo-input') as HTMLInputElement;
        if (addInput) addInput.value = '';
    };

    const handleAddPhoto = async (e: React.ChangeEvent<HTMLInputElement>) => {
        console.log('PartBlock handleAddPhoto: START - event target files:', e.target.files);
        const file = e.target.files?.[0];
        console.log('PartBlock handleAddPhoto: file selected:', file?.name, 'size:', file?.size, 'type:', file?.type);
        if (file) {
            console.log('PartBlock handleAddPhoto: file validation passed');
            // Validate file type - only allow images
            if (!file.type.startsWith('image/')) {
                console.error('PartBlock handleAddPhoto: invalid file type:', file.type);
                alert('Please select a valid image file.');
                return;
            }

            // Validate file size (max 5MB to prevent memory issues)
            const maxSize = 5 * 1024 * 1024; // 5MB
            if (file.size > maxSize) {
                console.error('PartBlock handleAddPhoto: file too large:', file.size, 'max:', maxSize);
                alert('File size must be less than 5MB.');
                return;
            }

            console.log('PartBlock handleAddPhoto: uploading photo directly');
            try {
                await partsApi.uploadPhoto(part.id, file);
                setPhotoUploadTimestamp(Date.now());
                void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
                console.log('PartBlock handleAddPhoto: photo uploaded successfully');
            } catch (error) {
                console.error('PartBlock handleAddPhoto: upload failed:', error);
                alert('Failed to upload photo. Please try again.');
            }
        } else {
            console.warn('PartBlock handleAddPhoto: no file selected');
        }
        console.log('PartBlock handleAddPhoto: END');
        // Reset the input
        e.target.value = '';
    };

    const handleDeletePhoto = async (photoPath: string) => {
        if (!confirm('Вы уверены, что хотите удалить это фото?')) {
            return;
        }

        console.log('PartBlock handleDeletePhoto: START - photoPath:', photoPath);
        try {
            const response = await fetch(`${API_BASE_URL}/api/v1/parts/${part.id}/photo?photo_url=${encodeURIComponent(photoPath)}`, {
                method: 'DELETE',
                headers: getAuthHeaders(),
                credentials: 'include',
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`HTTP ${response.status}: ${errorText}`);
            }

            const result = await response.json();
            console.log('Фото успешно удалено:', result);

            // Обновить локальное состояние
            setPhotoUploadTimestamp(Date.now());
            void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
        } catch (error) {
            console.error('Ошибка при удалении фото:', error);
            alert('Не удалось удалить фото. Попробуйте еще раз.');
        }
        console.log('PartBlock handleDeletePhoto: END');
    };

    const handleDelete = async () => {
        console.time('PartBlock.handleDelete');
        console.log('PartBlock handleDelete: START - part.id:', part.id, 'isDeleting:', isDeleting);

        if (isDeleting) {
            console.log('PartBlock handleDelete: Already deleting, ignoring click');
            return;
        }

        console.log(`PartBlock handleDelete: Starting delete for part ${part.id}`);
        setIsDeleting(true);

        try {
            console.log('PartBlock handleDelete: Calling deletePartMutation with:', part.id);
            await deletePartMutation.mutateAsync(part.id);
            console.log(`PartBlock handleDelete: Delete completed for part ${part.id}`);
            // Success - component will unmount as parent removes it from list
        } catch (error: unknown) {
            // Error occurred - component still mounted, reset state
            console.error(`PartBlock handleDelete: Error for part ${part.id}:`, error);
            console.error('PartBlock handleDelete: Error details:', error instanceof Error ? error.message : error);
            console.error('PartBlock handleDelete: Error stack:', error instanceof Error ? error.stack : 'No stack');
            setIsDeleting(false);
            throw error;
        }
        console.log('PartBlock handleDelete: END');
        console.timeEnd('PartBlock.handleDelete');
    };

    // Проверка безопасности - валидация данных детали для предотвращения ошибок выполнения
    if (!part) {
        console.error('PartBlock: Invalid part data', part);
        return null;
    }

    if (isLoading) {
        return (
            <motion.div
                className="border rounded-lg mb-2 p-4"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3 }}
            >
                <div className="flex items-center space-x-4">
                    <motion.div
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={{ opacity: 1, scale: 1 }}
                        transition={{ delay: 0.1, duration: 0.3 }}
                    >
                        <Skeleton className="h-4 w-48" />
                    </motion.div>
                    <motion.div
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={{ opacity: 1, scale: 1 }}
                        transition={{ delay: 0.2, duration: 0.3 }}
                    >
                        <Skeleton className="h-8 w-8 rounded" />
                    </motion.div>
                </div>
                <motion.div
                    className="mt-4 space-y-2"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 0.3, duration: 0.4 }}
                >
                    {[32, 40, 36, 28, 30, 34, 26].map((width, index) => (
                        <motion.div
                            key={index}
                            initial={{ opacity: 0, x: -20 }}
                            animate={{ opacity: 1, x: 0 }}
                            transition={{ delay: 0.4 + index * 0.1, duration: 0.3 }}
                        >
                            <Skeleton className={`h-4 w-${width}`} />
                        </motion.div>
                    ))}
                </motion.div>
            </motion.div>
        );
    }

    return (
        <motion.div
            className="bg-card border border-border rounded-lg mb-4 relative group hover:shadow-md transition-shadow p-4"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{
                duration: 1.0,
                ease: [0.25, 0.46, 0.45, 0.94],
                delay: 0.1
            }}
            layout
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

            {/* Элемент аккордеона для деталей детали */}
            <AccordionItem
                value={`part-${part.id}`}
                data-testid="part-accordion"
                ref={accordionRef}
                className="border-b last:border-b-0"
            >
                <AccordionTrigger
                    className="flex items-center justify-between w-full hover:bg-accent hover:text-accent-foreground pr-4 pointer-events-auto"
                >
                    <div className="flex items-center space-x-2 sm:space-x-3 flex-1">
                        <AnimatePresence initial={false}>
                            <motion.div
                                layoutId={`part-image-${part.id}`}
                                initial={{ opacity: 0, scale: 0.8 }}
                                animate={{ opacity: 1, scale: 1 }}
                                exit={{ opacity: 0, scale: 0.8 }}
                                transition={{ duration: 1.0, delay: 0.1 }}
                            >
                                <motion.img
                                    key={photoUploadTimestamp}
                                    src={(part.photos && part.photos.length > 0) ? `${API_BASE_URL}${part.photos[0]}?t=${photoUploadTimestamp}` : '/placeholder-part.svg'}
                                    alt={part.name || 'Изображение детали'}
                                    onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                        e.currentTarget.onerror = null;
                                        e.currentTarget.src = '/placeholder-part.svg';
                                    }}
                                    fetchPriority="high"
                                    loading="eager"
                                    className="w-12 h-12 sm:w-16 sm:h-16 object-cover rounded-lg border cursor-pointer ml-3 mr-3 hover:ring-2 hover:ring-primary/50 transition-all"
                                    whileHover={{ scale: 1.08, rotate: 3 }}
                                    whileTap={{ scale: 0.95 }}
                                    transition={{ duration: 0.1 }}
                                    onClick={(e) => handleOpenPhotoViewer(0, e)}
                                    title="Нажмите для увеличения фото"
                                />
                            </motion.div>
                        </AnimatePresence>
                        <div className="flex flex-col min-w-0 flex-1">
                            <span className="truncate text-sm sm:text-base font-semibold">{part.name || 'Unnamed Part'}</span>
                            {(part.brand || part.model) && (
                                <span className="text-xs sm:text-sm text-muted-foreground truncate">
                                    {part.brand && part.model ? `${part.brand} ${part.model}` : part.brand || part.model}
                                </span>
                            )}
                            <div className="flex flex-wrap gap-2 mt-1">
                                {part.category && (
                                    <span className="text-sm bg-secondary text-secondary-foreground border border-border px-3 py-1 rounded-md font-medium">{part.category}</span>
                                )}
                                <span className="text-sm bg-secondary text-secondary-foreground border border-border px-3 py-1 rounded-md font-medium">Кол: {part.quantity ?? 0}</span>
                                <span className="text-sm bg-primary text-primary-foreground dark:bg-emerald-500/20 dark:text-emerald-300 dark:border-emerald-500/30 border border-primary/20 px-3 py-1 rounded-md font-semibold">Цена: {part.price && Number(part.price) > 0 ? `₽${part.price}` : 'отсутствует'}</span>
                            </div>
                        </div>
                    </div>
                    {/* Кнопки действий */}
                    <div className="flex space-x-1 shrink-0" onClick={(e) => e.stopPropagation()}>
                        {/* Кнопка добавления в заказ */}
                        <motion.div
                            className="opacity-100 transition-opacity w-8 h-8 flex items-center justify-center rounded hover:bg-accent"
                            onClick={(e) => {
                                e.stopPropagation();
                                setIsOrderDialogOpen(true);
                            }}
                            transition={{ duration: 0.2 }}
                        >
                            <ShoppingCart className="h-4 w-4 cursor-pointer" />
                        </motion.div>

                        {/* Кнопка редактирования */}
                        {user?.role === 'admin' && (
                            <motion.div
                                className="opacity-100 transition-opacity w-8 h-8 flex items-center justify-center rounded hover:bg-accent"
                                onClick={(e) => {
                                    console.log('Edit button clicked for part:', part.id);
                                    partEdit.setIsEditing(true);
                                    e.stopPropagation();
                                }}
                                style={{ cursor: partEdit.isEditing ? 'not-allowed' : 'pointer' }}
                                whileHover={{ scale: 1.1, rotate: 10 }}
                                whileTap={{ scale: 0.9 }}
                                transition={{ duration: 0.1 }}
                            >
                                <Edit className="h-4 w-4 cursor-pointer" />
                            </motion.div>
                        )}

                        {/* Диалог подтверждения удаления */}
                        {user?.role === 'admin' && (
                            <AlertDialog>
                                <AlertDialogTrigger asChild>
                                    <motion.div
                                        className="w-8 h-8 flex items-center justify-center rounded hover:bg-accent pointer-events-auto cursor-pointer"
                                        transition={{ duration: 0.1 }}
                                        style={{ pointerEvents: isDeleting ? 'none' : 'auto' }}
                                    >
                                        <Trash2 className={`h-4 w-4 ${isDeleting ? 'text-muted' : 'text-destructive hover:text-destructive/90'}`} />
                                    </motion.div>
                                </AlertDialogTrigger>
                                <AlertDialogContent>
                                    <AlertDialogHeader>
                                        <AlertDialogTitle>Вы абсолютно уверены?</AlertDialogTitle>
                                        <AlertDialogDescription>
                                            Это действие нельзя отменить. Это навсегда удалит деталь
                                            и удалит её данные из нашего инвентаря.
                                        </AlertDialogDescription>
                                    </AlertDialogHeader>
                                    <AlertDialogFooter>
                                        <AlertDialogCancel>Отмена</AlertDialogCancel>
                                        <AlertDialogAction
                                            onClick={() => handleDelete()}
                                            disabled={isDeleting}
                                        >
                                            {isDeleting ? 'Удаление...' : 'Продолжить'}
                                        </AlertDialogAction>
                                    </AlertDialogFooter>
                                </AlertDialogContent>
                            </AlertDialog>
                        )}
                    </div>

                </AccordionTrigger>

                {/* Развёрнутый контент с деталями детали */}
                <AccordionContent>
                    <motion.div
                        className="hover:bg-accent/50 border-t border-border pt-3 px-4"
                        initial={{ opacity: 0, height: 0 }}
                        animate={{ opacity: 1, height: "auto" }}
                        exit={{ opacity: 0, height: 0 }}
                        transition={{ duration: 0.3, ease: "easeInOut" }}
                    >
                        <motion.div
                            className="space-y-1"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: 0.1, duration: 0.3 }}
                        >
                            {/* Отображение фото в развёрнутом виде */}
                            {(part.photos && part.photos.length > 0) && (
                                <AnimatePresence>
                                    <motion.div
                                        className="mb-3"
                                        initial={{ opacity: 0, y: 20 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        exit={{ opacity: 0, y: -20 }}
                                        transition={{ duration: 1.0, delay: 0.1 }}
                                    >
                                        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                                             {part.photos.map((photoPath, index) => (
                                                <div key={index} className="relative group">
                                                    <button
                                                        type="button"
                                                        onClick={(e) => handleOpenPhotoViewer(index, e)}
                                                        className="bg-transparent border-none p-0 pointer-events-auto w-full group/img focus:outline-none"
                                                        title="Нажмите для увеличения фото"
                                                    >
                                                        <motion.img
                                                            src={`${API_BASE_URL}${photoPath}?t=${partEdit.photoUpload.uploadTimestamp}`}
                                                            alt={`${part.name} - фото ${index + 1}`}
                                                            className="w-full h-24 object-cover rounded-lg border cursor-pointer group-hover/img:ring-2 group-hover/img:ring-primary/50 group-hover/img:brightness-105 transition-all"
                                                            transition={{ duration: 0.2 }}
                                                            onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                                                e.currentTarget.onerror = null;
                                                                e.currentTarget.src = '/placeholder-part.svg';
                                                            }}
                                                        />
                                                    </button>
                                                    {/* Кнопки действий над фото */}
                                                    {user?.role === 'admin' && (
                                                        <div className="absolute top-1 right-1 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                                            <button
                                                                type="button"
                                                                onClick={(e) => {
                                                                    e.stopPropagation();
                                                                    const photoUrl = photoPath.startsWith('http') || photoPath.startsWith('data:')
                                                                        ? photoPath
                                                                        : `${API_BASE_URL}${photoPath.startsWith('/') ? photoPath : `/${photoPath}`}`;
                                                                    setTempImageSrc(photoUrl);
                                                                    setPhotoToReplace(photoPath);
                                                                    setShowCropper(true);
                                                                }}
                                                                className="bg-neutral-900/80 hover:bg-neutral-800 text-white rounded-full w-6 h-6 flex items-center justify-center shadow transition-colors"
                                                                title="Редактировать фото"
                                                            >
                                                                <Crop className="w-3 h-3" />
                                                            </button>
                                                            <button
                                                                type="button"
                                                                onClick={(e) => {
                                                                    e.stopPropagation();
                                                                    handleDeletePhoto(photoPath);
                                                                }}
                                                                className="bg-red-500 hover:bg-red-600 text-white rounded-full w-6 h-6 flex items-center justify-center shadow transition-colors"
                                                                title="Удалить фото"
                                                            >
                                                                <X className="w-3 h-3" />
                                                            </button>
                                                        </div>
                                                    )}
                                                </div>
                                            ))}
                                        </div>

                                        {/* Кнопка добавления фото */}
                                        {user?.role === 'admin' && (
                                            <motion.div
                                                className="mt-3"
                                                initial={{ opacity: 0, y: 20 }}
                                                animate={{ opacity: 1, y: 0 }}
                                                transition={{ duration: 0.3, delay: 0.2 }}
                                            >
                                                <label className="inline-block">
                                                    <input
                                                        type="file"
                                                        accept="image/*"
                                                        onChange={handleAddPhoto}
                                                        className="hidden"
                                                    />
                                                    <Button
                                                        variant="outline"
                                                        size="sm"
                                                        className="cursor-pointer"
                                                        asChild
                                                    >
                                                        <span>
                                                            <Plus className="w-4 h-4 mr-2" />
                                                            Добавить фото
                                                        </span>
                                                    </Button>
                                                </label>
                                            </motion.div>
                                        )}
                                    </motion.div>
                                </AnimatePresence>
                            )}
                            <div className="w-full space-y-4 pt-2">
                                <h4 className="font-semibold text-xs tracking-wider uppercase text-muted-foreground flex items-center gap-2">
                                    <span className="h-px bg-border flex-1"></span>
                                    <span>Характеристики запчасти</span>
                                    <span className="h-px bg-border flex-1"></span>
                                </h4>

                                {/* Описание запчасти */}
                                {part.description && (
                                    <motion.div
                                        initial={{ opacity: 0, y: 10 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        transition={{ duration: 0.2 }}
                                        className="bg-muted/40 border border-border/60 rounded-xl p-3 text-sm"
                                    >
                                        <div className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground mb-1">
                                            Описание
                                        </div>
                                        <div className="text-foreground whitespace-pre-line leading-relaxed">
                                            {part.description}
                                        </div>
                                    </motion.div>
                                )}

                                {/* Информация о дефекте */}
                                {(part.category !== 'Автохимия и масла' && part.category !== 'Аксессуары и тюннинг') && part.defect && (
                                    <motion.div
                                        initial={{ opacity: 0, y: 10 }}
                                        animate={{ opacity: 1, y: 0 }}
                                        transition={{ duration: 0.2 }}
                                        className="bg-amber-500/10 border border-amber-500/25 text-amber-950 dark:text-amber-200 rounded-xl p-3 text-sm flex items-start gap-2.5"
                                    >
                                        <div className="flex-1">
                                            <div className="text-[11px] font-bold uppercase tracking-wider text-amber-700 dark:text-amber-400 mb-0.5">
                                                Дефект
                                            </div>
                                            <div className="font-medium">{part.defect}</div>
                                        </div>
                                    </motion.div>
                                )}

                                {/* Адаптивная сетка карточек характеристик */}
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
                                            { code: 'vin', label: 'VIN / Номер кузова', value: part.vin },
                                            { code: 'car_release_period', label: 'Период выпуска автомобиля', value: part.car_release_period },
                                            { code: 'drive', label: 'Привод', value: part.drive },
                                            { code: 'color', label: 'Цвет кузовных деталей', value: part.color },
                                            { code: 'front_rear', label: 'Перед/зад', value: formatPosValue('front_rear', part.front_rear) },
                                            { code: 'left_right', label: 'Право/лево', value: formatPosValue('left_right', part.left_right) },
                                            { code: 'top_bottom', label: 'Верх/низ', value: part.top_bottom },
                                            { code: 'number', label: 'Номер', value: part.number },
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
                                            .filter(item => {
                                                if (!item.value) return false;
                                                return true;
                                            })
                                            .map((item, idx) => (
                                                <motion.div
                                                    key={item.label}
                                                    initial={{ opacity: 0, y: 10 }}
                                                    animate={{ opacity: 1, y: 0 }}
                                                    transition={{ delay: Math.min(idx * 0.03, 0.4), duration: 0.2 }}
                                                    className="bg-muted/40 hover:bg-muted/70 border border-border/60 rounded-xl p-2.5 flex flex-col justify-between transition-colors min-h-[58px]"
                                                >
                                                    <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground truncate" title={item.label}>
                                                        {item.label}
                                                    </span>
                                                    <span className="text-sm font-semibold text-foreground break-words mt-0.5 line-clamp-2" title={String(item.value)}>
                                                        {item.value}
                                                    </span>
                                                </motion.div>
                                            ));
                                    })()}
                                </div>
                            </div>
                        </motion.div>
                    </motion.div>
                </AccordionContent>
            </AccordionItem>

            {/* Диалог редактирования.
                Монтируется только на время редактирования: PartBlock рисуется на
                каждую строку инвентаря, а EditPartDialog — это семьсот строк разметки
                и собственные запросы к каталогам. Держать его смонтированным для
                каждой запчасти незачем. */}
            {partEdit.isEditing && (
            <EditPartDialog
                partEdit={partEdit}
                part={part}
                onPhotoChange={handlePhotoChange}
                onCrop={(src, pathToReplace) => {
                    if (src) {
                        setTempImageSrc(src);
                        setPhotoToReplace(pathToReplace || null);
                        setShowCropper(true);
                    }
                }}
                onDeletePhoto={async (photoPath?: string) => {
                    // Если есть новое загруженное фото (не сохраненное), просто сбросить его
                    if (partEdit.photoUpload.photoFile) {
                        partEdit.photoUpload.resetPhoto();
                        partEdit.updateFormField('photo', '');
                        setOriginalFile(null);
                        return;
                    }

                    // Удаляем конкретное фото или первое
                    const targetPhoto = photoPath || (part.photos && part.photos.length > 0 ? part.photos[0] : '');

                    if (!targetPhoto) {
                        alert('Нет фото для удаления');
                        return;
                    }

                    if (!confirm('Вы уверены, что хотите удалить это фото?')) {
                        return;
                    }

                    try {
                        const deleteUrl = `${API_BASE_URL}/api/v1/parts/${part.id}/photo?photo_url=${encodeURIComponent(targetPhoto)}`;
                        console.log('Отправка запроса на удаление:', deleteUrl);

                        const response = await fetch(deleteUrl, {
                            method: 'DELETE',
                            headers: getAuthHeaders(),
                            credentials: 'include',
                        });

                        if (!response.ok) {
                            const errorText = await response.text();
                            console.error('Ошибка ответа сервера:', errorText);
                            throw new Error(`HTTP ${response.status}: ${errorText}`);
                        }

                        // Обновить локальное состояние
                        partEdit.updateFormField('photo', '');
                        setPhotoUploadTimestamp(Date.now());
                        void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
                    } catch (error) {
                        console.error('Ошибка при удалении фото:', error);
                        alert('Не удалось удалить фото. Попробуйте еще раз.');
                    }
                }}
                originalFile={originalFile}
            />
            )}

            {/* Диалог заказа */}
            <PartOrderDialog
                part={part}
                isOpen={isOrderDialogOpen}
                onOpenChange={setIsOrderDialogOpen}
            />

            {showCropper && (
                <ImageEditor
                    src={tempImageSrc}
                    onEditComplete={handleCropComplete}
                    onCancel={handleCropCancel}
                    aspect={null} // Free aspect ratio for parts photos
                />
            )}

            {/* Просмотр и масштабирование фотографий детали */}
            <PartPhotoViewer
                isOpen={isPhotoViewerOpen}
                onClose={() => setIsPhotoViewerOpen(false)}
                photos={part.photos || []}
                initialIndex={photoViewerIndex}
                partName={part.name}
                partSubtitle={part.brand && part.model ? `${part.brand} ${part.model}` : part.brand || part.model}
                uploadTimestamp={partEdit.photoUpload.uploadTimestamp || photoUploadTimestamp}
            />
        </motion.div>
    );
}

export default React.memo(PartBlock);
