/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { Dialog, DialogContent, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { 
    ZoomIn, 
    ZoomOut, 
    RotateCcw, 
    ChevronLeft, 
    ChevronRight, 
    X, 
    ExternalLink
} from 'lucide-react';
import { API_BASE_URL } from '@/lib/api';

interface PartPhotoViewerProps {
    isOpen: boolean;
    onClose: () => void;
    photos: string[];
    initialIndex?: number;
    partName?: string;
    partSubtitle?: string;
    uploadTimestamp?: number;
}

export default function PartPhotoViewer({
    isOpen,
    onClose,
    photos,
    initialIndex = 0,
    partName = 'Деталь',
    partSubtitle,
    uploadTimestamp
}: PartPhotoViewerProps) {
    const validPhotos = photos && photos.length > 0 ? photos : ['/placeholder-part.svg'];
    const [currentIndex, setCurrentIndex] = useState<number>(initialIndex);
    const [scale, setScale] = useState<number>(1);
    const [pan, setPan] = useState<{ x: number; y: number }>({ x: 0, y: 0 });
    const [isDragging, setIsDragging] = useState<boolean>(false);
    
    const viewportRef = useRef<HTMLDivElement>(null);
    const dragStartRef = useRef<{ x: number; y: number }>({ x: 0, y: 0 });
    const panStartRef = useRef<{ x: number; y: number }>({ x: 0, y: 0 });
    const touchDistanceRef = useRef<number | null>(null);
    const lastTapRef = useRef<number>(0);

    // Sync initialIndex when dialog opens
    useEffect(() => {
        if (isOpen) {
            const index = Math.max(0, Math.min(initialIndex, validPhotos.length - 1));
            setCurrentIndex(index);
            setScale(1);
            setPan({ x: 0, y: 0 });
        }
    }, [isOpen, initialIndex, validPhotos.length]);

    // Reset zoom when switching photo
    const changePhoto = useCallback((newIndex: number) => {
        if (newIndex >= 0 && newIndex < validPhotos.length) {
            setCurrentIndex(newIndex);
            setScale(1);
            setPan({ x: 0, y: 0 });
        }
    }, [validPhotos.length]);

    const handlePrev = useCallback(() => {
        if (currentIndex > 0) {
            changePhoto(currentIndex - 1);
        } else {
            changePhoto(validPhotos.length - 1);
        }
    }, [currentIndex, validPhotos.length, changePhoto]);

    const handleNext = useCallback(() => {
        if (currentIndex < validPhotos.length - 1) {
            changePhoto(currentIndex + 1);
        } else {
            changePhoto(0);
        }
    }, [currentIndex, validPhotos.length, changePhoto]);

    const handleZoomIn = useCallback(() => {
        setScale(prev => {
            const next = Math.min(prev * 1.35, 6);
            return Number(next.toFixed(2));
        });
    }, []);

    const handleZoomOut = useCallback(() => {
        setScale(prev => {
            const next = Math.max(prev / 1.35, 1);
            if (next === 1) {
                setPan({ x: 0, y: 0 });
            }
            return Number(next.toFixed(2));
        });
    }, []);

    const handleResetZoom = useCallback(() => {
        setScale(1);
        setPan({ x: 0, y: 0 });
    }, []);

    // Format photo URL
    const formatPhotoUrl = (path: string) => {
        if (!path || path === '/placeholder-part.svg') return '/placeholder-part.svg';
        if (path.startsWith('http://') || path.startsWith('https://') || path.startsWith('data:')) {
            return path;
        }
        const base = API_BASE_URL.endsWith('/') ? API_BASE_URL.slice(0, -1) : API_BASE_URL;
        const cleanPath = path.startsWith('/') ? path : `/${path}`;
        const url = `${base}${cleanPath}`;
        return uploadTimestamp ? `${url}?t=${uploadTimestamp}` : url;
    };

    const currentPhotoPath = validPhotos[currentIndex] || '';
    const currentPhotoUrl = formatPhotoUrl(currentPhotoPath);

    // Non-passive wheel zoom
    useEffect(() => {
        const el = viewportRef.current;
        if (!el || !isOpen) return;

        const handleWheel = (e: WheelEvent) => {
            e.preventDefault();
            const rect = el.getBoundingClientRect();
            const mouseX = e.clientX - rect.left - rect.width / 2;
            const mouseY = e.clientY - rect.top - rect.height / 2;

            const zoomFactor = e.deltaY < 0 ? 1.25 : 0.8;
            setScale(prevScale => {
                const nextScale = Math.min(Math.max(prevScale * zoomFactor, 1), 6);
                if (nextScale <= 1.01) {
                    setPan({ x: 0, y: 0 });
                    return 1;
                }
                setPan(prevPan => ({
                    x: mouseX - (mouseX - prevPan.x) * (nextScale / prevScale),
                    y: mouseY - (mouseY - prevPan.y) * (nextScale / prevScale),
                }));
                return Number(nextScale.toFixed(2));
            });
        };

        el.addEventListener('wheel', handleWheel, { passive: false });
        return () => el.removeEventListener('wheel', handleWheel);
    }, [isOpen]);

    // Keyboard navigation
    useEffect(() => {
        if (!isOpen) return;
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'ArrowLeft') {
                handlePrev();
            } else if (e.key === 'ArrowRight') {
                handleNext();
            } else if (e.key === '+' || e.key === '=') {
                handleZoomIn();
            } else if (e.key === '-') {
                handleZoomOut();
            } else if (e.key === '0') {
                handleResetZoom();
            } else if (e.key === 'Escape') {
                if (scale > 1.05) {
                    handleResetZoom();
                } else {
                    onClose();
                }
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [isOpen, scale, handlePrev, handleNext, handleZoomIn, handleZoomOut, handleResetZoom, onClose]);

    // Pointer drag for panning
    const handlePointerDown = (e: React.PointerEvent) => {
        if (scale <= 1) return;
        setIsDragging(true);
        dragStartRef.current = { x: e.clientX, y: e.clientY };
        panStartRef.current = { ...pan };
        try {
            (e.target as HTMLElement).setPointerCapture(e.pointerId);
        } catch {
            // ignore
        }
    };

    const handlePointerMove = (e: React.PointerEvent) => {
        if (!isDragging || scale <= 1) return;
        const dx = e.clientX - dragStartRef.current.x;
        const dy = e.clientY - dragStartRef.current.y;
        setPan({
            x: panStartRef.current.x + dx,
            y: panStartRef.current.y + dy,
        });
    };

    const handlePointerUp = (e: React.PointerEvent) => {
        setIsDragging(false);
        try {
            (e.target as HTMLElement).releasePointerCapture(e.pointerId);
        } catch {
            // ignore
        }
    };

    // Double click / tap to zoom
    const handleDoubleClick = (e: React.MouseEvent) => {
        if (scale > 1) {
            handleResetZoom();
        } else {
            const el = viewportRef.current;
            if (!el) return;
            const rect = el.getBoundingClientRect();
            const mouseX = e.clientX - rect.left - rect.width / 2;
            const mouseY = e.clientY - rect.top - rect.height / 2;
            const targetScale = 2.5;
            setScale(targetScale);
            setPan({
                x: -mouseX * (targetScale - 1),
                y: -mouseY * (targetScale - 1),
            });
        }
    };

    // Touch pinch-to-zoom
    const handleTouchStart = (e: React.TouchEvent) => {
        if (e.touches.length === 2) {
            const dist = Math.hypot(
                e.touches[0].clientX - e.touches[1].clientX,
                e.touches[0].clientY - e.touches[1].clientY
            );
            touchDistanceRef.current = dist;
        } else if (e.touches.length === 1) {
            // Double-tap detection
            const now = Date.now();
            if (now - lastTapRef.current < 300) {
                if (scale > 1) {
                    handleResetZoom();
                } else {
                    setScale(2.5);
                }
                lastTapRef.current = 0;
            } else {
                lastTapRef.current = now;
            }
        }
    };

    const handleTouchMove = (e: React.TouchEvent) => {
        if (e.touches.length === 2 && touchDistanceRef.current !== null) {
            const dist = Math.hypot(
                e.touches[0].clientX - e.touches[1].clientX,
                e.touches[0].clientY - e.touches[1].clientY
            );
            const ratio = dist / touchDistanceRef.current;
            setScale(prev => {
                const next = Math.min(Math.max(prev * ratio, 1), 6);
                if (next <= 1.01) {
                    setPan({ x: 0, y: 0 });
                    return 1;
                }
                return Number(next.toFixed(2));
            });
            touchDistanceRef.current = dist;
        }
    };

    const handleTouchEnd = () => {
        touchDistanceRef.current = null;
    };

    return (
        <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <DialogContent 
                showCloseButton={false}
                className="max-w-[98vw] sm:max-w-[98vw] w-[98vw] h-[95vh] p-0 border border-white/10 bg-neutral-950 text-white rounded-2xl overflow-hidden shadow-2xl flex flex-col justify-between select-none"
                onClick={(e) => e.stopPropagation()}
            >
                <DialogTitle className="sr-only">Просмотр фото детали</DialogTitle>
                <DialogDescription className="sr-only">Масштабирование и просмотр фотографий</DialogDescription>

                {/* Верхняя панель: название, счётчик и кнопки управления зумом */}
                <div className="flex items-center justify-between px-4 py-3 bg-neutral-900/90 backdrop-blur-md border-b border-white/10 z-20 shrink-0">
                    <div className="flex items-center gap-3 min-w-0 flex-1 mr-4">
                        <div className="flex flex-col min-w-0">
                            <span className="text-sm sm:text-base font-semibold truncate text-white">
                                {partName}
                            </span>
                            {partSubtitle && (
                                <span className="text-xs text-neutral-400 truncate">
                                    {partSubtitle}
                                </span>
                            )}
                        </div>
                        {validPhotos.length > 1 && (
                            <span className="shrink-0 text-xs px-2.5 py-1 rounded-full bg-neutral-800 text-neutral-300 border border-white/10 font-medium">
                                {currentIndex + 1} / {validPhotos.length}
                            </span>
                        )}
                    </div>

                    <div className="flex items-center gap-1 sm:gap-2 shrink-0">
                        {/* Зум - */}
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-neutral-300 hover:text-white hover:bg-neutral-800 rounded-lg"
                            onClick={handleZoomOut}
                            disabled={scale <= 1}
                            title="Уменьшить (-)"
                        >
                            <ZoomOut className="h-4 w-4" />
                        </Button>

                        {/* Индикатор масштаба */}
                        <button
                            type="button"
                            onClick={handleResetZoom}
                            className="px-2 py-1 text-xs font-mono font-medium rounded-md bg-neutral-800 hover:bg-neutral-700 text-neutral-200 border border-white/10 transition-colors"
                            title="Сбросить масштаб (0)"
                        >
                            {Math.round(scale * 100)}%
                        </button>

                        {/* Зум + */}
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-neutral-300 hover:text-white hover:bg-neutral-800 rounded-lg"
                            onClick={handleZoomIn}
                            disabled={scale >= 6}
                            title="Увеличить (+)"
                        >
                            <ZoomIn className="h-4 w-4" />
                        </Button>

                        {/* Сброс */}
                        {scale > 1 && (
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-neutral-300 hover:text-white hover:bg-neutral-800 rounded-lg"
                                onClick={handleResetZoom}
                                title="Сбросить масштаб (1:1)"
                            >
                                <RotateCcw className="h-4 w-4" />
                            </Button>
                        )}

                        {/* Открыть оригинал */}
                        {currentPhotoUrl && currentPhotoUrl !== '/placeholder-part.svg' && (
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-neutral-300 hover:text-white hover:bg-neutral-800 rounded-lg hidden sm:flex"
                                onClick={() => window.open(currentPhotoUrl, '_blank')}
                                title="Открыть оригинал в новой вкладке"
                            >
                                <ExternalLink className="h-4 w-4" />
                            </Button>
                        )}

                        {/* Закрыть */}
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-neutral-300 hover:text-red-400 hover:bg-neutral-800 rounded-lg ml-1"
                            onClick={onClose}
                            title="Закрыть (Esc)"
                        >
                            <X className="h-5 w-5" />
                        </Button>
                    </div>
                </div>

                {/* Основная рабочая область (Viewport для изображения) */}
                <div 
                    ref={viewportRef}
                    className="relative flex-1 w-full h-full overflow-hidden flex items-center justify-center bg-neutral-950 touch-none"
                    style={{
                        cursor: scale > 1 ? (isDragging ? 'grabbing' : 'grab') : 'zoom-in'
                    }}
                    onPointerDown={handlePointerDown}
                    onPointerMove={handlePointerMove}
                    onPointerUp={handlePointerUp}
                    onDoubleClick={handleDoubleClick}
                    onTouchStart={handleTouchStart}
                    onTouchMove={handleTouchMove}
                    onTouchEnd={handleTouchEnd}
                >
                    {/* Кнопка Предыдущее фото */}
                    {validPhotos.length > 1 && (
                        <button
                            type="button"
                            className="absolute left-3 top-1/2 -translate-y-1/2 z-30 p-2.5 sm:p-3 rounded-full bg-neutral-900/80 hover:bg-neutral-800 text-white border border-white/15 shadow-xl transition-transform active:scale-95"
                            onClick={(e) => {
                                e.stopPropagation();
                                handlePrev();
                            }}
                            title="Предыдущее фото (←)"
                        >
                            <ChevronLeft className="h-5 w-5 sm:h-6 sm:w-6" />
                        </button>
                    )}

                    {/* Контейнер увеличенного фото */}
                    <div 
                        className="transition-transform duration-75 will-change-transform flex items-center justify-center max-w-full max-h-full"
                        style={{
                            transform: `translate(${pan.x}px, ${pan.y}px) scale(${scale})`,
                            transition: isDragging ? 'none' : 'transform 0.15s ease-out'
                        }}
                    >
                        <img
                            src={currentPhotoUrl}
                            alt={`${partName} - фото ${currentIndex + 1}`}
                            className="max-h-[75vh] max-w-[92vw] object-contain rounded pointer-events-none drop-shadow-2xl"
                            draggable={false}
                            onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                e.currentTarget.onerror = null;
                                e.currentTarget.src = '/placeholder-part.svg';
                            }}
                        />
                    </div>

                    {/* Кнопка Следующее фото */}
                    {validPhotos.length > 1 && (
                        <button
                            type="button"
                            className="absolute right-3 top-1/2 -translate-y-1/2 z-30 p-2.5 sm:p-3 rounded-full bg-neutral-900/80 hover:bg-neutral-800 text-white border border-white/15 shadow-xl transition-transform active:scale-95"
                            onClick={(e) => {
                                e.stopPropagation();
                                handleNext();
                            }}
                            title="Следующее фото (→)"
                        >
                            <ChevronRight className="h-5 w-5 sm:h-6 sm:w-6" />
                        </button>
                    )}
                </div>

                {/* Нижняя панель: Лента миниатюр и подсказка */}
                <div className="px-4 py-2.5 bg-neutral-900/90 backdrop-blur-md border-t border-white/10 flex flex-col items-center gap-1.5 shrink-0 z-20">
                    {validPhotos.length > 1 && (
                        <div className="flex items-center gap-2 overflow-x-auto max-w-full py-1 px-2 scrollbar-thin">
                            {validPhotos.map((photo, idx) => {
                                const thumbUrl = formatPhotoUrl(photo);
                                const isActive = idx === currentIndex;
                                return (
                                    <button
                                        key={idx}
                                        type="button"
                                        onClick={() => changePhoto(idx)}
                                        className={`relative shrink-0 w-12 h-12 sm:w-14 sm:h-14 rounded-lg overflow-hidden border transition-all ${
                                            isActive 
                                                ? 'border-emerald-400 ring-2 ring-emerald-400/50 scale-105 shadow-md' 
                                                : 'border-white/20 opacity-60 hover:opacity-100 hover:border-white/50'
                                        }`}
                                    >
                                        <img
                                            src={thumbUrl}
                                            alt={`Миниатюра ${idx + 1}`}
                                            className="w-full h-full object-cover"
                                            onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                                e.currentTarget.onerror = null;
                                                e.currentTarget.src = '/placeholder-part.svg';
                                            }}
                                        />
                                    </button>
                                );
                            })}
                        </div>
                    )}
                    <span className="text-[11px] text-neutral-400 text-center">
                        Колёсико мыши: масштаб • Перетаскивание: перемещение • Двойной клик: приблизить • Esc: сброс/закрыть
                    </span>
                </div>
            </DialogContent>
        </Dialog>
    );
}
