/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState } from 'react';
import { useCreateCustomer, useUpdateCustomer } from '@/hooks/useCustomers';
import type { CustomerWithStats, CustomerCategory } from '@/lib/types';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

export interface CustomerFormModalProps {
  isOpen: boolean;
  customer?: CustomerWithStats | null;
  onClose: () => void;
}

export const CustomerFormModal: React.FC<CustomerFormModalProps> = ({ isOpen, customer, onClose }) => {
  const isEditing = Boolean(customer);
  const createMutation = useCreateCustomer();
  const updateMutation = useUpdateCustomer();

  const [name, setName] = useState(customer?.name || '');
  const [phone, setPhone] = useState(customer?.phone || '');
  const [city, setCity] = useState(customer?.city || '');
  const [preferredTk, setPreferredTk] = useState(customer?.preferred_tk || '');
  const [passportOrInn, setPassportOrInn] = useState(customer?.passport_or_inn || '');
  const [category, setCategory] = useState<CustomerCategory>(customer?.category || 'regular');
  const [discountPercent, setDiscountPercent] = useState<number>(customer?.discount_percent || 0);
  const [notes, setNotes] = useState(customer?.notes || '');
  const [errorMsg, setErrorMsg] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() && !phone.trim()) {
      setErrorMsg('Укажите ФИО или телефон клиента');
      return;
    }

    try {
      if (isEditing && customer) {
        await updateMutation.mutateAsync({
          id: customer.id,
          data: {
            name: name.trim(),
            phone: phone.trim(),
            city: city.trim(),
            preferred_tk: preferredTk.trim(),
            passport_or_inn: passportOrInn.trim(),
            category,
            discount_percent: Number(discountPercent) || 0,
            notes: notes.trim(),
          },
        });
      } else {
        await createMutation.mutateAsync({
          name: name.trim() || phone.trim(),
          phone: phone.trim(),
          city: city.trim(),
          preferred_tk: preferredTk.trim(),
          passport_or_inn: passportOrInn.trim(),
          category,
          discount_percent: Number(discountPercent) || 0,
          notes: notes.trim(),
        });
      }
      onClose();
    } catch (err: unknown) {
      setErrorMsg(err instanceof Error ? err.message : 'Ошибка сохранения клиента');
    }
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
        <form onSubmit={handleSubmit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>{isEditing ? `Редактирование клиента #${customer?.id}` : 'Новый клиент'}</DialogTitle>
            <DialogDescription className="text-xs">
              {isEditing
                ? 'Обновите контактные данные, категорию или реквизиты для ТК'
                : 'Заполните карточку покупателя для быстрого оформления заказов'}
            </DialogDescription>
          </DialogHeader>

          {errorMsg && (
            <div className="p-2.5 bg-destructive/10 border border-destructive/20 text-destructive rounded text-xs">
              {errorMsg}
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label className="text-xs">Телефон *</Label>
              <Input
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+7 (999) 000-00-00"
                required
              />
            </div>

            <div className="space-y-1">
              <Label className="text-xs">ФИО / Организация</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Иванов Иван или СТО"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label className="text-xs">Город доставки</Label>
              <Input
                value={city}
                onChange={(e) => setCity(e.target.value)}
                placeholder="Новосибирск, Красноярск..."
              />
            </div>

            <div className="space-y-1">
              <Label className="text-xs">ТК / Адрес терминала</Label>
              <Input
                value={preferredTk}
                onChange={(e) => setPreferredTk(e.target.value)}
                placeholder="Энергия (терминал 1) / СДЭК"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label className="text-xs">Категория клиента</Label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value as CustomerCategory)}
                className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-xs shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <option value="regular">👤 Обычный розничный</option>
                <option value="vip">⭐ Постоянный / СТО</option>
                <option value="wholesale">🏢 Оптовик / Перекуп</option>
                <option value="blacklist">⛔ Чёрный список / Проблемный</option>
              </select>
            </div>

            <div className="space-y-1">
              <Label className="text-xs">Скидка клиента (%)</Label>
              <Input
                type="number"
                min="0"
                max="100"
                value={discountPercent}
                onChange={(e) => setDiscountPercent(Number(e.target.value))}
                placeholder="0"
              />
            </div>
          </div>

          <div className="space-y-1">
            <Label className="text-xs">Данные для ТК (Паспорт серия/номер или ИНН)</Label>
            <Input
              value={passportOrInn}
              onChange={(e) => setPassportOrInn(e.target.value)}
              placeholder="Серия 1234 № 567890 или ИНН организации"
            />
          </div>

          <div className="space-y-1">
            <Label className="text-xs">Заметки / Примечание {category === 'blacklist' && '(Причина внесения в ЧС)'}</Label>
            <Textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder={
                category === 'blacklist'
                  ? 'Укажите причину: например, не забрал груз из ТК, необоснованный возврат...'
                  : 'Особые пожелания, упаковка, скидки...'
              }
              rows={2}
            />
          </div>

          <DialogFooter className="pt-2">
            <Button type="button" variant="outline" size="sm" onClick={onClose} disabled={isPending}>
              Отмена
            </Button>
            <Button type="submit" size="sm" disabled={isPending}>
              {isPending ? 'Сохранение...' : isEditing ? 'Сохранить изменения' : 'Создать клиента'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
