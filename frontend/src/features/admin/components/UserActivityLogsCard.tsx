/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Activity } from 'lucide-react';
import type { UserActivityLog, UserActivityAction } from '@/lib/types';
import type { AdminUser, ActivityFilters, UserActivityResourceType } from '../types';

interface UserActivityLogsCardProps {
  logs: UserActivityLog[];
  loading: boolean;
  users: AdminUser[];
  onApplyFilters: (filters: ActivityFilters) => Promise<void>;
  onClearFilters: () => void;
}

export const UserActivityLogsCard: React.FC<UserActivityLogsCardProps> = ({
  logs,
  loading,
  users,
  onApplyFilters,
  onClearFilters,
}) => {
  const [selectedUser, setSelectedUser] = useState<string>('');
  const [selectedAction, setSelectedAction] = useState<string>('');
  const [selectedResourceType, setSelectedResourceType] = useState<string>('');
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');
  const [usefulOnly, setUsefulOnly] = useState<boolean>(true);

  const handleApply = async () => {
    const filters: ActivityFilters = {};
    if (selectedUser && selectedUser !== 'all') filters.user_id = parseInt(selectedUser);
    if (selectedAction && selectedAction !== 'all') filters.action = selectedAction as UserActivityAction;
    if (selectedResourceType && selectedResourceType !== 'all') filters.resource_type = selectedResourceType as UserActivityResourceType;
    if (startDate) filters.start_date = startDate;
    if (endDate) filters.end_date = endDate;
    if (usefulOnly) filters.useful_only = true;

    await onApplyFilters(filters);
  };

  const handleClear = () => {
    setSelectedUser('');
    setSelectedAction('');
    setSelectedResourceType('');
    setStartDate('');
    setEndDate('');
    setUsefulOnly(true);
    onClearFilters();
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Activity className="w-5 h-5 mr-2" />
          User Activity Logs ({logs.length})
        </CardTitle>
        <CardDescription>
          Подробное логирование действий пользователей в системе
        </CardDescription>
      </CardHeader>
      <CardContent>
        {/* Filters */}
        <div className="mb-4 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
            <div>
              <Label htmlFor="user-filter">Пользователь</Label>
              <Select value={selectedUser} onValueChange={setSelectedUser}>
                <SelectTrigger id="user-filter">
                  <SelectValue placeholder="Все пользователи" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Все пользователи</SelectItem>
                  {users.map((user) => (
                    <SelectItem key={user.id} value={user.id.toString()}>
                      {user.name} ({user.email})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="action-filter">Действие</Label>
              <Select value={selectedAction} onValueChange={setSelectedAction}>
                <SelectTrigger id="action-filter">
                  <SelectValue placeholder="Все действия" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Все действия</SelectItem>
                  <SelectItem value="login">Вход</SelectItem>
                  <SelectItem value="logout">Выход</SelectItem>
                  <SelectItem value="create_part">Создание запчасти</SelectItem>
                  <SelectItem value="update_part">Обновление запчасти</SelectItem>
                  <SelectItem value="delete_part">Удаление запчасти</SelectItem>
                  <SelectItem value="create_order">Создание заказа</SelectItem>
                  <SelectItem value="update_order">Обновление заказа</SelectItem>
                  <SelectItem value="delete_order">Удаление заказа</SelectItem>
                  <SelectItem value="create_user">Создание пользователя</SelectItem>
                  <SelectItem value="update_user">Обновление пользователя</SelectItem>
                  <SelectItem value="delete_user">Удаление пользователя</SelectItem>
                  <SelectItem value="upload_photo">Загрузка фото</SelectItem>
                  <SelectItem value="delete_photo">Удаление фото</SelectItem>
                  <SelectItem value="delete_zero_quantity_parts">Удаление запчастей с quantity=0</SelectItem>
                  <SelectItem value="mark_part_for_deletion">Отметка для удаления</SelectItem>
                  <SelectItem value="create_defect_report">Создание дефектной ведомости</SelectItem>
                  <SelectItem value="update_defect_report">Обновление дефектной ведомости</SelectItem>
                  <SelectItem value="delete_defect_report">Удаление дефектной ведомости</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="resource-filter">Тип ресурса</Label>
              <Select value={selectedResourceType} onValueChange={setSelectedResourceType}>
                <SelectTrigger id="resource-filter">
                  <SelectValue placeholder="Все типы" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Все типы</SelectItem>
                  <SelectItem value="part">Запчасти</SelectItem>
                  <SelectItem value="order">Заказы</SelectItem>
                  <SelectItem value="user">Пользователи</SelectItem>
                  <SelectItem value="photo">Фото</SelectItem>
                  <SelectItem value="system">Система</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="start-date">Дата от</Label>
              <Input
                id="start-date"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </div>

            <div>
              <Label htmlFor="end-date">Дата до</Label>
              <Input
                id="end-date"
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
              />
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button onClick={handleApply} disabled={loading}>
              {loading ? 'Применение...' : 'Применить фильтры'}
            </Button>
            <Button variant="outline" onClick={handleClear}>
              Очистить
            </Button>
            <label className="flex items-center gap-2 ml-auto text-sm text-gray-600 cursor-pointer select-none">
              <Checkbox
                checked={usefulOnly}
                onCheckedChange={(checked) => setUsefulOnly(checked === true)}
              />
              Только полезные данные
            </label>
          </div>
        </div>

        {/* Activity Logs */}
        {loading ? (
          <p className="text-center py-4">Загрузка логов активности...</p>
        ) : logs.length > 0 ? (
          <div className="space-y-3 max-h-96 overflow-y-auto">
            {logs.map((log) => (
              <div key={log.id} className="flex items-start space-x-3 p-3 border rounded-lg bg-card">
                <div className="flex-1">
                  <div className="flex items-center space-x-2 mb-1">
                    <span
                      className={`px-2 py-1 text-xs rounded-full ${
                        log.action === 'login'
                          ? 'bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-300'
                          : log.action === 'logout'
                          ? 'bg-muted text-muted-foreground'
                          : log.action.includes('create')
                          ? 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300'
                          : log.action.includes('update')
                          ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-300'
                          : log.action.includes('delete')
                          ? 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300'
                          : 'bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-300'
                      }`}
                    >
                      {log.action.replace('_', ' ').toUpperCase()}
                    </span>
                    <span className="text-sm font-medium">{log.user_name}</span>
                    <span className="text-xs text-muted-foreground">({log.user_email})</span>
                  </div>
                  <p className="text-sm text-foreground/90 mb-1">{log.details}</p>
                  <div className="flex items-center space-x-4 text-xs text-muted-foreground">
                    <span>Тип: {log.resource_type}</span>
                    {log.resource_id && <span>ID: {log.resource_id}</span>}
                    <span>{new Date(log.created_at).toLocaleString()}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-center py-8 text-gray-500">Логи активности недоступны</p>
        )}
      </CardContent>
    </Card>
  );
};
