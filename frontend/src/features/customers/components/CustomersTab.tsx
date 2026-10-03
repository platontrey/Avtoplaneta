import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  useCustomers,
  useCustomer,
  useCreateCustomer,
  useUpdateCustomer,
  useDeleteCustomer,
} from '@/hooks/useCustomers';
import type { Customer, CustomerWithStats, CustomerCategory } from '@/lib/types';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
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
  ExternalLink,
  AlertTriangle,
  ShoppingBag,
  Truck,
  MapPin,
  Phone,
  FileText,
} from 'lucide-react';

export const CATEGORY_LABELS: Record<CustomerCategory, { label: string; badgeClass: string; icon: string }> = {
  regular: {
    label: 'Обычный',
    badgeClass: 'bg-zinc-500/15 text-zinc-700 dark:text-zinc-300 border-zinc-500/30',
    icon: '👤',
  },
  vip: {
    label: 'VIP / СТО',
    badgeClass: 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30',
    icon: '⭐',
  },
  wholesale: {
    label: 'Оптовик',
    badgeClass: 'bg-blue-500/15 text-blue-700 dark:text-blue-300 border-blue-500/30',
    icon: '🏢',
  },
  blacklist: {
    label: 'Чёрный список',
    badgeClass: 'bg-rose-500/15 text-rose-700 dark:text-rose-300 border-rose-500/30 font-semibold',
    icon: '⛔',
  },
};

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

interface CustomerDetailsModalProps {
  customerId: number;
  onClose: () => void;
  onEdit: (c: CustomerWithStats) => void;
  onCopyTK: (c: Customer) => void;
  copiedId: number | null;
}

const CustomerDetailsModal: React.FC<CustomerDetailsModalProps> = ({
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

interface CustomerFormModalProps {
  isOpen: boolean;
  customer?: CustomerWithStats | null;
  onClose: () => void;
}

const CustomerFormModal: React.FC<CustomerFormModalProps> = ({ isOpen, customer, onClose }) => {
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
