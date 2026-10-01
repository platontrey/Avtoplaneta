import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/customer.dart';

class CustomerFilter {
  final String search;
  final String category; // '', 'regular', 'vip', 'wholesale', 'blacklist'

  const CustomerFilter({
    this.search = '',
    this.category = '',
  });

  CustomerFilter copyWith({
    String? search,
    String? category,
  }) {
    return CustomerFilter(
      search: search ?? this.search,
      category: category ?? this.category,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is CustomerFilter &&
          runtimeType == other.runtimeType &&
          search == other.search &&
          category == other.category;

  @override
  int get hashCode => search.hashCode ^ category.hashCode;
}

final customerFilterProvider = StateProvider<CustomerFilter>((ref) {
  return const CustomerFilter();
});

final customersProvider = FutureProvider<List<Customer>>((ref) async {
  final filter = ref.watch(customerFilterProvider);
  final queryParams = <String, dynamic>{};
  if (filter.search.trim().isNotEmpty) {
    queryParams['search'] = filter.search.trim();
  }
  if (filter.category.trim().isNotEmpty) {
    queryParams['category'] = filter.category.trim();
  }

  final response = await apiClient.dio.get(
    '/orders/customers',
    queryParameters: queryParams,
  );

  final data = response.data;
  if (data is Map) {
    final parsed = CustomersResponse.fromJson(Map<String, dynamic>.from(data));
    return parsed.customers;
  }
  if (data is List) {
    return data
        .whereType<Map>()
        .map((c) => Customer.fromJson(Map<String, dynamic>.from(c)))
        .toList();
  }
  return const [];
});

final customerDetailsProvider =
    FutureProvider.family<CustomerDetails, int>((ref, id) async {
  final response = await apiClient.dio.get('/orders/customers/$id');
  final data = response.data;
  if (data is Map) {
    return CustomerDetails.fromJson(Map<String, dynamic>.from(data));
  }
  throw Exception('Неверный формат данных клиента');
});

class CustomerService {
  static Future<Customer> createCustomer({
    required String name,
    required String phone,
    String city = '',
    String preferredTk = '',
    String tkDetails = '',
    String notes = '',
    String category = 'regular',
    double discountPercent = 0.0,
  }) async {
    final response = await apiClient.dio.post(
      '/orders/customers',
      data: {
        'name': name.trim(),
        'phone': phone.trim(),
        'city': city.trim(),
        'preferred_tk': preferredTk.trim(),
        'tk_details': tkDetails.trim(),
        'notes': notes.trim(),
        'category': category,
        'discount_percent': discountPercent,
      },
    );
    return Customer.fromJson(Map<String, dynamic>.from(response.data as Map));
  }

  static Future<Customer> updateCustomer({
    required int id,
    required String name,
    required String phone,
    String city = '',
    String preferredTk = '',
    String tkDetails = '',
    String notes = '',
    String category = 'regular',
    double discountPercent = 0.0,
  }) async {
    final response = await apiClient.dio.put(
      '/orders/customers/$id',
      data: {
        'name': name.trim(),
        'phone': phone.trim(),
        'city': city.trim(),
        'preferred_tk': preferredTk.trim(),
        'tk_details': tkDetails.trim(),
        'notes': notes.trim(),
        'category': category,
        'discount_percent': discountPercent,
      },
    );
    return Customer.fromJson(Map<String, dynamic>.from(response.data as Map));
  }

  static Future<void> deleteCustomer(int id) async {
    await apiClient.dio.delete('/orders/customers/$id');
  }
}
