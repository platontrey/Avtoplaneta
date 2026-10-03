/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { SearchableSelect } from '@/components/ui/searchable-select';
import { Package, Trash2 } from 'lucide-react';
import type { SupplierBatch } from '../types';

interface ZeroQuantityPartsCardProps {
  supplierBatches: SupplierBatch[];
  selectedSupplierCode: string;
  onSelectSupplierCode: (code: string) => void;
  onDeleteZeroParts: () => Promise<void>;
  onRefreshCodes: () => Promise<void>;
  loading: boolean;
}

export const ZeroQuantityPartsCard: React.FC<ZeroQuantityPartsCardProps> = ({
  supplierBatches,
  selectedSupplierCode,
  onSelectSupplierCode,
  onDeleteZeroParts,
  onRefreshCodes,
  loading,
}) => {
  return (
    <Card className="mt-8">
      <CardHeader>
        <CardTitle className="flex items-center">
          <Package className="w-5 h-5 mr-2" />
          Управление запчастями ({supplierBatches.length})
        </CardTitle>
        <CardDescription>
          Массовые операции с запчастями и дефектными ведомостями
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="p-4 border rounded-lg bg-yellow-50 dark:bg-yellow-950/20 border-yellow-200 dark:border-yellow-900">
            <h4 className="font-medium text-yellow-800 dark:text-yellow-300 mb-2">
              Удаление шаблонных запчастей (quantity = 0)
            </h4>
            <p className="text-sm text-yellow-700 dark:text-yellow-400 mb-4">
              Эта операция удалит все незаполненные запчасти с количеством 0 для выбранной дефектной ведомости (кода поставки).
              Используйте, когда уверены, что все необходимые запчасти из дефектной ведомости уже оприходованы.
            </p>
            <div className="space-y-4">
              <div>
                <Label htmlFor="supplier-code-select" className="mb-2 block">
                  Выберите дефектную ведомость / код поставки ({supplierBatches.length} доступно)
                </Label>
                <SearchableSelect
                  value={selectedSupplierCode}
                  onValueChange={onSelectSupplierCode}
                  options={supplierBatches.map((batch) => ({
                    value: batch.code,
                    label: batch.label || batch.code,
                  }))}
                  placeholder={
                    supplierBatches.length > 0
                      ? 'Выберите дефектную ведомость...'
                      : 'Нет ведомостей с нулевыми запчастями'
                  }
                  searchPlaceholder="Поиск по марке, модели или коду..."
                  emptyMessage="Ведомость не найдена"
                  disabled={supplierBatches.length === 0 || loading}
                  className="w-full"
                />
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <Button
                  variant="destructive"
                  onClick={onDeleteZeroParts}
                  disabled={loading || !selectedSupplierCode}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  {loading ? 'Удаление...' : 'Удалить запчасти с quantity = 0'}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={onRefreshCodes}
                  disabled={loading}
                >
                  Обновить список
                </Button>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
