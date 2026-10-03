/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { UserActivityAction } from '@/lib/types';

export interface AdminUser {
  id: number;
  email: string;
  name: string;
  initials?: string;
  inn?: string;
  provider: string;
  role: string;
}

export interface ServerStatus {
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

export interface ServerLogEntry {
  timestamp: string;
  level: string;
  message: string;
}

export type UserActivityResourceType =
  | 'part'
  | 'order'
  | 'user'
  | 'photo'
  | 'system';

export interface ActivityFilters {
  user_id?: number;
  action?: UserActivityAction;
  resource_type?: UserActivityResourceType;
  start_date?: string;
  end_date?: string;
  useful_only?: boolean;
}

export interface SupplierBatch {
  code: string;
  label: string;
  zero_count?: number;
}
