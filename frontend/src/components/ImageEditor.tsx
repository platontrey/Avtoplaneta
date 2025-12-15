/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useCallback, useEffect, useState } from 'react';
import ImageCropper from './ImageCropper';

interface ImageEditorProps {
    src: string | File;
    onEditComplete: (editedImageBlob: Blob) => void;
    onCancel: () => void;
    aspect?: number | null;
}

export default function ImageEditor({ src, onEditComplete, onCancel, aspect }: ImageEditorProps) {
    const [processedSrc, setProcessedSrc] = useState<string>('');

    // Convert src to data URL for ImageCropper
    useEffect(() => {
        console.log('ImageEditor received src:', src, 'type:', typeof src, 'instanceof File:', src instanceof File);
        if (src instanceof File) {
            console.log('Converting File to data URL');
            const reader = new FileReader();
            reader.onload = (e) => {
                const dataUrl = e.target?.result as string;
                console.log('File converted to data URL, length:', dataUrl.length);
                setProcessedSrc(dataUrl);
            };
            reader.readAsDataURL(src);
        } else if (typeof src === 'string') {
            if (src.startsWith('data:')) {
                console.log('Using data URL src directly:', src);
                setProcessedSrc(src);
            } else {
                console.log('Fetching URL to blob:', src);
                fetch(src, { mode: 'cors' })
                    .then(response => response.blob())
                    .then(blob => {
                        const reader = new FileReader();
                        reader.onload = (e) => {
                            const dataUrl = e.target?.result as string;
                            console.log('URL converted to data URL, length:', dataUrl.length);
                            setProcessedSrc(dataUrl);
                        };
                        reader.readAsDataURL(blob);
                    })
                    .catch(error => {
                        console.error('Failed to fetch image:', error);
                        // Fallback to direct URL
                        setProcessedSrc(src);
                    });
            }
        }
    }, [src]);

    const handleCropComplete = useCallback((croppedImageBlob: Blob) => {
        console.log('ImageEditor handleCropComplete called with blob size:', croppedImageBlob.size);
        onEditComplete(croppedImageBlob);
    }, [onEditComplete]);

    if (!processedSrc) {
        return null; // Or a loading component
    }

    return (
        <ImageCropper
            src={processedSrc}
            onCropComplete={handleCropComplete}
            onCancel={onCancel}
            aspect={aspect}
        />
    );
}
