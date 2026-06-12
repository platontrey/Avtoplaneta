/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useRef, useCallback, useEffect } from 'react';
import ReactCrop, { centerCrop, makeAspectCrop } from 'react-image-crop';
import type { Crop, PixelCrop } from 'react-image-crop';
import 'react-image-crop/dist/ReactCrop.css';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Slider } from '@/components/ui/slider';
import {
    Crop as CropIcon,
    RotateCw,
    RotateCcw,
    FlipHorizontal,
    FlipVertical,
    Brush,
    Eraser,
    Sparkles,
    Move,
    Undo2,
    Redo2,
    Check,
    RefreshCw,
    ZoomIn,
    ZoomOut
} from 'lucide-react';

interface ImageCropperProps {
    src: string;
    onCropComplete: (croppedImageBlob: Blob) => void;
    onCancel: () => void;
    aspect?: number | null;
}

type EditMode = 'view' | 'crop' | 'brush' | 'blur' | 'eraser' | 'pan';

interface HistoryState {
    main: string;  // DataURL
    clean: string; // DataURL
    blur: string;  // DataURL
}

const BRUSH_COLORS = [
    { value: '#EF4444', label: 'Красный' },
    { value: '#EAB308', label: 'Желтый' },
    { value: '#22C55E', label: 'Зеленый' },
    { value: '#3B82F6', label: 'Синий' },
    { value: '#FFFFFF', label: 'Белый' },
    { value: '#000000', label: 'Черный' },
];

