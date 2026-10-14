/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Filter, Check, RotateCcw } from 'lucide-react';
import type { Part } from '@/lib/types';

interface SelectByCriteriaDialogProps {
  isOpen: boolean;
  onClose: () => void;
  parts: Part[];
  onSelectMatching: (matchedIds: number[]) => void;
}

const CATEGORIES = [
  "Все категории",
  "Тормоза",
  "Двигатель",
  "Подвеска",
  "Подвеска ДВС/КПП",
  "Подвеска передних колес",
  "Подвеска задних колес",
  "Электрика",
  "Кузов",
  "Кузов снаружи",
  "Интерьер",
  "Трансмиссия",
  "Система охлаждения и отопления",
  "Система выхлопа (Глушитель)",
  "Система рулевого управления",
  "Рулевое управление",
  "Система фильтрации (Фильтры)",
  "Шины и диски",
  "Автохимия и масла",
  "Аксессуары и тюннинг",
  "Другое"
];

export default function SelectByCriteriaDialog({
  isOpen,
  onClose,
  parts,
  onSelectMatching,
}: SelectByCriteriaDialogProps) {
  const [minPrice, setMinPrice] = useState<string>('');
  const [maxPrice, setMaxPrice] = useState<string>('');
  const [category, setCategory] = useState<string>('Все категории');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive'>('all');

  const handleReset = () => {
    setMinPrice('');
    setMaxPrice('');
    setCategory('Все категории');
    setStatusFilter('all');
  };

  const handleApply = () => {
    const min = minPrice ? parseFloat(minPrice.replace(',', '.')) : undefined;
    const max = maxPrice ? parseFloat(maxPrice.replace(',', '.')) : undefined;
    const selectedCategory = category !== 'Все категории' ? category : undefined;

    const matched = parts.filter((part) => {
      // Фильтр по минимальной цене
      if (min !== undefined && !isNaN(min)) {
        if ((part.price ?? 0) < min) return false;
      }
      // Фильтр по максимальной цене
      if (max !== undefined && !isNaN(max)) {
        if ((part.price ?? 0) > max) return false;
      }
      // Фильтр по категории
      if (selectedCategory && part.category !== selectedCategory) {
        return false;
      }
      // Фильтр по статусу
      if (statusFilter === 'active' && part.status === false) {
        return false;
      }
      if (statusFilter === 'inactive' && part.status !== false) {
        return false;
      }
      return true;
    });

    onSelectMatching(matched.map((p) => p.id));
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[460px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5 text-primary" />
            Выбрать запчасти по критерию
          </DialogTitle>
          <DialogDescription>
            Задайте параметры для автоматического выделения запчастей в текущем списке.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          {/* Диапазон цен */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">Диапазон цен (₽)</Label>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Input
                  type="number"
                  placeholder="От (₽)"
                  value={minPrice}
                  onChange={(e) => setMinPrice(e.target.value)}
                />
              </div>
              <div>
                <Input
                  type="number"
                  placeholder="До (₽)"
                  value={maxPrice}
                  onChange={(e) => setMaxPrice(e.target.value)}
                />
              </div>
            </div>
          </div>

          {/* Категория */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">Категория</Label>
            <Select value={category} onValueChange={setCategory}>
              <SelectTrigger>
                <SelectValue placeholder="Все категории" />
              </SelectTrigger>
              <SelectContent className="max-h-60">
                {CATEGORIES.map((cat) => (
                  <SelectItem key={cat} value={cat}>
                    {cat}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Статус */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">Статус</Label>
            <Select value={statusFilter} onValueChange={(val) => setStatusFilter(val as 'all' | 'active' | 'inactive')}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Все статусы</SelectItem>
                <SelectItem value="active">Только активные</SelectItem>
                <SelectItem value="inactive">Только неактивные</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter className="flex sm:justify-between items-center gap-2 pt-2">
          <Button type="button" variant="ghost" size="sm" onClick={handleReset} className="gap-1.5 text-xs text-muted-foreground">
            <RotateCcw className="h-3.5 w-3.5" />
            Сбросить
          </Button>
          <div className="flex gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Отмена
            </Button>
            <Button type="button" onClick={handleApply} className="gap-1.5">
              <Check className="h-4 w-4" />
              Выделить
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
