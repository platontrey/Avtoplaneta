/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Link } from 'react-router-dom';
import { useCustomer } from '@/hooks/useCustomers';
import type { Customer, CustomerWithStats } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { AlertTriangle, ShoppingBag, Copy, Check, Edit, ExternalLink } from 'lucide-react';
import { CATEGORY_LABELS } from '../constants';

export interface CustomerDetailsModalProps {
  customerId: number;
  onClose: () => void;
  onEdit: (c: CustomerWithStats) => void;
  onCopyTK: (c: Customer) => void;
  copiedId: number | null;
}

export const CustomerDetailsModal: React.FC<CustomerDetailsModalProps> = ({
  customerId,
  onClose,
  onEdit,
  onCopyTK,
  copiedId,
}) => {
  const { data: customer, isLoading, error } = useCustomer(customerId);

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
        {isLoading ? (
          <div className="space-y-4 py-4">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        ) : error || !customer ? (
          <div className="py-6 text-center space-y-2">
            <p className="text-destructive font-semibold">Не удалось загрузить данные клиента</p>
            <Button variant="outline" size="sm" onClick={onClose}>
              Закрыть
            </Button>
          </div>
        ) : (
          <>
            <DialogHeader>
              <div className="flex items-center justify-between gap-2 pr-6">
                <div>
                  <DialogTitle className="text-xl font-bold flex items-center gap-2">
                    <span>{customer.name || customer.phone}</span>
                    <Badge
                      variant="outline"
                      className={`text-xs ${
                        (CATEGORY_LABELS[customer.category] || CATEGORY_LABELS.regular).badgeClass
                      }`}
                    >
                      {(CATEGORY_LABELS[customer.category] || CATEGORY_LABELS.regular).icon}{' '}
                      {(CATEGORY_LABELS[customer.category] || CATEGORY_LABELS.regular).label}
                    </Badge>
                  </DialogTitle>
                  <DialogDescription className="text-xs text-muted-foreground mt-0.5">
                    ID клиента: #{customer.id} · Телефон: {customer.phone}
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>

            {/* Предупреждение о ЧС */}
            {customer.category === 'blacklist' && (
              <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg text-sm text-rose-700 dark:text-rose-300 flex items-start gap-2.5">
                <AlertTriangle className="h-5 w-5 shrink-0 text-rose-600 mt-0.5" />
                <div>
                  <div className="font-semibold">Клиент в чёрном списке!</div>
                  <div className="text-xs mt-0.5">
                    {customer.notes || 'Причина не указана. Будьте бдительны при оформлении.'}
                  </div>
                </div>
              </div>
            )}

            {/* Реквизиты и данные */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 p-3.5 bg-muted/40 rounded-lg text-xs">
              <div>
                <span className="text-muted-foreground">Город: </span>
                <span className="font-medium text-foreground">{customer.city || 'Не указан'}</span>
              </div>
              <div>
                <span className="text-muted-foreground">ТК / ПВЗ: </span>
                <span className="font-medium text-foreground">{customer.preferred_tk || 'Не указана'}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Паспорт / ИНН: </span>
                <span className="font-medium text-foreground">{customer.passport_or_inn || 'Не указаны'}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Персональная скидка: </span>
                <span className="font-semibold text-emerald-600 dark:text-emerald-400">
                  {customer.discount_percent}%
                </span>
              </div>
              <div>
                <span className="text-muted-foreground">Всего заказов: </span>
                <span className="font-semibold text-foreground">{customer.total_orders}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Общая сумма выкупа: </span>
                <span className="font-semibold text-foreground">
                  ₽{Number(customer.total_spent || 0).toLocaleString('ru-RU')}
                </span>
              </div>
            </div>

            {customer.notes && customer.category !== 'blacklist' && (
              <div className="text-xs p-3 bg-card border rounded-lg">
                <span className="font-semibold text-muted-foreground block mb-0.5">Примечание:</span>
                <span className="text-foreground">{customer.notes}</span>
              </div>
            )}

            {/* Кнопка копирования для ТК */}
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                className="gap-2 flex-1 text-xs"
                onClick={() => onCopyTK(customer)}
              >
                {copiedId === customer.id ? (
                  <>
                    <Check className="h-4 w-4 text-emerald-600" />
                    <span className="text-emerald-600 font-semibold">Реквизиты скопированы в буфер</span>
                  </>
                ) : (
                  <>
                    <Copy className="h-4 w-4 text-muted-foreground" />
                    <span>Скопировать реквизиты для ТК</span>
                  </>
                )}
              </Button>
              <Button variant="outline" size="sm" onClick={() => onEdit(customer)} className="gap-1.5 text-xs">
                <Edit className="h-3.5 w-3.5" />
                <span>Редактировать</span>
              </Button>
            </div>

            {/* История заказов клиента */}
            <div className="space-y-2 pt-2">
              <h3 className="font-bold text-sm flex items-center gap-1.5">
                <ShoppingBag className="h-4 w-4 text-primary" />
                <span>История заказов ({customer.orders?.length || 0})</span>
              </h3>

              {!customer.orders || customer.orders.length === 0 ? (
                <div className="py-6 text-center text-xs text-muted-foreground border rounded-lg">
                  У этого клиента пока нет сохранённых заказов
                </div>
              ) : (
                <div className="rounded-md border overflow-x-auto max-h-60 overflow-y-auto">
                  <Table>
                    <TableHeader>
                      <TableRow className="text-xs">
                        <TableHead>Заказ</TableHead>
                        <TableHead>Запчасть</TableHead>
                        <TableHead>Сумма</TableHead>
                        <TableHead>Статус</TableHead>
                        <TableHead>Дата</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {customer.orders.map((o) => {
                        const total = o.total_amount || 0;
                        return (
                          <TableRow key={o.id} className="text-xs">
                            <TableCell className="font-medium">#{o.id}</TableCell>
                            <TableCell className="max-w-[200px] truncate" title={o.part}>
                              {o.items && o.items.length > 0 ? (
                                <div className="space-y-0.5">
                                  {o.items.map((it) => (
                                    <div key={it.id || it.part_id}>
                                      {it.part_id > 0 ? (
                                        <Link
                                          to={`/inventory?partId=${it.part_id}`}
                                          className="text-primary hover:underline inline-flex items-center gap-1"
                                        >
                                          <span>{it.quantity}× {it.part_name || it.part_name_snapshot || `Запчасть #${it.part_id}`}</span>
                                          <ExternalLink className="h-3 w-3 opacity-60" />
                                        </Link>
                                      ) : (
                                        <span>{it.quantity}× {it.part_name || it.part_name_snapshot}</span>
                                      )}
                                    </div>
                                  ))}
                                </div>
                              ) : (
                                <span>{o.part}</span>
                              )}
                            </TableCell>
                            <TableCell className="font-semibold whitespace-nowrap">
                              ₽{Number(total).toLocaleString('ru-RU')}
                            </TableCell>
                            <TableCell>
                              <Badge
                                variant={o.auto_deleted || o.status === 'green' ? 'default' : 'secondary'}
                                className="text-[10px] py-0 px-1.5"
                              >
                                {o.auto_deleted || o.status === 'green' ? 'Завершён' : o.status}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-muted-foreground whitespace-nowrap text-[11px]">
                              {o.created_at_formatted || o.created_at?.slice(0, 10)}
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button onClick={onClose} variant="outline" size="sm">
                Закрыть
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
};
