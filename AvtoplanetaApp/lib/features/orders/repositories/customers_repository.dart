import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/api_endpoints.dart';
import '../../../core/errors/failure.dart';
import '../../../core/models/customer.dart';

final customersRepositoryProvider = Provider<CustomersRepository>((ref) {
  return CustomersRepository(apiClient);
});

/// Репозиторий для управления клиентами (CRM), инкапсулирующий сетевые запросы и типизирующий ошибки.
class CustomersRepository {
  final Dio _dio;

  CustomersRepository([ApiClient? client, Dio? dio])
      : _dio = dio ?? (client ?? apiClient).dio;

  /// Получение списка клиентов с поиском и фильтром по категории.
  Future<List<Customer>> fetchCustomers({
    String? search,
    String? category,
  }) async {
    try {
      final queryParams = <String, dynamic>{};
      if (search != null && search.trim().isNotEmpty) {
        queryParams['search'] = search.trim();
      }
      if (category != null && category.trim().isNotEmpty) {
        queryParams['category'] = category.trim();
      }

      final response = await _dio.get(
        ApiEndpoints.customers,
        queryParameters: queryParams.isEmpty ? null : queryParams,
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
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Получение детальной карточки клиента с историей заказов и LTV.
  Future<CustomerDetails> fetchCustomerById(int id) async {
    try {
      final response = await _dio.get(ApiEndpoints.customerById(id));
      final data = response.data;
      if (data is Map) {
        return CustomerDetails.fromJson(Map<String, dynamic>.from(data));
      }
      throw const ValidationFailure('Неверный формат данных клиента');
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Создание нового клиента.
  Future<Customer> createCustomer({
    required String name,
    required String phone,
    String city = '',
    String preferredTk = '',
    String tkDetails = '',
    String notes = '',
    String category = 'regular',
    double discountPercent = 0.0,
  }) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.customers,
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
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Обновление существующего клиента.
  Future<Customer> updateCustomer({
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
    try {
      final response = await _dio.put(
        ApiEndpoints.customerById(id),
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
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Удаление клиента.
  Future<void> deleteCustomer(int id) async {
    try {
      await _dio.delete(ApiEndpoints.customerById(id));
    } catch (e) {
      throw AppFailure.from(e);
    }
  }
}
