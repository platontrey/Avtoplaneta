/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useEffect, useMemo } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { getAuthHeaders } from "@/lib/csrf";
import { logUserActivity } from "@/features/admin/api/adminApi";
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/hooks/useParts';
import { API_BASE_URL } from '@/lib/api';
import type { Part } from '@/lib/types';
import { ArrowRight, Calculator, CheckCircle2, SlidersHorizontal } from "lucide-react";

export type PriceOperationMode = 'PRICE_OPERATION_MODE_SET' | 'PRICE_OPERATION_MODE_PERCENT' | 'PRICE_OPERATION_MODE_OFFSET';

interface BulkEditDialogProps {
  isOpen: boolean;
  onClose: () => void;
  selectedPartIds: number[];
  selectedParts?: Part[];
  activeFilters?: Record<string, string>;
  initialScopeType?: 'ids' | 'filter';
  onSuccess: () => void;
}

interface BulkEditForm {
  category?: string;
  priceValue: string;
  priceMode: PriceOperationMode;
  location?: string;
  status: boolean;
  salesman?: string;
  // Флаги применения полей
  applyCategory: boolean;
  applyPrice: boolean;
  applyLocation: boolean;
  applyStatus: boolean;
  applySalesman: boolean;
}

export default function BulkEditDialog({
  isOpen,
  onClose,
  selectedPartIds,
  selectedParts = [],
  activeFilters,
  initialScopeType,
  onSuccess,
}: BulkEditDialogProps) {
  const queryClient = useQueryClient();
  const [isLoading, setIsLoading] = useState(false);
  const [users, setUsers] = useState<{ id: number; name: string; email: string }[]>([]);
  const [scopeType, setScopeType] = useState<'ids' | 'filter'>(initialScopeType || 'ids');

  const [form, setForm] = useState<BulkEditForm>({
    category: '',
    priceValue: '',
    priceMode: 'PRICE_OPERATION_MODE_SET',
    location: '',
    status: true,
    salesman: '',
    applyCategory: false,
    applyPrice: false,
    applyLocation: false,
    applyStatus: false,
    applySalesman: false,
  });

  // Загружаем список пользователей при открытии
  useEffect(() => {
    if (isOpen) {
      fetchUsers();
      if (initialScopeType) {
        setScopeType(initialScopeType);
      }
    }
  }, [isOpen, initialScopeType]);

  const fetchUsers = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users`, {
        headers: getAuthHeaders(),
        credentials: 'include',
      });

      if (response.ok) {
        const data = await response.json();
        setUsers(data.users || []);
      }
    } catch (error) {
      console.error('Failed to fetch users:', error);
    }
  };

  const updateForm = <K extends keyof BulkEditForm>(field: K, value: BulkEditForm[K]) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  // Расчет предпросмотра цен для первых 3 запчастей
  const previewParts = useMemo(() => {
    if (!form.applyPrice || !form.priceValue || selectedParts.length === 0) return [];
    const val = parseFloat(form.priceValue.replace(',', '.'));
    if (isNaN(val)) return [];

    return selectedParts.slice(0, 3).map((part) => {
      const oldPrice = part.price ?? 0;
      let newPrice = oldPrice;
      if (form.priceMode === 'PRICE_OPERATION_MODE_SET') {
        newPrice = Math.max(0, val);
      } else if (form.priceMode === 'PRICE_OPERATION_MODE_PERCENT') {
        newPrice = Math.max(0, Math.round(oldPrice * (1 + val / 100)));
      } else if (form.priceMode === 'PRICE_OPERATION_MODE_OFFSET') {
        newPrice = Math.max(0, Math.round(oldPrice + val));
      }
      return {
        id: part.id,
        name: part.name,
        oldPrice,
        newPrice,
        diff: newPrice - oldPrice,
      };
    });
  }, [form.applyPrice, form.priceValue, form.priceMode, selectedParts]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const hasAnyField =
        form.applyPrice ||
        form.applyCategory ||
        form.applyLocation ||
        form.applyStatus ||
        form.applySalesman;

      if (!hasAnyField) {
        alert('Выберите хотя бы одно поле для изменения');
        setIsLoading(false);
        return;
      }

      // Формируем операции строго по gRPC / Protobuf спецификации
      const operations: Record<string, unknown> = {};

      if (form.applyPrice) {
        const val = parseFloat(form.priceValue.replace(',', '.'));
        if (isNaN(val)) {
          alert('Введите корректное числовое значение для цены');
          setIsLoading(false);
          return;
        }
        operations.price = {
          mode: form.priceMode,
          value: val,
        };
      }

      if (form.applyCategory && form.category) {
        operations.category = form.category;
      }

      if (form.applyLocation && form.location) {
        operations.location = form.location;
      }

      if (form.applySalesman && form.salesman) {
        operations.salesman = form.salesman;
      }

      if (form.applyStatus) {
        operations.status = form.status;
      }

      // Формируем область применения (scope)
      const scope: Record<string, unknown> = {
        type: scopeType === 'filter' ? 'BULK_SCOPE_TYPE_FILTER' : 'BULK_SCOPE_TYPE_IDS',
      };

      if (scopeType === 'filter') {
        scope.filter = activeFilters || {};
      } else {
        scope.ids = selectedPartIds;
      }

      const payload = {
        scope,
        operations,
      };

      const response = await fetch(`${API_BASE_URL}/api/v1/admin/parts/bulk-update`, {
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
        throw new Error(errorData.error || 'Failed to update parts');
      }

      const result = await response.json();

      // Логируем действие
      try {
        await logUserActivity({
          action: 'bulk_update_parts',
          resource_type: 'part',
          details: `Массовое обновление запчастей (${scopeType === 'filter' ? 'по фильтру' : `${selectedPartIds.length} шт.`}). Изменения: ${Object.keys(operations).join(', ')}`,
        });
      } catch (logError) {
        console.warn('Failed to log bulk update activity:', logError);
      }

      // Инвалидируем кэш списков деталей
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });

      alert(`Успешно обновлено запчастей: ${result.updated_count ?? selectedPartIds.length}`);
      onSuccess();
    } catch (error) {
      console.error('Error bulk updating parts:', error);
      alert('Ошибка при массовом обновлении запчастей');
    } finally {
      setIsLoading(false);
    }
  };

  const countDisplay = scopeType === 'filter' ? 'по текущему фильтру' : `${selectedPartIds.length} выбранных`;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[560px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <SlidersHorizontal className="h-5 w-5 text-primary" />
            Массовое редактирование запчастей
          </DialogTitle>
          <DialogDescription>
            Изменения будут применены ко всем {countDisplay} запчастям.
          </DialogDescription>
        </DialogHeader>

        {/* Переключатель области применения, если есть активные фильтры */}
        {activeFilters && Object.keys(activeFilters).length > 0 && (
          <div className="bg-muted/50 p-3 rounded-lg border flex flex-col gap-2">
            <span className="text-xs font-medium text-muted-foreground">Область применения:</span>
            <div className="flex gap-2">
              <Button
                type="button"
                size="sm"
                variant={scopeType === 'ids' ? 'default' : 'outline'}
                onClick={() => setScopeType('ids')}
                className="text-xs h-8"
              >
                Выбранные вручную ({selectedPartIds.length})
              </Button>
              <Button
                type="button"
                size="sm"
                variant={scopeType === 'filter' ? 'default' : 'outline'}
                onClick={() => setScopeType('filter')}
                className="text-xs h-8"
              >
                Все по активному фильтру
              </Button>
            </div>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4 pt-1">
          {/* Блок изменения цены */}
          <div className="rounded-lg border p-3.5 space-y-3 bg-card shadow-xs">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="applyPrice"
                  checked={form.applyPrice}
                  onCheckedChange={(checked) => updateForm('applyPrice', Boolean(checked))}
                />
                <Label htmlFor="applyPrice" className="text-sm font-semibold cursor-pointer">
                  Изменение цены
                </Label>
              </div>
              {form.applyPrice && (
                <Badge variant="outline" className="text-xs">
                  {form.priceMode === 'PRICE_OPERATION_MODE_SET' && 'Точное значение'}
                  {form.priceMode === 'PRICE_OPERATION_MODE_PERCENT' && 'Процент (+/-)'}
                  {form.priceMode === 'PRICE_OPERATION_MODE_OFFSET' && 'Смещение (+/- ₽)'}
                </Badge>
              )}
            </div>

            {form.applyPrice && (
              <div className="space-y-3 pt-1">
                <Tabs
                  value={form.priceMode}
                  onValueChange={(val) => updateForm('priceMode', val as PriceOperationMode)}
                >
                  <TabsList className="grid grid-cols-3 w-full">
                    <TabsTrigger value="PRICE_OPERATION_MODE_SET" className="text-xs">
                      Фиксированная
                    </TabsTrigger>
                    <TabsTrigger value="PRICE_OPERATION_MODE_PERCENT" className="text-xs">
                      Процент %
                    </TabsTrigger>
                    <TabsTrigger value="PRICE_OPERATION_MODE_OFFSET" className="text-xs">
                      Смещение ₽
                    </TabsTrigger>
                  </TabsList>
                </Tabs>

                <div>
                  <Label className="text-xs text-muted-foreground mb-1 block">
                    {form.priceMode === 'PRICE_OPERATION_MODE_SET' && 'Новая цена для всех запчастей (₽)'}
                    {form.priceMode === 'PRICE_OPERATION_MODE_PERCENT' && 'Процент изменения (например: 10 для +10% или -15 для скидки)'}
                    {form.priceMode === 'PRICE_OPERATION_MODE_OFFSET' && 'Сумма изменения (например: 500 для +500 ₽ или -300 для снижения)'}
                  </Label>
                  <Input
                    type="number"
                    step={form.priceMode === 'PRICE_OPERATION_MODE_PERCENT' ? '1' : '50'}
                    placeholder={
                      form.priceMode === 'PRICE_OPERATION_MODE_SET'
                        ? 'Например: 2500'
                        : form.priceMode === 'PRICE_OPERATION_MODE_PERCENT'
                        ? 'Например: 15 или -10'
                        : 'Например: 500 или -200'
                    }
                    value={form.priceValue}
                    onChange={(e) => updateForm('priceValue', e.target.value)}
                  />
                </div>

                {/* Живой предпросмотр цен для выбранных */}
                {previewParts.length > 0 && (
                  <div className="mt-2 p-2.5 rounded-md bg-muted/60 border text-xs space-y-1.5">
                    <div className="flex items-center gap-1.5 font-medium text-foreground">
                      <Calculator className="h-3.5 w-3.5 text-primary" />
                      Предпросмотр расчета цен:
                    </div>
                    {previewParts.map((item) => (
                      <div key={item.id} className="flex items-center justify-between text-muted-foreground">
                        <span className="truncate max-w-[200px]" title={item.name}>
                          {item.name}
                        </span>
                        <div className="flex items-center gap-1.5 font-mono">
                          <span>{item.oldPrice.toLocaleString('ru-RU')} ₽</span>
                          <ArrowRight className="h-3 w-3 text-primary inline" />
                          <span className="font-semibold text-foreground">
                            {item.newPrice.toLocaleString('ru-RU')} ₽
                          </span>
                          {item.diff !== 0 && (
                            <span className={item.diff > 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}>
                              ({item.diff > 0 ? `+${item.diff.toLocaleString('ru-RU')}` : item.diff.toLocaleString('ru-RU')} ₽)
                            </span>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Категория */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="applyCategory"
                checked={form.applyCategory}
                onCheckedChange={(checked) => updateForm('applyCategory', Boolean(checked))}
              />
              <Label htmlFor="applyCategory" className="text-sm font-semibold cursor-pointer">
                Категория
              </Label>
            </div>
            {form.applyCategory && (
              <Select value={form.category || ""} onValueChange={(value) => updateForm('category', value)}>
                <SelectTrigger>
                  <SelectValue placeholder="Выберите категорию" />
                </SelectTrigger>
                <SelectContent className="max-h-60">
                  <SelectItem value="Тормоза">Тормоза</SelectItem>
                  <SelectItem value="Двигатель">Двигатель</SelectItem>
                  <SelectItem value="Подвеска">Подвеска</SelectItem>
                  <SelectItem value="Подвеска ДВС/КПП">Подвеска ДВС/КПП</SelectItem>
                  <SelectItem value="Подвеска передних колес">Подвеска передних колес</SelectItem>
                  <SelectItem value="Подвеска задних колес">Подвеска задних колес</SelectItem>
                  <SelectItem value="Электрика">Электрика</SelectItem>
                  <SelectItem value="Кузов">Кузов</SelectItem>
                  <SelectItem value="Кузов снаружи">Кузов снаружи</SelectItem>
                  <SelectItem value="Интерьер">Интерьер</SelectItem>
                  <SelectItem value="Трансмиссия">Трансмиссия</SelectItem>
                  <SelectItem value="Система охлаждения и отопления">Система охлаждения и отопления</SelectItem>
                  <SelectItem value="Система выхлопа (Глушитель)">Система выхлопа (Глушитель)</SelectItem>
                  <SelectItem value="Система рулевого управления">Система рулевого управления</SelectItem>
                  <SelectItem value="Рулевое управление">Рулевое управление</SelectItem>
                  <SelectItem value="Система фильтрации (Фильтры)">Система фильтрации (Фильтры)</SelectItem>
                  <SelectItem value="Шины и диски">Шины и диски</SelectItem>
                  <SelectItem value="Автохимия и масла">Автохимия и масла</SelectItem>
                  <SelectItem value="Аксессуары и тюннинг">Аксессуары и тюннинг</SelectItem>
                  <SelectItem value="Другое">Другое</SelectItem>
                </SelectContent>
              </Select>
            )}
          </div>

          {/* Местоположение */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="applyLocation"
                checked={form.applyLocation}
                onCheckedChange={(checked) => updateForm('applyLocation', Boolean(checked))}
              />
              <Label htmlFor="applyLocation" className="text-sm font-semibold cursor-pointer">
                Местоположение / Склад
              </Label>
            </div>
            {form.applyLocation && (
              <Input
                placeholder="Введите местоположение (ряд, полка, склад)"
                value={form.location || ''}
                onChange={(e) => updateForm('location', e.target.value)}
              />
            )}
          </div>

          {/* Статус */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="applyStatus"
                checked={form.applyStatus}
                onCheckedChange={(checked) => updateForm('applyStatus', Boolean(checked))}
              />
              <Label htmlFor="applyStatus" className="text-sm font-semibold cursor-pointer">
                Статус активности
              </Label>
            </div>
            {form.applyStatus && (
              <Select
                value={form.status ? "active" : "inactive"}
                onValueChange={(value) => updateForm('status', value === "active")}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">Активный</SelectItem>
                  <SelectItem value="inactive">Неактивный</SelectItem>
                </SelectContent>
              </Select>
            )}
          </div>

          {/* Продавец */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="applySalesman"
                checked={form.applySalesman}
                onCheckedChange={(checked) => updateForm('applySalesman', Boolean(checked))}
              />
              <Label htmlFor="applySalesman" className="text-sm font-semibold cursor-pointer">
                Ответственный продавец
              </Label>
            </div>
            {form.applySalesman && (
              <Select value={form.salesman || ""} onValueChange={(value) => updateForm('salesman', value)}>
                <SelectTrigger>
                  <SelectValue placeholder="Выберите продавца" />
                </SelectTrigger>
                <SelectContent>
                  {users.map((user) => (
                    <SelectItem key={user.id} value={user.name}>
                      {user.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          <DialogFooter className="pt-2">
            <Button type="button" variant="outline" onClick={onClose} disabled={isLoading}>
              Отмена
            </Button>
            <Button type="submit" disabled={isLoading} className="gap-2">
              {isLoading ? (
                'Применение...'
              ) : (
                <>
                  <CheckCircle2 className="h-4 w-4" />
                  Применить изменения
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
