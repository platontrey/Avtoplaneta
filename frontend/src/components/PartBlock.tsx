/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import React, { useState, useRef, useEffect } from "react";
import { AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Trash2, Edit, ShoppingCart, Plus, X } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { motion, AnimatePresence } from "framer-motion";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
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
    const [showCropper, setShowCropper] = useState(false);
    const [tempImageSrc, setTempImageSrc] = useState<string | File>("");
    const [originalFile, setOriginalFile] = useState<File | null>(null);
    const queryClient = useQueryClient();

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

        // Upload cropped photo directly
        try {
            await partsApi.uploadPhoto(part.id, croppedFile);
            setPhotoUploadTimestamp(Date.now());
            void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
        } catch (error) {
            console.error('PartBlock handleCropComplete: upload failed:', error);
            alert('Failed to upload cropped photo. Please try again.');
        }

        setShowCropper(false);
        setTempImageSrc("");
    };

    const handleCropCancel = () => {
        setShowCropper(false);
        setTempImageSrc("");
        // Reset the input
        const input = document.getElementById('photo') as HTMLInputElement;
        if (input) input.value = '';
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
            const response = await fetch(`${API_BASE_URL}/api/deletepartphoto/${part.id}?photo=${encodeURIComponent(photoPath)}`, {
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
                                <Dialog>
                                    <DialogTrigger asChild>
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
                                            className="w-12 h-12 sm:w-16 sm:h-16 object-cover rounded-lg border cursor-pointer ml-3 mr-3"
                                            whileHover={{ scale: 1.1, rotate: 5 }}
                                            whileTap={{ scale: 0.95 }}
                                            transition={{ duration: 0.1 }}
                                        />
                                    </DialogTrigger>
                                    <DialogContent className="max-w-4xl" onClick={(e) => e.stopPropagation()}>
                                        <DialogTitle>{part.name}</DialogTitle>
                                        <DialogDescription>Изображения детали</DialogDescription>
                                        {(part.photos && part.photos.length > 0) ? (
                                            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                                                {part.photos.map((photoPath, index) => (
                                                    <motion.img
                                                        key={index}
                                                        src={`${API_BASE_URL}${photoPath}?t=${partEdit.photoUpload.uploadTimestamp}`}
                                                        alt={`${part.name} - фото ${index + 1}`}
                                                        className="w-full h-auto max-h-48 object-contain rounded-lg border"
                                                        onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                                            e.currentTarget.onerror = null;
                                                            e.currentTarget.src = '/placeholder-part.svg';
                                                        }}
                                                        initial={{ opacity: 0, scale: 0.9 }}
                                                        animate={{ opacity: 1, scale: 1 }}
                                                        transition={{ duration: 0.3, delay: index * 0.1 }}
                                                    />
                                                ))}
                                            </div>
                                        ) : (
                                            <motion.img
                                                key={partEdit.photoUpload.forceRefresh}
                                                src="/placeholder-part.svg"
                                                alt={part.name}
                                                className="w-full h-auto max-h-[80vh] object-contain"
                                                initial={{ opacity: 0, scale: 0.9 }}
                                                animate={{ opacity: 1, scale: 1 }}
                                                transition={{ duration: 0.3 }}
                                            />
                                        )}
                                    </DialogContent>
                                </Dialog>
                            </motion.div>
                        </AnimatePresence>
                        <div className="flex flex-col min-w-0 flex-1">
                            <span className="truncate text-sm sm:text-base font-semibold">{part.name || 'Unnamed Part'}</span>
                            {(part.brand || part.model) && (
                                <span className="text-xs sm:text-sm text-muted-foreground truncate">
                                    {part.brand && part.model ? `${part.brand} ${part.model}` : part.brand || part.model}
                                </span>
                            )}
                            <div className="flex gap-2 mt-1">
                                {part.category && (
                                    <span className="text-sm bg-white border border-border px-3 py-1 rounded-md font-medium">{part.category}</span>
                                )}
                                <span className="text-sm bg-white border border-border px-3 py-1 rounded-md font-medium">Кол: {part.quantity ?? 0}</span>
                                <span className="text-sm bg-primary text-primary-foreground px-3 py-1 rounded-md font-medium">Цена: {part.price ? `₽${part.price}` : 'TBD'}</span>
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
                                                    <Dialog>
                                                        <DialogTrigger asChild>
                                                            <button onClick={(e) => e.stopPropagation()} className="bg-transparent border-none p-0 pointer-events-auto w-full">
                                                                <motion.img
                                                                    src={`${API_BASE_URL}${photoPath}?t=${partEdit.photoUpload.uploadTimestamp}`}
                                                                    alt={`${part.name} - фото ${index + 1}`}
                                                                    className="w-full h-24 object-cover rounded-lg border cursor-pointer"
                                                                    transition={{ duration: 0.2 }}
                                                                    onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                                                        e.currentTarget.onerror = null;
                                                                        e.currentTarget.src = '/placeholder-part.svg';
                                                                    }}
                                                                />
                                                            </button>
                                                        </DialogTrigger>
                                                        <DialogContent className="max-w-4xl" onClick={(e) => e.stopPropagation()}>
                                                            <DialogTitle>{part.name} - Фото {index + 1}</DialogTitle>
                                                            <DialogDescription>Изображение детали</DialogDescription>
                                                            <motion.img
                                                                src={`${API_BASE_URL}${photoPath}?t=${partEdit.photoUpload.uploadTimestamp}`}
                                                                alt={`${part.name} - фото ${index + 1}`}
                                                                onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                                                    e.currentTarget.onerror = null;
                                                                    e.currentTarget.src = '/placeholder-part.svg';
                                                                }}
                                                                className="w-full h-auto max-h-[80vh] object-contain"
                                                                style={{ minHeight: '300px' }}
                                                                initial={{ opacity: 0, scale: 0.9 }}
                                                                animate={{ opacity: 1, scale: 1 }}
                                                                transition={{ duration: 1.0, delay: 0.1 }}
                                                                onClick={(e) => e.stopPropagation()}
                                                            />
                                                        </DialogContent>
                                                    </Dialog>
                                                    {/* Кнопка удаления фото */}
                                                    {user?.role === 'admin' && (
                                                        <button
                                                            onClick={(e) => {
                                                                e.stopPropagation();
                                                                handleDeletePhoto(photoPath);
                                                            }}
                                                            className="absolute top-1 right-1 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-600"
                                                            title="Удалить фото"
                                                        >
                                                            <X className="w-3 h-3" />
                                                        </button>
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
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div className="space-y-3">
                                    <h4 className="font-semibold text-sm text-muted-foreground mb-3">ХАРАКТЕРИСТИКИ:</h4>
                                    <div className="grid grid-cols-1 gap-3">
                                        {part.brand && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.15, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-muted-foreground">Бренд</div>
                                                <div>{part.brand}</div>
                                            </motion.div>
                                        )}
                                        {part.model && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.2, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Модель</div>
                                                <div>{part.model}</div>
                                            </motion.div>
                                        )}
                                        {part.vin && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.25, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">VIN</div>
                                                <div>{part.vin}</div>
                                            </motion.div>
                                        )}
                                        <motion.div
                                            initial={{ opacity: 0, x: -20 }}
                                            animate={{ opacity: 1, x: 0 }}
                                            transition={{ delay: 0.3, duration: 0.3 }}
                                            className="text-sm"
                                        >
                                            <div className="font-medium text-gray-600">Количество</div>
                                            <div>{part.quantity ?? 0}</div>
                                        </motion.div>
                                        {part.description && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.35, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Описание</div>
                                                <div>{part.description}</div>
                                            </motion.div>
                                        )}
                                        {part.category && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.4, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Категория</div>
                                                <div>{part.category}</div>
                                            </motion.div>
                                        )}
                                        {part.location && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.45, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Местоположение</div>
                                                <div>{part.location}</div>
                                            </motion.div>
                                        )}
                                        {part.salesman && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.5, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Продавец</div>
                                                <div>{part.salesman}</div>
                                            </motion.div>
                                        )}
                                        {/* Характеристики запчасти */}
                                        {/* Common fields */}
                                        {part.manufacturer && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.55, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Производитель</div>
                                                <div>{part.manufacturer}</div>
                                            </motion.div>
                                        )}
                                        {part.manufacturer_code && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.6, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Код производителя</div>
                                                <div>{part.manufacturer_code}</div>
                                            </motion.div>
                                        )}
                                        {part.oem_code && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.65, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">OEM код</div>
                                                <div>{part.oem_code}</div>
                                            </motion.div>
                                        )}
                                        {part.supplier_code && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.7, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Код поставки</div>
                                                <div>{part.supplier_code}</div>
                                            </motion.div>
                                        )}
                                        {part.condition && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.75, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Состояние</div>
                                                <div>{part.condition}</div>
                                            </motion.div>
                                        )}
                                        {part.wear_percentage && (
                                            <motion.div
                                                initial={{ opacity: 0, x: -20 }}
                                                animate={{ opacity: 1, x: 0 }}
                                                transition={{ delay: 0.8, duration: 0.3 }}
                                                className="text-sm"
                                            >
                                                <div className="font-medium text-gray-600">Процент износа</div>
                                                <div>{part.wear_percentage}%</div>
                                            </motion.div>
                                        )}

                                        {/* Category-specific fields */}
                                        {(part.category === 'Двигатель' || part.category === 'Трансмиссия') && (
                                            <>
                                                {part.engine_brand && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.85, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Марка двигателя</div>
                                                        <div>{part.engine_brand}</div>
                                                    </motion.div>
                                                )}
                                                {part.transmission && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.9, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Трансмиссия</div>
                                                        <div>{part.transmission}</div>
                                                    </motion.div>
                                                )}
                                                {part.drive && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.95, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Привод</div>
                                                        <div>{part.drive}</div>
                                                    </motion.div>
                                                )}
                                            </>
                                        )}

                                        {(part.category === 'Кузов' || part.category === 'Кузов снаружи' || part.category === 'Интерьер') && (
                                            <>
                                                {part.body_brand && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.85, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Марка кузова</div>
                                                        <div>{part.body_brand}</div>
                                                    </motion.div>
                                                )}
                                                {part.color && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.9, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Цвет</div>
                                                        <div>{part.color}</div>
                                                    </motion.div>
                                                )}
                                            </>
                                        )}

                                        {part.category === 'Шины и диски' && (
                                            <>
                                                {part.diameter && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.85, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Диаметр</div>
                                                        <div>{part.diameter}</div>
                                                    </motion.div>
                                                )}
                                                {part.width && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.9, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Ширина</div>
                                                        <div>{part.width}</div>
                                                    </motion.div>
                                                )}
                                                {part.profile && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 0.95, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Профиль</div>
                                                        <div>{part.profile}</div>
                                                    </motion.div>
                                                )}
                                                {part.tire_quantity && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.0, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Количество шин</div>
                                                        <div>{part.tire_quantity}</div>
                                                    </motion.div>
                                                )}
                                                {part.drilling && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.05, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Сверловка</div>
                                                        <div>{part.drilling}</div>
                                                    </motion.div>
                                                )}
                                                {part.offset && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.1, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Вылет</div>
                                                        <div>{part.offset}</div>
                                                    </motion.div>
                                                )}
                                                {part.center_hole_diameter && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.15, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Диаметр ЦО</div>
                                                        <div>{part.center_hole_diameter}</div>
                                                    </motion.div>
                                                )}
                                                {part.tire_model && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.2, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Модель шины</div>
                                                        <div>{part.tire_model}</div>
                                                    </motion.div>
                                                )}
                                                {part.season && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.25, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Сезон</div>
                                                        <div>{part.season}</div>
                                                    </motion.div>
                                                )}
                                            </>
                                        )}

                                        {/* Position fields for most categories */}
                                        {(part.category !== 'Автохимия и масла' && part.category !== 'Аксессуары и тюннинг' && part.category !== 'Другое') && (
                                            <>
                                                {part.front_rear && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.3, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Перед/зад</div>
                                                        <div>{part.front_rear}</div>
                                                    </motion.div>
                                                )}
                                                {part.left_right && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.35, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Право/лево</div>
                                                        <div>{part.left_right}</div>
                                                    </motion.div>
                                                )}
                                                {part.top_bottom && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.4, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Верх/низ</div>
                                                        <div>{part.top_bottom}</div>
                                                    </motion.div>
                                                )}
                                                {part.number && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.45, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Номер</div>
                                                        <div>{part.number}</div>
                                                    </motion.div>
                                                )}
                                            </>
                                        )}

                                        {/* Defect field for most categories */}
                                        {(part.category !== 'Автохимия и масла' && part.category !== 'Аксессуары и тюннинг') && (
                                            <>
                                                {part.defect && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.5, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Дефект</div>
                                                        <div>{part.defect}</div>
                                                    </motion.div>
                                                )}
                                            </>
                                        )}

                                        {/* Car release date for most categories */}
                                        {(part.category !== 'Автохимия и масла' && part.category !== 'Аксессуары и тюннинг' && part.category !== 'Другое') && (
                                            <>
                                                {part.car_release_date && (
                                                    <motion.div
                                                        initial={{ opacity: 0, x: -20 }}
                                                        animate={{ opacity: 1, x: 0 }}
                                                        transition={{ delay: 1.55, duration: 0.3 }}
                                                        className="text-sm"
                                                    >
                                                        <div className="font-medium text-gray-600">Дата выпуска автомобиля</div>
                                                        <div>{part.car_release_date}</div>
                                                    </motion.div>
                                                )}
                                            </>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </motion.div>
                    </motion.div>
                </AccordionContent>
            </AccordionItem>

            {/* Диалог редактирования */}
            <EditPartDialog
                partEdit={partEdit}
                part={part}
                onPhotoChange={handlePhotoChange}
                onCrop={(src) => {
                    if (src) {
                        setTempImageSrc(src);
                        setShowCropper(true);
                    }
                }}
                onDeletePhoto={async () => {
                    console.log('Удаление фото начато для детали:', part.id);
                    console.log('Текущее состояние part.photos:', part.photos);
                    console.log('Текущее photoPreview:', partEdit.photoUpload.photoPreview);
                    console.log('Есть ли новое фото (photoFile):', !!partEdit.photoUpload.photoFile);

                    // Если есть новое загруженное фото (не сохраненное), просто сбросить его
                    if (partEdit.photoUpload.photoFile) {
                        console.log('Сброс нового загруженного фото');
                        partEdit.photoUpload.resetPhoto();
                        partEdit.updateFormField('photo', '');
                        return;
                    }

                    // Иначе удаляем фото из базы данных
                    const photoToDelete = (part.photos && part.photos.length > 0) ? part.photos[0] : '';

                    console.log('photoToDelete (первое фото из массива):', photoToDelete);

                    if (!photoToDelete) {
                        alert('Нет фото для удаления');
                        return;
                    }

                    try {
                        const deleteUrl = `${API_BASE_URL}/api/deletepartphoto/${part.id}?photo=${encodeURIComponent(photoToDelete)}`;
                        console.log('Отправка запроса на удаление:', deleteUrl);

                        const response = await fetch(deleteUrl, {
                            method: 'DELETE',
                            headers: getAuthHeaders(),
                            credentials: 'include',
                        });

                        console.log('Ответ от сервера status:', response.status, 'ok:', response.ok);

                        if (!response.ok) {
                            const errorText = await response.text();
                            console.error('Ошибка ответа сервера:', errorText);
                            throw new Error(`HTTP ${response.status}: ${errorText}`);
                        }

                        const result = await response.json();
                        console.log('Фото успешно удалено, результат:', result);

                        // Обновить локальное состояние
                        partEdit.updateFormField('photo', '');
                        setPhotoUploadTimestamp(Date.now());
                        // Invalidate queries to refresh the UI
                        void queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
                    } catch (error) {
                        console.error('Ошибка при удалении фото:', error);
                        alert('Не удалось удалить фото. Попробуйте еще раз.');
                        throw error;
                    }
                }}
                originalFile={originalFile}
            />

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
        </motion.div>
    );
}

export default React.memo(PartBlock);