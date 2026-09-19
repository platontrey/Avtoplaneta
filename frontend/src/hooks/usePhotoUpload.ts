/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState, useCallback, useEffect } from 'react';
import { getAuthHeaders } from '@/lib/csrf';
import { API_BASE_URL } from '@/lib/api';

interface UsePhotoUploadOptions {
  initialPhoto?: string;
  onPhotoUpload?: (photoPath: string) => void;
  onUploadComplete?: () => void;
}

interface UsePhotoUploadReturn {
  currentPhoto: string | undefined;
  photoFile: File | null;
  photoPreview: string | null;
  isUploading: boolean;
  uploadTimestamp: number;
  forceRefresh: number;
  handlePhotoChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  uploadPhoto: (partId: number) => Promise<string | null>;
  resetPhoto: () => void;
  updateCurrentPhoto: (photoPath: string) => void;
}

/**
 * Custom hook for handling photo upload functionality
 * Separates photo logic from display components
 */
export function usePhotoUpload(options: UsePhotoUploadOptions = {}): UsePhotoUploadReturn {
  const { initialPhoto, onPhotoUpload, onUploadComplete } = options;

  const [currentPhoto, setCurrentPhoto] = useState<string | undefined>(initialPhoto);
  const [photoFile, setPhotoFile] = useState<File | null>(null);
  const [photoPreview, setPhotoPreview] = useState<string | null>(
    initialPhoto ? `${API_BASE_URL}${initialPhoto}?t=${Date.now()}` : null
  );
  const [isUploading, setIsUploading] = useState(false);
  const [uploadTimestamp, setUploadTimestamp] = useState<number>(Date.now());
  const [forceRefresh, setForceRefresh] = useState<number>(0);

  // Update currentPhoto and photoPreview when initialPhoto changes
  useEffect(() => {
    console.log('usePhotoUpload: initialPhoto changed to:', initialPhoto, 'API_BASE_URL:', API_BASE_URL);
    setCurrentPhoto(initialPhoto);
    const preview = initialPhoto ? `${API_BASE_URL}${initialPhoto}?t=${Date.now()}` : null;
    console.log('usePhotoUpload: setting photoPreview to:', preview);
    setPhotoPreview(preview);
  }, [initialPhoto]);

  // Cleanup object URL on unmount to prevent memory leaks
  useEffect(() => {
    return () => {
      if (photoPreview && photoPreview.startsWith('blob:')) {
        URL.revokeObjectURL(photoPreview);
      }
    };
  }, [photoPreview]);

  /**
   * Handles photo file selection with validation
   */
  const handlePhotoChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    console.log('usePhotoUpload handlePhotoChange: file selected:', file?.name, 'size:', file?.size, 'type:', file?.type);
    if (!file) return;

    // Validate file type - only allow images
    if (!file.type.startsWith('image/')) {
      alert('Please select a valid image file.');
      return;
    }

    // Validate file size (max 5MB to prevent memory issues)
    const maxSize = 5 * 1024 * 1024; // 5MB
    if (file.size > maxSize) {
      alert('File size must be less than 5MB.');
      return;
    }

    console.log('usePhotoUpload handlePhotoChange: file name without extension:', file.name.replace(/\.[^/.]+$/, ""));

    // Clean up previous object URL to prevent memory leaks
    if (photoPreview && photoPreview.startsWith('blob:')) {
      URL.revokeObjectURL(photoPreview);
    }

    // Create object URL for better performance than base64 encoding
    const objectUrl = URL.createObjectURL(file);
    setPhotoFile(file);
    setPhotoPreview(objectUrl);
  }, [photoPreview]);

  /**
   * Uploads the selected photo to the server
   */
  const uploadPhoto = useCallback(async (partId: number): Promise<string | null> => {
    if (!photoFile) {
      console.log('No photo file selected for upload');
      return null;
    }

    console.log('Starting photo upload for part:', partId, 'File:', photoFile.name, 'Size:', photoFile.size, 'Type:', photoFile.type);
    console.log('File object details:', photoFile);
    setIsUploading(true);
    const photoFormData = new FormData();
    photoFormData.append("photo", photoFile);

    // Log FormData contents
    console.log('FormData contents:');
    for (const [key, value] of photoFormData.entries()) {
      console.log(`${key}:`, value, 'Type:', value instanceof File ? 'File' : typeof value);
      if (value instanceof File) {
        console.log('File details:', { name: value.name, size: value.size, type: value.type, lastModified: value.lastModified });
      }
    }

    const csrfToken = getAuthHeaders()['X-CSRF-Token'] || '';
    const uploadUrl = `${API_BASE_URL}/api/v1/uploadpartphoto/${partId}`;

    console.log('Photo upload URL:', uploadUrl);
    console.log('CSRF Token present:', !!csrfToken, 'Token length:', csrfToken.length);

    try {
      console.log('Making photo upload request...');
      const photoResponse = await fetch(uploadUrl, {
        method: "POST",
        // Don't set Content-Type header for FormData - let browser set it with boundary
        headers: {
          'X-CSRF-Token': csrfToken,
        },
        credentials: 'include',
        body: photoFormData,
      });

      console.log('Photo upload response status:', photoResponse.status, 'OK:', photoResponse.ok);
      console.log('Response headers:', Object.fromEntries(photoResponse.headers.entries()));

      if (photoResponse.ok) {
        const result = await photoResponse.json();
        console.log('DEBUG: Photo upload successful, result:', result);
        console.log('DEBUG: Setting currentPhoto to:', result.photo);
        setCurrentPhoto(result.photo);
        console.log('DEBUG: currentPhoto set, calling onPhotoUpload with:', result.photo);
        onPhotoUpload?.(result.photo);
        // Force refresh of photo preview to bypass cache
        setUploadTimestamp(Date.now());
        setForceRefresh(prev => prev + 1);
        // Call the upload complete callback
        onUploadComplete?.();
        return result.photo;
      } else {
        const errorText = await photoResponse.text();
        console.error("Photo upload failed with status:", photoResponse.status, "Error:", errorText);
        alert('Failed to upload photo. Please try again.');
        return null;
      }
    } catch (error) {
      console.error('Network error during photo upload:', error);
      alert('Failed to upload photo. Please try again.');
      return null;
    } finally {
      setIsUploading(false);
    }
  }, [photoFile, onPhotoUpload, onUploadComplete]);

  /**
   * Updates the current photo path
   */
  const updateCurrentPhoto = useCallback((photoPath: string) => {
    console.log('updateCurrentPhoto called with:', photoPath);
    setCurrentPhoto(photoPath);
    setPhotoPreview(`${API_BASE_URL}${photoPath}?t=${Date.now()}`);
    setUploadTimestamp(Date.now());
    setForceRefresh(prev => prev + 1);
  }, []);

  /**
   * Resets the photo state
   */
  const resetPhoto = useCallback(() => {
    if (photoPreview && photoPreview.startsWith('blob:')) {
      URL.revokeObjectURL(photoPreview);
    }
    setPhotoFile(null);
    setPhotoPreview(initialPhoto ? `${API_BASE_URL}${initialPhoto}?t=${Date.now()}` : null);
    setCurrentPhoto(initialPhoto);
  }, [photoPreview, initialPhoto]);

  return {
    currentPhoto,
    photoFile,
    photoPreview,
    isUploading,
    uploadTimestamp,
    forceRefresh,
    handlePhotoChange,
    uploadPhoto,
    resetPhoto,
    updateCurrentPhoto,
  };
}