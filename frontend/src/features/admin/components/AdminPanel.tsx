/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React, { useState, useEffect, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { UserPlus } from 'lucide-react';
import { getAuthHeaders } from '@/lib/csrf';
import { useAuth } from '@/hooks/useAuth';
import { getUserActivityLogs, logUserActivity } from '@/features/admin/api/adminApi';
import type { UserActivityLog } from '@/lib/types';
import PushNotifications from '@/components/PushNotifications';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Bell } from 'lucide-react';
import { API_BASE_URL } from '@/lib/api';

import type { AdminUser, ServerStatus, ServerLogEntry, ActivityFilters, SupplierBatch } from '../types';
import { AdminSkeleton } from './AdminSkeleton';
import { ServerStatusCard } from './ServerStatusCard';
import { ServerLogsCard } from './ServerLogsCard';
import { UserActivityLogsCard } from './UserActivityLogsCard';
import { UsersManagementCard } from './UsersManagementCard';
import { AddUserDialog, EditUserDialog } from './UserDialogs';
import { ZeroQuantityPartsCard } from './ZeroQuantityPartsCard';
import { ImageEditorTestCard } from './ImageEditorTestCard';

export const AdminPanel: React.FC = () => {
  const { user } = useAuth();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showAddUser, setShowAddUser] = useState(false);
  const [editingUser, setEditingUser] = useState<AdminUser | null>(null);

  const [serverStatus, setServerStatus] = useState<ServerStatus | null>(null);
  const [logs, setLogs] = useState<ServerLogEntry[]>([]);
  const [statusLoading, setStatusLoading] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);

  const [activityLogs, setActivityLogs] = useState<UserActivityLog[]>([]);
  const [activityLoading, setActivityLoading] = useState(false);

  const [supplierBatches, setSupplierBatches] = useState<SupplierBatch[]>([]);
  const [selectedSupplierCode, setSelectedSupplierCode] = useState<string>('');

  const fetchUsers = useCallback(async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users`, {
        credentials: 'include',
      });

      if (!response.ok) {
        setError('Failed to fetch users');
        return;
      }

      const data = await response.json();
      setUsers(data.users || []);

      await logUserActivity({
        action: 'view_users',
        resource_type: 'user',
        resource_id: undefined,
        details: 'Просмотр списка пользователей в админ панели',
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchSupplierCodes = useCallback(async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/supplier-codes`, {
        credentials: 'include',
      });

      if (!response.ok) return;

      const data = await response.json();
      const codes: string[] = data.supplier_codes || data.codes || [];
      if (Array.isArray(data.batches) && data.batches.length > 0) {
        setSupplierBatches(data.batches);
      } else {
        setSupplierBatches(codes.map((code: string) => ({ code, label: code })));
      }
    } catch (err) {
      console.error('Failed to fetch supplier codes:', err);
    }
  }, []);

  const fetchServerStatus = useCallback(async () => {
    try {
      setStatusLoading(true);
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/status`, {
        credentials: 'include',
      });

      if (!response.ok) return;

      const data = await response.json();
      setServerStatus(data);
    } catch (err) {
      console.error('Failed to fetch server status:', err);
    } finally {
      setStatusLoading(false);
    }
  }, []);

  const fetchServerLogs = useCallback(async () => {
    try {
      setLogsLoading(true);
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/logs`, {
        credentials: 'include',
      });

      if (!response.ok) return;

      const data = await response.json();
      setLogs(data.logs || []);
    } catch (err) {
      console.error('Failed to fetch server logs:', err);
    } finally {
      setLogsLoading(false);
    }
  }, []);

  const fetchActivityLogs = useCallback(async (filters?: ActivityFilters) => {
    try {
      setActivityLoading(true);
      const fetchedLogs = await getUserActivityLogs(filters || { useful_only: true });
      setActivityLogs(fetchedLogs);
    } catch (err) {
      console.error('Failed to fetch activity logs:', err);
      setActivityLogs([]);
    } finally {
      setActivityLoading(false);
    }
  }, []);

  useEffect(() => {
    const initializePanel = async () => {
      void fetchUsers();
      void fetchServerStatus();
      void fetchServerLogs();
      void fetchActivityLogs();
      void fetchSupplierCodes();

      await logUserActivity({
        action: 'access_admin_panel',
        resource_type: 'system',
        resource_id: undefined,
        details: 'Доступ к панели администратора',
      });
    };

    void initializePanel();
  }, [fetchUsers, fetchServerStatus, fetchServerLogs, fetchActivityLogs, fetchSupplierCodes]);

  const handleApplyActivityFilters = async (filters: ActivityFilters) => {
    await fetchActivityLogs(filters);

    await logUserActivity({
      action: 'view_activity_logs',
      resource_type: 'system',
      resource_id: undefined,
      details: 'Применение фильтров для просмотра логов активности пользователей',
    });
  };

  const handleClearActivityFilters = () => {
    void fetchActivityLogs({ useful_only: true });
  };

  const handleAddUserSubmit = async (newUser: {
    email: string;
    name: string;
    initials: string;
    inn: string;
    password: string;
    role: string;
  }) => {
    setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users`, {
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
      void fetchUsers();

      await logUserActivity({
        action: 'create_user',
        resource_type: 'user',
        resource_id: undefined,
        details: `Создан новый пользователь: ${newUser.email}`,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add user');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateUserSubmit = async (
    userId: number,
    editForm: {
      name: string;
      email: string;
      initials: string;
      inn: string;
      role: string;
    }
  ) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users/${userId}`, {
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
      void fetchUsers();

      await logUserActivity({
        action: 'update_user',
        resource_type: 'user',
        resource_id: userId,
        details: `Обновлен пользователь: ${editForm.email}`,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update user');
    }
  };

  const handleDeleteUser = async (userId: number) => {
    if (!window.confirm('Вы уверены, что хотите удалить этого пользователя?')) {
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users/${userId}`, {
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

      void fetchUsers();

      await logUserActivity({
        action: 'delete_user',
        resource_type: 'user',
        resource_id: userId,
        details: `Удален пользователь с ID: ${userId}`,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete user');
    }
  };

  const handleDeleteZeroQuantityParts = async () => {
    if (!selectedSupplierCode) {
      setError('Пожалуйста, выберите код поставки');
      return;
    }

    if (
      !window.confirm(
        `Вы уверены, что хотите удалить все запчасти с количеством 0 для кода поставки "${selectedSupplierCode}"? Это действие нельзя отменить.`
      )
    ) {
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/parts/zero-quantity`, {
        method: 'POST',
        headers: {
          ...getAuthHeaders(),
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ supplier_code: selectedSupplierCode }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to delete zero quantity parts');
        return;
      }

      const result = await response.json();
      alert(`Удалено ${result.deleted_count} запчастей для кода поставки "${selectedSupplierCode}"`);
      setSelectedSupplierCode('');
      void fetchServerStatus();
      void fetchSupplierCodes();

      await logUserActivity({
        action: 'delete_zero_quantity_parts',
        resource_type: 'system',
        resource_id: undefined,
        details: `Удалено ${result.deleted_count} запчастей с quantity=0 для кода поставки "${selectedSupplierCode}"`,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete zero quantity parts');
    } finally {
      setLoading(false);
    }
  };

  if (loading && users.length === 0) {
    return <AdminSkeleton />;
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4 sm:mb-6">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold">Панель администратора</h1>
          <p className="text-gray-600 text-sm sm:text-base">Управление пользователями и настройками системы</p>
        </div>
        {user?.role === 'admin' && (
          <Button onClick={() => setShowAddUser(true)}>
            <UserPlus className="w-4 h-4 mr-2" />
            Добавить пользователя
          </Button>
        )}
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-md">
          <p className="text-red-600">{error}</p>
        </div>
      )}

      {/* Server Status and Logs */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <ServerStatusCard status={serverStatus} loading={statusLoading} />
        <ServerLogsCard logs={logs} loading={logsLoading} />
      </div>

      {/* User Activity Logs */}
      <UserActivityLogsCard
        logs={activityLogs}
        loading={activityLoading}
        users={users}
        onApplyFilters={handleApplyActivityFilters}
        onClearFilters={handleClearActivityFilters}
      />

      {/* Users Management */}
      <UsersManagementCard
        users={users}
        onEditUser={(u) => setEditingUser(u)}
        onDeleteUser={handleDeleteUser}
      />

      {/* Parts Management (Zero quantity parts deletion) */}
      <ZeroQuantityPartsCard
        supplierBatches={supplierBatches}
        selectedSupplierCode={selectedSupplierCode}
        onSelectSupplierCode={setSelectedSupplierCode}
        onDeleteZeroParts={handleDeleteZeroQuantityParts}
        onRefreshCodes={fetchSupplierCodes}
        loading={loading}
      />

      {/* Image Editor Test */}
      <ImageEditorTestCard />

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

      {/* Dialogs */}
      <AddUserDialog
        isOpen={showAddUser}
        onOpenChange={setShowAddUser}
        onSubmit={handleAddUserSubmit}
        loading={loading}
      />

      <EditUserDialog
        user={editingUser}
        isOpen={editingUser !== null}
        onClose={() => setEditingUser(null)}
        onSubmit={handleUpdateUserSubmit}
        loading={loading}
      />
    </div>
  );
};

export default AdminPanel;
