/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState } from 'react';
import { useCustomers, useDeleteCustomer } from '@/hooks/useCustomers';
import type { CustomerWithStats } from '@/lib/types';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
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
import {
  Search,
  Plus,
  Users,
  Copy,
  Check,
  Edit,
  Trash2,
  ShoppingBag,
  Truck,
  MapPin,
  Phone,
  FileText,
} from 'lucide-react';
import { CATEGORY_LABELS } from '../constants';
import { CustomerDetailsModal } from './CustomerDetailsModal';
import { CustomerFormModal } from './CustomerFormModal';

export { CATEGORY_LABELS };

export const CustomersTab: React.FC = () => {
  const [search, setSearch] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [detailsCustomerId, setDetailsCustomerId] = useState<number | null>(null);
  const [editingCustomer, setEditingCustomer] = useState<CustomerWithStats | null>(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [customerToDelete, setCustomerToDelete] = useState<CustomerWithStats | null>(null);
  const [copiedId, setCopiedId] = useState<number | null>(null);

  const { data, isLoading, error, refetch } = useCustomers({
    category: selectedCategory === 'all' ? '' : selectedCategory,
    q: search.trim(),
  });

  const deleteMutation = useDeleteCustomer();
  const customers = data?.customers || [];

  const handleCopyTK = (c: {
    id: number;
    name: string;
    phone: string;
    city?: string;
    preferred_tk?: string;
    passport_or_inn?: string;
  }) => {
    const lines = [
      `Получатель: ${c.name}`,
      `Телефон: ${c.phone}`,
      c.city ? `Город: ${c.city}` : '',
      c.preferred_tk ? `ТК / Терминал: ${c.preferred_tk}` : '',
      c.passport_or_inn ? `Данные: ${c.passport_or_inn}` : '',
    ].filter(Boolean);

    navigator.clipboard.writeText(lines.join('\n'));
    setCopiedId(c.id);
    setTimeout(() => setCopiedId(null), 2500);
  };

  const handleDeleteConfirm = async () => {
    if (!customerToDelete) return;
    try {
      await deleteMutation.mutateAsync(customerToDelete.id);
      setCustomerToDelete(null);
    } catch (e) {
      console.error('Failed to delete customer', e);
    }
  };

  return (
    <div className="space-y-4">
      {/* Поиск, фильтры и кнопка создания */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Поиск по ФИО, телефону, городу, ТК..."
            className="pl-9"
          />
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <div className="inline-flex rounded-lg border bg-muted p-1 text-xs">
            <Button
              variant={selectedCategory === 'all' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setSelectedCategory('all')}
              className="h-7 text-xs"
            >
              Все ({customers.length})
            </Button>
            <Button
              variant={selectedCategory === 'vip' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setSelectedCategory('vip')}
              className="h-7 text-xs"
            >
              ⭐ VIP / СТО
            </Button>
            <Button
              variant={selectedCategory === 'wholesale' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setSelectedCategory('wholesale')}
              className="h-7 text-xs"
            >
              🏢 Оптовики
            </Button>
            <Button
              variant={selectedCategory === 'blacklist' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setSelectedCategory('blacklist')}
              className="h-7 text-xs text-rose-600 dark:text-rose-400"
            >
              ⛔ ЧС
            </Button>
          </div>

          <Button onClick={() => setIsCreateOpen(true)} size="sm" className="gap-1.5 h-9">
            <Plus className="h-4 w-4" />
            <span>Новый клиент</span>
          </Button>
        </div>
      </div>

      {/* Ошибка или загрузка */}
      {isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-14 w-full rounded-lg" />
          <Skeleton className="h-48 w-full rounded-lg" />
        </div>
      ) : error ? (
        <div className="p-4 bg-destructive/10 text-destructive border border-destructive/20 rounded-lg flex items-center justify-between">
          <span>Не удалось загрузить клиентов: {error instanceof Error ? error.message : 'Ошибка'}</span>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Повторить
          </Button>
        </div>
      ) : customers.length === 0 ? (
        <Card className="border-dashed">
          <CardContent className="py-12 flex flex-col items-center justify-center text-center space-y-3">
            <Users className="h-10 w-10 text-muted-foreground opacity-40" />
            <div className="space-y-1">
              <p className="font-semibold text-base">Клиенты не найдены</p>
              <p className="text-xs text-muted-foreground max-w-sm">
                {search || selectedCategory !== 'all'
                  ? 'По заданным фильтрам клиентов не найдено. Попробуйте изменить параметры поиска.'
                  : 'Клиенты создаются автоматически при оформлении заказов, либо добавьте первого вручную.'}
              </p>
            </div>
            <Button onClick={() => setIsCreateOpen(true)} variant="outline" size="sm" className="gap-1.5">
              <Plus className="h-4 w-4" />
              Добавить клиента
            </Button>
          </CardContent>
        </Card>
      ) : (
        <>
          {/* Десктопная таблица */}
          <div className="hidden md:block rounded-md border overflow-x-auto bg-card">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Клиент / Телефон</TableHead>
                  <TableHead>Категория</TableHead>
                  <TableHead>Город и ТК</TableHead>
                  <TableHead>Скидка</TableHead>
                  <TableHead>Покупки / LTV</TableHead>
                  <TableHead className="text-right">Действия</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {customers.map((c) => {
                  const cat = CATEGORY_LABELS[c.category] || CATEGORY_LABELS.regular;
                  const isCopied = copiedId === c.id;

                  return (
                    <TableRow key={c.id} className="hover:bg-muted/40 transition-colors">
                      <TableCell className="font-medium">
                        <div
                          onClick={() => setDetailsCustomerId(c.id)}
                          className="cursor-pointer hover:underline text-foreground flex items-center gap-1.5"
                        >
                          <span className="font-semibold">{c.name || c.phone}</span>
                        </div>
                        <div className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                          <Phone className="h-3 w-3" />
                          <span>{c.phone}</span>
                        </div>
                      </TableCell>

                      <TableCell>
                        <Badge variant="outline" className={`text-xs gap-1 py-0.5 px-2 ${cat.badgeClass}`}>
                          <span>{cat.icon}</span>
                          <span>{cat.label}</span>
                        </Badge>
                        {c.category === 'blacklist' && c.notes && (
                          <div className="text-[11px] text-rose-600 dark:text-rose-400 mt-1 line-clamp-1 max-w-[180px]">
                            ⚠️ {c.notes}
                          </div>
                        )}
                      </TableCell>

                      <TableCell>
                        <div className="text-sm">
                          {c.city ? (
                            <span className="inline-flex items-center gap-1 font-medium">
                              <MapPin className="h-3.5 w-3.5 text-muted-foreground" />
                              {c.city}
                            </span>
                          ) : (
                            <span className="text-muted-foreground text-xs">—</span>
                          )}
                        </div>
                        {c.preferred_tk && (
                          <div className="text-xs text-muted-foreground inline-flex items-center gap-1 mt-0.5">
                            <Truck className="h-3 w-3" />
                            <span>{c.preferred_tk}</span>
                          </div>
                        )}
                      </TableCell>

                      <TableCell>
                        {c.discount_percent > 0 ? (
                          <Badge variant="secondary" className="font-semibold text-emerald-600 dark:text-emerald-400">
                            {c.discount_percent}%
                          </Badge>
                        ) : (
                          <span className="text-xs text-muted-foreground">0%</span>
                        )}
                      </TableCell>

                      <TableCell>
                        <div className="text-sm font-semibold">
                          ₽{Number(c.total_spent || 0).toLocaleString('ru-RU')}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          Заказов: {c.total_orders || 0}
                        </div>
                      </TableCell>

                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          <Button
                            variant="outline"
                            size="sm"
                            title="Скопировать реквизиты для ТК"
                            onClick={() => handleCopyTK(c)}
                            className="h-8 px-2 gap-1 text-xs"
                          >
                            {isCopied ? (
                              <>
                                <Check className="h-3.5 w-3.5 text-emerald-600" />
                                <span className="text-emerald-600">Скопировано</span>
                              </>
                            ) : (
                              <>
                                <Copy className="h-3.5 w-3.5 text-muted-foreground" />
                                <span>ТК</span>
                              </>
                            )}
                          </Button>

                          <Button
                            variant="outline"
                            size="sm"
                            title="Карточка и история заказов"
                            onClick={() => setDetailsCustomerId(c.id)}
                            className="h-8 px-2 text-xs"
                          >
                            <FileText className="h-3.5 w-3.5" />
                          </Button>

                          <Button
                            variant="ghost"
                            size="sm"
                            title="Редактировать"
                            onClick={() => setEditingCustomer(c)}
                            className="h-8 w-8 p-0"
                          >
                            <Edit className="h-3.5 w-3.5" />
                          </Button>

                          <Button
                            variant="ghost"
                            size="sm"
                            title="Удалить"
                            onClick={() => setCustomerToDelete(c)}
                            className="h-8 w-8 p-0 text-destructive hover:bg-destructive/10"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>

          {/* Мобильный вид (карточки) */}
          <div className="grid grid-cols-1 gap-3 md:hidden">
            {customers.map((c) => {
              const cat = CATEGORY_LABELS[c.category] || CATEGORY_LABELS.regular;
              const isCopied = copiedId === c.id;

              return (
                <Card key={c.id} className="p-4 space-y-3">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <div
                        onClick={() => setDetailsCustomerId(c.id)}
                        className="font-semibold text-base hover:underline cursor-pointer"
                      >
                        {c.name || c.phone}
                      </div>
                      <div className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                        <Phone className="h-3 w-3" />
                        <span>{c.phone}</span>
                      </div>
                    </div>
                    <Badge variant="outline" className={`text-xs gap-1 py-0.5 px-2 ${cat.badgeClass}`}>
                      <span>{cat.icon}</span>
                      <span>{cat.label}</span>
                    </Badge>
                  </div>

                  {c.category === 'blacklist' && c.notes && (
                    <div className="p-2 rounded bg-rose-500/10 border border-rose-500/20 text-xs text-rose-700 dark:text-rose-300">
                      ⚠️ <strong>Чёрный список:</strong> {c.notes}
                    </div>
                  )}

                  <div className="grid grid-cols-2 gap-2 text-xs pt-1 border-t">
                    <div>
                      <span className="text-muted-foreground">Город / ТК: </span>
                      <span className="font-medium text-foreground">
                        {c.city || '—'} {c.preferred_tk ? `(${c.preferred_tk})` : ''}
                      </span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Выкуп (LTV): </span>
                      <span className="font-semibold text-foreground">
                        ₽{Number(c.total_spent || 0).toLocaleString('ru-RU')} ({c.total_orders || 0} зак.)
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 pt-1">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleCopyTK(c)}
                      className="flex-1 text-xs h-8 gap-1.5"
                    >
                      {isCopied ? <Check className="h-3.5 w-3.5 text-emerald-600" /> : <Copy className="h-3.5 w-3.5" />}
                      <span>{isCopied ? 'Скопировано' : 'Реквизиты ТК'}</span>
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setDetailsCustomerId(c.id)}
                      className="flex-1 text-xs h-8 gap-1.5"
                    >
                      <ShoppingBag className="h-3.5 w-3.5" />
                      <span>Заказы</span>
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setEditingCustomer(c)}
                      className="h-8 w-8 p-0"
                    >
                      <Edit className="h-3.5 w-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setCustomerToDelete(c)}
                      className="h-8 w-8 p-0 text-destructive"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </Card>
              );
            })}
          </div>
        </>
      )}

      {/* Модальное окно: Детали клиента и история заказов */}
      {detailsCustomerId && (
        <CustomerDetailsModal
          customerId={detailsCustomerId}
          onClose={() => setDetailsCustomerId(null)}
          onEdit={(c) => {
            setDetailsCustomerId(null);
            setEditingCustomer(c);
          }}
          onCopyTK={handleCopyTK}
          copiedId={copiedId}
        />
      )}

      {/* Модальное окно: Создание / Редактирование клиента */}
      {(isCreateOpen || editingCustomer) && (
        <CustomerFormModal
          isOpen={isCreateOpen || Boolean(editingCustomer)}
          customer={editingCustomer}
          onClose={() => {
            setIsCreateOpen(false);
            setEditingCustomer(null);
          }}
        />
      )}

      {/* Подтверждение удаления */}
      <AlertDialog open={Boolean(customerToDelete)} onOpenChange={(open) => !open && setCustomerToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Удаление карточки клиента</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите удалить клиента <strong>{customerToDelete?.name || customerToDelete?.phone}</strong>?
              История ранее оформленных заказов останется в базе, но связь с клиентом будет удалена.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Удалить
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};
