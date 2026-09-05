/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState, useCallback, useEffect } from 'react';
import { useUpdatePart } from './useParts';
import { usePhotoUpload } from './usePhotoUpload';
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/features/parts/hooks/usePartsQueries';
import { logUserActivity } from '@/features/admin/api/adminApi';
import type { Part } from '@/features/parts/types';

interface UsePartEditOptions {
  initialPart: Part;
}

export interface UsePartEditReturn {
  editForm: {
    name: string;
    quantity: string;
    description: string;
    category: string;
    price: string;
    salesman: string;
    seller_id?: number;
    location: string;
    address: string;
    status: boolean;
    brand: string;
    model: string;
    vin: string;
    photo?: string;
    // Характеристики запчасти
    body_brand?: string;
    engine_brand?: string;
    car_release_date?: string;
    front_rear?: string;
    left_right?: string;
    top_bottom?: string;
    number?: string;
    manufacturer?: string;
    manufacturer_code?: string;
    oem_code?: string;
    color?: string;
    condition?: string;
    supplier_code?: string;
    defect?: string;
    transmission?: string;
    transmission_model?: string;
    drive?: string;
    wear_percentage?: string;
    season?: string;
    diameter?: string;
    width?: string;
    profile?: string;
    tire_quantity?: string;
    drilling?: string;
    offset?: string;
    center_hole_diameter?: string;
    tire_model?: string;
  };
  isEditing: boolean;
  setIsEditing: (editing: boolean) => void;
  handleEdit: () => Promise<void>;
  updateFormField: (field: keyof UsePartEditReturn['editForm'], value: string | boolean | number | undefined) => void;
  isLoading: boolean;
  photoUpload: ReturnType<typeof usePhotoUpload>;
  uploadPhoto?: (partId: number) => Promise<string | null>;
  shouldDeletePhoto: boolean;
  setShouldDeletePhoto: (should: boolean) => void;
}

/**
 * Custom hook for handling part editing logic with React Query
 * Separates edit logic from display components
 */
