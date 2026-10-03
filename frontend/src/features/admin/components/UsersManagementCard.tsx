/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Users, Edit, Trash2 } from 'lucide-react';
import type { AdminUser } from '../types';

interface UsersManagementCardProps {
  users: AdminUser[];
  onEditUser: (user: AdminUser) => void;
  onDeleteUser: (userId: number) => void;
}

export const UsersManagementCard: React.FC<UsersManagementCardProps> = ({
  users,
  onEditUser,
  onDeleteUser,
}) => {
  return (
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
                    <p className="text-sm text-muted-foreground">{user.email}</p>
                    {user.initials && <p className="text-sm text-muted-foreground">Инициалы: {user.initials}</p>}
                    {user.inn && <p className="text-sm text-muted-foreground">ИНН: {user.inn}</p>}
                  </div>
                  <div className="flex space-x-2">
                    <span
                      className={`px-2 py-1 text-xs rounded-full ${
                        user.provider === 'google'
                          ? 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300'
                          : 'bg-muted text-foreground'
                      }`}
                    >
                      {user.provider}
                    </span>
                    <span
                      className={`px-2 py-1 text-xs rounded-full ${
                        user.role === 'admin'
                          ? 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300'
                          : user.role === 'manager'
                          ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-300'
                          : 'bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-300'
                      }`}
                    >
                      {user.role}
                    </span>
                  </div>
                </div>
                <div className="flex space-x-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onEditUser(user)}
                    className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 hover:bg-blue-50 dark:hover:bg-blue-950/30"
                  >
                    <Edit className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onDeleteUser(user.id)}
                    className="text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 hover:bg-red-50 dark:hover:bg-red-950/30"
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
  );
};
