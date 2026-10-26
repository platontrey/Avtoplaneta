/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import React, { useState, useEffect, useMemo } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { getAuthHeaders } from "@/lib/csrf";
import { safeFetchJson, API_BASE_URL } from "@/lib/api";
import { formatCarReleasePeriod } from "@/lib/utils";
import { logUserActivity } from "@/features/admin/api/adminApi";
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/hooks/useParts';
import { usePartCatalog } from '@/features/catalog/usePartCatalog';
import { useVehicleOptions } from '@/features/vehicles/useVehicleCatalog';
import type { Part } from '@/lib/types';
import { ArrowRight, Calculator, CheckCircle2, RotateCcw, Search, SlidersHorizontal } from "lucide-react";

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

type FieldSection = 'basic' | 'specs' | 'tires';

interface FieldValuesState {
  priceValue: string;
  priceMode: PriceOperationMode;
  quantity: string;
  name: string;
  brand: string;
  model: string;
  vin: string;
  car_release_date: string;
  car_release_period: string;
  category: string;
  description: string;
  location: string;
  address: string;
  salesman: string;
  status: boolean;
  body_brand: string;
  engine_brand: string;
  transmission: string;
  transmission_model: string;
  drive: string;
  color: string;
  condition: string;
  defect: string;
  supplier_code: string;
  front_rear: string;
  left_right: string;
  top_bottom: string;
  number: string;
  manufacturer: string;
  manufacturer_code: string;
  oem_code: string;
  wear_percentage: string;
  season: string;
  diameter: string;
  width: string;
  profile: string;
  tire_quantity: string;
  drilling: string;
  offset: string;
  center_hole_diameter: string;
  tire_model: string;
}

export type BulkFieldKey =
  | 'price'
  | 'quantity'
  | 'name'
  | 'brand'
  | 'model'
  | 'vin'
  | 'car_release_date'
  | 'car_release_period'
  | 'category'
  | 'description'
  | 'location'
  | 'address'
  | 'salesman'
  | 'status'
  | 'body_brand'
  | 'engine_brand'
  | 'transmission'
  | 'transmission_model'
  | 'drive'
  | 'color'
  | 'condition'
  | 'defect'
  | 'supplier_code'
  | 'front_rear'
  | 'left_right'
  | 'top_bottom'
  | 'number'
  | 'manufacturer'
  | 'manufacturer_code'
  | 'oem_code'
  | 'wear_percentage'
  | 'season'
  | 'diameter'
  | 'width'
  | 'profile'
  | 'tire_quantity'
  | 'drilling'
  | 'offset'
  | 'center_hole_diameter'
  | 'tire_model';

interface FieldMeta {
  key: BulkFieldKey;
  label: string;
  section: FieldSection;
  keywords: string;
}

