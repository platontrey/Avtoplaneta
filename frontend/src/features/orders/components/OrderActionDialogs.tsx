/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import type { Order } from '@/lib/types';
import { calcOrderTotal } from '../utils/orderHelpers';

interface CompleteOrderDialogProps {
  order: Order | null;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (order: Order) => void;
}

export const CompleteOrderDialog: React.FC<CompleteOrderDialogProps> = ({
  order,
  isOpen,
  onClose,
  onConfirm,
}) => {
  return (
    <AlertDialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Завершить сделку по заказу #{order?.id}?</AlertDialogTitle>
          <AlertDialogDescription>
            Заказанное количество будет списано со склада (если остаток запчасти исчерпан — карточка запчасти и её фото будут удалены из инвентаря), а сам заказ перейдёт во вкладку «История продаж» и учтётся в статистике на сумму{' '}
            <strong>₽{order ? calcOrderTotal(order).toLocaleString('ru-RU') : 0}</strong>.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Отмена</AlertDialogCancel>
          <AlertDialogAction
            onClick={() => {
              if (order) onConfirm(order);
            }}
          >
            Завершить продажу
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};

interface DeleteOrderDialogProps {
  order: Order | null;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (order: Order) => void;
}

export const DeleteOrderDialog: React.FC<DeleteOrderDialogProps> = ({
  order,
  isOpen,
  onClose,
  onConfirm,
}) => {
  return (
    <AlertDialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Отменить и удалить заказ #{order?.id}?</AlertDialogTitle>
          <AlertDialogDescription>
            Заказ #{order?.id} ({order?.buyer_number}) будет удалён из активных заказов, а запчасти останутся на складе.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Назад</AlertDialogCancel>
          <AlertDialogAction
            onClick={() => {
              if (order) onConfirm(order);
            }}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            Удалить заказ
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};