export default function ImageCropper({ src, onCropComplete, onCancel, aspect: initialAspect }: ImageCropperProps) {
    const [mode, setMode] = useState<EditMode>('view');
    const [zoom, setZoom] = useState(1);
    const [pan, setPan] = useState({ x: 0, y: 0 });
    const [isPanning, setIsPanning] = useState(false);
    const [panStart, setPanStart] = useState({ x: 0, y: 0 });
    const [spacePressed, setSpacePressed] = useState(false);

    // Drawing settings
    const [brushColor, setBrushColor] = useState('#EF4444');
    const [brushSize, setBrushSize] = useState(10);
    const [blurSize, setBlurSize] = useState(25);
    const [eraserSize, setEraserSize] = useState(15);
    
    // Crop states
    const [crop, setCrop] = useState<Crop>();
    const [completedCrop, setCompletedCrop] = useState<PixelCrop>();
    const [aspectPreset, setAspectPreset] = useState<number | null | undefined>(initialAspect);

    // Drawing state
    const [isDrawing, setIsDrawing] = useState(false);
    const [lastPos, setLastPos] = useState({ x: 0, y: 0 });

    // History stacks for undo/redo
    const [history, setHistory] = useState<HistoryState[]>([]);
    const [redoStack, setRedoStack] = useState<HistoryState[]>([]);

    // Canvas refs
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const cleanCanvasRef = useRef<HTMLCanvasElement>(null);
    const blurCanvasRef = useRef<HTMLCanvasElement>(null);
    
    // Cached image natural dimensions
    const imgWidthRef = useRef<number>(0);
    const imgHeightRef = useRef<number>(0);

    // Track spacebar for panning shortcut
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.code === 'Space') {
                setSpacePressed(true);
                // Prevent scrolling page when space is pressed
                if (document.activeElement?.tagName !== 'INPUT' && document.activeElement?.tagName !== 'TEXTAREA') {
                    e.preventDefault();
                }
            }
        };
        const handleKeyUp = (e: KeyboardEvent) => {
            if (e.code === 'Space') {
                setSpacePressed(false);
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        window.addEventListener('keyup', handleKeyUp);
        return () => {
            window.removeEventListener('keydown', handleKeyDown);
            window.removeEventListener('keyup', handleKeyUp);
        };
    }, []);

    // Helper: Initialize all canvases with image data
    const initCanvases = useCallback((img: HTMLImageElement) => {
        const width = img.naturalWidth;
        const height = img.naturalHeight;
        imgWidthRef.current = width;
        imgHeightRef.current = height;

        // Setup main editing canvas
        const canvas = canvasRef.current;
        if (canvas) {
            canvas.width = width;
            canvas.height = height;
            const ctx = canvas.getContext('2d');
            ctx?.drawImage(img, 0, 0);
        }

        // Setup clean reference canvas (for eraser tool)
        const cleanCanvas = cleanCanvasRef.current;
        if (cleanCanvas) {
            cleanCanvas.width = width;
            cleanCanvas.height = height;
            const ctx = cleanCanvas.getContext('2d');
            ctx?.drawImage(img, 0, 0);
        }

        // Setup blurred reference canvas (for blur/redaction brush)
        const blurCanvas = blurCanvasRef.current;
        if (blurCanvas) {
            blurCanvas.width = width;
            blurCanvas.height = height;
            const ctx = blurCanvas.getContext('2d');
            if (ctx) {
                ctx.filter = 'blur(20px)';
                ctx.drawImage(img, 0, 0);
            }
        }

        setHistory([]);
        setRedoStack([]);
    }, []);

    // Load initial image
    useEffect(() => {
        if (!src) return;
        const img = new Image();
        img.crossOrigin = 'anonymous';
        img.onload = () => {
            initCanvases(img);
        };
        img.src = src;
    }, [src, initCanvases]);

    // Set crop aspect preset and update crop selection
    const applyAspectPreset = (preset: number | null) => {
        setAspectPreset(preset);
        const canvas = canvasRef.current;
        if (!canvas) return;

        const w = canvas.clientWidth;
        const h = canvas.clientHeight;

        let newCrop;
        if (preset !== null && preset !== undefined) {
            newCrop = centerCrop(
                makeAspectCrop(
                    {
                        unit: '%',
                        width: 90,
                    },
                    preset,
                    w,
                    h
                ),
                w,
                h
            );
        } else {
            newCrop = {
                unit: '%' as const,
                x: 5,
                y: 5,
                width: 90,
                height: 90,
            };
        }
        setCrop(newCrop);
    };

    // Initialize crop selection when entering crop mode
    useEffect(() => {
        if (mode === 'crop') {
            // Reset zoom/pan so user can crop cleanly
            setZoom(1);
            setPan({ x: 0, y: 0 });
            
            // Wait brief moment for canvas to finish layout transitions
            setTimeout(() => {
                applyAspectPreset(aspectPreset ?? null);
            }, 100);
        }
    }, [mode]);

    // Handle mouse wheel zoom
    const handleWheel = (e: React.WheelEvent) => {
        if (mode === 'crop') return; // Disable zoom in crop mode
        e.preventDefault();
        const delta = e.deltaY > 0 ? -0.05 : 0.05;
        setZoom(prev => Math.max(0.2, Math.min(3, prev + delta)));
    };

    // Save state helper for undo/redo
    const saveHistory = useCallback(() => {
        const canvas = canvasRef.current;
        const clean = cleanCanvasRef.current;
        const blur = blurCanvasRef.current;
        if (canvas && clean && blur) {
            const newState: HistoryState = {
                main: canvas.toDataURL('image/jpeg', 0.95),
                clean: clean.toDataURL('image/jpeg', 0.95),
                blur: blur.toDataURL('image/jpeg', 0.95),
            };
            setHistory(prev => [...prev, newState]);
            setRedoStack([]); // clear redo stack on new operation
        }
    }, []);

    // Load canvas states from history
    const loadState = (state: HistoryState) => {
        const loadToCanvas = (canvas: HTMLCanvasElement | null, dataUrl: string) => {
            if (!canvas) return;
            const ctx = canvas.getContext('2d');
            const img = new Image();
            img.onload = () => {
                canvas.width = img.width;
                canvas.height = img.height;
                ctx?.drawImage(img, 0, 0);
            };
            img.src = dataUrl;
        };

        loadToCanvas(canvasRef.current, state.main);
        loadToCanvas(cleanCanvasRef.current, state.clean);
        loadToCanvas(blurCanvasRef.current, state.blur);
    };

    // Undo action
    const handleUndo = () => {
        if (history.length === 0) return;
        const canvas = canvasRef.current;
        const clean = cleanCanvasRef.current;
        const blur = blurCanvasRef.current;
        if (!canvas || !clean || !blur) return;

        const previousState = history[history.length - 1];
        const currentState: HistoryState = {
            main: canvas.toDataURL('image/jpeg', 0.95),
            clean: clean.toDataURL('image/jpeg', 0.95),
            blur: blur.toDataURL('image/jpeg', 0.95),
        };

        setRedoStack(prev => [...prev, currentState]);
        setHistory(prev => prev.slice(0, -1));
        loadState(previousState);
    };

    // Redo action
    const handleRedo = () => {
        if (redoStack.length === 0) return;
        const canvas = canvasRef.current;
        const clean = cleanCanvasRef.current;
        const blur = blurCanvasRef.current;
        if (!canvas || !clean || !blur) return;

        const nextState = redoStack[redoStack.length - 1];
        const currentState: HistoryState = {
            main: canvas.toDataURL('image/jpeg', 0.95),
            clean: clean.toDataURL('image/jpeg', 0.95),
            blur: blur.toDataURL('image/jpeg', 0.95),
        };

        setHistory(prev => [...prev, currentState]);
        setRedoStack(prev => prev.slice(0, -1));
        loadState(nextState);
    };

    // Reset all changes
    const handleReset = () => {
        if (window.confirm('Вы уверены, что хотите сбросить все изменения?')) {
            const img = new Image();
            img.crossOrigin = 'anonymous';
            img.onload = () => {
                initCanvases(img);
                setZoom(1);
                setPan({ x: 0, y: 0 });
            };
            img.src = src;
        }
    };

    // Rotate canvases by 90 degrees (CW/CCW)
    const rotateCanvases = (clockwise = true) => {
        saveHistory();

        const rotateSingleCanvas = (canvas: HTMLCanvasElement | null) => {
            if (!canvas) return;
            const temp = document.createElement('canvas');
            temp.width = canvas.height;
            temp.height = canvas.width;
            const ctx = temp.getContext('2d');
            if (ctx) {
                ctx.translate(temp.width / 2, temp.height / 2);
                ctx.rotate((clockwise ? 90 : -90) * Math.PI / 180);
                ctx.drawImage(canvas, -canvas.width / 2, -canvas.height / 2);
            }
            canvas.width = temp.width;
            canvas.height = temp.height;
            const mainCtx = canvas.getContext('2d');
            mainCtx?.drawImage(temp, 0, 0);
        };

        rotateSingleCanvas(canvasRef.current);
        rotateSingleCanvas(cleanCanvasRef.current);
        rotateSingleCanvas(blurCanvasRef.current);
        
        // Swap bounds refs
        const w = imgWidthRef.current;
        imgWidthRef.current = imgHeightRef.current;
        imgHeightRef.current = w;
    };

    // Flip canvases (horizontal/vertical)
    const flipCanvases = (horizontal = true) => {
        saveHistory();

        const flipSingleCanvas = (canvas: HTMLCanvasElement | null) => {
            if (!canvas) return;
            const temp = document.createElement('canvas');
            temp.width = canvas.width;
            temp.height = canvas.height;
            const ctx = temp.getContext('2d');
            if (ctx) {
                ctx.translate(temp.width / 2, temp.height / 2);
                if (horizontal) {
                    ctx.scale(-1, 1);
                } else {
                    ctx.scale(1, -1);
                }
                ctx.drawImage(canvas, -canvas.width / 2, -canvas.height / 2);
            }
            const mainCtx = canvas.getContext('2d');
            mainCtx?.drawImage(temp, 0, 0);
        };

        flipSingleCanvas(canvasRef.current);
        flipSingleCanvas(cleanCanvasRef.current);
        flipSingleCanvas(blurCanvasRef.current);
    };

    // Map pointer coordinates to natural high-res image coordinate system
    const getCanvasCoords = (clientX: number, clientY: number) => {
        const canvas = canvasRef.current;
        if (!canvas) return null;
        
        const rect = canvas.getBoundingClientRect();
        if (rect.width === 0 || rect.height === 0) return null;

        // Find position inside the bounding rect, factoring zoom
        const x = (clientX - rect.left) * (canvas.width / rect.width);
        const y = (clientY - rect.top) * (canvas.height / rect.height);

        return { x, y };
    };

    // Render brush/eraser/blur path at original high-res coordinates
    const drawAtCoords = (x1: number, y1: number, x2: number, y2: number) => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        ctx.save();
        ctx.lineCap = 'round';
        ctx.lineJoin = 'round';

        if (mode === 'brush') {
            ctx.beginPath();
            ctx.strokeStyle = brushColor;
            ctx.lineWidth = brushSize;
            ctx.moveTo(x1, y1);
            ctx.lineTo(x2, y2);
            ctx.stroke();
        } else if (mode === 'blur') {
            const blurCanvas = blurCanvasRef.current;
            if (blurCanvas) {
                const dist = Math.hypot(x2 - x1, y2 - y1);
                const steps = Math.ceil(dist / 2);
                for (let i = 0; i <= steps; i++) {
                    const t = steps === 0 ? 1 : i / steps;
                    const cx = x1 + (x2 - x1) * t;
                    const cy = y1 + (y2 - y1) * t;
                    
                    ctx.save();
                    ctx.beginPath();
                    ctx.arc(cx, cy, blurSize / 2, 0, Math.PI * 2);
                    ctx.clip();
                    ctx.drawImage(blurCanvas, 0, 0);
                    ctx.restore();
                }
            }
        } else if (mode === 'eraser') {
            const cleanCanvas = cleanCanvasRef.current;
            if (cleanCanvas) {
                const dist = Math.hypot(x2 - x1, y2 - y1);
                const steps = Math.ceil(dist / 2);
                for (let i = 0; i <= steps; i++) {
                    const t = steps === 0 ? 1 : i / steps;
                    const cx = x1 + (x2 - x1) * t;
                    const cy = y1 + (y2 - y1) * t;
                    
                    ctx.save();
                    ctx.beginPath();
                    ctx.arc(cx, cy, eraserSize / 2, 0, Math.PI * 2);
                    ctx.clip();
                    ctx.drawImage(cleanCanvas, 0, 0);
                    ctx.restore();
                }
            }
        }

        ctx.restore();
    };

    // Mouse events on display canvas
    const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
        if (mode === 'view' || mode === 'pan' || spacePressed) {
            setIsPanning(true);
            setPanStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
            return;
        }

        if (mode === 'crop') return;

        // Prevent browser selection/drag ghosting
        e.preventDefault();

        const coords = getCanvasCoords(e.clientX, e.clientY);
        if (coords) {
            saveHistory();
            setIsDrawing(true);
            setLastPos(coords);
            drawAtCoords(coords.x, coords.y, coords.x, coords.y);
        }
    };

    const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
        if (isPanning) {
            setPan({
                x: e.clientX - panStart.x,
                y: e.clientY - panStart.y,
            });
            return;
        }

        if (!isDrawing) return;

        // Prevent browser selection/drag ghosting
        e.preventDefault();

        const coords = getCanvasCoords(e.clientX, e.clientY);
        if (coords && lastPos) {
            drawAtCoords(lastPos.x, lastPos.y, coords.x, coords.y);
            setLastPos(coords);
        }
    };

    const handleMouseUpOrLeave = () => {
        setIsDrawing(false);
        setIsPanning(false);
    };

    // Crop application logic
    const handleApplyCrop = () => {
        if (!completedCrop || !canvasRef.current) return;

        const canvas = canvasRef.current;
        const displayedWidth = canvas.clientWidth;
        const displayedHeight = canvas.clientHeight;

        const scaleX = canvas.width / displayedWidth;
        const scaleY = canvas.height / displayedHeight;

        const cropX = completedCrop.x * scaleX;
        const cropY = completedCrop.y * scaleY;
        const cropW = completedCrop.width * scaleX;
        const cropH = completedCrop.height * scaleY;

        if (cropW <= 0 || cropH <= 0) return;

        saveHistory();

        const performCrop = (c: HTMLCanvasElement) => {
            const temp = document.createElement('canvas');
            temp.width = cropW;
            temp.height = cropH;
            const ctx = temp.getContext('2d');
            ctx?.drawImage(c, cropX, cropY, cropW, cropH, 0, 0, cropW, cropH);
            return temp;
        };

        const croppedMain = performCrop(canvas);
        const croppedClean = cleanCanvasRef.current ? performCrop(cleanCanvasRef.current) : null;
        const croppedBlur = blurCanvasRef.current ? performCrop(blurCanvasRef.current) : null;

        const replaceContent = (target: HTMLCanvasElement, source: HTMLCanvasElement) => {
            target.width = source.width;
            target.height = source.height;
            const ctx = target.getContext('2d');
            ctx?.drawImage(source, 0, 0);
        };

        replaceContent(canvas, croppedMain);
        if (croppedClean && cleanCanvasRef.current) {
            replaceContent(cleanCanvasRef.current, croppedClean);
        }
        if (croppedBlur && blurCanvasRef.current) {
            replaceContent(blurCanvasRef.current, croppedBlur);
        }

        imgWidthRef.current = cropW;
        imgHeightRef.current = cropH;

        setCrop(undefined);
        setCompletedCrop(undefined);
        setMode('view');
    };

    // Finalize editing and export Blob
    const handleSave = () => {
        const canvas = canvasRef.current;
        if (canvas) {
            canvas.toBlob(
                (blob) => {
                    if (blob) {
                        onCropComplete(blob);
                    }
                },
                'image/jpeg',
                0.95
            );
        }
    };

    return (
        <Dialog open={true} onOpenChange={onCancel}>
            <DialogContent className="max-w-6xl w-full h-[90vh] flex flex-col bg-zinc-950 text-zinc-100 border-zinc-800 p-0 overflow-hidden">
                
                {/* Header with global options */}
                <DialogHeader className="p-4 border-b border-zinc-800 flex flex-row items-center justify-between space-y-0 h-16">
                    <DialogTitle className="text-lg font-semibold tracking-wide bg-gradient-to-r from-blue-400 to-indigo-400 bg-clip-text text-transparent">
                        Редактирование изображения
                    </DialogTitle>
                    
                    <div className="flex items-center gap-2">
                        <Button 
                            variant="ghost" 
                            size="icon" 
                            onClick={handleUndo} 
                            disabled={history.length === 0}
                            title="Отменить действие"
                            className="text-zinc-400 hover:text-zinc-200 disabled:opacity-30 disabled:hover:text-zinc-400 transition-all"
                        >
                            <Undo2 className="h-5 w-5" />
                        </Button>
                        <Button 
                            variant="ghost" 
                            size="icon" 
                            onClick={handleRedo} 
                            disabled={redoStack.length === 0}
                            title="Вернуть действие"
                            className="text-zinc-400 hover:text-zinc-200 disabled:opacity-30 disabled:hover:text-zinc-400 transition-all"
                        >
                            <Redo2 className="h-5 w-5" />
                        </Button>
                        <div className="w-px h-6 bg-zinc-800 mx-1" />
                        <Button 
                            variant="ghost" 
                            size="sm" 
                            onClick={handleReset}
                            className="text-rose-400 hover:text-rose-300 hover:bg-rose-950/20"
                        >
                            <RefreshCw className="h-4 w-4 mr-2" /> Сбросить все
                        </Button>
                    </div>
                </DialogHeader>

                {/* Main Workspace Area */}
                <div className="flex-1 flex overflow-hidden">
                    
                    {/* Left Toolbar (Mode selectors) */}
                    <div className="w-16 border-r border-zinc-800 bg-zinc-900/50 flex flex-col items-center py-4 gap-4">
                        <Button
                            variant={mode === 'view' ? 'default' : 'ghost'}
                            size="icon"
                            onClick={() => setMode('view')}
                            title="Режим просмотра и навигации"
                            className={`h-11 w-11 rounded-xl transition-all duration-200 ${mode === 'view' ? 'bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-600/20' : 'text-zinc-400 hover:text-zinc-200'}`}
                        >
                            <Move className="h-5 w-5" />
                        </Button>
                        <Button
                            variant={mode === 'crop' ? 'default' : 'ghost'}
                            size="icon"
                            onClick={() => setMode('crop')}
                            title="Обрезать изображение"
                            className={`h-11 w-11 rounded-xl transition-all duration-200 ${mode === 'crop' ? 'bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-600/20' : 'text-zinc-400 hover:text-zinc-200'}`}
                        >
                            <CropIcon className="h-5 w-5" />
                        </Button>
                        <Button
                            variant={mode === 'brush' ? 'default' : 'ghost'}
                            size="icon"
                            onClick={() => setMode('brush')}
                            title="Рисование маркером"
                            className={`h-11 w-11 rounded-xl transition-all duration-200 ${mode === 'brush' ? 'bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-600/20' : 'text-zinc-400 hover:text-zinc-200'}`}
                        >
                            <Brush className="h-5 w-5" />
                        </Button>
                        <Button
                            variant={mode === 'blur' ? 'default' : 'ghost'}
                            size="icon"
                            onClick={() => setMode('blur')}
                            title="Размытие / Цензурирование деталей"
                            className={`h-11 w-11 rounded-xl transition-all duration-200 ${mode === 'blur' ? 'bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-600/20' : 'text-zinc-400 hover:text-zinc-200'}`}
                        >
                            <Sparkles className="h-5 w-5" />
                        </Button>
                        <Button
                            variant={mode === 'eraser' ? 'default' : 'ghost'}
                            size="icon"
                            onClick={() => setMode('eraser')}
                            title="Ластик"
                            className={`h-11 w-11 rounded-xl transition-all duration-200 ${mode === 'eraser' ? 'bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-600/20' : 'text-zinc-400 hover:text-zinc-200'}`}
                        >
                            <Eraser className="h-5 w-5" />
                        </Button>
                    </div>

                    {/* Center Canvas Viewport */}
                    <div className="flex-1 bg-zinc-950 flex flex-col justify-between p-4 relative overflow-hidden">
                        
                        {/* Scale / Coordinate information bar */}
                        <div className="absolute top-4 left-6 z-10 bg-zinc-900/80 backdrop-blur border border-zinc-800/80 px-3 py-1.5 rounded-full text-xs text-zinc-400 font-mono shadow-md">
                            Размер: {imgWidthRef.current}x{imgHeightRef.current}px • Масштаб: {Math.round(zoom * 100)}%
                            {spacePressed && <span className="text-emerald-400 ml-2">• Навигация (Пробел)</span>}
                        </div>

                        {/* Viewport container */}
                        <div 
                            className="flex-1 flex items-center justify-center relative overflow-hidden" 
                            onWheel={handleWheel}
                        >
                            <div
                                style={{
                                    transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
                                    transformOrigin: 'center center',
                                    transition: isPanning ? 'none' : 'transform 0.12s ease-out',
                                    cursor: isPanning || spacePressed ? 'grabbing' : (mode === 'pan' || mode === 'view' ? 'grab' : 'crosshair'),
                                    willChange: 'transform'
                                }}
                            >
                                {mode === 'crop' ? (
                                    <ReactCrop
                                        crop={crop}
                                        onChange={setCrop}
                                        onComplete={setCompletedCrop}
                                        aspect={aspectPreset ?? undefined}
                                        className="max-w-full max-h-[50vh]"
                                    >
                                        <canvas 
                                            ref={canvasRef} 
                                            draggable={false}
                                            onDragStart={(e) => e.preventDefault()}
                                            className="max-w-[70vw] max-h-[55vh] shadow-2xl block select-none" 
                                            style={{ touchAction: 'none', userSelect: 'none', WebkitUserSelect: 'none' }}
                                        />
                                    </ReactCrop>
                                ) : (
                                    <canvas 
                                        ref={canvasRef} 
                                        draggable={false}
                                        onDragStart={(e) => e.preventDefault()}
                                        className="max-w-[70vw] max-h-[55vh] shadow-2xl block select-none"
                                        style={{ touchAction: 'none', userSelect: 'none', WebkitUserSelect: 'none' }}
                                        onMouseDown={handleMouseDown}
                                        onMouseMove={handleMouseMove}
                                        onMouseUp={handleMouseUpOrLeave}
                                        onMouseLeave={handleMouseUpOrLeave}
                                    />
                                )}
                            </div>
                        </div>

                        {/* Floating bottom zoom bar */}
                        <div className="flex justify-center items-center gap-3 py-2 z-10">
                            <Button 
                                variant="ghost" 
                                size="icon" 
                                onClick={() => setZoom(prev => Math.max(0.2, prev - 0.15))}
                                className="h-8 w-8 text-zinc-400 hover:text-zinc-200"
                            >
                                <ZoomOut className="h-4 w-4" />
                            </Button>
                            <Slider
                                value={[zoom * 100]}
                                onValueChange={(val) => setZoom(val[0] / 100)}
                                min={20}
                                max={300}
                                step={5}
                                className="w-40"
                            />
                            <Button 
                                variant="ghost" 
                                size="icon" 
                                onClick={() => setZoom(prev => Math.min(3, prev + 0.15))}
                                className="h-8 w-8 text-zinc-400 hover:text-zinc-200"
                            >
                                <ZoomIn className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>

                    {/* Right Contextual Control Menu */}
                    <div className="w-72 border-l border-zinc-800 bg-zinc-900/50 p-6 flex flex-col justify-between overflow-y-auto">
                        
                        {/* Action parameters based on active edit mode */}
                        <div className="space-y-6">
                            
                            {/* Mode: VIEW & PAN */}
                            {mode === 'view' && (
                                <div className="space-y-4">
                                    <h3 className="text-sm font-semibold text-zinc-300 uppercase tracking-wider">Инструменты поворота</h3>
                                    <div className="grid grid-cols-2 gap-2">
                                        <Button variant="outline" size="sm" onClick={() => rotateCanvases(false)} className="border-zinc-800 bg-zinc-900/40 hover:bg-zinc-800/80">
                                            <RotateCcw className="h-4 w-4 mr-2" /> -90°
                                        </Button>
                                        <Button variant="outline" size="sm" onClick={() => rotateCanvases(true)} className="border-zinc-800 bg-zinc-900/40 hover:bg-zinc-800/80">
                                            <RotateCw className="h-4 w-4 mr-2" /> +90°
                                        </Button>
                                    </div>
                                    <div className="grid grid-cols-2 gap-2 mt-2">
                                        <Button variant="outline" size="sm" onClick={() => flipCanvases(true)} className="border-zinc-800 bg-zinc-900/40 hover:bg-zinc-800/80" title="Отразить горизонтально">
                                            <FlipHorizontal className="h-4 w-4 mr-2" /> По гориз.
                                        </Button>
                                        <Button variant="outline" size="sm" onClick={() => flipCanvases(false)} className="border-zinc-800 bg-zinc-900/40 hover:bg-zinc-800/80" title="Отразить вертикально">
                                            <FlipVertical className="h-4 w-4 mr-2" /> По вертик.
                                        </Button>
                                    </div>
                                    <div className="bg-zinc-950/40 rounded-xl p-3 border border-zinc-800/60 mt-6 text-xs text-zinc-400 leading-relaxed">
                                        💡 В режиме просмотра вы можете перетаскивать изображение мышкой для навигации. Используйте колесико мыши или слайдер внизу для масштабирования.
                                    </div>
                                </div>
                            )}

                            {/* Mode: CROP */}
                            {mode === 'crop' && (
                                <div className="space-y-4">
                                    <h3 className="text-sm font-semibold text-zinc-300 uppercase tracking-wider">Пропорции обрезки</h3>
                                    <div className="flex flex-col gap-2">
                                        <Button 
                                            variant={aspectPreset === null ? 'default' : 'outline'} 
                                            size="sm" 
                                            onClick={() => applyAspectPreset(null)}
                                            className={aspectPreset === null ? 'bg-indigo-600 hover:bg-indigo-700' : 'border-zinc-800 bg-zinc-900/40 hover:bg-zinc-800'}
                                        >
                                            Свободная форма
                                        </Button>
                                        <Button 
                                            variant={aspectPreset === 1 ? 'default' : 'outline'} 
                                            size="sm" 
                                            onClick={() => applyAspectPreset(1)}
                                            className={aspectPreset === 1 ? 'bg-indigo-600 hover:bg-indigo-700' : 'border-zinc-800 bg-zinc-900/40 hover:bg-zinc-800'}
                                        >
                                            Квадрат (1:1)
                                        </Button>
                                        <Button 
                                            variant={aspectPreset === 4/3 ? 'default' : 'outline'} 
                                            size="sm" 
                                            onClick={() => applyAspectPreset(4/3)}
                                            className={aspectPreset === 4/3 ? 'bg-indigo-600 hover:bg-indigo-700' : 'border-zinc-800 bg-zinc-900/40 hover:bg-zinc-800'}
                                        >
                                            Формат 4:3
                                        </Button>
                                        <Button 
                                            variant={aspectPreset === 16/9 ? 'default' : 'outline'} 
                                            size="sm" 
                                            onClick={() => applyAspectPreset(16/9)}
                                            className={aspectPreset === 16/9 ? 'bg-indigo-600 hover:bg-indigo-700' : 'border-zinc-800 bg-zinc-900/40 hover:bg-zinc-800'}
                                        >
                                            Широкий экран (16:9)
                                        </Button>
                                    </div>

                                    <Button 
                                        onClick={handleApplyCrop} 
                                        disabled={!completedCrop} 
                                        className="w-full bg-emerald-600 hover:bg-emerald-700 font-semibold shadow-lg shadow-emerald-900/20 mt-6"
                                    >
                                        <Check className="h-4 w-4 mr-2" /> Подтвердить обрезку
                                    </Button>
                                </div>
                            )}

                            {/* Mode: BRUSH */}
                            {mode === 'brush' && (
                                <div className="space-y-4">
                                    <h3 className="text-sm font-semibold text-zinc-300 uppercase tracking-wider">Настройка маркера</h3>
                                    
                                    <div className="space-y-2">
                                        <label className="text-xs text-zinc-400">Толщина кисти: {brushSize}px</label>
                                        <Slider
                                            value={[brushSize]}
                                            onValueChange={(val) => setBrushSize(val[0])}
                                            min={2}
                                            max={50}
                                            step={1}
                                        />
                                    </div>

                                    <div className="space-y-2 pt-2">
                                        <label className="text-xs text-zinc-400">Цвет маркера:</label>
                                        <div className="grid grid-cols-3 gap-2">
                                            {BRUSH_COLORS.map(c => (
                                                <button
                                                    key={c.value}
                                                    onClick={() => setBrushColor(c.value)}
                                                    className={`h-9 rounded-lg border-2 flex items-center justify-center transition-all ${brushColor === c.value ? 'border-indigo-500 scale-105 shadow-md shadow-indigo-500/20' : 'border-zinc-800 hover:border-zinc-600'}`}
                                                    style={{ backgroundColor: c.value }}
                                                    title={c.label}
                                                >
                                                    {brushColor === c.value && (
                                                        <Check className={`h-4 w-4 ${c.value === '#FFFFFF' || c.value === '#EAB308' ? 'text-zinc-950' : 'text-white'}`} />
                                                    )}
                                                </button>
                                            ))}
                                            {/* Custom color picker palette */}
                                            <div className="relative h-9 rounded-lg border-2 border-zinc-800 hover:border-zinc-500 transition-all flex items-center justify-center bg-zinc-900/60 overflow-hidden cursor-pointer" title="Выбрать другой цвет">
                                                <input 
                                                    type="color" 
                                                    value={brushColor} 
                                                    onChange={(e) => setBrushColor(e.target.value)}
                                                    className="absolute inset-0 w-full h-full p-0 border-0 cursor-pointer opacity-0" 
                                                />
                                                <div 
                                                    className={`h-5 w-5 rounded-md border ${!BRUSH_COLORS.some(c => c.value === brushColor) ? 'border-indigo-500 scale-105' : 'border-zinc-700'}`} 
                                                    style={{ 
                                                        backgroundColor: !BRUSH_COLORS.some(c => c.value === brushColor) ? brushColor : 'transparent',
                                                        backgroundImage: BRUSH_COLORS.some(c => c.value === brushColor) ? 'linear-gradient(to right, red, orange, yellow, green, blue, indigo, violet)' : 'none'
                                                    }} 
                                                />
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <div className="bg-zinc-950/40 rounded-xl p-3 border border-zinc-800/60 mt-4 text-xs text-zinc-400 leading-relaxed">
                                        💡 Удерживайте Пробел, чтобы временно включить перетаскивание при рисовании.
                                    </div>
                                </div>
                            )}

                            {/* Mode: BLUR */}
                            {mode === 'blur' && (
                                <div className="space-y-4">
                                    <h3 className="text-sm font-semibold text-zinc-300 uppercase tracking-wider">Цензура / Размытие</h3>
                                    
                                    <div className="space-y-2">
                                        <label className="text-xs text-zinc-400">Размер кисти размытия: {blurSize}px</label>
                                        <Slider
                                            value={[blurSize]}
                                            onValueChange={(val) => setBlurSize(val[0])}
                                            min={5}
                                            max={80}
                                            step={1}
                                        />
                                    </div>
                                    
                                    <div className="bg-zinc-950/40 rounded-xl p-3 border border-zinc-800/60 mt-4 text-xs text-zinc-400 leading-relaxed">
                                        🕵️‍♂️ Замазывайте кистью номера, надписи, ценники или дефекты. Область будет автоматически размыта.
                                    </div>
                                </div>
                            )}

                            {/* Mode: ERASER */}
                            {mode === 'eraser' && (
                                <div className="space-y-4">
                                    <h3 className="text-sm font-semibold text-zinc-300 uppercase tracking-wider">Ластик восстановления</h3>
                                    
                                    <div className="space-y-2">
                                        <label className="text-xs text-zinc-400">Размер ластика: {eraserSize}px</label>
                                        <Slider
                                            value={[eraserSize]}
                                            onValueChange={(val) => setEraserSize(val[0])}
                                            min={5}
                                            max={60}
                                            step={1}
                                        />
                                    </div>
                                    
                                    <div className="bg-zinc-950/40 rounded-xl p-3 border border-zinc-800/60 mt-4 text-xs text-zinc-400 leading-relaxed">
                                        🧯 Ластик восстанавливает исходную фотографию под проведенной кистью (стирает нарисованные маркеры и размытие).
                                    </div>
                                </div>
                            )}
                        </div>

                        {/* Hidden reference canvases used as backing filters */}
                        <canvas ref={cleanCanvasRef} className="hidden" />
                        <canvas ref={blurCanvasRef} className="hidden" />

                        {/* Right sidebar bottom save controls */}
                        <div className="space-y-2 pt-4 border-t border-zinc-800">
                            <Button 
                                variant="outline" 
                                onClick={onCancel} 
                                className="w-full border-zinc-800 bg-zinc-900/30 text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100"
                            >
                                Отмена
                            </Button>
                            <Button 
                                onClick={handleSave} 
                                className="w-full bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white font-semibold shadow-lg shadow-indigo-900/30"
                            >
                                Сохранить
                            </Button>
                        </div>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}