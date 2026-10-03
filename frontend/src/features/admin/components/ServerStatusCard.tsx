/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Server } from 'lucide-react';
import type { ServerStatus } from '../types';

interface ServerStatusCardProps {
  status: ServerStatus | null;
  loading: boolean;
}

export const ServerStatusCard: React.FC<ServerStatusCardProps> = ({ status, loading }) => {
  return (
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
        {loading ? (
          <p className="text-center py-4">Загрузка статуса сервера...</p>
        ) : status ? (
          <div className="space-y-4">
            <div>
              <h4 className="font-medium mb-2">Сервер</h4>
              <div className="space-y-1 text-sm">
                <p>
                  <span className="font-medium">Status:</span>{' '}
                  <span className="text-green-600">{status.server?.status || 'Unknown'}</span>
                </p>
                <p>
                  <span className="font-medium">Go Version:</span>{' '}
                  {status.server?.go_version || 'Unknown'}
                </p>
                <p>
                  <span className="font-medium">OS:</span> {status.server?.os || 'Unknown'}
                </p>
                <p>
                  <span className="font-medium">Architecture:</span>{' '}
                  {status.server?.arch || 'Unknown'}
                </p>
              </div>
            </div>
            <div>
              <h4 className="font-medium mb-2">База данных</h4>
              <div className="space-y-1 text-sm">
                <p>
                  <span className="font-medium">Status:</span>{' '}
                  <span className="text-green-600">{status.database?.status || 'Unknown'}</span>
                </p>
                <p>
                  <span className="font-medium">Total Parts:</span>{' '}
                  {status.database?.total_parts || 0}
                </p>
                <p>
                  <span className="font-medium">Total Users:</span>{' '}
                  {status.database?.total_users || 0}
                </p>
              </div>
            </div>
            <div className="text-xs text-gray-500">
              Last updated:{' '}
              {status.timestamp ? new Date(status.timestamp).toLocaleString() : 'Unknown'}
            </div>
          </div>
        ) : (
          <p className="text-center py-4 text-red-500">Не удалось загрузить статус сервера</p>
        )}
      </CardContent>
    </Card>
  );
};