const FIELD_METAS: FieldMeta[] = [
  // Основные
  { key: 'price', label: 'Изменение цены', section: 'basic', keywords: 'цена стоимость прайс скидка наценка руб' },
  { key: 'quantity', label: 'Количество (шт.)', section: 'basic', keywords: 'количество остаток шт склад' },
  { key: 'name', label: 'Название детали', section: 'basic', keywords: 'название наименование имя деталь' },
  { key: 'brand', label: 'Бренд автомобиля', section: 'basic', keywords: 'бренд марка авто' },
  { key: 'model', label: 'Модель автомобиля', section: 'basic', keywords: 'модель авто' },
  { key: 'vin', label: 'VIN / Номер кузова', section: 'basic', keywords: 'вин vin кузов фрейм frame' },
  { key: 'car_release_date', label: 'Год выпуска', section: 'basic', keywords: 'год выпуска дата' },
  { key: 'car_release_period', label: 'Период выпуска', section: 'basic', keywords: 'период выпуска поколение года' },
  { key: 'category', label: 'Категория', section: 'basic', keywords: 'категория раздел группа' },
  { key: 'description', label: 'Описание / Заметки', section: 'basic', keywords: 'описание заметки комментарий примечание' },
  { key: 'location', label: 'Местоположение (стеллаж / полка)', section: 'basic', keywords: 'местоположение локация стеллаж полка ряд ячейка' },
  { key: 'address', label: 'Адрес склада', section: 'basic', keywords: 'адрес склад томск кузовлево профсоюзная' },
  { key: 'salesman', label: 'Ответственный продавец', section: 'basic', keywords: 'продавец менеджер сотрудник' },
  { key: 'status', label: 'Статус активности', section: 'basic', keywords: 'статус активность активный неактивный' },

  // Характеристики
  { key: 'body_brand', label: 'Марка кузова', section: 'specs', keywords: 'марка кузова кузов код' },
  { key: 'engine_brand', label: 'Марка двигателя', section: 'specs', keywords: 'марка двигателя двс мотор' },
  { key: 'transmission', label: 'Тип трансмиссии (КПП)', section: 'specs', keywords: 'трансмиссия кпп акпп мкпп вариатор робот' },
  { key: 'transmission_model', label: 'Модель трансмиссии', section: 'specs', keywords: 'модель трансмиссии номер кпп' },
  { key: 'drive', label: 'Привод', section: 'specs', keywords: 'привод передний задний полный 4wd awd' },
  { key: 'color', label: 'Цвет деталей', section: 'specs', keywords: 'цвет краска оттенок' },
  { key: 'condition', label: 'Состояние', section: 'specs', keywords: 'состояние бу б/у новое новый' },
  { key: 'defect', label: 'Дефект', section: 'specs', keywords: 'дефект повреждение царапина вмятина' },
  { key: 'supplier_code', label: 'Код поставщика / партии', section: 'specs', keywords: 'код поставщика партия ведомость поставка' },
  { key: 'front_rear', label: 'Перед / зад', section: 'specs', keywords: 'перед зад расположение f r' },
  { key: 'left_right', label: 'Право / лево', section: 'specs', keywords: 'право лево сторона l r' },
  { key: 'top_bottom', label: 'Верх / низ', section: 'specs', keywords: 'верх низ' },
  { key: 'number', label: 'Номер детали', section: 'specs', keywords: 'номер детали артикул' },
  { key: 'manufacturer', label: 'Производитель', section: 'specs', keywords: 'производитель завод фирма' },
  { key: 'manufacturer_code', label: 'Код производителя', section: 'specs', keywords: 'код производителя кросс' },
  { key: 'oem_code', label: 'OEM код', section: 'specs', keywords: 'oem оем оригинальный номер код' },
  { key: 'wear_percentage', label: 'Процент износа (%)', section: 'specs', keywords: 'процент износа износ остаток' },

  // Шины и диски
  { key: 'season', label: 'Сезон шин', section: 'tires', keywords: 'сезон зимние летние всесезонные шины резина' },
  { key: 'diameter', label: 'Диаметр (R)', section: 'tires', keywords: 'диаметр радиус r15 r16 r17 диски шины' },
  { key: 'width', label: 'Ширина', section: 'tires', keywords: 'ширина профиль диск шина' },
  { key: 'profile', label: 'Профиль (%)', section: 'tires', keywords: 'профиль высота шина' },
  { key: 'tire_quantity', label: 'Количество шин / дисков в комплекте', section: 'tires', keywords: 'количество шин дисков комплект' },
  { key: 'drilling', label: 'Сверловка (PCD)', section: 'tires', keywords: 'сверловка pcd разболтовка' },
  { key: 'offset', label: 'Вылет (ET)', section: 'tires', keywords: 'вылет et' },
  { key: 'center_hole_diameter', label: 'Диаметр ЦО (DIA)', section: 'tires', keywords: 'диаметр цо dia ступица центральное отверстие' },
  { key: 'tire_model', label: 'Модель шины / диска', section: 'tires', keywords: 'модель шины резины диска' },
];

const DEFAULT_CATEGORIES = [
  'Тормоза', 'Двигатель', 'Подвеска', 'Подвеска ДВС/КПП', 'Подвеска передних колес',
  'Подвеска задних колес', 'Электрика', 'Кузов', 'Кузов снаружи', 'Интерьер',
  'Трансмиссия', 'Система охлаждения и отопления', 'Система выхлопа (Глушитель)',
  'Система рулевого управления', 'Рулевое управление', 'Система фильтрации (Фильтры)',
  'Шины и диски', 'Автохимия и масла', 'Аксессуары и тюннинг', 'Другое',
];

const DEFAULT_COLORS = [
  'Черный', 'Белый', 'Серебристый', 'Серый', 'Темно-серый', 'Светло-серый',
  'Синий', 'Темно-синий', 'Светло-синий', 'Красный', 'Темно-красный', 'Бордовый',
  'Зеленый', 'Темно-зеленый', 'Светло-зеленый', 'Желтый', 'Оранжевый', 'Фиолетовый',
  'Коричневый', 'Бежевый', 'Золотой', 'Бронзовый', 'Перламутровый', 'Металлик', 'Матовый',
];

