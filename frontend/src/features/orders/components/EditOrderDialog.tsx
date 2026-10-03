/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Trash2 } from 'lucide-react';
import type { Order, OrderItem } from '@/lib/types';
import { PAYMENT_OPTIONS, WAREHOUSE_OPTIONS, DELIVERY_OPTIONS } from '../constants';

interface EditOrderDialogProps {
  order: Order | null;
  isOpen: boolean;
  onClose: () => void;
  onSave: (orderId: number, details: {
    buyer_number: string;
    order_number: string;
    source: string;
    payment_status: string;
    warehouse_status: string;
    delivery_method: string;
    transport_company: string;
    tracking_number: string;
    notes: string;
    discount: number;
    items: OrderItem[];
  }) => Promise<void>;
  onDeleteItem: (orderId: number, itemId: number) => Promise<void>;
  isSaving: boolean;
}

export const EditOrderDialog: React.FC<EditOrderDialogProps> = ({
  order,
  isOpen,
  onClose,
  onSave,
  onDeleteItem,
  isSaving,
}) => {
  const [buyer, setBuyer] = useState('');
  const [orderNumber, setOrderNumber] = useState('');
  const [source, setSource] = useState('drom');
  const [payment, setPayment] = useState('unpaid');
  const [warehouse, setWarehouse] = useState('inspecting');
  const [delivery, setDelivery] = useState('pickup');
  const [tk, setTk] = useState('');
  const [tracking, setTracking] = useState('');
  const [notes, setNotes] = useState('');
  const [discount, setDiscount] = useState(0);
  const [items, setItems] = useState<OrderItem[]>([]);

  useEffect(() => {
    if (order) {
      setBuyer(order.buyer_number || '');
      setOrderNumber(order.order_number || '');
      setSource(order.source || 'drom');
      setPayment(order.payment_status || 'unpaid');
      setWarehouse(order.warehouse_status || 'inspecting');
      setDelivery(order.delivery_method || 'pickup');
      setTk(order.transport_company || '');
      setTracking(order.tracking_number || '');
      setNotes(order.notes || '');
      setDiscount(order.discount || 0);
      setItems(order.items ? order.items.map((it) => ({ ...it })) : []);
    }
  }, [order]);

  const handleSave = async () => {
    if (!order) return;
    await onSave(order.id, {
      buyer_number: buyer,
      order_number: orderNumber,
      source,
      payment_status: payment,
      warehouse_status: warehouse,
      delivery_method: delivery,
      transport_company: tk,
      tracking_number: tracking,
      notes,
      discount: Number(discount) || 0,
      items,
    });
  };

  const handleRemoveItem = async (itemId: number) => {
    if (!order) return;
    await onDeleteItem(order.id, itemId);
    setItems((prev) => prev.filter((it) => it.id !== itemId));
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-[620px] max-h-[92vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Редактирование заказа #{order?.id}</DialogTitle>
          <DialogDescription>
            Изменение реквизитов покупателя, логистики (ТК и трек-номера), скидки и цен позиций.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="space-y-1.5">
              <Label>Покупатель / Телефон</Label>
              <Input value={buyer} onChange={(e) => setBuyer(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label>№ сделки (Дром / Авито)</Label>
              <Input value={orderNumber} onChange={(e) => setOrderNumber(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label>Источник</Label>
              <Select value={source} onValueChange={setSource}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="drom">Дром</SelectItem>
                  <SelectItem value="avito">Авито</SelectItem>
                  <SelectItem value="messenger">Мессенджер</SelectItem>
                  <SelectItem value="pickup">Самовывоз</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="space-y-1.5">
              <Label>💰 Оплата</Label>
              <Select value={payment} onValueChange={setPayment}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {PAYMENT_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>🏭 Склад</Label>
              <Select value={warehouse} onValueChange={setWarehouse}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {WAREHOUSE_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>🚚 Способ выдачи</Label>
              <Select value={delivery} onValueChange={setDelivery}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {DELIVERY_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label>Транспортная компания / Город</Label>
              <Input
                value={tk}
                onChange={(e) => setTk(e.target.value)}
                placeholder="Например: СДЭК, Энергия..."
              />
            </div>
            <div className="space-y-1.5">
              <Label>Трек-номер (накладная ТК)</Label>
              <Input
                value={tracking}
                onChange={(e) => setTracking(e.target.value)}
                placeholder="Введите номер отслеживания"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="sm:col-span-2 space-y-1.5">
              <Label>Комментарий / Заметка для склада</Label>
              <Input
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="Примечания к сборке, упаковке или отправке"
              />
            </div>
            <div className="space-y-1.5">
              <Label>Скидка на заказ (₽)</Label>
              <Input
                type="number"
                min={0}
                value={discount}
                onChange={(e) => setDiscount(Math.max(0, Number(e.target.value) || 0))}
              />
            </div>
          </div>

          {/* Позиции заказа */}
          {items.length > 0 && (
            <div className="border-t pt-3 space-y-2">
              <Label className="text-sm font-medium block">Позиции заказа (торг / количество):</Label>
              {items.map((item, idx) => (
                <div
                  key={item.id}
                  className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-2.5 rounded-lg border bg-muted/40"
                >
                  <div className="text-sm font-medium truncate flex-1">
                    {item.part_name || item.part_name_snapshot || `Запчасть #${item.part_id}`}
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex items-center gap-1">
                      <span className="text-xs text-muted-foreground">Кол-во:</span>
                      <Input
                        type="number"
                        min={1}
                        value={item.quantity}
                        onChange={(e) => {
                          const val = Math.max(1, Number(e.target.value) || 1);
                          setItems((prev) =>
                            prev.map((it, i) => (i === idx ? { ...it, quantity: val } : it))
                          );
                        }}
                        className="w-16 h-8 text-xs"
                      />
                    </div>
                    <div className="flex items-center gap-1">
                      <span className="text-xs text-muted-foreground">Цена ₽:</span>
                      <Input
                        type="number"
                        min={0}
                        value={item.price || 0}
                        onChange={(e) => {
                          const val = Math.max(0, Number(e.target.value) || 0);
                          setItems((prev) =>
                            prev.map((it, i) => (i === idx ? { ...it, price: val } : it))
                          );
                        }}
                        className="w-24 h-8 text-xs"
                      />
                    </div>
                    {items.length > 1 && order && !order.auto_deleted && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="text-destructive h-8 px-2"
                        onClick={() => handleRemoveItem(item.id)}
                        title="Убрать позицию из заказа"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Отмена
          </Button>
          <Button onClick={handleSave} disabled={isSaving}>
            {isSaving ? 'Сохранение...' : 'Сохранить изменения'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
