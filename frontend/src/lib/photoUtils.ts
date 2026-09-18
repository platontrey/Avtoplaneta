/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

/**
 * Преобразует произвольный Blob (jpeg, webp и т.д.) в формат PNG.
 * W3C Clipboard API требует строгий MIME-тип 'image/png' для записи в системный буфер обмена.
 */
export async function convertBlobToPng(blob: Blob): Promise<Blob> {
    if (blob.type === 'image/png') return blob;

    return new Promise((resolve) => {
        const img = new Image();
        img.crossOrigin = 'anonymous';
        const url = URL.createObjectURL(blob);

        img.onload = () => {
            URL.revokeObjectURL(url);
            const canvas = document.createElement('canvas');
            canvas.width = img.naturalWidth || img.width;
            canvas.height = img.naturalHeight || img.height;
            const ctx = canvas.getContext('2d');
            if (!ctx) {
                resolve(blob);
                return;
            }
            ctx.drawImage(img, 0, 0);
            canvas.toBlob((pngBlob) => {
                resolve(pngBlob || blob);
            }, 'image/png');
        };

        img.onerror = () => {
            URL.revokeObjectURL(url);
            resolve(blob);
        };

        img.src = url;
    });
}

/**
 * Получает Blob из canvas
 */
export function canvasToBlob(canvas: HTMLCanvasElement, type = 'image/png', quality = 0.95): Promise<Blob> {
    return new Promise((resolve, reject) => {
        canvas.toBlob((blob) => {
            if (blob) resolve(blob);
            else reject(new Error('Не удалось экспортировать canvas в Blob'));
        }, type, quality);
    });
}

/**
 * Загружает изображение по URL в виде Blob
 */
export async function fetchImageBlob(imageUrl: string): Promise<Blob> {
    const res = await fetch(imageUrl, { mode: 'cors' });
    if (!res.ok) {
        throw new Error(`Ошибка загрузки изображения: HTTP ${res.status}`);
    }
    return await res.blob();
}

/**
 * Проверяет, поддерживает ли текущее устройство/браузер нативный шеринг файлов (Web Share API)
 */
export function canShareFiles(): boolean {
    if (typeof navigator === 'undefined' || !navigator.share || !navigator.canShare) {
        return false;
    }
    try {
        const testFile = new File(['test'], 'test.png', { type: 'image/png' });
        return navigator.canShare({ files: [testFile] });
    } catch {
        return false;
    }
}

/**
 * Копирует изображение в системный буфер обмена.
 * Возвращает 'image', если изображение скопировано как картинка,
 * или 'text', если был скопирован URL (fallback).
 */
export async function copyPhotoToClipboard(
    source: string | Blob | HTMLCanvasElement
): Promise<'image' | 'text'> {
    let pngBlob: Blob | null = null;
    let fallbackText: string | null = null;

    if (source instanceof HTMLCanvasElement) {
        pngBlob = await canvasToBlob(source, 'image/png');
    } else if (source instanceof Blob) {
        pngBlob = await convertBlobToPng(source);
    } else if (typeof source === 'string') {
        fallbackText = source;
        try {
            const rawBlob = await fetchImageBlob(source);
            pngBlob = await convertBlobToPng(rawBlob);
        } catch (fetchErr) {
            console.warn('Не удалось загрузить blob изображения для буфера, копируем URL:', fetchErr);
        }
    }

    // 1. Попытка записать бинарное изображение в буфер обмена
    if (pngBlob && typeof ClipboardItem !== 'undefined' && navigator.clipboard?.write) {
        try {
            const item = new ClipboardItem({ 'image/png': pngBlob });
            await navigator.clipboard.write([item]);
            return 'image';
        } catch (clipErr) {
            console.warn('Запись изображения в ClipboardItem не удалась, используем fallback:', clipErr);
        }
    }

    // 2. Fallback: копирование текстовой ссылки
    if (fallbackText && navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(fallbackText);
        return 'text';
    }

    throw new Error('Копирование в буфер обмена не поддерживается вашим браузером');
}

/**
 * Копирует прямую ссылку на фото в буфер обмена
 */
export async function copyPhotoUrlToClipboard(url: string): Promise<void> {
    if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(url);
        return;
    }
    const textArea = document.createElement('textarea');
    textArea.value = url;
    textArea.style.position = 'fixed';
    textArea.style.opacity = '0';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    document.execCommand('copy');
    document.body.removeChild(textArea);
}

/**
 * Открывает системное меню шеринга (iOS/Android) для отправки фото в мессенджеры/сохранения
 */
export async function sharePhoto(
    source: string | Blob | HTMLCanvasElement,
    options?: { title?: string; text?: string; filename?: string }
): Promise<boolean> {
    if (typeof navigator === 'undefined' || !navigator.share) {
        return false;
    }

    const title = options?.title || 'Фото детали';
    const text = options?.text || '';
    const filename = options?.filename || 'part-photo.png';

    try {
        let blob: Blob;
        if (source instanceof HTMLCanvasElement) {
            blob = await canvasToBlob(source, 'image/png');
        } else if (source instanceof Blob) {
            blob = source;
        } else {
            blob = await fetchImageBlob(source);
        }

        const ext = blob.type === 'image/jpeg' ? 'jpg' : 'png';
        const file = new File([blob], filename.endsWith(`.${ext}`) ? filename : `${filename}.${ext}`, {
            type: blob.type || 'image/png',
        });

        if (navigator.canShare && navigator.canShare({ files: [file] })) {
            await navigator.share({
                files: [file],
                title,
                text,
            });
            return true;
        } else {
            // Если файлы не поддерживаются, делимся ссылкой
            const url = typeof source === 'string' ? source : '';
            if (url) {
                await navigator.share({ title, text, url });
                return true;
            }
        }
    } catch (err) {
        if ((err as Error).name === 'AbortError') {
            return false; // Пользователь сам закрыл системное меню
        }
        console.warn('Ошибка при вызове navigator.share:', err);
    }

    return false;
}

/**
 * Скачивает изображение на устройство
 */
export async function downloadPhoto(
    source: string | Blob | HTMLCanvasElement,
    filename = 'part-photo.jpg'
): Promise<void> {
    let blob: Blob;

    if (source instanceof HTMLCanvasElement) {
        blob = await canvasToBlob(source, 'image/jpeg', 0.95);
    } else if (source instanceof Blob) {
        blob = source;
    } else {
        blob = await fetchImageBlob(source);
    }

    const objectUrl = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = objectUrl;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(objectUrl);
}
