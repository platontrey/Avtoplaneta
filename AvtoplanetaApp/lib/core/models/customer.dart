import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'order.dart';

class Customer {
  final int id;
  final String name;
  final String phone;
  final String city;
  final String preferredTk;
  final String tkDetails;
  final String notes;
  final String category; // 'regular', 'vip', 'wholesale', 'blacklist'
  final double discountPercent;
  final int totalOrders;
  final double totalSpent;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  const Customer({
    required this.id,
    required this.name,
    required this.phone,
    this.city = '',
    this.preferredTk = '',
    this.tkDetails = '',
    this.notes = '',
    this.category = 'regular',
    this.discountPercent = 0.0,
    this.totalOrders = 0,
    this.totalSpent = 0.0,
    this.createdAt,
    this.updatedAt,
  });

  factory Customer.fromJson(Map<String, dynamic> json) {
    return Customer(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: json['name'] as String? ?? '',
      phone: json['phone'] as String? ?? '',
      city: json['city'] as String? ?? '',
      preferredTk: json['preferred_tk'] as String? ?? '',
      tkDetails: json['tk_details'] as String? ?? '',
      notes: json['notes'] as String? ?? '',
      category: json['category'] as String? ?? 'regular',
      discountPercent: (json['discount_percent'] as num?)?.toDouble() ?? 0.0,
      totalOrders: (json['total_orders'] as num?)?.toInt() ?? 0,
      totalSpent: (json['total_spent'] as num?)?.toDouble() ?? 0.0,
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'].toString())
          : null,
      updatedAt: json['updated_at'] != null
          ? DateTime.tryParse(json['updated_at'].toString())
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'phone': phone,
      'city': city,
      'preferred_tk': preferredTk,
      'tk_details': tkDetails,
      'notes': notes,
      'category': category,
      'discount_percent': discountPercent,
    };
  }

  bool get isBlacklist => category == 'blacklist';
  bool get isVip => category == 'vip';
  bool get isWholesale => category == 'wholesale';

  String get categoryLabel {
    switch (category) {
      case 'vip':
        return 'СТО / VIP';
      case 'wholesale':
        return 'Оптовик';
      case 'blacklist':
        return 'Черный список ⚠️';
      case 'regular':
      default:
        return 'Обычный';
    }
  }

  Color get categoryColor {
    switch (category) {
      case 'vip':
        return const Color(0xFFD97706); // amber-600
      case 'wholesale':
        return const Color(0xFF2563EB); // blue-600
      case 'blacklist':
        return const Color(0xFFDC2626); // red-600
      case 'regular':
      default:
        return const Color(0xFF71717A); // zinc-500
    }
  }

  IconData get categoryIcon {
    switch (category) {
      case 'vip':
        return LucideIcons.crown;
      case 'wholesale':
        return LucideIcons.building;
      case 'blacklist':
        return LucideIcons.triangle_alert;
      case 'regular':
      default:
        return LucideIcons.user;
    }
  }
}

class CustomerDetails {
  final Customer customer;
  final List<Order> orders;

  const CustomerDetails({
    required this.customer,
    required this.orders,
  });

  factory CustomerDetails.fromJson(Map<String, dynamic> json) {
    final customerData = json['customer'] != null
        ? Map<String, dynamic>.from(json['customer'] as Map)
        : <String, dynamic>{};
    final ordersRaw = json['orders'] as List<dynamic>? ?? [];
    return CustomerDetails(
      customer: Customer.fromJson(customerData),
      orders: ordersRaw
          .whereType<Map>()
          .map((e) => Order.fromJson(Map<String, dynamic>.from(e)))
          .toList(),
    );
  }
}

class CustomersResponse {
  final List<Customer> customers;
  final int total;

  const CustomersResponse({
    required this.customers,
    required this.total,
  });

  factory CustomersResponse.fromJson(Map<String, dynamic> json) {
    final listRaw = json['customers'] as List<dynamic>? ?? [];
    return CustomersResponse(
      customers: listRaw
          .whereType<Map>()
          .map((e) => Customer.fromJson(Map<String, dynamic>.from(e)))
          .toList(),
      total: (json['total'] as num?)?.toInt() ?? listRaw.length,
    );
  }
}
