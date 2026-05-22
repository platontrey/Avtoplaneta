/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { useState } from "react";
import { AlertDialog, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { AlertTriangle } from "lucide-react";
import { getAuthHeaders } from "@/lib/csrf";
import { logUserActivity } from "@/features/admin/api/adminApi";
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/features/parts/hooks/usePartsQueries';
import { API_BASE_URL } from '@/lib/api';

interface BulkDeleteDialogProps {
  isOpen: boolean;
  onClose: () => void;
  selectedPartIds: number[];
  onSuccess: () => void;
}

export default function BulkDeleteDialog({ isOpen, onClose, selectedPartIds, onSuccess }: BulkDeleteDialogProps) {
  const queryClient = useQueryClient();
  const [isLoading, setIsLoading] = useState(false);

  const handleDelete = async () => {
    setIsLoading(true);

    try {
      // Отправляем запрос на массовое удаление
      const response = await fetch(`${API_BASE_URL}/api/admin/bulk-delete-parts`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
        },
        credentials: 'include',
        body: JSON.stringify({
          ids: selectedPartIds,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to delete parts');
      }

      const result = await response.json();

      // Логируем массовое удаление
      try {
        await logUserActivity({
          action: 'bulk_delete_parts',
          resource_type: 'part',
          details: `Массовое удаление ${selectedPartIds.length} запчастей (ID: ${selectedPartIds.join(', ')})`,
        });
      } catch (logError) {
        console.warn('Failed to log bulk delete activity:', logError);
      }

      // Обновляем кэш
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });

      alert(`Успешно удалено ${result.deleted_count} запчастей`);
      onSuccess();
    } catch (error) {
      console.error('Error bulk deleting parts:', error);
      alert('Ошибка при массовом удалении запчастей');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog open={isOpen} onOpenChange={onClose}>
      <AlertDialogContent className="sm:max-w-[400px]">
        <AlertDialogHeader>
          <div className="flex items-center gap-3">
            <AlertTriangle className="h-6 w-6 text-red-500" />
            <div>
              <AlertDialogTitle className="text-red-600">Удалить запчасти</AlertDialogTitle>
              <AlertDialogDescription className="mt-2">
                Вы уверены, что хотите удалить {selectedPartIds.length} выбранных запчастей?
                <br />
                <strong className="text-red-600">Это действие нельзя отменить.</strong>
              </AlertDialogDescription>
            </div>
          </div>
        </AlertDialogHeader>

        <div className="py-4">
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <p className="text-sm text-red-800">
              Будут удалены запчасти с ID: {selectedPartIds.join(', ')}
            </p>
          </div>
        </div>

        <AlertDialogFooter>
          <Button type="button" variant="outline" onClick={onClose} disabled={isLoading}>
            Отмена
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={handleDelete}
            disabled={isLoading}
          >
            {isLoading ? 'Удаление...' : `Удалить ${selectedPartIds.length} запчастей`}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}