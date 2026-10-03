/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FileText } from 'lucide-react';
import type { ServerLogEntry } from '../types';

interface ServerLogsCardProps {
  logs: ServerLogEntry[];
  loading: boolean;
}

export const ServerLogsCard: React.FC<ServerLogsCardProps> = ({ logs, loading }) => {
  return (
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
        {loading ? (
          <p className="text-center py-4">Загрузка логов...</p>
        ) : logs.length > 0 ? (
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {logs.map((log, index) => (
              <div key={index} className="flex items-start space-x-2 text-sm">
                <span
                  className={`px-2 py-1 text-xs rounded ${
                    log.level === 'ERROR'
                      ? 'bg-red-100 text-red-800'
                      : log.level === 'WARN'
                      ? 'bg-yellow-100 text-yellow-800'
                      : 'bg-blue-100 text-blue-800'
                  }`}
                >
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
  );
};
