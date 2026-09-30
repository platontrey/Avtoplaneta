/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useState, useEffect } from "react";
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createOrder } from '@/features/orders/api/ordersApi';
import { sanitizeText, sanitizeNumber, INPUT_LIMITS } from "@/hooks/usePartValidation";
import { Zap } from "lucide-react";
import type { Part } from '@/features/parts/types';

interface BulkOrderDialogProps {
  isOpen: boolean;
  onClose: () => void;
  selectedParts: Part[];
  onSuccess: () => void;
}

export default function BulkOrderDialog({ isOpen, onClose, selectedParts, onSuccess }: BulkOrderDialogProps) {
  const [buyerNumber, setBuyerNumber] = useState('');
  const [orderNumber, setOrderNumber] = useState('');
  const [source, setSource] = useState('drom');
  const [paymentStatus, setPaymentStatus] = useState('unpaid');
  const [warehouseStatus, setWarehouseStatus] = useState('inspecting');
  const [deliveryMethod, setDeliveryMethod] = useState('tk');
  const [transportCompany, setTransportCompany] = useState('');
  const [notes, setNotes] = useState('');
  const [discount, setDiscount] = useState(0);
  const [quickSale, setQuickSale] = useState(false);
  const [quantities, setQuantities] = useState<Record<number, number>>({});
  const [prices, setPrices] = useState<Record<number, number>>({});

  const queryClient = useQueryClient();

  useEffect(() => {
    if (isOpen && selectedParts.length > 0) {
      const initialQuantities: Record<number, number> = {};
      const initialPrices: Record<number, number> = {};
      selectedParts.forEach((part) => {
        initialQuantities[part.id] = 1;
        initialPrices[part.id] = Number(part.price) || 0;
      });
      setQuantities(initialQuantities);
      setPrices(initialPrices);
    }
  }, [isOpen, selectedParts]);

  const createBulkOrderMutation = useMutation({
    mutationFn: async () => {
      const items = selectedParts.map((part) => ({
        part_id: part.id,
        quantity: Math.max(1, quantities[part.id] || 1),
        price: prices[part.id] ?? (Number(part.price) || 0),
      }));

      return createOrder({
        customer_id: 0,
        order_number: orderNumber.trim(),
        source: quickSale ? 'pickup' : source,
        part: selectedParts.map((p) => p.name).join(', '),
        part_id: selectedParts[0]?.id,
        buyer_number: buyerNumber.trim() || (quickSale ? 'Самовывоз' : ''),
        payment_status: quickSale ? 'paid' : paymentStatus,
        warehouse_status: quickSale ? 'ready' : warehouseStatus,
        delivery_method: quickSale ? 'pickup' : deliveryMethod,
        transport_company: transportCompany.trim(),
        notes: notes.trim(),
        discount: Math.max(0, Number(discount) || 0),
        quick_sale: quickSale,
        items,
      });
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['orders'] });
      void queryClient.invalidateQueries({ queryKey: ['orders', 'completed'] });
      void queryClient.invalidateQueries({ queryKey: ['parts'] });
      void queryClient.invalidateQueries({ queryKey: ['inventory'] });
      void queryClient.invalidateQueries({ queryKey: ['statistics'] });
      onSuccess();
      handleClose();
    },
    onError: (error) => {
      console.error('Failed to create bulk order:', error);
      alert('Ошибка при создании заказа');
    },
  });

  const handleClose = () => {
    setBuyerNumber('');
    setOrderNumber('');
    setSource('drom');
    setPaymentStatus('unpaid');
    setWarehouseStatus('inspecting');
    setDeliveryMethod('tk');
    setTransportCompany('');
    setNotes('');
    setDiscount(0);
    setQuickSale(false);
    setQuantities({});
    setPrices({});
    onClose();
  };

  const subtotal = selectedParts.reduce(
    (sum, part) => sum + (prices[part.id] ?? (Number(part.price) || 0)) * (quantities[part.id] || 1),
    0
  );
  const totalAmount = Math.max(0, subtotal - (Number(discount) || 0));
  const isValid = (quickSale || buyerNumber.trim().length > 0) && selectedParts.length > 0;

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[640px] max-h-[92vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-lg sm:text-xl">
            Оформление заказа ({selectedParts.length} запч.)
          </DialogTitle>
          <DialogDescription className="text-sm">
            Проверьте состав заказа, укажите контакт покупателя или проведите быструю продажу с места.
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-2 pt-1">
          <Button
            type="button"
            variant={!quickSale ? 'default' : 'outline'}
            size="sm"
            onClick={() => setQuickSale(false)}
          >
            📦 В заказ (В работу / ТК)
          </Button>
          <Button
            type="button"
            variant={quickSale ? 'default' : 'outline'}
            size="sm"
            onClick={() => setQuickSale(true)}
            className="gap-1.5"
          >
            <Zap className="h-4 w-4" />
            Быстрая продажа с места
          </Button>
        </div>

        <div className="space-y-4 py-2">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="bulk_buyer">
                Покупатель (телефон / имя) {quickSale ? '(необязательно)' : '*'}
              </Label>
              <Input
                id="bulk_buyer"
                value={buyerNumber}
                onChange={(e) => setBuyerNumber(sanitizeText(e.target.value))}
                placeholder={quickSale ? 'Самовывоз' : '+7 999 ... или имя клиента'}
                maxLength={INPUT_LIMITS.TEXT_MAX_LENGTH}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="bulk_order_num">№ сделки (Дром / Авито)</Label>
              <Input
                id="bulk_order_num"
                value={orderNumber}
                onChange={(e) => setOrderNumber(sanitizeText(e.target.value))}
                placeholder="Например: 14829104"
              />
            </div>
          </div>

          {!quickSale && (
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
              <div className="space-y-1">
                <Label className="text-xs">Источник</Label>
                <Select value={source} onValueChange={setSource}>
                  <SelectTrigger className="h-9 text-xs">
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
              <div className="space-y-1">
                <Label className="text-xs">💰 Оплата</Label>
                <Select value={paymentStatus} onValueChange={setPaymentStatus}>
                  <SelectTrigger className="h-9 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="unpaid">Не оплачен</SelectItem>
                    <SelectItem value="prepaid">Предоплата</SelectItem>
                    <SelectItem value="paid">Оплачен</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1">
                <Label className="text-xs">🏭 Склад</Label>
                <Select value={warehouseStatus} onValueChange={setWarehouseStatus}>
                  <SelectTrigger className="h-9 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="inspecting">Проверка / Фото</SelectItem>
                    <SelectItem value="transfer">Перемещение</SelectItem>
                    <SelectItem value="ready">Готов</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1">
                <Label className="text-xs">🚚 Получение</Label>
                <Select value={deliveryMethod} onValueChange={setDeliveryMethod}>
                  <SelectTrigger className="h-9 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="tk">Отправка ТК</SelectItem>
                    <SelectItem value="pickup">Самовывоз</SelectItem>
                    <SelectItem value="city">По городу</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          {!quickSale && deliveryMethod === 'tk' && (
            <div className="space-y-1.5">
              <Label htmlFor="bulk_tk">Транспортная компания / Город</Label>
              <Input
                id="bulk_tk"
                value={transportCompany}
                onChange={(e) => setTransportCompany(sanitizeText(e.target.value))}
                placeholder="Например: СДЭК, Энергия, ПЭК"
              />
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="sm:col-span-2 space-y-1.5">
              <Label htmlFor="bulk_notes">Комментарий / Заметка для склада</Label>
              <Input
                id="bulk_notes"
                value={notes}
                onChange={(e) => setNotes(sanitizeText(e.target.value))}
                placeholder="Упаковка, доп. фото, условия выдачи..."
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="bulk_discount">Скидка на заказ (₽)</Label>
              <Input
                id="bulk_discount"
                type="number"
                min={0}
                value={discount}
                onChange={(e) => setDiscount(Math.max(0, Number(e.target.value) || 0))}
              />
            </div>
          </div>

          <div className="border-t pt-3">
            <Label className="text-sm font-medium mb-2 block">Позиции заказа:</Label>
            <div className="space-y-2 max-h-56 overflow-y-auto pr-1">
              {selectedParts.map((part) => (
                <div
                  key={part.id}
                  className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-2.5 bg-muted/50 border border-border/50 rounded-lg"
                >
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-sm truncate">{part.name}</p>
                    <p className="text-xs text-muted-foreground">
                      Склад: {part.location || '—'} · В наличии: {part.quantity ?? 0} шт.
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex items-center gap-1">
                      <Label className="text-xs text-muted-foreground">Кол-во:</Label>
                      <Input
                        type="number"
                        min={1}
                        max={INPUT_LIMITS.QUANTITY_MAX}
                        value={quantities[part.id] || 1}
                        onChange={(e) =>
                          setQuantities((prev) => ({
                            ...prev,
                            [part.id]: Math.max(1, sanitizeNumber(e.target.value)),
                          }))
                        }
                        className="w-16 h-8 text-xs"
                      />
                    </div>
                    <div className="flex items-center gap-1">
                      <Label className="text-xs text-muted-foreground">Цена ₽:</Label>
                      <Input
                        type="number"
                        min={0}
                        value={prices[part.id] ?? (Number(part.price) || 0)}
                        onChange={(e) =>
                          setPrices((prev) => ({
                            ...prev,
                            [part.id]: Math.max(0, Number(e.target.value) || 0),
                          }))
                        }
                        className="w-24 h-8 text-xs"
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="flex items-center justify-between rounded-lg bg-primary/10 px-4 py-2.5 text-sm font-semibold">
            <span>Итого к оплате:</span>
            <span className="text-base">₽{totalAmount.toLocaleString('ru-RU')}</span>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            Отмена
          </Button>
          <Button
            onClick={() => createBulkOrderMutation.mutate()}
            disabled={!isValid || createBulkOrderMutation.isPending}
          >
            {createBulkOrderMutation.isPending
              ? 'Оформление...'
              : quickSale
              ? `Продать за ₽${totalAmount.toLocaleString('ru-RU')}`
              : 'Создать заказ'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}