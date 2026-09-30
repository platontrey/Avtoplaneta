/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, ShoppingCart, Zap, AlertTriangle } from "lucide-react";
import { useOrderDialog } from "@/hooks/useOrderDialog";
import type { OrderChoice } from "@/hooks/useOrderDialog";
import { sanitizeText, sanitizeNumber, INPUT_LIMITS } from "@/hooks/usePartValidation";
import type { Part } from "@/features/parts/types";
import type { Order } from "@/lib/types";

interface PartOrderDialogProps {
  part: Part;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

function PartOrderDialog({ part, isOpen, onOpenChange }: PartOrderDialogProps) {
  const catalogPrice = Number(part.price) || 0;
  const {
    orderChoice,
    selectedOrderId,
    orderForm,
    existingOrders,
    createOrderMutation,
    addToExistingOrderMutation,
    selectChoice,
    updateForm,
    submitNewOrder,
    submitQuickSale,
    submitAddToExisting,
    setSelectedOrderId,
  } = useOrderDialog(part.name || '', catalogPrice);

  // Проверяем, есть ли эта запчасть уже в активном заказе
  const activeOrdersWithPart = (existingOrders || []).filter((o) =>
    o.part_id === part.id || (o.items && o.items.some((it) => it.part_id === part.id))
  );

  const handleFormUpdate = (field: string, value: string) => {
    let sanitizedValue: string | number = value;

    switch (field) {
      case 'buyer_number':
      case 'order_number':
      case 'transport_company':
      case 'notes':
        sanitizedValue = sanitizeText(value);
        break;
      case 'quantity':
        sanitizedValue = Math.max(1, sanitizeNumber(value));
        break;
      case 'price':
        sanitizedValue = Math.max(0, Number(value) || 0);
        break;
    }

    updateForm(field as keyof typeof orderForm, sanitizedValue);
  };

  const totalAmount = (Number(orderForm.price) || 0) * (Number(orderForm.quantity) || 1);

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[560px] max-h-[92vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-lg sm:text-xl">Оформление сделки</DialogTitle>
          <DialogDescription className="text-sm flex flex-wrap items-center gap-2 pt-1">
            <span className="font-medium text-foreground">{part.name}</span>
            <Badge variant="outline">Склад: {part.location || '—'}</Badge>
            <Badge variant="secondary">В наличии: {part.quantity ?? 0} шт.</Badge>
            <Badge variant="default">Прайс: ₽{catalogPrice.toLocaleString('ru-RU')}</Badge>
          </DialogDescription>
        </DialogHeader>

        {activeOrdersWithPart.length > 0 && (
          <div className="flex items-start gap-2 rounded-lg border border-amber-500/40 bg-amber-500/10 p-3 text-xs sm:text-sm text-amber-700 dark:text-amber-300">
            <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5" />
            <div>
              <strong>Внимание:</strong> эта деталь уже числится в активном заказе{' '}
              {activeOrdersWithPart.map((o) => `#${o.id} (${o.buyer_number}, продавец ${o.seller})`).join(', ')}.
            </div>
          </div>
        )}

        {/* Переключатель режимов */}
        <div className="grid grid-cols-3 gap-2 pt-1">
          <Button
            type="button"
            variant={orderChoice === 'new' ? 'default' : 'outline'}
            size="sm"
            onClick={() => selectChoice('new' as OrderChoice)}
            className="gap-1.5 text-xs sm:text-sm"
          >
            <Plus className="h-4 w-4" />
            <span>В заказ</span>
          </Button>
          <Button
            type="button"
            variant={orderChoice === 'quick' ? 'default' : 'outline'}
            size="sm"
            onClick={() => selectChoice('quick' as OrderChoice)}
            className="gap-1.5 text-xs sm:text-sm"
          >
            <Zap className="h-4 w-4" />
            <span>Быстрая продажа</span>
          </Button>
          <Button
            type="button"
            variant={orderChoice === 'existing' ? 'default' : 'outline'}
            size="sm"
            onClick={() => selectChoice('existing' as OrderChoice)}
            className="gap-1.5 text-xs sm:text-sm"
          >
            <ShoppingCart className="h-4 w-4" />
            <span>К заказу</span>
          </Button>
        </div>

        {orderChoice === 'quick' ? (
          <div className="space-y-4 py-2">
            <div className="rounded-lg bg-muted/50 border p-3 text-xs text-muted-foreground">
              ⚡ <strong>Продажа с места (Самовывоз):</strong> деталь сразу спишется со склада (или удалится при исчерпании остатка), а сделка сразу попадёт в историю завершённых продаж и статистику.
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="quick_qty">Количество (шт.)</Label>
                <Input
                  id="quick_qty"
                  type="number"
                  min={1}
                  max={INPUT_LIMITS.QUANTITY_MAX}
                  value={orderForm.quantity}
                  onChange={(e) => handleFormUpdate('quantity', e.target.value)}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="quick_price">Цена продажи за шт. (₽)</Label>
                <Input
                  id="quick_price"
                  type="number"
                  min={0}
                  value={orderForm.price}
                  onChange={(e) => handleFormUpdate('price', e.target.value)}
                />
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="quick_buyer">Покупатель / Контакт (необязательно)</Label>
                <Input
                  id="quick_buyer"
                  value={orderForm.buyer_number}
                  onChange={(e) => handleFormUpdate('buyer_number', e.target.value)}
                  placeholder="Самовывоз / телефон"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="quick_notes">Комментарий</Label>
                <Input
                  id="quick_notes"
                  value={orderForm.notes}
                  onChange={(e) => handleFormUpdate('notes', e.target.value)}
                  placeholder="Например: скидка на месте"
                />
              </div>
            </div>

            <div className="flex items-center justify-between rounded-lg bg-primary/10 px-4 py-2.5 text-sm font-semibold">
              <span>Итого к зачислению в продажи:</span>
              <span className="text-base">₽{totalAmount.toLocaleString('ru-RU')}</span>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Отмена
              </Button>
              <Button
                onClick={() => submitQuickSale(part.id, () => onOpenChange(false))}
                disabled={createOrderMutation.isPending}
                className="gap-1.5"
              >
                <Zap className="h-4 w-4" />
                {createOrderMutation.isPending ? 'Проведение...' : `Продать за ₽${totalAmount.toLocaleString('ru-RU')}`}
              </Button>
            </DialogFooter>
          </div>
        ) : orderChoice === 'existing' ? (
          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <Label>Выберите активный заказ *</Label>
              <Select
                value={selectedOrderId?.toString() || ""}
                onValueChange={(value) => setSelectedOrderId(parseInt(value))}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Выберите заказ из списка" />
                </SelectTrigger>
                <SelectContent>
                  {existingOrders && existingOrders.length > 0 ? (
                    existingOrders.map((order: Order) => (
                      <SelectItem key={order.id} value={order.id.toString()}>
                        #{order.id} · {order.buyer_number} {order.order_number ? `(№ ${order.order_number})` : ''} — {order.part}
                      </SelectItem>
                    ))
                  ) : (
                    <SelectItem value="no-orders" disabled>
                      Нет активных заказов
                    </SelectItem>
                  )}
                </SelectContent>
              </Select>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="add_quantity">Количество (шт.)</Label>
                <Input
                  id="add_quantity"
                  type="number"
                  min={1}
                  max={INPUT_LIMITS.QUANTITY_MAX}
                  value={orderForm.quantity}
                  onChange={(e) => handleFormUpdate('quantity', e.target.value)}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="add_price">Цена за шт. (₽)</Label>
                <Input
                  id="add_price"
                  type="number"
                  min={0}
                  value={orderForm.price}
                  onChange={(e) => handleFormUpdate('price', e.target.value)}
                />
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Отмена
              </Button>
              <Button
                onClick={() => submitAddToExisting(part.id, () => onOpenChange(false))}
                disabled={!selectedOrderId || addToExistingOrderMutation.isPending}
              >
                {addToExistingOrderMutation.isPending ? 'Добавление...' : 'Добавить в заказ'}
              </Button>
            </DialogFooter>
          </div>
        ) : (
          <div className="space-y-3.5 py-2">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="buyer_number">Покупатель (телефон / имя) *</Label>
                <Input
                  id="buyer_number"
                  value={orderForm.buyer_number}
                  onChange={(e) => handleFormUpdate('buyer_number', e.target.value)}
                  placeholder="+7 999 ... или имя клиента"
                  maxLength={INPUT_LIMITS.TEXT_MAX_LENGTH}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="order_number">№ сделки (Дром / Авито)</Label>
                <Input
                  id="order_number"
                  value={orderForm.order_number}
                  onChange={(e) => handleFormUpdate('order_number', e.target.value)}
                  placeholder="Например: 14829104"
                />
              </div>
            </div>

            <div className="grid grid-cols-3 gap-3">
              <div className="space-y-1.5">
                <Label>Источник</Label>
                <Select value={orderForm.source} onValueChange={(v) => updateForm('source', v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="drom">Дром</SelectItem>
                    <SelectItem value="avito">Авито</SelectItem>
                    <SelectItem value="messenger">Мессенджер / Звонок</SelectItem>
                    <SelectItem value="pickup">Самовывоз</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="quantity">Кол-во (шт.)</Label>
                <Input
                  id="quantity"
                  type="number"
                  min={1}
                  max={INPUT_LIMITS.QUANTITY_MAX}
                  value={orderForm.quantity}
                  onChange={(e) => handleFormUpdate('quantity', e.target.value)}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="price">Цена за шт. (₽)</Label>
                <Input
                  id="price"
                  type="number"
                  min={0}
                  value={orderForm.price}
                  onChange={(e) => handleFormUpdate('price', e.target.value)}
                />
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="space-y-1.5">
                <Label>💰 Оплата</Label>
                <Select value={orderForm.payment_status} onValueChange={(v) => updateForm('payment_status', v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="unpaid">Не оплачен (Бронь)</SelectItem>
                    <SelectItem value="prepaid">Предоплата</SelectItem>
                    <SelectItem value="paid">Оплачен полностью</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-1.5">
                <Label>🏭 Задача складу</Label>
                <Select value={orderForm.warehouse_status} onValueChange={(v) => updateForm('warehouse_status', v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="inspecting">Проверка / Фото</SelectItem>
                    <SelectItem value="transfer">Перемещение</SelectItem>
                    <SelectItem value="ready">Готов к выдаче</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-1.5">
                <Label>🚚 Получение</Label>
                <Select value={orderForm.delivery_method} onValueChange={(v) => updateForm('delivery_method', v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="tk">Отправка ТК</SelectItem>
                    <SelectItem value="pickup">Самовывоз</SelectItem>
                    <SelectItem value="city">Доставка по городу</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {orderForm.delivery_method === 'tk' && (
              <div className="space-y-1.5">
                <Label htmlFor="transport_company">Транспортная компания / Город</Label>
                <Input
                  id="transport_company"
                  value={orderForm.transport_company}
                  onChange={(e) => handleFormUpdate('transport_company', e.target.value)}
                  placeholder="Например: СДЭК, Энергия, ПЭК (г. Новосибирск)"
                />
              </div>
            )}

            <div className="space-y-1.5">
              <Label htmlFor="notes">Комментарий / Заметка для склада</Label>
              <Input
                id="notes"
                value={orderForm.notes}
                onChange={(e) => handleFormUpdate('notes', e.target.value)}
                placeholder="Например: сфоткать крепления, упаковать в картон..."
              />
            </div>

            <div className="flex items-center justify-between rounded-lg bg-muted px-4 py-2 text-sm">
              <span className="text-muted-foreground">Сумма заказа:</span>
              <span className="font-semibold">₽{totalAmount.toLocaleString('ru-RU')}</span>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Отмена
              </Button>
              <Button
                onClick={() => submitNewOrder(part.id, () => onOpenChange(false))}
                disabled={!orderForm.buyer_number.trim() || createOrderMutation.isPending}
              >
                {createOrderMutation.isPending ? 'Создание...' : 'Создать заказ'}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

export default PartOrderDialog;