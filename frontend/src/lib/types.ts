/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

export interface Part {
  id: number;
  name: string;
  quantity: number;
  description?: string;
  category?: string;
  price?: number;
  salesman?: string;
  location?: string;
  status?: boolean;
  brand?: string;
  model?: string;
  photo?: string;
  inn?: string;
  vin?: string;
  to_delete_at_formatted?: string;
  time_until_deletion?: string;
  body_brand?: string;
  engine_brand?: string;
  car_release_date?: string;
  front_rear?: string;
  left_right?: string;
  top_bottom?: string;
  number?: string;
  manufacturer?: string;
  manufacturer_code?: string;
  oem_code?: string;
  color?: string;
  condition?: string;
  supplier_code?: string;
  defect?: string;
  transmission?: string;
  drive?: string;
  wear_percentage?: string;
  season?: string;
  diameter?: string;
  width?: string;
  profile?: string;
  tire_quantity?: string;
  drilling?: string;
  offset?: string;
  center_hole_diameter?: string;
  tire_model?: string;
}

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
  categories: CategoryCount[];
  monthlySales: MonthlySales[];
}

export interface OrderItem {
  id: number;
  order_id: number;
  part_id: number;
  quantity: number;
}

export interface Order {
  id: number;
  customer_id: number;
  seller_id: number;
  seller: string;
  part: string;
  location?: string;
  buyer_number: string;
  order_number: string;
  status: string;
  status_text: string;
  created_at: string;
  created_at_formatted?: string;
  time_ago?: string;
  items: OrderItem[];
}

export interface User {
  id: number;
  email: string;
  name: string;
  provider: string;
  role: string;
}

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