export function usePartEdit(options: UsePartEditOptions): UsePartEditReturn {
  const { initialPart } = options;

  const [isEditing, setIsEditing] = useState(false);
  const [shouldDeletePhoto, setShouldDeletePhoto] = useState(false);
  const [editForm, setEditForm] = useState({
    name: initialPart.name || '',
    quantity: initialPart.quantity?.toString() || '0',
    description: initialPart.description || '',
    category: initialPart.category || '',
    price: (initialPart.price || 0).toString(),
    salesman: initialPart.salesman || '',
    seller_id: initialPart.seller_id,
    location: initialPart.location || '',
    address: initialPart.address || '',
    status: initialPart.status ?? true,
    brand: initialPart.brand || '',
    model: initialPart.model || '',
    vin: initialPart.vin || initialPart.body_brand || '',
    photo: initialPart.photo || '',
    // Характеристики запчасти
    body_brand: initialPart.body_brand || '',
    engine_brand: initialPart.engine_brand || '',
    car_release_date: initialPart.car_release_date || '',
    front_rear: initialPart.front_rear || '',
    left_right: initialPart.left_right || '',
    top_bottom: initialPart.top_bottom || '',
    number: initialPart.number || '',
    manufacturer: initialPart.manufacturer || '',
    manufacturer_code: initialPart.manufacturer_code || '',
    oem_code: initialPart.oem_code || '',
    color: initialPart.color || '',
    condition: initialPart.condition || '',
    supplier_code: initialPart.supplier_code || '',
    defect: initialPart.defect || '',
    transmission: initialPart.transmission || '',
    transmission_model: initialPart.transmission_model || '',
    drive: initialPart.drive || '',
    wear_percentage: initialPart.wear_percentage || '',
    season: initialPart.season || '',
    diameter: initialPart.diameter || '',
    width: initialPart.width || '',
    profile: initialPart.profile || '',
    tire_quantity: initialPart.tire_quantity || '',
    drilling: initialPart.drilling || '',
    offset: initialPart.offset || '',
    center_hole_diameter: initialPart.center_hole_diameter || '',
    tire_model: initialPart.tire_model || '',
  });

  // Update editForm when initialPart changes
  useEffect(() => {
    setEditForm({
      name: initialPart.name || '',
      quantity: initialPart.quantity?.toString() || '0',
      description: initialPart.description || '',
      category: initialPart.category || '',
      price: (initialPart.price || 0).toString(),
      salesman: initialPart.salesman || '',
      seller_id: initialPart.seller_id,
      location: initialPart.location || '',
      address: initialPart.address || '',
      status: initialPart.status ?? true,
      brand: initialPart.brand || '',
      model: initialPart.model || '',
      vin: initialPart.vin || initialPart.body_brand || '',
      photo: initialPart.photo || '',
      // Характеристики запчасти
      body_brand: initialPart.body_brand || '',
      engine_brand: initialPart.engine_brand || '',
      car_release_date: initialPart.car_release_date || '',
      front_rear: initialPart.front_rear || '',
      left_right: initialPart.left_right || '',
      top_bottom: initialPart.top_bottom || '',
      number: initialPart.number || '',
      manufacturer: initialPart.manufacturer || '',
      manufacturer_code: initialPart.manufacturer_code || '',
      oem_code: initialPart.oem_code || '',
      color: initialPart.color || '',
      condition: initialPart.condition || '',
      supplier_code: initialPart.supplier_code || '',
      defect: initialPart.defect || '',
      transmission: initialPart.transmission || '',
      transmission_model: initialPart.transmission_model || '',
      drive: initialPart.drive || '',
      wear_percentage: initialPart.wear_percentage || '',
      season: initialPart.season || '',
      diameter: initialPart.diameter || '',
      width: initialPart.width || '',
      profile: initialPart.profile || '',
      tire_quantity: initialPart.tire_quantity || '',
      drilling: initialPart.drilling || '',
      offset: initialPart.offset || '',
      center_hole_diameter: initialPart.center_hole_diameter || '',
      tire_model: initialPart.tire_model || '',
    });
  }, [initialPart]);

  const queryClient = useQueryClient();
  const updatePartMutation = useUpdatePart();
  const photoUpload = usePhotoUpload({
    initialPhoto: initialPart.photo,
    onUploadComplete: () => {
      // Invalidate parts list cache when photo upload completes
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
      queryClient.refetchQueries({ queryKey: partsKeys.lists() });
    },
  });

  /**
   * Updates a specific form field
   */
  const updateFormField = useCallback((field: keyof UsePartEditReturn['editForm'], value: string | boolean | number | undefined) => {
    setEditForm(prev => ({
      ...prev,
      [field]: value
    }));
  }, []);

  /**
   * Handles the edit operation for a part
   */
  const handleEdit = useCallback(async () => {
    try {
      // Validate required fields
      if (!editForm.name.trim()) {
        alert('Part name is required.');
        return;
      }
      if (!editForm.category) {
        alert('Category is required.');
        return;
      }
      const quantity = parseInt(editForm.quantity);
      if (isNaN(quantity) || quantity < 0) {
        alert('Quantity must be at least 0.');
        return;
      }
      const price = parseFloat(editForm.price);
      if (isNaN(price) || price < 0) {
        alert('Price must be a valid number >= 0.');
        return;
      }

      // Upload photo if a new file is selected
      let photoPath = editForm.photo || '';
      console.log('DEBUG: Initial photoPath from editForm.photo:', photoPath);
      if (photoUpload.photoFile) {
        console.log('DEBUG: Uploading new photo before saving part, photoFile present');
        const newPhotoPath = await photoUpload.uploadPhoto(initialPart.id);
        console.log('DEBUG: New photo path from upload:', newPhotoPath);
        if (!newPhotoPath) {
          alert('Failed to upload photo. Please try again.');
          return;
        }
        photoPath = newPhotoPath;
        // Update editForm.photo to reflect the new photo
        setEditForm(prev => ({ ...prev, photo: newPhotoPath }));
        console.log('DEBUG: Final photoPath after upload:', photoPath);
      } else {
        console.log('DEBUG: No new photo file selected, using existing photoPath');
      }

      const updatedPart: Partial<Part> = {
        name: editForm.name,
        quantity: parseInt(editForm.quantity) || 0,
        description: editForm.description,
        category: editForm.category,
        price: parseFloat(editForm.price) || 0,
        salesman: editForm.salesman,
        seller_id: editForm.seller_id,
        location: editForm.location,
        address: editForm.address,
        status: editForm.status,
        brand: editForm.brand,
        model: editForm.model,
        vin: editForm.vin,
        // Характеристики запчасти
        body_brand: editForm.body_brand,
        engine_brand: editForm.engine_brand,
        car_release_date: editForm.car_release_date,
        front_rear: editForm.front_rear,
        left_right: editForm.left_right,
        top_bottom: editForm.top_bottom,
        number: editForm.number,
        manufacturer: editForm.manufacturer,
        manufacturer_code: editForm.manufacturer_code,
        oem_code: editForm.oem_code,
        color: editForm.color,
        condition: editForm.condition,
        supplier_code: editForm.supplier_code,
        defect: editForm.defect,
        transmission: editForm.transmission,
        transmission_model: editForm.transmission_model,
        drive: editForm.drive,
        wear_percentage: editForm.wear_percentage,
        season: editForm.season,
        diameter: editForm.diameter,
        width: editForm.width,
        profile: editForm.profile,
        tire_quantity: editForm.tire_quantity,
        drilling: editForm.drilling,
        offset: editForm.offset,
        center_hole_diameter: editForm.center_hole_diameter,
        tire_model: editForm.tire_model,
      };

      // Only include photo if it has a value
      if (photoPath) {
        updatedPart.photo = photoPath;
      }

      console.log('DEBUG: Updated part data:', updatedPart);
      console.log('DEBUG: Photo field value in updatedPart:', updatedPart.photo);
      console.log('DEBUG: editForm.photo after photo upload:', editForm.photo);
      console.log('DEBUG: Calling updatePart with id:', initialPart.id, 'data:', updatedPart);
      console.log('DEBUG: Photo path being sent in updatePart:', photoPath);

      await updatePartMutation.mutateAsync({
        id: initialPart.id,
        data: updatedPart,
      });

      console.log('DEBUG: updatePartMutation completed successfully');

      // Update photo preview if photo was changed
      if (photoPath && photoPath !== initialPart.photo) {
        console.log('DEBUG: Updating photo preview to:', photoPath);
        photoUpload.updateCurrentPhoto(photoPath);
      }

      // Log user activity for part update
      try {
        await logUserActivity({
          action: 'update_part',
          resource_type: 'part',
          resource_id: initialPart.id,
          details: `Обновлена запчасть "${editForm.name}" (ID: ${initialPart.id})`,
        });
        console.log('DEBUG: User activity logged successfully');
      } catch (logError) {
        console.warn('Failed to log user activity:', logError);
        // Don't fail the operation if logging fails
      }

      // Force immediate cache invalidation and refetch to ensure UI updates
      console.log('DEBUG: Invalidating and refetching parts cache');
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });
      await queryClient.refetchQueries({ queryKey: partsKeys.lists() });
      console.log('DEBUG: Cache invalidation completed');

      console.log('Part update completed, closing dialog');
      setIsEditing(false);
      setShouldDeletePhoto(false); // Reset flag
    } catch (error: unknown) {
      console.error('Error updating part:', error);
      alert('An error occurred while updating the part. Please try again.');
      setIsEditing(false);
      setShouldDeletePhoto(false); // Reset flag on error
    }
  }, [editForm.name, editForm.quantity, editForm.description, editForm.category, editForm.price, editForm.salesman, editForm.location, editForm.address, editForm.status, editForm.brand, editForm.model, editForm.photo, editForm.body_brand, editForm.engine_brand, editForm.car_release_date, editForm.front_rear, editForm.left_right, editForm.top_bottom, editForm.number, editForm.manufacturer, editForm.manufacturer_code, editForm.oem_code, editForm.color, editForm.condition, editForm.supplier_code, editForm.defect, editForm.transmission, editForm.transmission_model, editForm.drive, editForm.wear_percentage, editForm.season, editForm.diameter, editForm.width, editForm.profile, editForm.tire_quantity, editForm.drilling, editForm.offset, editForm.center_hole_diameter, editForm.tire_model, initialPart.id, updatePartMutation, queryClient, photoUpload]);


  return {
    editForm,
    isEditing,
    setIsEditing,
    handleEdit,
    updateFormField,
    isLoading: updatePartMutation.isPending,
    photoUpload,
    uploadPhoto: photoUpload.uploadPhoto,
    shouldDeletePhoto,
    setShouldDeletePhoto,
  };
}
