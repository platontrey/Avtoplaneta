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
  aspect?: number | null; // aspect ratio, e.g., 1 for square, 16/9 for landscape, null for free
}

export default function ImageCropper({ src, onCropComplete, onCancel, aspect }: ImageCropperProps) {
  const [crop, setCrop] = useState<Crop>();
  const [completedCrop, setCompletedCrop] = useState<PixelCrop>();
  const [zoom, setZoom] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [isCropMode, setIsCropMode] = useState(true); // true for crop, false for view
  const imgRef = useRef<HTMLImageElement>(null);

  const onImageLoad = useCallback((e: React.SyntheticEvent<HTMLImageElement>) => {
    const { width, height } = e.currentTarget;
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
      // Free aspect ratio - create a crop that covers most of the image
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

  const getCroppedImg = useCallback(
    (image: HTMLImageElement, crop: PixelCrop): Promise<Blob> => {
      console.log('getCroppedImg called with crop:', crop, 'zoom:', zoom);
      console.log('Image dimensions: natural', image.naturalWidth, 'x', image.naturalHeight, 'display', image.width, 'x', image.height);

      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      if (!ctx) {
        throw new Error('No 2d context');
      }

      // Account for zoom and position in crop coordinates
      const adjustedCrop = {
        x: (crop.x - position.x) / zoom,
        y: (crop.y - position.y) / zoom,
        width: crop.width / zoom,
        height: crop.height / zoom,
      };

      const scaleX = image.naturalWidth / image.width;
      const scaleY = image.naturalHeight / image.height;
      console.log('Scale factors:', scaleX, scaleY, 'adjusted crop:', adjustedCrop);

      canvas.width = adjustedCrop.width;
      canvas.height = adjustedCrop.height;
      console.log('Canvas dimensions:', canvas.width, 'x', canvas.height);

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
      console.log('drawImage called with params:', adjustedCrop.x * scaleX, adjustedCrop.y * scaleY, adjustedCrop.width * scaleX, adjustedCrop.height * scaleY, 0, 0, adjustedCrop.width, adjustedCrop.height);

      return new Promise((resolve, reject) => {
        canvas.toBlob((blob) => {
          if (blob) {
            console.log('Cropped image blob created, size:', blob.size, 'type:', blob.type);
            resolve(blob);
          } else {
            console.error('canvas.toBlob returned null blob');
            reject(new Error('Failed to create blob from canvas'));
          }
        }, 'image/jpeg', 0.95);
      });
    },
    [zoom, position.x, position.y]
  );


  const handleWheel = useCallback((e: React.WheelEvent) => {
    if (!isCropMode) {
      e.preventDefault();
      const delta = e.deltaY > 0 ? -0.05 : 0.05;
      setZoom(prev => Math.max(0.1, Math.min(2, prev + delta)));
    }
  }, [isCropMode]);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (!isCropMode && e.button === 0) { // Left mouse button in view mode
      e.preventDefault();
      setIsDragging(true);
      setDragStart({ x: e.clientX - position.x, y: e.clientY - position.y });
    }
  }, [isCropMode, position]);


  // Global mouse listeners for dragging
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

  const handleCropComplete = useCallback(async () => {
    if (completedCrop && imgRef.current) {
      console.log('handleCropComplete: completedCrop:', completedCrop);
      if (completedCrop.width === 0 || completedCrop.height === 0) {
        console.error('Crop dimensions are zero');
        return;
      }
      try {
        const croppedImageBlob = await getCroppedImg(imgRef.current, completedCrop);
        onCropComplete(croppedImageBlob);
      } catch (error) {
        console.error('Error cropping image:', error);
        alert('Ошибка при обрезке изображения: ' + (error instanceof Error ? error.message : String(error)));
      }
    } else {
      console.error('No completedCrop or imgRef.current');
    }
  }, [completedCrop, getCroppedImg, onCropComplete]);

  return (
    <Dialog open={true} onOpenChange={onCancel}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Обрезать изображение</DialogTitle>
          <div className="flex gap-2 mt-2">
            <Button
              variant={isCropMode ? "default" : "outline"}
              size="sm"
              onClick={() => setIsCropMode(true)}
            >
              Режим обрезки
            </Button>
            <Button
              variant={!isCropMode ? "default" : "outline"}
              size="sm"
              onClick={() => setIsCropMode(false)}
            >
              Режим просмотра
            </Button>
          </div>
        </DialogHeader>
        <div
          className="flex justify-center overflow-hidden"
          onWheel={handleWheel}
          onMouseDown={handleMouseDown}
          style={{ cursor: isDragging ? 'grabbing' : (isCropMode ? 'default' : 'grab'), maxHeight: '400px' }}
        >
          {isCropMode ? (
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
                  willChange: 'transform'
                }}
              />
            </ReactCrop>
          ) : (
            <img
              ref={imgRef}
              src={src}
              onLoad={onImageLoad}
              alt="Crop preview"
              className="max-w-full max-h-96 object-contain"
              style={{
                transform: `translate(${position.x}px, ${position.y}px) scale(${zoom})`,
                transformOrigin: 'center center',
                transition: isDragging ? 'none' : 'transform 0.1s ease-out',
                willChange: 'transform'
              }}
            />
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>
            Отмена
          </Button>
          {isCropMode && (
            <Button onClick={handleCropComplete} disabled={!completedCrop}>
              Применить обрезку
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};