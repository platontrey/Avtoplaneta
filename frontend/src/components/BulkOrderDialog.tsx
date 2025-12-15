/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useState, useEffect } from "react";
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createOrder } from '@/features/orders/api/ordersApi';
import { sanitizeText, sanitizeNumber, sanitizeCustomerId, INPUT_LIMITS } from "@/hooks/usePartValidation";
import type { Part } from '@/features/parts/types';

interface BulkOrderDialogProps {
  isOpen: boolean;
  onClose: () => void;
  selectedParts: Part[];
  onSuccess: () => void;
}

interface OrderItem {
  part_id: number;
  quantity: number;
}

export default function BulkOrderDialog({ isOpen, onClose, selectedParts, onSuccess }: BulkOrderDialogProps) {
  const [customerId, setCustomerId] = useState('');
  const [buyerNumber, setBuyerNumber] = useState('');
  const [quantities, setQuantities] = useState<Record<number, number>>({});

  const queryClient = useQueryClient();

  // Инициализируем количества при открытии диалога
  useEffect(() => {
    if (isOpen && selectedParts.length > 0) {
      const initialQuantities: Record<number, number> = {};
      selectedParts.forEach(part => {
        initialQuantities[part.id] = 1;
      });
      setQuantities(initialQuantities);
    }
  }, [isOpen, selectedParts]);

  const createBulkOrderMutation = useMutation({
    mutationFn: async (orderData: { customer_id: number; buyer_number: string; items: OrderItem[] }) => {
      const items = selectedParts.map(part => ({
        part_id: part.id,
        quantity: quantities[part.id] || 1
      }));

      return createOrder({
        customer_id: orderData.customer_id,
        order_number: '',
        part: selectedParts.map(p => p.name).join(', '),
        buyer_number: orderData.buyer_number,
        items
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      queryClient.invalidateQueries({ queryKey: ['parts'] });
      onSuccess();
      handleClose();
    },
    onError: (error) => {
      console.error('Failed to create bulk order:', error);
      alert('Ошибка при создании заказа');
    },
  });

  const handleClose = () => {
    setCustomerId('');
    setBuyerNumber('');
    setQuantities({});
    onClose();
  };

  const handleQuantityChange = (partId: number, value: string) => {
    const sanitized = sanitizeNumber(value);
    setQuantities(prev => ({
      ...prev,
      [partId]: sanitized
    }));
  };

  const handleCustomerIdChange = (value: string) => {
    setCustomerId(sanitizeCustomerId(value));
  };

  const handleBuyerNumberChange = (value: string) => {
    setBuyerNumber(sanitizeText(value));
  };

  const handleSubmit = () => {
    if (!customerId || !buyerNumber) {
      alert('Заполните все обязательные поля');
      return;
    }

    const items = selectedParts.map(part => ({
      part_id: part.id,
      quantity: quantities[part.id] || 1
    }));

    createBulkOrderMutation.mutate({
      customer_id: parseInt(customerId),
      buyer_number: buyerNumber,
      items
    });
  };

  const isValid = customerId && buyerNumber && selectedParts.length > 0;

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-lg sm:text-xl">Массовый заказ запчастей</DialogTitle>
          <DialogDescription className="text-sm sm:text-base">
            Создание заказа для {selectedParts.length} выбранных запчастей
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Общие поля заказа */}
          <div className="grid gap-4">
            <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
              <Label htmlFor="customer_id" className="sm:text-right text-sm">
                ID клиента *
              </Label>
              <Input
                id="customer_id"
                type="number"
                value={customerId}
                onChange={(e) => handleCustomerIdChange(e.target.value)}
                className="sm:col-span-3"
                placeholder="Введите ID клиента"
                maxLength={INPUT_LIMITS.CUSTOMER_ID_MAX_LENGTH}
              />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">
              <Label htmlFor="buyer_number" className="sm:text-right text-sm">
                № покупателя *
              </Label>
              <Input
                id="buyer_number"
                value={buyerNumber}
                onChange={(e) => handleBuyerNumberChange(e.target.value)}
                className="sm:col-span-3"
                placeholder="Введите номер покупателя"
                maxLength={INPUT_LIMITS.TEXT_MAX_LENGTH}
              />
            </div>
          </div>

          {/* Список запчастей с количествами */}
          <div className="border-t pt-4">
            <Label className="text-sm font-medium mb-3 block">Запчасти для заказа:</Label>
            <div className="space-y-3 max-h-60 overflow-y-auto">
              {selectedParts.map((part) => (
                <div key={part.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-sm truncate">{part.name}</p>
                    <p className="text-xs text-gray-600">{part.category}</p>
                  </div>
                  <div className="flex items-center gap-2 ml-4">
                    <Label htmlFor={`quantity-${part.id}`} className="text-xs">
                      Кол-во:
                    </Label>
                    <Input
                      id={`quantity-${part.id}`}
                      type="number"
                      min={INPUT_LIMITS.QUANTITY_MIN}
                      max={INPUT_LIMITS.QUANTITY_MAX}
                      value={quantities[part.id] || 1}
                      onChange={(e) => handleQuantityChange(part.id, e.target.value)}
                      className="w-20 h-8 text-xs"
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            Отмена
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!isValid || createBulkOrderMutation.isPending}
          >
            {createBulkOrderMutation.isPending ? 'Создание...' : 'Создать заказ'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}