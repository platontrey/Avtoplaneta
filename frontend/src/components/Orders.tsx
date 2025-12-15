import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useOrders } from '../hooks/useOrders';
import { updateOrderStatus, deleteOrder, completeOrder } from '../features/orders/api/ordersApi';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { Order } from '../lib/types';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Badge } from './ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';
import { Skeleton } from './ui/skeleton';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from './ui/alert-dialog';
import { Plus } from 'lucide-react';
// import { motion, AnimatePresence } from 'framer-motion'; // Закомментировано, так как не используется

const Orders = () => {
    const navigate = useNavigate();
    const { data: orders, isLoading: ordersLoading, error: ordersError } = useOrders();
    const queryClient = useQueryClient();
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [orderToDelete, setOrderToDelete] = useState<Order | null>(null);
    const [completeDialogOpen, setCompleteDialogOpen] = useState(false);
    const [orderToComplete, setOrderToComplete] = useState<Order | null>(null);


    const updateStatusMutation = useMutation({
        mutationFn: ({ orderId, status }: { orderId: number; status: string }) =>
            updateOrderStatus(orderId, status),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['orders']});
        },
    });

    const deleteOrderMutation = useMutation({
        mutationFn: deleteOrder,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['orders'] });
            setDeleteDialogOpen(false);
            setOrderToDelete(null);
        },
    });

    const completeOrderMutation = useMutation({
        mutationFn: completeOrder,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['orders'] });
            queryClient.invalidateQueries({ queryKey: ['parts'] });
            queryClient.invalidateQueries({ queryKey: ['inventory'] });
            setCompleteDialogOpen(false);
            setOrderToComplete(null);
        },
    });


    if (ordersLoading) {
        return (
            <div className="container mx-auto p-4">
                <div className="space-y-6">
                    {/* Скелетон заголовка */}
                    <div className="space-y-2">
                        <Skeleton className="h-8 w-48" />
                        <Skeleton className="h-4 w-64" />
                    </div>

                    {/* Скелетон карточки создания заказа */}
                    <div className="bg-white border border-gray-200 rounded-lg shadow-sm">
                        <div className="p-6 border-b border-gray-200">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <Skeleton className="h-5 w-5" />
                                    <Skeleton className="h-6 w-48" />
                                </div>
                                <Skeleton className="h-8 w-20" />
                            </div>
                        </div>
                    </div>

                    {/* Скелетон таблицы заказов */}
                    <div className="bg-white border border-gray-200 rounded-lg shadow-sm">
                        <div className="p-6 border-b border-gray-200">
                            <Skeleton className="h-6 w-56" />
                        </div>
                        <div className="p-6">
                            <div className="rounded-md border">
                                {/* Скелетон заголовка таблицы */}
                                <div className="border-b border-gray-200">
                                    <div className="grid grid-cols-7 gap-4 p-4">
                                        {Array.from({ length: 7 }).map((_, index) => (
                                            <Skeleton key={index} className="h-4 w-full" />
                                        ))}
                                    </div>
                                </div>
                                {/* Скелетон строк таблицы */}
                        {Array.from({ length: 5 }).map((_, rowIndex) => (
                            <div key={rowIndex} className="border-b border-gray-200 last:border-b-0">
                                <div className="grid grid-cols-8 gap-4 p-4">
                                    <Skeleton className="h-4 w-16" />
                                    <Skeleton className="h-4 w-20" />
                                    <Skeleton className="h-4 w-24" />
                                    <Skeleton className="h-4 w-18" />
                                    <Skeleton className="h-4 w-20" />
                                    <Skeleton className="h-6 w-32" />
                                    <Skeleton className="h-8 w-24" />
                                    <Skeleton className="h-8 w-16" />
                                </div>
                            </div>
                        ))}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    if (ordersError) {
        return <div className="container mx-auto p-4">Ошибка загрузки заказов</div>;
    }


    return (
        <div className="container mx-auto p-4">
            <h1 className="text-2xl font-bold mb-4">Заказы</h1>

            <Card className="mb-6">
                <CardHeader
                    className="cursor-pointer hover:bg-accent/50 transition-colors"
                    onClick={() => navigate('/inventory')}
                >
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <Plus className="h-5 w-5" />
                            <CardTitle>Создать новый заказ</CardTitle>
                        </div>
                        <Button variant="ghost" size="sm" className="gap-2">
                            <span className="text-sm">Перейти в инвентарь</span>
                        </Button>
                    </div>
                </CardHeader>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Таблица статусов заказов</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>№ заказа</TableHead>
                                    <TableHead>Продавец</TableHead>
                                    <TableHead>Запчасть</TableHead>
                                    <TableHead>Местоположение</TableHead>
                                    <TableHead>№ покупателя</TableHead>
                                    <TableHead>Статус</TableHead>
                                    <TableHead>Создано</TableHead>
                                    <TableHead>Действия</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {!orders || orders.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                                            Заказов пока нет. Создайте первый заказ выше!
                                        </TableCell>
                                    </TableRow>
                                ) : orders.map((order: Order) => (
                                    <TableRow key={order.id}>
                                        <TableCell className="font-medium">{order.id}</TableCell>
                                        <TableCell>{order.seller}</TableCell>
                                        <TableCell>
                                            <div className="space-y-1">
                                                {order.items && order.items.length > 0 ? (
                                                    order.items.map((item) => (
                                                        <div key={item.id} className="text-sm">
                                                            {item.quantity}x {order.part}
                                                        </div>
                                                    ))
                                                ) : (
                                                    <div className="text-sm">
                                                        {order.part || 'Нет деталей'}
                                                    </div>
                                                )}
                                            </div>
                                        </TableCell>
                                        <TableCell>{order.location || 'Неизвестно'}</TableCell>
                                        <TableCell>{order.buyer_number}</TableCell>
                                        <TableCell>
                                            <div className="flex items-center space-x-2">
                                                <Badge variant={
                                                    order.status === 'red' ? 'destructive' :
                                                        order.status === 'green' ? 'default' :
                                                            'secondary'
                                                }>
                                                    {order.status_text}
                                                </Badge>
                                                <Select
                                                    value={order.status}
                                                    onValueChange={(value) => {
                                                        updateStatusMutation.mutate({
                                                            orderId: order.id,
                                                            status: value
                                                        });
                                                    }}
                                                    disabled={updateStatusMutation.isPending}
                                                >
                                                    <SelectTrigger className="w-48">
                                                        <SelectValue />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        <SelectItem value="red">🔴 Нужен транспорт</SelectItem>
                                                        <SelectItem value="brown">🟤 Ожидание ответа</SelectItem>
                                                        <SelectItem value="yellow">🟡 Нужна доставка</SelectItem>
                                                        <SelectItem value="green">🟢 Доставлено</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="text-sm">
                                                <div>{order.created_at_formatted}</div>
                                                <div className="text-gray-500">{order.time_ago}</div>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex gap-2">
                                                <AlertDialog open={completeDialogOpen && orderToComplete?.id === order.id} onOpenChange={setCompleteDialogOpen}>
                                                    <AlertDialogTrigger asChild>
                                                        <Button
                                                            variant="default"
                                                            size="sm"
                                                            onClick={() => {
                                                                setOrderToComplete(order);
                                                                setCompleteDialogOpen(true);
                                                            }}
                                                            disabled={completeOrderMutation.isPending}
                                                        >
                                                            {completeOrderMutation.isPending ? 'Завершение...' : 'Завершить'}
                                                        </Button>
                                                    </AlertDialogTrigger>
                                                    <AlertDialogContent>
                                                        <AlertDialogHeader>
                                                            <AlertDialogTitle>Подтверждение завершения продажи</AlertDialogTitle>
                                                            <AlertDialogDescription>
                                                                Вы уверены, что хотите завершить продажу по заказу {orderToComplete?.id}? Заказ будет отмечен как проданный и учтён в статистике продаж.
                                                            </AlertDialogDescription>
                                                        </AlertDialogHeader>
                                                        <AlertDialogFooter>
                                                            <AlertDialogCancel onClick={() => setCompleteDialogOpen(false)}>Отмена</AlertDialogCancel>
                                                            <AlertDialogAction
                                                                onClick={() => {
                                                                    if (orderToComplete) {
                                                                        completeOrderMutation.mutate(orderToComplete.id);
                                                                    }
                                                                }}
                                                            >
                                                                Завершить продажу
                                                            </AlertDialogAction>
                                                        </AlertDialogFooter>
                                                    </AlertDialogContent>
                                                </AlertDialog>
                                                <AlertDialog open={deleteDialogOpen && orderToDelete?.id === order.id} onOpenChange={setDeleteDialogOpen}>
                                                    <AlertDialogTrigger asChild>
                                                        <Button
                                                            variant="destructive"
                                                            size="sm"
                                                            onClick={() => {
                                                                setOrderToDelete(order);
                                                                setDeleteDialogOpen(true);
                                                            }}
                                                            disabled={deleteOrderMutation.isPending}
                                                        >
                                                            {deleteOrderMutation.isPending ? 'Удаление...' : 'Удалить'}
                                                        </Button>
                                                    </AlertDialogTrigger>
                                                    <AlertDialogContent>
                                                        <AlertDialogHeader>
                                                            <AlertDialogTitle>Подтверждение удаления</AlertDialogTitle>
                                                            <AlertDialogDescription>
                                                                Вы уверены, что хотите удалить заказ {orderToDelete?.id}? Количество запчастей будет восстановлено.
                                                            </AlertDialogDescription>
                                                        </AlertDialogHeader>
                                                        <AlertDialogFooter>
                                                            <AlertDialogCancel onClick={() => setDeleteDialogOpen(false)}>Отмена</AlertDialogCancel>
                                                            <AlertDialogAction
                                                                onClick={() => {
                                                                    if (orderToDelete) {
                                                                        deleteOrderMutation.mutate(orderToDelete.id);
                                                                    }
                                                                }}
                                                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                                            >
                                                                Удалить
                                                            </AlertDialogAction>
                                                        </AlertDialogFooter>
                                                    </AlertDialogContent>
                                                </AlertDialog>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};

export default Orders;