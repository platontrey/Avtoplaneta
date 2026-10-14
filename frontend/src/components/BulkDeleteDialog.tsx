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
import { partsKeys } from '@/hooks/useParts';
import { API_BASE_URL } from '@/lib/api';

interface BulkDeleteDialogProps {
  isOpen: boolean;
  onClose: () => void;
  selectedPartIds: number[];
  scopeType?: 'ids' | 'filter';
  activeFilters?: Record<string, string>;
  onSuccess: () => void;
}

export default function BulkDeleteDialog({
  isOpen,
  onClose,
  selectedPartIds,
  scopeType = 'ids',
  activeFilters,
  onSuccess,
}: BulkDeleteDialogProps) {
  const queryClient = useQueryClient();
  const [isLoading, setIsLoading] = useState(false);
  const [confirmEntireCatalog, setConfirmEntireCatalog] = useState(false);

  const isFilterScope = scopeType === 'filter';
  const isBlankFilter = isFilterScope && (!activeFilters || Object.keys(activeFilters).length === 0);

  const handleDelete = async () => {
    setIsLoading(true);

    try {
      const payload = isFilterScope
        ? {
            scope: {
              type: 'BULK_SCOPE_TYPE_FILTER',
              filter: activeFilters || {},
            },
            confirm_entire_catalog: confirmEntireCatalog,
          }
        : {
            scope: {
              type: 'BULK_SCOPE_TYPE_IDS',
              ids: selectedPartIds,
            },
            ids: selectedPartIds,
          };

      const response = await fetch(`${API_BASE_URL}/api/v1/admin/parts/bulk-delete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
        },
        credentials: 'include',
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to delete parts');
      }

      const result = await response.json();

      try {
        await logUserActivity({
          action: 'bulk_delete_parts',
          resource_type: 'part',
          details: isFilterScope
            ? `Массовое удаление запчастей по фильтру (${result.deleted_count || 0} шт.)`
            : `Массовое удаление ${selectedPartIds.length} запчастей (ID: ${selectedPartIds.join(', ')})`,
        });
      } catch (logError) {
        console.warn('Failed to log bulk delete activity:', logError);
      }

      // Обновляем кэш
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });

      alert(`Успешно удалено ${result.deleted_count ?? selectedPartIds.length} запчастей`);
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
      <AlertDialogContent className="sm:max-w-[420px]">
        <AlertDialogHeader>
          <div className="flex items-center gap-3">
            <AlertTriangle className="h-6 w-6 text-red-500 shrink-0" />
            <div>
              <AlertDialogTitle className="text-red-600">
                {isFilterScope ? 'Удалить запчасти по запросу' : 'Удалить запчасти'}
              </AlertDialogTitle>
              <AlertDialogDescription className="mt-2">
                {isFilterScope ? (
                  <>
                    Вы уверены, что хотите удалить <strong>все запчасти</strong>, соответствующие текущему поисковому запросу и фильтрам?
                  </>
                ) : (
                  <>
                    Вы уверены, что хотите удалить {selectedPartIds.length} выбранных запчастей?
                  </>
                )}
                <br />
                <strong className="text-red-600">Это действие нельзя отменить.</strong>
              </AlertDialogDescription>
            </div>
          </div>
        </AlertDialogHeader>

        <div className="py-4">
          <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-900 rounded-lg p-3">
            <p className="text-sm text-red-800 dark:text-red-300">
              {isFilterScope ? (
                'Операция удалит все детали по текущему поисковому фильтру на сервере без необходимости их загрузки.'
              ) : selectedPartIds.length <= 6 ? (
                `Будут удалены запчасти с ID: ${selectedPartIds.join(', ')}`
              ) : (
                `Будут удалены запчасти с ID: ${selectedPartIds.slice(0, 6).join(', ')} ... и ещё ${selectedPartIds.length - 6}`
              )}
            </p>
          </div>

          {isBlankFilter && (
            <div className="mt-3 p-3 border border-red-400 bg-red-100/90 dark:bg-red-950/60 rounded-lg">
              <p className="text-xs font-bold text-red-900 dark:text-red-200 uppercase mb-1">
                ⚠️ Внимание: Фильтры поиска не заданы!
              </p>
              <p className="text-xs text-red-800 dark:text-red-300 mb-2.5 leading-relaxed">
                Поскольку фильтры не применены, операция затронет <strong>весь склад целиком</strong>.
              </p>
              <label className="flex items-center gap-2 cursor-pointer text-xs font-semibold text-red-900 dark:text-red-100 select-none">
                <input
                  type="checkbox"
                  checked={confirmEntireCatalog}
                  onChange={(e) => setConfirmEntireCatalog(e.target.checked)}
                  className="rounded text-red-600 focus:ring-red-500 h-4 w-4"
                />
                <span>Я подтверждаю удаление всего склада</span>
              </label>
            </div>
          )}
        </div>

        <AlertDialogFooter>
          <Button type="button" variant="outline" onClick={onClose} disabled={isLoading}>
            Отмена
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={handleDelete}
            disabled={isLoading || (isBlankFilter && !confirmEntireCatalog)}
          >
            {isLoading
              ? 'Удаление...'
              : isFilterScope
              ? 'Удалить всё по запросу'
              : `Удалить ${selectedPartIds.length} запчастей`}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
