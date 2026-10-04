/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

export type { Part, PartFormData, PartFilters } from '@/features/parts/types';

export interface CategoryCount {
  name: string;
  count: number;
}

export interface MonthlySales {
  month: string;
  sales: number;
}

export interface StatisticsResponse {
  totalParts: number;
  totalQuantity: number;
  totalValue: number;
  totalEarnings: number;
  partsGrowth?: number;
  quantityGrowth?: number;
  valueGrowth?: number;
  earningsGrowth?: number;
  categories: CategoryCount[];
  monthlySales: MonthlySales[];
}

export interface OrderItem {
  id: number;
  order_id: number;
  part_id: number;
  part_name?: string;
  part_name_snapshot?: string;
  location?: string;
  quantity: number;
  price?: number;
}

export interface Order {
  id: number;
  customer_id: number;
  seller_id: number;
  seller: string;
  part: string;
  part_id?: number;
  location?: string;
  buyer_number: string;
  order_number: string;
  source?: string;
  status: string;
  status_text: string;
  payment_status?: string;
  warehouse_status?: string;
  delivery_method?: string;
  transport_company?: string;
  tracking_number?: string;
  notes?: string;
  discount?: number;
  total_amount?: number;
  auto_deleted?: boolean;
  created_at: string;
  created_at_formatted?: string;
  time_ago?: string;
  completed_at?: string;
  completed_at_formatted?: string;
  items: OrderItem[];
}

export type { User, LoginCredentials, AuthState } from '@/features/auth/types';

export interface UserActivityLog {
  id: number;
  user_id: number;
  user_name: string;
  user_email: string;
  action: string;
  resource_type: string;
  resource_id?: number;
  details?: string;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
  created_at_formatted?: string;
}

export type UserActivityAction =
    | 'part'
    | 'order'
    | 'user'
    | 'photo'
    | 'search'
    | 'system'
    | 'login'
    | 'logout'
    | 'create_part'
    | 'update_part'
    | 'delete_part'
    | 'bulk_update_parts'
    | 'bulk_delete_parts'
    | 'delete_zero_quantity_parts'
    | 'create_order'
    | 'update_order'
    | 'delete_order'
    | 'create_user'
    | 'update_user'
    | 'delete_user'
    | 'upload_photo'
    | 'delete_photo'
    | 'search_parts'
    | 'view_part'
    | 'view_users'
    | 'view_activity_logs'
    | 'access_admin_panel'
    | 'export_data'
    | 'change_password'
    | 'update_profile'
    | 'create_defect_report'
    | 'update_defect_report'
    | 'delete_defect_report'
    | 'mark_part_for_deletion';

export type UserActivityResourceType =
    | 'part'
    | 'order'
    | 'user'
    | 'photo'
    | 'system';

export type CustomerCategory = 'regular' | 'vip' | 'wholesale' | 'blacklist';

export interface Customer {
  id: number;
  name: string;
  phone: string;
  city: string;
  preferred_tk: string;
  passport_or_inn: string;
  category: CustomerCategory;
  discount_percent: number;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface CustomerWithStats extends Customer {
  total_orders: number;
  total_spent: number;
  last_order_at?: string;
}

export interface CustomerDetails extends CustomerWithStats {
  orders: Order[];
}

export interface CreateCustomerInput {
  name: string;
  phone: string;
  city?: string;
  preferred_tk?: string;
  passport_or_inn?: string;
  category?: CustomerCategory;
  discount_percent?: number;
  notes?: string;
}

export interface UpdateCustomerInput {
  name?: string;
  phone?: string;
  city?: string;
  preferred_tk?: string;
  passport_or_inn?: string;
  category?: CustomerCategory;
  discount_percent?: number;
  notes?: string;
}

