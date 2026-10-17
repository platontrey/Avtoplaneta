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
import { safeFetchJson, API_BASE_URL } from "@/lib/api";
import { logUserActivity } from "@/features/admin/api/adminApi";
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/hooks/useParts';
import { usePartCatalog } from '@/features/catalog/usePartCatalog';
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
  address?: string;
  status: boolean;
  salesman?: string;
  bodyBrand?: string;
  engineBrand?: string;
  applyCategory: boolean;
  applyPrice: boolean;
  applyLocation: boolean;
  applyAddress: boolean;
  applyStatus: boolean;
  applySalesman: boolean;
  applyBodyBrand: boolean;
  applyEngineBrand: boolean;
}

const DEFAULT_CATEGORIES = [
  'Тормоза', 'Двигатель', 'Подвеска', 'Подвеска ДВС/КПП', 'Подвеска передних колес',
  'Подвеска задних колес', 'Электрика', 'Кузов', 'Кузов снаружи', 'Интерьер',
  'Трансмиссия', 'Система охлаждения и отопления', 'Система выхлопа (Глушитель)',
  'Система рулевого управления', 'Рулевое управление', 'Система фильтрации (Фильтры)',
  'Шины и диски', 'Автохимия и масла', 'Аксессуары и тюннинг', 'Другое',
];

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
  const { data: partCatalog } = usePartCatalog();
  const [isLoading, setIsLoading] = useState(false);
  const [users, setUsers] = useState<{ id: number; name: string; email: string }[]>([]);
  const [scopeType, setScopeType] = useState<'ids' | 'filter'>(initialScopeType || 'ids');
  const [confirmEntireCatalog, setConfirmEntireCatalog] = useState(false);

  const isBlankFilter = scopeType === 'filter' && (!activeFilters || Object.keys(activeFilters).length === 0);

  const [form, setForm] = useState<BulkEditForm>({
    category: '',
    priceValue: '',
    priceMode: 'PRICE_OPERATION_MODE_SET',
    location: '',
    address: '',
    status: true,
    salesman: '',
    bodyBrand: '',
    engineBrand: '',
    applyCategory: false,
    applyPrice: false,
    applyLocation: false,
    applyAddress: false,
    applyStatus: false,
    applySalesman: false,
    applyBodyBrand: false,
    applyEngineBrand: false,
  });

  const categories = useMemo(() => {
    const fromCatalog = partCatalog?.part_form_categories?.map((c) => c.name) ?? [];
    return Array.from(new Set([...fromCatalog, ...DEFAULT_CATEGORIES])).sort((a, b) => a.localeCompare(b, 'ru'));
  }, [partCatalog]);

  useEffect(() => {
    if (isOpen) {
      safeFetchJson<{ users?: { id: number; name: string; email: string }[] }>(`${API_BASE_URL}/api/v1/admin/users`, {
        headers: getAuthHeaders(),
        credentials: 'include',
      })
        .then((data) => setUsers(data.users || []))
        .catch((err) => console.error('Failed to fetch users:', err));

      if (initialScopeType) {
        setScopeType(initialScopeType);
      }
    }
  }, [isOpen, initialScopeType]);

  const updateForm = <K extends keyof BulkEditForm>(field: K, value: BulkEditForm[K]) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

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
        form.applyPrice || form.applyCategory || form.applyLocation || form.applyAddress ||
        form.applyStatus || form.applySalesman || form.applyBodyBrand || form.applyEngineBrand;

      if (!hasAnyField) {
        alert('Выберите хотя бы одно поле для изменения');
        setIsLoading(false);
        return;
      }

      const operations: Record<string, unknown> = {};

      if (form.applyPrice) {
        const val = parseFloat(form.priceValue.replace(',', '.'));
        if (isNaN(val)) {
          alert('Введите корректное числовое значение для цены');
          setIsLoading(false);
          return;
        }
        operations.price = { mode: form.priceMode, value: val };
      }

      if (form.applyCategory && form.category) operations.category = form.category;
      if (form.applyLocation && form.location) operations.location = form.location;
      if (form.applyAddress && form.address) operations.address = form.address;
      if (form.applySalesman && form.salesman) operations.salesman = form.salesman;
      if (form.applyStatus) operations.status = form.status;
      if (form.applyBodyBrand) operations.body_brand = form.bodyBrand || '';
      if (form.applyEngineBrand) operations.engine_brand = form.engineBrand || '';

      const scope: Record<string, unknown> = {
        type: scopeType === 'filter' ? 'BULK_SCOPE_TYPE_FILTER' : 'BULK_SCOPE_TYPE_IDS',
        ...(scopeType === 'filter' ? { filter: activeFilters || {} } : { ids: selectedPartIds }),
      };

      const result = await safeFetchJson<{ updated_count?: number }>(`${API_BASE_URL}/api/v1/admin/parts/bulk-update`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
        },
        credentials: 'include',
        body: JSON.stringify({ scope, operations, confirm_entire_catalog: confirmEntireCatalog }),
      });

      logUserActivity({
        action: 'bulk_update_parts',
        resource_type: 'part',
        details: `Массовое обновление запчастей (${scopeType === 'filter' ? 'по фильтру' : `${selectedPartIds.length} шт.`}). Изменения: ${Object.keys(operations).join(', ')}`,
      }).catch(console.warn);

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
          {/* Изменение цены */}
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
                <Tabs value={form.priceMode} onValueChange={(val) => updateForm('priceMode', val as PriceOperationMode)}>
                  <TabsList className="grid grid-cols-3 w-full">
                    <TabsTrigger value="PRICE_OPERATION_MODE_SET" className="text-xs">Фиксированная</TabsTrigger>
                    <TabsTrigger value="PRICE_OPERATION_MODE_PERCENT" className="text-xs">Процент %</TabsTrigger>
                    <TabsTrigger value="PRICE_OPERATION_MODE_OFFSET" className="text-xs">Смещение ₽</TabsTrigger>
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

                {previewParts.length > 0 && (
                  <div className="mt-2 p-2.5 rounded-md bg-muted/60 border text-xs space-y-1.5">
                    <div className="flex items-center gap-1.5 font-medium text-foreground">
                      <Calculator className="h-3.5 w-3.5 text-primary" />
                      Предпросмотр расчета цен:
                    </div>
                    {previewParts.map((item) => (
                      <div key={item.id} className="flex items-center justify-between text-muted-foreground">
                        <span className="truncate max-w-[200px]" title={item.name}>{item.name}</span>
                        <div className="flex items-center gap-1.5 font-mono">
                          <span>{item.oldPrice.toLocaleString('ru-RU')} ₽</span>
                          <ArrowRight className="h-3 w-3 text-primary inline" />
                          <span className="font-semibold text-foreground">{item.newPrice.toLocaleString('ru-RU')} ₽</span>
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
              <Checkbox id="applyCategory" checked={form.applyCategory} onCheckedChange={(c) => updateForm('applyCategory', Boolean(c))} />
              <Label htmlFor="applyCategory" className="text-sm font-semibold cursor-pointer">Категория</Label>
            </div>
            {form.applyCategory && (
              <Select value={form.category || ""} onValueChange={(val) => updateForm('category', val)}>
                <SelectTrigger><SelectValue placeholder="Выберите категорию" /></SelectTrigger>
                <SelectContent className="max-h-60">
                  {categories.map((cat) => (
                    <SelectItem key={cat} value={cat}>{cat}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          {/* Местоположение */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox id="applyLocation" checked={form.applyLocation} onCheckedChange={(c) => updateForm('applyLocation', Boolean(c))} />
              <Label htmlFor="applyLocation" className="text-sm font-semibold cursor-pointer">Местоположение</Label>
            </div>
            {form.applyLocation && (
              <Input
                placeholder="Стеллаж / ряд / полка"
                value={form.location || ''}
                onChange={(e) => updateForm('location', e.target.value)}
              />
            )}
          </div>

          {/* Адрес склада */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox id="applyAddress" checked={form.applyAddress} onCheckedChange={(c) => updateForm('applyAddress', Boolean(c))} />
              <Label htmlFor="applyAddress" className="text-sm font-semibold cursor-pointer">Адрес склада</Label>
            </div>
            {form.applyAddress && (
              <Input
                placeholder="Например, г. Владивосток, ул. Снеговая, 2Б"
                value={form.address || ''}
                onChange={(e) => updateForm('address', e.target.value)}
              />
            )}
          </div>

          {/* Статус */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox id="applyStatus" checked={form.applyStatus} onCheckedChange={(c) => updateForm('applyStatus', Boolean(c))} />
              <Label htmlFor="applyStatus" className="text-sm font-semibold cursor-pointer">Статус активности</Label>
            </div>
            {form.applyStatus && (
              <Select value={form.status ? "active" : "inactive"} onValueChange={(val) => updateForm('status', val === "active")}>
                <SelectTrigger><SelectValue /></SelectTrigger>
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
              <Checkbox id="applySalesman" checked={form.applySalesman} onCheckedChange={(c) => updateForm('applySalesman', Boolean(c))} />
              <Label htmlFor="applySalesman" className="text-sm font-semibold cursor-pointer">Ответственный продавец</Label>
            </div>
            {form.applySalesman && (
              <Select value={form.salesman || ""} onValueChange={(val) => updateForm('salesman', val)}>
                <SelectTrigger><SelectValue placeholder="Выберите продавца" /></SelectTrigger>
                <SelectContent>
                  {users.map((user) => (
                    <SelectItem key={user.id} value={user.name}>{user.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          {/* Марка кузова */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox id="applyBodyBrand" checked={form.applyBodyBrand} onCheckedChange={(c) => updateForm('applyBodyBrand', Boolean(c))} />
              <Label htmlFor="applyBodyBrand" className="text-sm font-semibold cursor-pointer">Марка кузова</Label>
            </div>
            {form.applyBodyBrand && (
              <Input
                placeholder="Введите марку / номер кузова (например: AE100, W204, E39)"
                value={form.bodyBrand || ''}
                onChange={(e) => updateForm('bodyBrand', e.target.value)}
              />
            )}
          </div>

          {/* Марка двигателя */}
          <div className="rounded-lg border p-3.5 space-y-2 bg-card shadow-xs">
            <div className="flex items-center space-x-2">
              <Checkbox id="applyEngineBrand" checked={form.applyEngineBrand} onCheckedChange={(c) => updateForm('applyEngineBrand', Boolean(c))} />
              <Label htmlFor="applyEngineBrand" className="text-sm font-semibold cursor-pointer">Марка двигателя</Label>
            </div>
            {form.applyEngineBrand && (
              <Input
                placeholder="Введите марку / модель двигателя (например: 1JZ-GTE, 2GR-FE, M54)"
                value={form.engineBrand || ''}
                onChange={(e) => updateForm('engineBrand', e.target.value)}
              />
            )}
          </div>

          {isBlankFilter && (
            <div className="p-3 border border-orange-400 bg-orange-100/90 dark:bg-orange-950/60 rounded-lg">
              <p className="text-xs font-bold text-orange-900 dark:text-orange-200 uppercase mb-1">
                ⚠️ Внимание: Фильтры поиска не заданы!
              </p>
              <p className="text-xs text-orange-800 dark:text-orange-300 mb-2.5 leading-relaxed">
                Поскольку фильтры не применены, изменения будут применены ко <strong>всему складу целиком</strong>.
              </p>
              <label className="flex items-center gap-2 cursor-pointer text-xs font-semibold text-orange-900 dark:text-orange-100 select-none">
                <input
                  type="checkbox"
                  checked={confirmEntireCatalog}
                  onChange={(e) => setConfirmEntireCatalog(e.target.checked)}
                  className="rounded text-orange-600 focus:ring-orange-500 h-4 w-4"
                />
                <span>Я подтверждаю массовое изменение всего склада</span>
              </label>
            </div>
          )}

          <DialogFooter className="pt-2">
            <Button type="button" variant="outline" onClick={onClose} disabled={isLoading}>
              Отмена
            </Button>
            <Button
              type="submit"
              disabled={isLoading || (isBlankFilter && !confirmEntireCatalog)}
              className="gap-2"
            >
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
