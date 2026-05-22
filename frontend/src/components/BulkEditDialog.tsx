/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import React, { useState, useEffect } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { getAuthHeaders } from "@/lib/csrf";
import { logUserActivity } from "@/features/admin/api/adminApi";
import { useQueryClient } from '@tanstack/react-query';
import { partsKeys } from '@/features/parts/hooks/usePartsQueries';
import { API_BASE_URL } from '@/lib/api';

interface BulkEditDialogProps {
  isOpen: boolean;
  onClose: () => void;
  selectedPartIds: number[];
  onSuccess: () => void;
}

interface BulkEditForm {
  category?: string;
  price?: string;
  location?: string;
  status?: boolean;
  salesman?: string;
  // Поля для применения изменений
  applyCategory: boolean;
  applyPrice: boolean;
  applyLocation: boolean;
  applyStatus: boolean;
  applySalesman: boolean;
}

export default function BulkEditDialog({ isOpen, onClose, selectedPartIds, onSuccess }: BulkEditDialogProps) {
  const queryClient = useQueryClient();
  const [isLoading, setIsLoading] = useState(false);
  const [users, setUsers] = useState<{id: number, name: string, email: string}[]>([]);
  const [form, setForm] = useState<BulkEditForm>({
    category: '',
    price: '',
    location: '',
    status: true,
    salesman: '',
    applyCategory: false,
    applyPrice: false,
    applyLocation: false,
    applyStatus: false,
    applySalesman: false,
  });

  // Загружаем список пользователей при открытии диалога
  useEffect(() => {
    if (isOpen) {
      fetchUsers();
    }
  }, [isOpen]);

  const fetchUsers = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/admin/users`, {
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      // Собираем только те поля, которые отмечены для применения
      const updates: Partial<BulkEditForm> = {};

      if (form.applyCategory && form.category) updates.category = form.category;
      if (form.applyPrice && form.price) updates.price = form.price;
      if (form.applyLocation && form.location) updates.location = form.location;
      if (form.applyStatus !== undefined) updates.status = form.status;
      if (form.applySalesman && form.salesman) updates.salesman = form.salesman;

      if (Object.keys(updates).length === 0) {
        alert('Выберите хотя бы одно поле для изменения');
        setIsLoading(false);
        return;
      }

      // Формируем массив обновлений для каждого ID
      const bulkUpdates = selectedPartIds.map(partId => ({
        id: partId,
        ...updates,
      }));

      // Отправляем запрос на массовое обновление
      const response = await fetch(`${API_BASE_URL}/api/admin/bulk-update-parts`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
        },
        credentials: 'include',
        body: JSON.stringify(bulkUpdates),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to update parts');
      }

      const result = await response.json();

      // Логируем массовое обновление
      try {
        await logUserActivity({
          action: 'bulk_update_parts',
          resource_type: 'part',
          details: `Массовое обновление ${selectedPartIds.length} запчастей. Изменения: ${Object.keys(updates).join(', ')}`,
        });
      } catch (logError) {
        console.warn('Failed to log bulk update activity:', logError);
      }

      // Обновляем кэш
      queryClient.invalidateQueries({ queryKey: partsKeys.lists() });

      alert(`Успешно обновлено ${result.updated_count} запчастей`);
      onSuccess();
    } catch (error) {
      console.error('Error bulk updating parts:', error);
      alert('Ошибка при массовом обновлении запчастей');
    } finally {
      setIsLoading(false);
    }
  };

  const updateForm = (field: keyof BulkEditForm, value: string | boolean) => {
    setForm(prev => ({ ...prev, [field]: value }));
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Массовое редактирование запчастей</DialogTitle>
          <DialogDescription>
            Выберите поля для изменения. Изменения будут применены ко всем {selectedPartIds.length} выбранным запчастям.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Категория */}
          <div className="flex items-center space-x-2">
            <Checkbox
              id="applyCategory"
              checked={form.applyCategory}
              onCheckedChange={(checked) => updateForm('applyCategory', checked as boolean)}
            />
            <Label htmlFor="applyCategory" className="text-sm font-medium">Категория</Label>
          </div>
          {form.applyCategory && (
            <Select value={form.category || ""} onValueChange={(value) => updateForm('category', value)}>
              <SelectTrigger>
                <SelectValue placeholder="Выберите категорию" />
              </SelectTrigger>
              <SelectContent>
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

          {/* Цена */}
          <div className="flex items-center space-x-2">
            <Checkbox
              id="applyPrice"
              checked={form.applyPrice}
              onCheckedChange={(checked) => updateForm('applyPrice', checked as boolean)}
            />
            <Label htmlFor="applyPrice" className="text-sm font-medium">Цена (₽)</Label>
          </div>
          {form.applyPrice && (
            <Input
              type="number"
              step="100"
              placeholder="Введите цену"
              value={form.price}
              onChange={(e) => updateForm('price', e.target.value)}
            />
          )}

          {/* Местоположение */}
          <div className="flex items-center space-x-2">
            <Checkbox
              id="applyLocation"
              checked={form.applyLocation}
              onCheckedChange={(checked) => updateForm('applyLocation', checked as boolean)}
            />
            <Label htmlFor="applyLocation" className="text-sm font-medium">Местоположение</Label>
          </div>
          {form.applyLocation && (
            <Input
              placeholder="Введите местоположение"
              value={form.location}
              onChange={(e) => updateForm('location', e.target.value)}
            />
          )}

          {/* Статус */}
          <div className="flex items-center space-x-2">
            <Checkbox
              id="applyStatus"
              checked={form.applyStatus}
              onCheckedChange={(checked) => updateForm('applyStatus', checked as boolean)}
            />
            <Label htmlFor="applyStatus" className="text-sm font-medium">Статус</Label>
          </div>
          {form.applyStatus && (
            <Select value={form.status ? "active" : "inactive"} onValueChange={(value) => updateForm('status', value === "active")}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="active">Активный</SelectItem>
                <SelectItem value="inactive">Неактивный</SelectItem>
              </SelectContent>
            </Select>
          )}

          {/* Продавец */}
          <div className="flex items-center space-x-2">
            <Checkbox
              id="applySalesman"
              checked={form.applySalesman}
              onCheckedChange={(checked) => updateForm('applySalesman', checked as boolean)}
            />
            <Label htmlFor="applySalesman" className="text-sm font-medium">Продавец</Label>
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

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Отмена
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Сохранение...' : 'Применить изменения'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
