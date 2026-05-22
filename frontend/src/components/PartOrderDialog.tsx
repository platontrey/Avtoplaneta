/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, ShoppingCart } from "lucide-react";
import { useOrderDialog } from "@/hooks/useOrderDialog";
import type { OrderChoice } from "@/hooks/useOrderDialog";
import { sanitizeText, sanitizeNumber, sanitizeCustomerId, INPUT_LIMITS } from "@/hooks/usePartValidation";
import type { Part } from "@/features/parts/types";
import type { Order } from "@/lib/types";

interface PartOrderDialogProps {
  part: Part;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

function PartOrderDialog({ part, isOpen, onOpenChange }: PartOrderDialogProps) {
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
    submitAddToExisting,
    setSelectedOrderId,
  } = useOrderDialog(part.name || '');

  const handleOpenChange = (open: boolean) => {
    onOpenChange(open);
  };

  const handleChoiceSelect = (choice: OrderChoice) => {
    selectChoice(choice);
  };

  const handleFormUpdate = (field: string, value: string) => {
    let sanitizedValue: string | number = value;

    switch (field) {
      case 'customer_id':
        sanitizedValue = sanitizeCustomerId(value);
        break;
      case 'buyer_number':
        sanitizedValue = sanitizeText(value);
        break;
      case 'quantity':
        sanitizedValue = sanitizeNumber(value);
        break;
    }

    updateForm(field as keyof typeof orderForm, sanitizedValue);
  };

  const handleSubmitNewOrder = () => {
    submitNewOrder(part.id);
  };

  const handleSubmitAddToExisting = () => {
    submitAddToExisting(part.id);
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[500px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-lg sm:text-xl">Добавить к заказу</DialogTitle>
          <DialogDescription className="text-sm sm:text-base">
            Выберите действие для детали: {part.name}
          </DialogDescription>
        </DialogHeader>

        {!orderChoice ? (
          <div className="grid gap-3 py-4">
            <Button
              onClick={() => handleChoiceSelect('new')}
              className="h-16 flex flex-col items-center justify-center gap-2"
              variant="outline"
            >
              <Plus className="h-6 w-6" />
              <span>Создать новый заказ</span>
            </Button>
            <Button
              onClick={() => handleChoiceSelect('existing')}
              className="h-16 flex flex-col items-center justify-center gap-2"
              variant="outline"
            >
              <ShoppingCart className="h-6 w-6" />
              <span>Добавить к существующему</span>
            </Button>
          </div>
        ) : orderChoice === 'new' ? (
          <>
            <div className="grid gap-3 sm:gap-4 py-4">
              <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                <Label htmlFor="customer_id" className="sm:text-right text-sm">
                  ID клиента
                </Label>
                <Input
                  id="customer_id"
                  type="number"
                  value={orderForm.customer_id}
                  onChange={(e) => handleFormUpdate('customer_id', e.target.value)}
                  className="sm:col-span-3"
                  placeholder="Введите ID клиента"
                  maxLength={INPUT_LIMITS.CUSTOMER_ID_MAX_LENGTH}
                />
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                <Label htmlFor="buyer_number" className="sm:text-right text-sm">
                  № покупателя
                </Label>
                <Input
                  id="buyer_number"
                  value={orderForm.buyer_number}
                  onChange={(e) => handleFormUpdate('buyer_number', e.target.value)}
                  className="sm:col-span-3"
                  placeholder="Введите номер покупателя"
                  maxLength={INPUT_LIMITS.TEXT_MAX_LENGTH}
                />
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                <Label htmlFor="quantity" className="sm:text-right text-sm">
                  Количество
                </Label>
                <Input
                  id="quantity"
                  type="number"
                  min={INPUT_LIMITS.QUANTITY_MIN}
                  max={INPUT_LIMITS.QUANTITY_MAX}
                  value={orderForm.quantity}
                  onChange={(e) => handleFormUpdate('quantity', e.target.value)}
                  className="sm:col-span-3"
                />
              </div>
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => handleChoiceSelect(null)}
                className="mr-2"
              >
                Назад
              </Button>
              <Button
                onClick={handleSubmitNewOrder}
                disabled={!orderForm.customer_id || !orderForm.buyer_number || createOrderMutation.isPending}
              >
                {createOrderMutation.isPending ? 'Создание...' : 'Создать заказ'}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <div className="grid gap-3 py-4">
              <div className="space-y-2">
                <Label>Выберите существующий заказ</Label>
                <Select value={selectedOrderId?.toString() || ""} onValueChange={(value) => setSelectedOrderId(parseInt(value))}>
                  <SelectTrigger>
                    <SelectValue placeholder="Выберите заказ" />
                  </SelectTrigger>
                  <SelectContent>
                    {existingOrders && existingOrders.length > 0 ? (
                      existingOrders.map((order: Order) => (
                        <SelectItem key={order.id} value={order.id.toString()}>
                          {order.order_number} - {order.part} ({order.buyer_number})
                        </SelectItem>
                      ))
                    ) : (
                      <SelectItem value="no-orders" disabled>
                        Нет доступных заказов
                      </SelectItem>
                    )}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
                <Label htmlFor="add_quantity" className="sm:text-right text-sm">
                  Количество
                </Label>
                <Input
                  id="add_quantity"
                  type="number"
                  min={INPUT_LIMITS.QUANTITY_MIN}
                  max={INPUT_LIMITS.QUANTITY_MAX}
                  value={orderForm.quantity}
                  onChange={(e) => handleFormUpdate('quantity', e.target.value)}
                  className="sm:col-span-3"
                />
              </div>
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => handleChoiceSelect(null)}
                className="mr-2"
              >
                Назад
              </Button>
              <Button
                onClick={handleSubmitAddToExisting}
                disabled={!selectedOrderId || addToExistingOrderMutation.isPending}
              >
                {addToExistingOrderMutation.isPending ? 'Добавление...' : 'Добавить к заказу'}
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

export default PartOrderDialog;