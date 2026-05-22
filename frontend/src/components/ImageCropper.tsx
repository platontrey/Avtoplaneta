/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useRef, useCallback, useEffect } from 'react';
import ReactCrop, { centerCrop, makeAspectCrop } from 'react-image-crop';
import type { Crop, PixelCrop } from 'react-image-crop';
import 'react-image-crop/dist/ReactCrop.css';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';

interface ImageCropperProps {
    src: string;
    onCropComplete: (croppedImageBlob: Blob) => void;
    onCancel: () => void;
    aspect?: number | null;
}

export default function ImageCropper({ src, onCropComplete, onCancel, aspect }: ImageCropperProps) {
    const [crop, setCrop] = useState<Crop>();
    const [completedCrop, setCompletedCrop] = useState<PixelCrop>();
    const [zoom, setZoom] = useState(1);
    const [position, setPosition] = useState({ x: 0, y: 0 });
    const [isDragging, setIsDragging] = useState(false);
    const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
    const [mode, setMode] = useState<'crop' | 'view' | 'draw'>('crop');
    const imgRef = useRef<HTMLImageElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const [isDrawing, setIsDrawing] = useState(false);
    const [lastPos, setLastPos] = useState({ x: 0, y: 0 });
    const [isSelecting, setIsSelecting] = useState(false);
    const [selection, setSelection] = useState<{start: {x: number, y: number}, end: {x: number, y: number}} | null>(null);

    useEffect(() => {
        const container = containerRef.current;
        if (container) {
            const handleWheelEvent = (e: WheelEvent) => {
                if (mode !== 'crop') {
                    e.preventDefault();
                    const delta = e.deltaY > 0 ? -0.05 : 0.05;
                    setZoom(prev => Math.max(0.1, Math.min(2, prev + delta)));
                }
            };

            container.addEventListener('wheel', handleWheelEvent, { passive: false });
            return () => container.removeEventListener('wheel', handleWheelEvent);
        }
    }, [mode]);


  const onImageLoad = useCallback((e: React.SyntheticEvent<HTMLImageElement>) => {
        const { width, height, naturalWidth, naturalHeight } = e.currentTarget;

        if (canvasRef.current) {
            canvasRef.current.width = naturalWidth;
            canvasRef.current.height = naturalHeight;
        }

        let crop;
        if (aspect !== null && aspect !== undefined) {
            crop = centerCrop(
                makeAspectCrop(
                    {
                        unit: '%',
                        width: 90,
                    },
                    aspect,
                    width,
                    height
                ),
                width,
                height
            );
        } else {
            crop = {
                unit: '%' as const,
                x: 5,
                y: 5,
                width: 90,
                height: 90,
            };
        }
        setCrop(crop);
    }, [aspect]);

    const cropImageWithWorker = useCallback(
        (image: HTMLImageElement, crop: PixelCrop): Promise<Blob> => {
            console.log('Starting worker for crop:', crop, 'image dimensions:', image.naturalWidth, 'x', image.naturalHeight);

            if (!window.OffscreenCanvas) {
                console.log('OffscreenCanvas not supported, falling back to main thread');
                return getCroppedImgFallback(image, crop);
            }

            return new Promise((resolve, reject) => {
                try {
                    const worker = new Worker(new URL('../workers/imageCropWorker.js', import.meta.url));
                    console.log('Worker created');

                    worker.postMessage({
                        imageSrc: src,
                        crop,
                        zoom,
                        position,
                        naturalWidth: image.naturalWidth,
                        naturalHeight: image.naturalHeight,
                        displayWidth: image.width,
                        displayHeight: image.height,
                    });
                    console.log('Message posted to worker');

                    worker.onmessage = function(e) {
                        console.log('Worker message received:', e.data);
                        const { success, blob, error } = e.data;
                        if (success) {
                            console.log('Cropped image blob created via worker, size:', blob.size, 'type:', blob.type);
                            resolve(blob);
                        } else {
                            console.error('Worker processing error:', error);
                            reject(new Error(error));
                        }
                        worker.terminate();
                    };

                    worker.onerror = function(error) {
                        console.error('Worker instantiation/error:', error);
                        reject(new Error('Worker failed'));
                        worker.terminate();
                    };
                } catch (error) {
                    console.error('Failed to create worker:', error);
                    reject(new Error('Worker creation failed'));
                }
            });
        },
        [src, zoom, position.x, position.y]
    );

    const getCroppedImgFallback = useCallback(
        (image: HTMLImageElement, crop: PixelCrop): Promise<Blob> => {
            console.log('Using fallback crop method');
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            if (!ctx) {
                throw new Error('No 2d context');
            }

            const adjustedCrop = {
                x: (crop.x - position.x) / zoom,
                y: (crop.y - position.y) / zoom,
                width: crop.width / zoom,
                height: crop.height / zoom,
            };

            const scaleX = image.naturalWidth / image.width;
            const scaleY = image.naturalHeight / image.height;

            canvas.width = adjustedCrop.width;
            canvas.height = adjustedCrop.height;

            ctx.drawImage(
                image,
                adjustedCrop.x * scaleX,
                adjustedCrop.y * scaleY,
                adjustedCrop.width * scaleX,
                adjustedCrop.height * scaleY,
                0,
                0,
                adjustedCrop.width,
                adjustedCrop.height
            );

            return new Promise((resolve, reject) => {
                canvas.toBlob((blob) => {
                    if (blob) {
                        console.log('Cropped image blob created via fallback, size:', blob.size, 'type:', blob.type);
                        resolve(blob);
                    } else {
                        reject(new Error('Failed to create blob from canvas'));
                    }
                }, 'image/jpeg', 0.95);
            });
        },
        [zoom, position.x, position.y]
    );

    const handleMouseDown = useCallback((e: React.MouseEvent) => {
        if (mode === 'view' && e.button === 0) {
            e.preventDefault();
            setIsDragging(true);
            setDragStart({ x: e.clientX - position.x, y: e.clientY - position.y });
        }
    }, [mode, position]);

    const startDrawing = useCallback((e: React.MouseEvent) => {
        if (mode !== 'draw') return;
        e.preventDefault();
        const canvas = canvasRef.current;
        if (!canvas || !imgRef.current) return;

        const img = imgRef.current;
        const canvasRect = canvas.getBoundingClientRect();

        const scaleX = img.naturalWidth / (img.width * zoom);
        const scaleY = img.naturalHeight / (img.height * zoom);

        // Координаты относительно canvas
        const x = (e.clientX - canvasRect.left) * scaleX;
        const y = (e.clientY - canvasRect.top) * scaleY;

        if (isDrawing) {
            // Continue drawing
            setLastPos({ x, y });
        } else {
            // Start potential selection
            setIsSelecting(true);
            setSelection({ start: { x, y }, end: { x, y } });
            setLastPos({ x, y });
        }
    }, [mode, zoom, isDrawing]);

    const draw = useCallback((e: React.MouseEvent) => {
        if (mode !== 'draw') return;
        e.preventDefault();
        const canvas = canvasRef.current;
        if (!canvas || !imgRef.current) return;

        const img = imgRef.current;
        const canvasRect = canvas.getBoundingClientRect();

        const scaleX = img.naturalWidth / (img.width * zoom);
        const scaleY = img.naturalHeight / (img.height * zoom);

        // Координаты относительно canvas
        const x = (e.clientX - canvasRect.left) * scaleX;
        const y = (e.clientY - canvasRect.top) * scaleY;

        if (isSelecting && selection) {
            // Update selection
            setSelection({ ...selection, end: { x, y } });
        } else if (isDrawing) {
            const ctx = canvas.getContext('2d');
            if (!ctx) return;
            ctx.beginPath();
            ctx.moveTo(lastPos.x, lastPos.y);
            ctx.lineTo(x, y);
            ctx.strokeStyle = 'red';
            ctx.lineWidth = 3;
            ctx.stroke();
            setLastPos({ x, y });
        }
    }, [mode, zoom, isSelecting, selection, isDrawing, lastPos]);

    const stopDrawing = useCallback(() => {
        if (isSelecting && selection) {
            const dx = selection.end.x - selection.start.x;
            const dy = selection.end.y - selection.start.y;
            const distance = Math.sqrt(dx * dx + dy * dy);
            if (distance < 5) {
                // Start drawing
                setIsDrawing(true);
                setIsSelecting(false);
                setSelection(null);
            } else {
                // Clear selection
                const canvas = canvasRef.current;
                if (canvas) {
                    const ctx = canvas.getContext('2d');
                    if (ctx) {
                        const x = Math.min(selection.start.x, selection.end.x);
                        const y = Math.min(selection.start.y, selection.end.y);
                        const w = Math.abs(selection.end.x - selection.start.x);
                        const h = Math.abs(selection.end.y - selection.start.y);
                        ctx.clearRect(x, y, w, h);
                    }
                }
                setIsSelecting(false);
                setSelection(null);
            }
        }
        setIsDrawing(false);
    }, [isSelecting, selection]);

    useEffect(() => {
        const handleGlobalMouseMove = (e: MouseEvent) => {
            if (isDragging) {
                setPosition({
                    x: e.clientX - dragStart.x,
                    y: e.clientY - dragStart.y,
                });
            }
        };

        const handleGlobalMouseUp = (e: MouseEvent) => {
            if (e.button === 0) {
                setIsDragging(false);
            }
        };

        if (isDragging) {
            document.addEventListener('mousemove', handleGlobalMouseMove);
            document.addEventListener('mouseup', handleGlobalMouseUp);
        }

        return () => {
            document.removeEventListener('mousemove', handleGlobalMouseMove);
            document.removeEventListener('mouseup', handleGlobalMouseUp);
        };
    }, [isDragging, dragStart]);

    const combineImageWithDrawing = useCallback(async (image: HTMLImageElement): Promise<Blob> => {
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        if (!ctx) throw new Error('No 2d context');

        canvas.width = image.naturalWidth;
        canvas.height = image.naturalHeight;

        image.crossOrigin = 'anonymous';

        ctx.drawImage(image, 0, 0);

        if (canvasRef.current) {
            ctx.drawImage(canvasRef.current, 0, 0);
        }

        return new Promise((resolve, reject) => {
            canvas.toBlob((blob) => {
                if (blob) {
                    resolve(blob);
                } else {
                    reject(new Error('Failed to create blob'));
                }
            }, 'image/jpeg', 0.95);
        });
    }, []);

    const handleEditComplete = useCallback(async () => {
        if (!imgRef.current) {
            console.error('No imgRef.current');
            return;
        }

        try {
            let resultBlob: Blob;
            if (mode === 'crop' && completedCrop) {
                if (completedCrop.width === 0 || completedCrop.height === 0) {
                    console.error('Crop dimensions are zero');
                    return;
                }
                resultBlob = await cropImageWithWorker(imgRef.current, completedCrop);
            } else {
                resultBlob = await combineImageWithDrawing(imgRef.current);
            }
            onCropComplete(resultBlob);
        } catch (error) {
            console.error('Error processing image:', error);
            alert('Ошибка при обработке изображения: ' + (error instanceof Error ? error.message : String(error)));
        }
    }, [mode, completedCrop, cropImageWithWorker, combineImageWithDrawing, onCropComplete]);

    return (
        <Dialog open={true} onOpenChange={onCancel}>
            <DialogContent className="max-w-6xl">
                <DialogHeader>
                    <DialogTitle>Обрезать изображение</DialogTitle>
                    <div className="flex gap-2 mt-2">
                        <Button
                            variant={mode === 'crop' ? "default" : "outline"}
                            size="sm"
                            onClick={() => setMode('crop')}
                        >
                            Режим обрезки
                        </Button>
                        <Button
                            variant={mode === 'view' ? "default" : "outline"}
                            size="sm"
                            onClick={() => setMode('view')}
                        >
                            Режим просмотра
                        </Button>
                        <Button
                            variant={mode === 'draw' ? "default" : "outline"}
                            size="sm"
                            onClick={() => setMode('draw')}
                        >
                            Режим рисования
                        </Button>
                    </div>
                </DialogHeader>
                <div
                    ref={containerRef}
                    className="flex justify-center overflow-hidden rounded-lg relative"
                    onMouseDown={handleMouseDown}
                    style={{ cursor: isDragging ? 'grabbing' : (mode === 'crop' ? 'default' : mode === 'view' ? 'grab' : 'crosshair'), maxHeight: '800px', backgroundColor: '#333' }}
                >
                  {mode === 'crop' ? (
                    <ReactCrop
                      crop={crop}
                      onChange={setCrop}
                      onComplete={setCompletedCrop}
                      {...(aspect !== null && aspect !== undefined ? { aspect } : {})}
                      className="max-w-full max-h-96"
                    >
                      <img
                        ref={imgRef}
                        src={src}
                        onLoad={onImageLoad}
                        alt="Crop preview"
                        className="max-w-full max-h-96 object-contain"
                        style={{
                          transform: `translate(${position.x}px, ${position.y}px) scale(${zoom})`,
                          transformOrigin: 'center center',
                          willChange: 'transform',
                          minWidth: '400px',
                          minHeight: '320px'
                        }}
                      />
                    </ReactCrop>
                  ) : (
                    <>
                      <img
                        ref={imgRef}
                        src={src}
                        onLoad={onImageLoad}
                        alt="Crop preview"
                        crossOrigin="anonymous"
                        className="max-w-full max-h-96 object-contain"
                        style={{
                          transform: `translate(${position.x}px, ${position.y}px) scale(${zoom})`,
                          transformOrigin: 'center center',
                          transition: isDragging ? 'none' : 'transform 0.1s ease-out',
                          willChange: 'transform',
                          minWidth: '400px',
                          minHeight: '320px',
                          pointerEvents: mode === 'draw' ? 'none' : 'auto'
                        }}
                      />
                      {mode === 'draw' && (
                        <>
                          <canvas
                            ref={canvasRef}
                            className="absolute pointer-events-auto z-10"
                            style={{
                              left: '50%',
                              top: '50%',
                              transform: `translate(-50%, -50%) translate(${position.x}px, ${position.y}px) scale(${zoom})`,
                              transformOrigin: 'center center',
                              willChange: 'transform',
                              width: imgRef.current?.width || 400,
                              height: imgRef.current?.height || 320
                            }}
                            onMouseDown={startDrawing}
                            onMouseMove={draw}
                            onMouseUp={stopDrawing}
                            onMouseLeave={stopDrawing}
                          />
                          {selection && (() => {
                            const canvas = canvasRef.current;
                            const container = containerRef.current;
                            if (!canvas || !container || !imgRef.current) return null;
                            const canvasRect = canvas.getBoundingClientRect();
                            const containerRect = container.getBoundingClientRect();
                            const img = imgRef.current;
                            const naturalWidth = img.naturalWidth;
                            const naturalHeight = img.naturalHeight;
                            const minX = Math.min(selection.start.x, selection.end.x);
                            const minY = Math.min(selection.start.y, selection.end.y);
                            const maxX = Math.max(selection.start.x, selection.end.x);
                            const maxY = Math.max(selection.start.y, selection.end.y);
                            const left = canvasRect.left - containerRect.left + (minX / naturalWidth) * canvasRect.width;
                            const top = canvasRect.top - containerRect.top + (minY / naturalHeight) * canvasRect.height;
                            const width = ((maxX - minX) / naturalWidth) * canvasRect.width;
                            const height = ((maxY - minY) / naturalHeight) * canvasRect.height;
                            return (
                              <div
                                className="absolute border-2 border-blue-500 bg-blue-200 bg-opacity-50 pointer-events-none z-20"
                                style={{
                                  left: `${left}px`,
                                  top: `${top}px`,
                                  width: `${width}px`,
                                  height: `${height}px`
                                }}
                              />
                            );
                          })()}
                        </>
                      )}
                    </>
                  )}
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={onCancel}>
                        Отмена
                    </Button>
                    <Button onClick={handleEditComplete} disabled={mode === 'crop' && !completedCrop}>
                        {mode === 'crop' ? 'Применить обрезку' : 'Применить изменения'}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}