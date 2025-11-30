/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

import React, { useState, useEffect } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Trash2, UserPlus, Users, Server, FileText, Edit, Activity, Package, Bell } from "lucide-react";
import { getAuthHeaders } from "@/lib/csrf";
import { useAuth } from "@/hooks/useAuth";
import { getUserActivityLogs } from "@/features/admin/api/adminApi";
import type { UserActivityLog, UserActivityAction, UserActivityResourceType } from "@/lib/types";
import PushNotifications from "./PushNotifications";

interface User {
  id: number;
  email: string;
  name: string;
  initials?: string;
  inn?: string;
  provider: string;
  role: string;
}

interface ServerStatus {
  server: {
    status: string;
    uptime: string;
    go_version: string;
    os: string;
    arch: string;
  };
  database: {
    status: string;
    total_parts: number;
    total_users: number;
  };
  timestamp: string;
}

interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
}

interface ActivityFilters {
  user_id?: number;
  action?: UserActivityAction;
  resource_type?: UserActivityResourceType;
  start_date?: string;
  end_date?: string;
}

export default function AdminPanel() {
  const { user } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showAddUser, setShowAddUser] = useState(false);
  const [newUser, setNewUser] = useState({
    email: '',
    name: '',
    initials: '',
    inn: '',
    password: '',
    role: 'operator',
  });
  const [serverStatus, setServerStatus] = useState<ServerStatus | null>(null);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [statusLoading, setStatusLoading] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [editForm, setEditForm] = useState({
    name: '',
    initials: '',
    inn: '',
    role: '',
  });
  const [activityLogs, setActivityLogs] = useState<UserActivityLog[]>([]);
  const [activityLoading, setActivityLoading] = useState(false);
  const [activityFilters, setActivityFilters] = useState<ActivityFilters>({});
  const [selectedUser, setSelectedUser] = useState<string>('');
  const [selectedAction, setSelectedAction] = useState<string>('');
  const [selectedResourceType, setSelectedResourceType] = useState<string>('');
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');
  const [supplierCodes, setSupplierCodes] = useState<string[]>([]);
  const [selectedSupplierCode, setSelectedSupplierCode] = useState<string>('');

  const fetchUsers = async () => {
    try {
      const response = await fetch('http://localhost:8083/admin/users', {
        credentials: 'include',
      });

      if (!response.ok) {
        setError('Failed to fetch users');
        return;
      }

      const data = await response.json();
      setUsers(data.users || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
    } finally {
      setLoading(false);
    }
  };

  const fetchSupplierCodes = async () => {
    try {
      const response = await fetch('http://localhost:8081/api/admin/supplier-codes', {
        credentials: 'include',
      });

      if (!response.ok) {
        console.error('Failed to fetch supplier codes');
        return;
      }

      const data = await response.json();
      setSupplierCodes(data.supplier_codes || []);
    } catch (err) {
      console.error('Failed to fetch supplier codes:', err);
    }
  };

  useEffect(() => {
    fetchUsers();
    fetchServerStatus();
    fetchServerLogs();
    fetchActivityLogs();
    fetchSupplierCodes();
  }, []);

  const fetchServerStatus = async () => {
    try {
      setStatusLoading(true);
      const response = await fetch('http://localhost:8083/admin/status', {
        credentials: 'include',
      });

      if (!response.ok) {
        console.error('Failed to fetch server status');
        return;
      }

      const data = await response.json();
      setServerStatus(data);
    } catch (err) {
      console.error('Failed to fetch server status:', err);
    } finally {
      setStatusLoading(false);
    }
  };

  const fetchServerLogs = async () => {
    try {
      setLogsLoading(true);
      const response = await fetch('http://localhost:8083/admin/logs', {
        credentials: 'include',
      });

      if (!response.ok) {
        console.error('Failed to fetch server logs');
        return;
      }

      const data = await response.json();
      setLogs(data.logs || []);
    } catch (err) {
      console.error('Failed to fetch server logs:', err);
    } finally {
      setLogsLoading(false);
    }
  };

  const fetchActivityLogs = async () => {
    try {
      setActivityLoading(true);
      console.log('Fetching activity logs with filters:', activityFilters);
      const logs = await getUserActivityLogs(activityFilters);
      console.log('Fetched activity logs:', logs);
      setActivityLogs(logs);
    } catch (err) {
      console.error('Failed to fetch activity logs:', err);
      setActivityLogs([]); // Set empty array on error
    } finally {
      setActivityLoading(false);
    }
  };

  const applyActivityFilters = () => {
    const filters: ActivityFilters = {};
    if (selectedUser && selectedUser !== 'all') filters.user_id = parseInt(selectedUser);
    if (selectedAction && selectedAction !== 'all') filters.action = selectedAction as UserActivityAction;
    if (selectedResourceType && selectedResourceType !== 'all') filters.resource_type = selectedResourceType as UserActivityResourceType;
    if (startDate) filters.start_date = startDate;
    if (endDate) filters.end_date = endDate;

    setActivityFilters(filters);
    fetchActivityLogs();
  };

  const clearActivityFilters = () => {
    setSelectedUser('');
    setSelectedAction('');
    setSelectedResourceType('');
    setStartDate('');
    setEndDate('');
    setActivityFilters({});
    fetchActivityLogs();
  };

  const handleAddUser = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await fetch('http://localhost:8083/admin/users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
        },
        credentials: 'include',
        body: JSON.stringify(newUser),
      });

      if (!response.ok) {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to add user');
        return;
      }

      setShowAddUser(false);
      setNewUser({ email: '', name: '', initials: '', inn: '', password: '', role: 'operator' });
      fetchUsers(); // Refresh the list
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add user');
    } finally {
      setLoading(false);
    }
  };

  const handleEditUser = (user: User) => {
    setEditingUser(user);
    setEditForm({
      name: user.name,
      initials: user.initials || '',
      inn: user.inn || '',
      role: user.role,
    });
  };

  const handleUpdateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingUser) return;

    try {
      const response = await fetch(`http://localhost:8083/admin/users/${editingUser.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
        },
        credentials: 'include',
        body: JSON.stringify(editForm),
      });

      if (!response.ok) {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to update user');
        return;
      }

      setEditingUser(null);
      fetchUsers(); // Refresh the list
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update user');
    }
  };

  const handleDeleteUser = async (userId: number) => {
    if (!window.confirm('Вы уверены, что хотите удалить этого пользователя?')) {
      return;
    }

    try {
      const response = await fetch(`http://localhost:8083/admin/users/${userId}`, {
        method: 'DELETE',
        headers: {
          'X-CSRF-Token': getAuthHeaders()['X-CSRF-Token'] || '',
        },
        credentials: 'include',
      });

      if (!response.ok) {
        setError('Failed to delete user');
        return;
      }

      fetchUsers(); // Refresh the list
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete user');
    }
  };

  const handleDeleteZeroQuantityParts = async () => {
    if (!selectedSupplierCode) {
      setError('Пожалуйста, выберите код поставки');
      return;
    }

    if (!window.confirm(`Вы уверены, что хотите удалить все запчасти с количеством 0 для кода поставки "${selectedSupplierCode}"? Это действие нельзя отменить.`)) {
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`http://localhost:8081/api/admin/delete-zero-quantity-parts/${encodeURIComponent(selectedSupplierCode)}`, {
        method: 'DELETE',
        headers: getAuthHeaders(),
        credentials: 'include',
      });

      if (!response.ok) {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to delete zero quantity parts');
        return;
      }

      const result = await response.json();
      alert(`Удалено ${result.deleted_count} запчастей для кода поставки "${selectedSupplierCode}"`);
      setSelectedSupplierCode(''); // Reset selection
      fetchServerStatus(); // Refresh status to show updated counts
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete zero quantity parts');
    } finally {
      setLoading(false);
    }
  };

  if (loading && users.length === 0) {
    return (
      <div className="max-w-7xl mx-auto px-6 py-4">
        <div className="space-y-6">
          {/* Header skeleton */}
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <Skeleton className="h-8 w-64" />
              <Skeleton className="h-4 w-80" />
            </div>
            <Skeleton className="h-10 w-48" />
          </div>

          {/* Status cards skeleton */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-white border border-gray-200 rounded-lg shadow-sm">
              <div className="p-6 border-b border-gray-200">
                <div className="flex items-center space-x-2">
                  <Skeleton className="h-5 w-5" />
                  <Skeleton className="h-6 w-32" />
                </div>
                <Skeleton className="h-4 w-48 mt-2" />
              </div>
              <div className="p-6 space-y-4">
                <div className="space-y-2">
                  <Skeleton className="h-4 w-16" />
                  <div className="space-y-1">
                    <Skeleton className="h-3 w-24" />
                    <Skeleton className="h-3 w-20" />
                    <Skeleton className="h-3 w-18" />
                    <Skeleton className="h-3 w-22" />
                  </div>
                </div>
                <div className="space-y-2">
                  <Skeleton className="h-4 w-24" />
                  <div className="space-y-1">
                    <Skeleton className="h-3 w-20" />
                    <Skeleton className="h-3 w-16" />
                    <Skeleton className="h-3 w-20" />
                  </div>
                </div>
                <Skeleton className="h-3 w-32" />
              </div>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg shadow-sm">
              <div className="p-6 border-b border-gray-200">
                <div className="flex items-center space-x-2">
                  <Skeleton className="h-5 w-5" />
                  <Skeleton className="h-6 w-24" />
                </div>
                <Skeleton className="h-4 w-56 mt-2" />
              </div>
              <div className="p-6">
                <div className="space-y-3">
                  {Array.from({ length: 4 }).map((_, index) => (
                    <div key={index} className="flex items-start space-x-2">
                      <Skeleton className="h-5 w-12" />
                      <div className="flex-1 space-y-1">
                        <Skeleton className="h-3 w-full" />
                        <Skeleton className="h-3 w-24" />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Users list skeleton */}
          <div className="bg-white border border-gray-200 rounded-lg shadow-sm">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center space-x-2">
                <Skeleton className="h-5 w-5" />
                <Skeleton className="h-6 w-24" />
              </div>
              <Skeleton className="h-4 w-64 mt-2" />
            </div>
            <div className="p-6 space-y-4">
              {Array.from({ length: 5 }).map((_, index) => (
                <div key={index} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center space-x-4">
                    <div className="space-y-2">
                      <Skeleton className="h-4 w-32" />
                      <Skeleton className="h-3 w-40" />
                      <Skeleton className="h-3 w-24" />
                      <Skeleton className="h-3 w-20" />
                    </div>
                    <div className="flex space-x-2">
                      <Skeleton className="h-5 w-16" />
                      <Skeleton className="h-5 w-12" />
                    </div>
                  </div>
                  <div className="flex space-x-2">
                    <Skeleton className="h-8 w-8" />
                    <Skeleton className="h-8 w-8" />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4 sm:mb-6">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold">Панель администратора</h1>
          <p className="text-gray-600 text-sm sm:text-base">Управление пользователями и настройками системы</p>
        </div>
        {user?.role === 'admin' && (
          <Dialog open={showAddUser} onOpenChange={setShowAddUser}>
            <DialogTrigger asChild>
              <Button>
                <UserPlus className="w-4 h-4 mr-2" />
                Добавить пользователя
              </Button>
            </DialogTrigger>
            <DialogContent>
            <DialogHeader>
              <DialogTitle>Добавить нового пользователя</DialogTitle>
              <DialogDescription>
                Создать новую учетную запись пользователя с email и паролем.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleAddUser} className="space-y-4">
              <div>
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={newUser.email}
                  onChange={(e) => setNewUser({...newUser, email: e.target.value})}
                  required
                />
              </div>
              <div>
                <Label htmlFor="name">Имя</Label>
                <Input
                  id="name"
                  value={newUser.name}
                  onChange={(e) => setNewUser({...newUser, name: e.target.value})}
                  required
                />
              </div>
              <div>
                <Label htmlFor="initials">Инициалы</Label>
                <Input
                  id="initials"
                  value={newUser.initials}
                  onChange={(e) => setNewUser({...newUser, initials: e.target.value})}
                />
              </div>
              <div>
                <Label htmlFor="inn">ИНН</Label>
                <Input
                  id="inn"
                  value={newUser.inn}
                  onChange={(e) => setNewUser({...newUser, inn: e.target.value})}
                />
              </div>
              <div>
                <Label htmlFor="password">Пароль</Label>
                <Input
                  id="password"
                  type="password"
                  value={newUser.password}
                  onChange={(e) => setNewUser({...newUser, password: e.target.value})}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="role">Роль</Label>
                <Select
                  value={newUser.role}
                  onValueChange={(value) => setNewUser({...newUser, role: value})}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Выберите роль" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="operator">Operator</SelectItem>
                    <SelectItem value="manager">Manager</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <DialogFooter>
                <Button type="submit" disabled={loading}>
                  {loading ? 'Создание...' : 'Создать пользователя'}
                </Button>
              </DialogFooter>
            </form>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-md">
          <p className="text-red-600">{error}</p>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        {/* Server Status Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center">
              <Server className="w-5 h-5 mr-2" />
              Server Status
            </CardTitle>
            <CardDescription>
              Текущая информация о сервере и базе данных
            </CardDescription>
          </CardHeader>
          <CardContent>
            {statusLoading ? (
              <p className="text-center py-4">Загрузка статуса сервера...</p>
            ) : serverStatus ? (
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium mb-2">Сервер</h4>
                  <div className="space-y-1 text-sm">
                    <p><span className="font-medium">Status:</span> <span className="text-green-600">{serverStatus.server.status}</span></p>
                    <p><span className="font-medium">Go Version:</span> {serverStatus.server.go_version}</p>
                    <p><span className="font-medium">OS:</span> {serverStatus.server.os}</p>
                    <p><span className="font-medium">Architecture:</span> {serverStatus.server.arch}</p>
                  </div>
                </div>
                <div>
                  <h4 className="font-medium mb-2">База данных</h4>
                  <div className="space-y-1 text-sm">
                    <p><span className="font-medium">Status:</span> <span className="text-green-600">{serverStatus.database.status}</span></p>
                    <p><span className="font-medium">Total Parts:</span> {serverStatus.database.total_parts}</p>
                    <p><span className="font-medium">Total Users:</span> {serverStatus.database.total_users}</p>
                  </div>
                </div>
                <div className="text-xs text-gray-500">
                  Last updated: {new Date(serverStatus.timestamp).toLocaleString()}
                </div>
              </div>
            ) : (
              <p className="text-center py-4 text-red-500">Не удалось загрузить статус сервера</p>
            )}
          </CardContent>
        </Card>

        {/* Server Logs Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center">
              <FileText className="w-5 h-5 mr-2" />
              Server Logs
            </CardTitle>
            <CardDescription>
              Недавняя активность сервера и события
            </CardDescription>
          </CardHeader>
          <CardContent>
            {logsLoading ? (
              <p className="text-center py-4">Загрузка логов...</p>
            ) : logs.length > 0 ? (
              <div className="space-y-2 max-h-64 overflow-y-auto">
                {logs.map((log, index) => (
                  <div key={index} className="flex items-start space-x-2 text-sm">
                    <span className={`px-2 py-1 text-xs rounded ${
                      log.level === 'ERROR' ? 'bg-red-100 text-red-800' :
                      log.level === 'WARN' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-blue-100 text-blue-800'
                    }`}>
                      {log.level}
                    </span>
                    <div className="flex-1">
                      <p>{log.message}</p>
                      <p className="text-xs text-gray-500">{new Date(log.timestamp).toLocaleString()}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-center py-4 text-gray-500">Логи недоступны</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* User Activity Logs Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Activity className="w-5 h-5 mr-2" />
            User Activity Logs ({activityLogs.length})
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

            <div className="flex gap-2">
              <Button onClick={applyActivityFilters} disabled={activityLoading}>
                {activityLoading ? 'Применение...' : 'Применить фильтры'}
              </Button>
              <Button variant="outline" onClick={clearActivityFilters}>
                Очистить
              </Button>
            </div>
          </div>

          {/* Activity Logs */}
          {activityLoading ? (
            <p className="text-center py-4">Загрузка логов активности...</p>
          ) : activityLogs.length > 0 ? (
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {activityLogs.map((log) => (
                <div key={log.id} className="flex items-start space-x-3 p-3 border rounded-lg bg-white">
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-1">
                      <span className={`px-2 py-1 text-xs rounded-full ${
                        log.action === 'login' ? 'bg-green-100 text-green-800' :
                        log.action === 'logout' ? 'bg-gray-100 text-gray-800' :
                        log.action.includes('create') ? 'bg-blue-100 text-blue-800' :
                        log.action.includes('update') ? 'bg-yellow-100 text-yellow-800' :
                        log.action.includes('delete') ? 'bg-red-100 text-red-800' :
                        'bg-purple-100 text-purple-800'
                      }`}>
                        {log.action.replace('_', ' ').toUpperCase()}
                      </span>
                      <span className="text-sm font-medium">{log.user_name}</span>
                      <span className="text-xs text-gray-500">({log.user_email})</span>
                    </div>
                    <p className="text-sm text-gray-700 mb-1">{log.details}</p>
                    <div className="flex items-center space-x-4 text-xs text-gray-500">
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

      {/* Edit User Dialog */}
      <Dialog open={editingUser !== null} onOpenChange={() => setEditingUser(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Редактировать пользователя</DialogTitle>
            <DialogDescription>
              Изменить данные пользователя {editingUser?.name}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleUpdateUser} className="space-y-4">
            <div>
              <Label htmlFor="edit-name">Имя</Label>
              <Input
                id="edit-name"
                value={editForm.name}
                onChange={(e) => setEditForm({...editForm, name: e.target.value})}
                required
              />
            </div>
            <div>
              <Label htmlFor="edit-initials">Инициалы</Label>
              <Input
                id="edit-initials"
                value={editForm.initials}
                onChange={(e) => setEditForm({...editForm, initials: e.target.value})}
              />
            </div>
            <div>
              <Label htmlFor="edit-inn">ИНН</Label>
              <Input
                id="edit-inn"
                value={editForm.inn}
                onChange={(e) => setEditForm({...editForm, inn: e.target.value})}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-role">Роль</Label>
              <Select
                value={editForm.role}
                onValueChange={(value) => setEditForm({...editForm, role: value})}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Выберите роль" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="operator">Operator</SelectItem>
                  <SelectItem value="manager">Manager</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditingUser(null)}>
                Отмена
              </Button>
              <Button type="submit" disabled={loading}>
                {loading ? 'Сохранение...' : 'Сохранить'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Card className="mt-8">
        <CardHeader>
          <CardTitle className="flex items-center">
            <Users className="w-5 h-5 mr-2" />
            Users ({users.length})
          </CardTitle>
          <CardDescription>
            Управление учетными записями пользователей и разрешениями
          </CardDescription>
        </CardHeader>
        <CardContent>
          {users.length === 0 ? (
            <p className="text-center py-8 text-gray-500">Пользователи не найдены</p>
          ) : (
            <div className="space-y-4">
              {users.map((user) => (
                <div key={user.id} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center space-x-4">
                    <div>
                        <p className="font-medium">{user.name}</p>
                        <p className="text-sm text-gray-500">{user.email}</p>
                        {user.initials && <p className="text-sm text-gray-500">Инициалы: {user.initials}</p>}
                        {user.inn && <p className="text-sm text-gray-500">ИНН: {user.inn}</p>}
                      </div>
                    <div className="flex space-x-2">
                      <span className={`px-2 py-1 text-xs rounded-full ${
                        user.provider === 'google'
                          ? 'bg-blue-100 text-blue-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {user.provider}
                      </span>
                      <span className={`px-2 py-1 text-xs rounded-full ${
                        user.role === 'admin' ? 'bg-red-100 text-red-800' :
                        user.role === 'manager' ? 'bg-yellow-100 text-yellow-800' :
                        'bg-green-100 text-green-800'
                      }`}>
                        {user.role}
                      </span>
                    </div>
                  </div>
                  <div className="flex space-x-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleEditUser(user)}
                      className="text-blue-600 hover:text-blue-700 hover:bg-blue-50"
                    >
                      <Edit className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDeleteUser(user.id)}
                      className="text-red-600 hover:text-red-700 hover:bg-red-50"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Parts Management Card */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle className="flex items-center">
            <Package className="w-5 h-5 mr-2" />
            Управление запчастями
          </CardTitle>
          <CardDescription>
            Массовые операции с запчастями
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="p-4 border rounded-lg bg-yellow-50 border-yellow-200">
              <h4 className="font-medium text-yellow-800 mb-2">Удаление шаблонных запчастей</h4>
              <p className="text-sm text-yellow-700 mb-4">
                Эта операция удалит все запчасти с количеством 0 для выбранного кода поставки.
                Используйте, когда уверены, что все необходимые запчасти из дефектной ведомости заполнены.
              </p>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="supplier-code-select">Выберите код поставки</Label>
                  <Select value={selectedSupplierCode} onValueChange={setSelectedSupplierCode}>
                    <SelectTrigger id="supplier-code-select">
                      <SelectValue placeholder="Выберите код поставки" />
                    </SelectTrigger>
                    <SelectContent>
                      {supplierCodes.map((code) => (
                        <SelectItem key={code} value={code}>
                          {code}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <Button
                  variant="destructive"
                  onClick={handleDeleteZeroQuantityParts}
                  disabled={loading || !selectedSupplierCode}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  {loading ? 'Удаление...' : 'Удалить запчасти с quantity = 0'}
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Push Notifications Card */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle className="flex items-center">
            <Bell className="w-5 h-5 mr-2" />
            Push Notifications
          </CardTitle>
          <CardDescription>
            Управление push-уведомлениями для PWA
          </CardDescription>
        </CardHeader>
        <CardContent>
          <PushNotifications />
        </CardContent>
      </Card>
    </div>
  );
}