const INITIAL_VALUES: FieldValuesState = {
  priceValue: '',
  priceMode: 'PRICE_OPERATION_MODE_SET',
  quantity: '1',
  name: '',
  brand: '',
  model: '',
  vin: '',
  car_release_date: '',
  car_release_period: '',
  category: '',
  description: '',
  location: '',
  address: '',
  salesman: '',
  status: true,
  body_brand: '',
  engine_brand: '',
  transmission: '',
  transmission_model: '',
  drive: '',
  color: '',
  condition: '',
  defect: '',
  supplier_code: '',
  front_rear: '',
  left_right: '',
  top_bottom: '',
  number: '',
  manufacturer: '',
  manufacturer_code: '',
  oem_code: '',
  wear_percentage: '',
  season: '',
  diameter: '',
  width: '',
  profile: '',
  tire_quantity: '',
  drilling: '',
  offset: '',
  center_hole_diameter: '',
  tire_model: '',
};

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

  const [activeTab, setActiveTab] = useState<'basic' | 'specs' | 'tires' | 'all'>('basic');
  const [fieldSearch, setFieldSearch] = useState('');
  const [showOnlyChecked, setShowOnlyChecked] = useState(false);

  const [values, setValues] = useState<FieldValuesState>(INITIAL_VALUES);
  const [checkedFields, setCheckedFields] = useState<Record<BulkFieldKey, boolean>>(
    () => Object.fromEntries(FIELD_METAS.map((m) => [m.key, false])) as Record<BulkFieldKey, boolean>
  );

  // Определяем контекстный бренд и модель (из формы или из первой выбранной детали/фильтра) для подсказок кузова и ДВС
  const effectiveBrand = useMemo(() => {
    if (values.brand) return values.brand;
    if (activeFilters?.brand) return activeFilters.brand;
    const firstWithBrand = selectedParts.find((p) => p.brand);
    return firstWithBrand?.brand || '';
  }, [values.brand, activeFilters, selectedParts]);

  const effectiveModel = useMemo(() => {
    if (values.model) return values.model;
    if (activeFilters?.model) return activeFilters.model;
    const firstWithModel = selectedParts.find((p) => p.model);
    return firstWithModel?.model || '';
  }, [values.model, activeFilters, selectedParts]);

  const { brandOptions, modelOptions, bodyOptions, engineOptions } = useVehicleOptions(effectiveBrand, effectiveModel);

  const isBlankFilter = scopeType === 'filter' && (!activeFilters || Object.keys(activeFilters).length === 0);

  const categories = useMemo(() => {
    const fromCatalog = partCatalog?.part_form_categories?.map((c) => c.name) ?? [];
    return Array.from(new Set([...fromCatalog, ...DEFAULT_CATEGORIES])).sort((a, b) => a.localeCompare(b, 'ru'));
  }, [partCatalog]);

  const transmissionOptions = useMemo(() => {
    return partCatalog?.attributes?.find((a) => a.code === 'transmission')?.options ?? ['АКПП', 'МКПП', 'Вариатор', 'Роботизированная'];
  }, [partCatalog]);

  const driveOptions = useMemo(() => {
    return partCatalog?.attributes?.find((a) => a.code === 'drive')?.options ?? ['Передний', 'Задний', 'Полный'];
  }, [partCatalog]);

  const conditionOptions = useMemo(() => {
    return partCatalog?.attributes?.find((a) => a.code === 'condition')?.options ?? ['Б/У', 'Новое'];
  }, [partCatalog]);

  const seasonOptions = useMemo(() => {
    return partCatalog?.attributes?.find((a) => a.code === 'season')?.options ?? ['Зимние', 'Летние', 'Всесезонные'];
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

  const updateValue = <K extends keyof FieldValuesState>(field: K, value: FieldValuesState[K]) => {
    setValues((prev) => ({ ...prev, [field]: value }));
  };

  const toggleField = (key: BulkFieldKey, checked: boolean) => {
    setCheckedFields((prev) => ({ ...prev, [key]: checked }));
  };

  const resetAllFields = () => {
    setCheckedFields(Object.fromEntries(FIELD_METAS.map((m) => [m.key, false])) as Record<BulkFieldKey, boolean>);
    setValues(INITIAL_VALUES);
    setShowOnlyChecked(false);
  };

  const checkedCount = useMemo(
    () => Object.values(checkedFields).filter(Boolean).length,
    [checkedFields]
  );

  const sectionCounts = useMemo(() => {
    const counts: Record<FieldSection, number> = { basic: 0, specs: 0, tires: 0 };
    for (const meta of FIELD_METAS) {
      if (checkedFields[meta.key]) {
        counts[meta.section]++;
      }
    }
    return counts;
  }, [checkedFields]);

  const visibleMetas = useMemo(() => {
    const q = fieldSearch.trim().toLowerCase();
    return FIELD_METAS.filter((meta) => {
      if (showOnlyChecked && !checkedFields[meta.key]) return false;
      if (q) {
        return (
          meta.label.toLowerCase().includes(q) ||
          meta.key.toLowerCase().includes(q) ||
          meta.keywords.toLowerCase().includes(q)
        );
      }
      if (activeTab === 'all') return true;
      return meta.section === activeTab;
    });
  }, [fieldSearch, showOnlyChecked, checkedFields, activeTab]);

  const previewParts = useMemo(() => {
    if (!checkedFields.price || !values.priceValue || selectedParts.length === 0) return [];
    const val = parseFloat(values.priceValue.replace(',', '.'));
    if (isNaN(val)) return [];

    return selectedParts.slice(0, 3).map((part) => {
      const oldPrice = part.price ?? 0;
      let newPrice = oldPrice;
      if (values.priceMode === 'PRICE_OPERATION_MODE_SET') {
        newPrice = Math.max(0, val);
      } else if (values.priceMode === 'PRICE_OPERATION_MODE_PERCENT') {
        newPrice = Math.max(0, Math.round(oldPrice * (1 + val / 100)));
      } else if (values.priceMode === 'PRICE_OPERATION_MODE_OFFSET') {
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
  }, [checkedFields.price, values.priceValue, values.priceMode, selectedParts]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      if (checkedCount === 0) {
        alert('Выберите хотя бы одно поле (строку) для изменения');
        setIsLoading(false);
        return;
      }

      const operations: Record<string, unknown> = {};

      if (checkedFields.price) {
        const val = parseFloat(values.priceValue.replace(',', '.'));
        if (isNaN(val)) {
          alert('Введите корректное числовое значение для цены');
          setIsLoading(false);
          return;
        }
        operations.price = { mode: values.priceMode, value: val };
      }

      if (checkedFields.quantity) {
        const qty = parseInt(values.quantity, 10);
        if (isNaN(qty) || qty < 0) {
          alert('Введите корректное неотрицательное количество (0 или больше)');
          setIsLoading(false);
          return;
        }
        operations.quantity = qty;
      }

      if (checkedFields.name) {
        if (!values.name.trim()) {
          alert('Название детали не может быть пустым');
          setIsLoading(false);
          return;
        }
        operations.name = values.name.trim();
      }

      if (checkedFields.status) {
        operations.status = values.status;
      }

      if (checkedFields.salesman) {
        operations.salesman = values.salesman;
        const matchedUser = users.find((u) => u.name === values.salesman);
        if (matchedUser) {
          operations.seller_id = matchedUser.id;
        } else if (!values.salesman) {
          operations.seller_id = 0;
        }
      }

      const stringFieldKeys: Exclude<BulkFieldKey, 'price' | 'quantity' | 'name' | 'status' | 'salesman'>[] = [
        'brand',
        'model',
        'vin',
        'car_release_date',
        'car_release_period',
        'category',
        'description',
        'location',
        'address',
        'body_brand',
        'engine_brand',
        'transmission',
        'transmission_model',
        'drive',
        'color',
        'condition',
        'defect',
        'supplier_code',
        'front_rear',
        'left_right',
        'top_bottom',
        'number',
        'manufacturer',
        'manufacturer_code',
        'oem_code',
        'wear_percentage',
        'season',
        'diameter',
        'width',
        'profile',
        'tire_quantity',
        'drilling',
        'offset',
        'center_hole_diameter',
        'tire_model',
      ];

      for (const key of stringFieldKeys) {
        if (checkedFields[key]) {
          operations[key] = values[key] ?? '';
        }
      }

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
      const msg = error instanceof Error ? error.message : 'Ошибка при массовом обновлении запчастей';
      alert(msg);
    } finally {
      setIsLoading(false);
    }
  };

  const renderFieldInput = (key: BulkFieldKey) => {
    switch (key) {
      case 'price':
        return (
          <div className="space-y-3 pt-1">
            <Tabs value={values.priceMode} onValueChange={(val) => updateValue('priceMode', val as PriceOperationMode)}>
              <TabsList className="grid grid-cols-3 w-full">
                <TabsTrigger value="PRICE_OPERATION_MODE_SET" className="text-xs">Фиксированная</TabsTrigger>
                <TabsTrigger value="PRICE_OPERATION_MODE_PERCENT" className="text-xs">Процент %</TabsTrigger>
                <TabsTrigger value="PRICE_OPERATION_MODE_OFFSET" className="text-xs">Смещение ₽</TabsTrigger>
              </TabsList>
            </Tabs>
            <div>
              <Label className="text-xs text-muted-foreground mb-1 block">
                {values.priceMode === 'PRICE_OPERATION_MODE_SET' && 'Новая цена для всех запчастей (₽)'}
                {values.priceMode === 'PRICE_OPERATION_MODE_PERCENT' && 'Процент изменения (например: 10 для +10% или -15 для скидки)'}
                {values.priceMode === 'PRICE_OPERATION_MODE_OFFSET' && 'Сумма изменения (например: 500 для +500 ₽ или -300 для снижения)'}
              </Label>
              <Input
                type="number"
                step={values.priceMode === 'PRICE_OPERATION_MODE_PERCENT' ? '1' : '50'}
                placeholder={
                  values.priceMode === 'PRICE_OPERATION_MODE_SET'
                    ? 'Например: 2500'
                    : values.priceMode === 'PRICE_OPERATION_MODE_PERCENT'
                    ? 'Например: 15 или -10'
                    : 'Например: 500 или -200'
                }
                value={values.priceValue}
                onChange={(e) => updateValue('priceValue', e.target.value)}
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
        );

      case 'quantity':
        return (
          <Input
            type="number"
            min={0}
            placeholder="Например: 1 (или 0)"
            value={values.quantity}
            onChange={(e) => updateValue('quantity', e.target.value)}
          />
        );

      case 'name':
        return (
          <Input
            placeholder="Например: Диск тормозной передний"
            value={values.name}
            onChange={(e) => updateValue('name', e.target.value)}
          />
        );

      case 'brand':
        return (
          <SearchableSelect
            value={values.brand}
            onValueChange={(val) => updateValue('brand', val)}
            options={brandOptions}
            placeholder="Выберите или введите бренд"
            searchPlaceholder="Поиск бренда..."
            emptyMessage="Бренд не найден — можно ввести свой"
            allowCustom
            className="h-10"
          />
        );

      case 'model':
        return (
          <SearchableSelect
            value={values.model}
            onValueChange={(val) => updateValue('model', val)}
            options={modelOptions}
            placeholder={effectiveBrand ? "Выберите или введите модель" : "Введите или выберите модель"}
            searchPlaceholder="Поиск модели..."
            emptyMessage="Модель не найдена — можно ввести свою"
            allowCustom
            className="h-10"
          />
        );

      case 'vin':
        return (
          <Input
            placeholder="VIN или номер кузова (оставьте пустым для очистки)"
            value={values.vin}
            onChange={(e) => updateValue('vin', e.target.value)}
          />
        );

      case 'car_release_date':
        return (
          <Input
            placeholder="Например: 2008"
            value={values.car_release_date}
            onChange={(e) => updateValue('car_release_date', e.target.value)}
          />
        );

      case 'car_release_period':
        return (
          <Input
            placeholder="Например: 2001-2007"
            value={values.car_release_period}
            onChange={(e) => updateValue('car_release_period', formatCarReleasePeriod(e.target.value))}
          />
        );

      case 'category':
        return (
          <Select value={values.category || "__clear__"} onValueChange={(val) => updateValue('category', val === "__clear__" ? "" : val)}>
            <SelectTrigger><SelectValue placeholder="Выберите категорию" /></SelectTrigger>
            <SelectContent className="max-h-60">
              <SelectItem value="__clear__">— Очистить категорию —</SelectItem>
              {categories.map((cat) => (
                <SelectItem key={cat} value={cat}>{cat}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        );

      case 'description':
        return (
          <Textarea
            placeholder="Дополнительные примечания к запчастям (оставьте пустым для очистки)"
            value={values.description}
            onChange={(e) => updateValue('description', e.target.value)}
            rows={3}
          />
        );

      case 'location':
        return (
          <Input
            placeholder="Стеллаж / ряд / полка (оставьте пустым для очистки)"
            value={values.location}
            onChange={(e) => updateValue('location', e.target.value)}
          />
        );

      case 'address':
        return (
          <SearchableSelect
            value={values.address}
            onValueChange={(val) => updateValue('address', val)}
            options={[
              { value: 'Томск, Профсоюзная 2/11', label: 'Томск, Профсоюзная 2/11' },
              { value: 'г. Томск, пос. Кузовлево, ул. Заречная, 36', label: 'г. Томск, пос. Кузовлево, ул. Заречная, 36' },
            ]}
            placeholder="Выберите или введите адрес склада"
            searchPlaceholder="Поиск или ввод адреса..."
            allowCustom
            className="h-10"
          />
        );

      case 'salesman':
        return (
          <Select value={values.salesman || "__clear__"} onValueChange={(val) => updateValue('salesman', val === "__clear__" ? "" : val)}>
            <SelectTrigger><SelectValue placeholder="Выберите продавца" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="__clear__">— Не назначен (очистить) —</SelectItem>
              {users.map((user) => (
                <SelectItem key={user.id} value={user.name}>{user.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        );

      case 'status':
        return (
          <Select value={values.status ? "active" : "inactive"} onValueChange={(val) => updateValue('status', val === "active")}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="active">Активный</SelectItem>
              <SelectItem value="inactive">Неактивный</SelectItem>
            </SelectContent>
          </Select>
        );

      case 'body_brand':
        return (
          <SearchableSelect
            value={values.body_brand}
            onValueChange={(val) => updateValue('body_brand', val)}
            options={bodyOptions}
            placeholder="Выберите или введите марку кузова (например: AE100, W204, E39)"
            searchPlaceholder="Поиск или ввод кузова..."
            emptyMessage="Кузов не найден — можно ввести свой"
            allowCustom
            className="h-10"
          />
        );

      case 'engine_brand':
        return (
          <SearchableSelect
            value={values.engine_brand}
            onValueChange={(val) => updateValue('engine_brand', val)}
            options={engineOptions}
            placeholder="Выберите или введите марку двигателя (например: 1JZ-GTE, 2GR-FE)"
            searchPlaceholder="Поиск или ввод двигателя..."
            emptyMessage="Двигатель не найден — можно ввести свой"
            allowCustom
            className="h-10"
          />
        );

      case 'transmission':
        return (
          <Select value={values.transmission || "__clear__"} onValueChange={(val) => updateValue('transmission', val === "__clear__" ? "" : val)}>
            <SelectTrigger><SelectValue placeholder="Выберите тип трансмиссии" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="__clear__">— Очистить —</SelectItem>
              {transmissionOptions.map((opt) => (
                <SelectItem key={opt} value={opt}>{opt}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        );

      case 'transmission_model':
        return (
          <Input
            placeholder="Модель / номер КПП"
            value={values.transmission_model}
            onChange={(e) => updateValue('transmission_model', e.target.value)}
          />
        );

      case 'drive':
        return (
          <Select value={values.drive || "__clear__"} onValueChange={(val) => updateValue('drive', val === "__clear__" ? "" : val)}>
            <SelectTrigger><SelectValue placeholder="Выберите привод" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="__clear__">— Очистить —</SelectItem>
              {driveOptions.map((opt) => (
                <SelectItem key={opt} value={opt}>{opt}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        );

      case 'color':
        return (
          <SearchableSelect
            value={values.color}
            onValueChange={(val) => updateValue('color', val)}
            options={DEFAULT_COLORS.map((c) => ({ value: c, label: c }))}
            placeholder="Выберите или введите цвет деталей"
            searchPlaceholder="Поиск или ввод цвета..."
            allowCustom
            className="h-10"
          />
        );

      case 'condition':
        return (
          <Select value={values.condition || "__clear__"} onValueChange={(val) => updateValue('condition', val === "__clear__" ? "" : val)}>
            <SelectTrigger><SelectValue placeholder="Выберите состояние" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="__clear__">— Очистить —</SelectItem>
              {conditionOptions.map((opt) => (
                <SelectItem key={opt} value={opt}>{opt}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        );

      case 'defect':
        return (
          <Input
            placeholder="Описание дефекта (оставьте пустым для очистки)"
            value={values.defect}
            onChange={(e) => updateValue('defect', e.target.value)}
          />
        );

      case 'supplier_code':
        return (
          <Input
            placeholder="Код поставщика / партии"
            value={values.supplier_code}
            onChange={(e) => updateValue('supplier_code', e.target.value)}
          />
        );

      case 'front_rear':
        return (
          <Select value={values.front_rear || "__clear__"} onValueChange={(val) => updateValue('front_rear', val === "__clear__" ? "" : val)}>
            <SelectTrigger><SelectValue placeholder="Выберите Перед / Зад" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="__clear__">— Очистить —</SelectItem>
              <SelectItem value="F">F (Перед)</SelectItem>
              <SelectItem value="R">R (Зад)</SelectItem>
            </SelectContent>
          </Select>
        );

      case 'left_right':
        return (
          <Select value={values.left_right || "__clear__"} onValueChange={(val) => updateValue('left_right', val === "__clear__" ? "" : val)}>
            <SelectTrigger><SelectValue placeholder="Выберите Лево / Право" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="__clear__">— Очистить —</SelectItem>
              <SelectItem value="L">L (Лево)</SelectItem>
              <SelectItem value="R">R (Право)</SelectItem>
            </SelectContent>
          </Select>
        );

      case 'top_bottom':
        return (
          <Input
            placeholder="Например: Верх / Низ"
            value={values.top_bottom}
            onChange={(e) => updateValue('top_bottom', e.target.value)}
          />
        );

      case 'number':
        return (
          <Input
            placeholder="Номер детали"
            value={values.number}
            onChange={(e) => updateValue('number', e.target.value)}
          />
        );

      case 'manufacturer':
        return (
          <Input
            placeholder="Производитель"
            value={values.manufacturer}
            onChange={(e) => updateValue('manufacturer', e.target.value)}
          />
        );

      case 'manufacturer_code':
        return (
          <Input
            placeholder="Код производителя"
            value={values.manufacturer_code}
            onChange={(e) => updateValue('manufacturer_code', e.target.value)}
          />
        );

      case 'oem_code':
        return (
          <Input
            placeholder="OEM код"
            value={values.oem_code}
            onChange={(e) => updateValue('oem_code', e.target.value)}
          />
        );

      case 'wear_percentage':
        return (
          <Input
            type="number"
            placeholder="Процент износа (0-100)"
            value={values.wear_percentage}
            onChange={(e) => updateValue('wear_percentage', e.target.value)}
          />
        );

      case 'season':
        return (
          <Select value={values.season || "__clear__"} onValueChange={(val) => updateValue('season', val === "__clear__" ? "" : val)}>
            <SelectTrigger><SelectValue placeholder="Выберите сезон" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="__clear__">— Очистить —</SelectItem>
              {seasonOptions.map((opt) => (
                <SelectItem key={opt} value={opt}>{opt}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        );

      case 'diameter':
        return (
          <Input
            placeholder="Например: R16"
            value={values.diameter}
            onChange={(e) => updateValue('diameter', e.target.value)}
          />
        );

      case 'width':
        return (
          <Input
            placeholder="Например: 205"
            value={values.width}
            onChange={(e) => updateValue('width', e.target.value)}
          />
        );

      case 'profile':
        return (
          <Input
            placeholder="Например: 55"
            value={values.profile}
            onChange={(e) => updateValue('profile', e.target.value)}
          />
        );

      case 'tire_quantity':
        return (
          <Input
            placeholder="Например: 4"
            value={values.tire_quantity}
            onChange={(e) => updateValue('tire_quantity', e.target.value)}
          />
        );

      case 'drilling':
        return (
          <Input
            placeholder="Например: 5x114.3"
            value={values.drilling}
            onChange={(e) => updateValue('drilling', e.target.value)}
          />
        );

      case 'offset':
        return (
          <Input
            placeholder="Например: ET45"
            value={values.offset}
            onChange={(e) => updateValue('offset', e.target.value)}
          />
        );

      case 'center_hole_diameter':
        return (
          <Input
            placeholder="Например: 67.1"
            value={values.center_hole_diameter}
            onChange={(e) => updateValue('center_hole_diameter', e.target.value)}
          />
        );

      case 'tire_model':
        return (
          <Input
            placeholder="Модель шины / диска"
            value={values.tire_model}
            onChange={(e) => updateValue('tire_model', e.target.value)}
          />
        );
    }
  };

  const countDisplay = scopeType === 'filter' ? 'по текущему фильтру' : `${selectedPartIds.length} выбранных`;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[680px] max-h-[90vh] flex flex-col overflow-hidden">
        <DialogHeader className="shrink-0">
          <DialogTitle className="flex items-center gap-2">
            <SlidersHorizontal className="h-5 w-5 text-primary" />
            Массовое редактирование запчастей
          </DialogTitle>
          <DialogDescription>
            Отметьте галочками любые строки (поля), которые нужно изменить для {countDisplay} запчастей.
          </DialogDescription>
        </DialogHeader>

        {activeFilters && Object.keys(activeFilters).length > 0 && (
          <div className="bg-muted/50 p-2.5 rounded-lg border flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 shrink-0">
            <span className="text-xs font-medium text-muted-foreground">Область применения:</span>
            <div className="flex gap-2">
              <Button
                type="button"
                size="sm"
                variant={scopeType === 'ids' ? 'default' : 'outline'}
                onClick={() => setScopeType('ids')}
                className="text-xs h-7"
              >
                Выбранные вручную ({selectedPartIds.length})
              </Button>
              <Button
                type="button"
                size="sm"
                variant={scopeType === 'filter' ? 'default' : 'outline'}
                onClick={() => setScopeType('filter')}
                className="text-xs h-7"
              >
                Все по активному фильтру
              </Button>
            </div>
          </div>
        )}

        {/* Панель быстрого поиска поля и вкладок */}
        <div className="space-y-2.5 shrink-0 pt-1">
          <div className="flex flex-col sm:flex-row gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                type="text"
                placeholder="Быстрый поиск строки (например: цвет, год, VIN, КПП, цена, кузов)..."
                value={fieldSearch}
                onChange={(e) => setFieldSearch(e.target.value)}
                className="pl-9 h-9 text-sm"
              />
            </div>
            <div className="flex items-center gap-1.5 shrink-0">
              <Button
                type="button"
                size="sm"
                variant={showOnlyChecked ? 'default' : 'outline'}
                onClick={() => setShowOnlyChecked((prev) => !prev)}
                className="h-9 text-xs"
              >
                Выбрано ({checkedCount})
              </Button>
              {checkedCount > 0 && (
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  onClick={resetAllFields}
                  className="h-9 px-2.5 text-xs text-muted-foreground hover:text-foreground"
                  title="Сбросить все отмеченные поля"
                >
                  <RotateCcw className="h-3.5 w-3.5" />
                </Button>
              )}
            </div>
          </div>

          {!fieldSearch.trim() && !showOnlyChecked && (
            <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)}>
              <TabsList className="grid grid-cols-4 w-full h-9">
                <TabsTrigger value="basic" className="text-xs gap-1">
                  Основные
                  {sectionCounts.basic > 0 && (
                    <Badge variant="secondary" className="h-4 px-1 text-[10px]">{sectionCounts.basic}</Badge>
                  )}
                </TabsTrigger>
                <TabsTrigger value="specs" className="text-xs gap-1">
                  Характеристики
                  {sectionCounts.specs > 0 && (
                    <Badge variant="secondary" className="h-4 px-1 text-[10px]">{sectionCounts.specs}</Badge>
                  )}
                </TabsTrigger>
                <TabsTrigger value="tires" className="text-xs gap-1">
                  Шины и диски
                  {sectionCounts.tires > 0 && (
                    <Badge variant="secondary" className="h-4 px-1 text-[10px]">{sectionCounts.tires}</Badge>
                  )}
                </TabsTrigger>
                <TabsTrigger value="all" className="text-xs">
                  Все ({FIELD_METAS.length})
                </TabsTrigger>
              </TabsList>
            </Tabs>
          )}
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col flex-1 min-h-0 space-y-3">
          <div className="flex-1 overflow-y-auto pr-1 space-y-2.5 py-1">
            {visibleMetas.length === 0 ? (
              <div className="text-center py-8 text-sm text-muted-foreground border rounded-lg bg-muted/20">
                {showOnlyChecked
                  ? 'Вы ещё не отметили ни одного поля для изменения'
                  : `Поля по запросу «${fieldSearch}» не найдены`}
              </div>
            ) : (
              visibleMetas.map((meta) => {
                const isChecked = checkedFields[meta.key];
                return (
                  <div
                    key={meta.key}
                    className={`rounded-lg border p-3 space-y-2.5 transition-colors ${
                      isChecked ? 'bg-primary/5 border-primary/40 shadow-xs' : 'bg-card'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2.5">
                        <Checkbox
                          id={`bulk-field-${meta.key}`}
                          checked={isChecked}
                          onCheckedChange={(checked) => toggleField(meta.key, Boolean(checked))}
                        />
                        <Label
                          htmlFor={`bulk-field-${meta.key}`}
                          className="text-sm font-semibold cursor-pointer select-none"
                        >
                          {meta.label}
                        </Label>
                      </div>
                      {meta.key === 'price' && isChecked && (
                        <Badge variant="outline" className="text-xs">
                          {values.priceMode === 'PRICE_OPERATION_MODE_SET' && 'Точное значение'}
                          {values.priceMode === 'PRICE_OPERATION_MODE_PERCENT' && 'Процент (+/-)'}
                          {values.priceMode === 'PRICE_OPERATION_MODE_OFFSET' && 'Смещение (+/- ₽)'}
                        </Badge>
                      )}
                    </div>

                    {isChecked && (
                      <div className="pt-0.5">
                        {renderFieldInput(meta.key)}
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>

          {isBlankFilter && (
            <div className="p-3 border border-orange-400 bg-orange-100/90 dark:bg-orange-950/60 rounded-lg shrink-0">
              <p className="text-xs font-bold text-orange-900 dark:text-orange-200 uppercase mb-1">
                ⚠️ Внимание: Фильтры поиска не заданы!
              </p>
              <p className="text-xs text-orange-800 dark:text-orange-300 mb-2 leading-relaxed">
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

          <DialogFooter className="pt-2 border-t shrink-0">
            <Button type="button" variant="outline" onClick={onClose} disabled={isLoading}>
              Отмена
            </Button>
            <Button
              type="submit"
              disabled={isLoading || checkedCount === 0 || (isBlankFilter && !confirmEntireCatalog)}
              className="gap-2"
            >
              {isLoading ? (
                'Применение...'
              ) : (
                <>
                  <CheckCircle2 className="h-4 w-4" />
                  Применить изменения ({checkedCount})
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
