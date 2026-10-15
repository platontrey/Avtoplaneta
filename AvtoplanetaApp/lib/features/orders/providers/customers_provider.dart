import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/models/customer.dart';
import '../repositories/customers_repository.dart';

export '../repositories/customers_repository.dart';

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

/// Провайдер списка клиентов с учетом фильтров поиска и категории
final customersProvider = FutureProvider<List<Customer>>((ref) async {
  final filter = ref.watch(customerFilterProvider);
  final repository = ref.watch(customersRepositoryProvider);

  return repository.fetchCustomers(
    search: filter.search,
    category: filter.category,
  );
});

/// Провайдер карточки клиента
final customerDetailsProvider =
    FutureProvider.family<CustomerDetails, int>((ref, id) async {
  final repository = ref.watch(customersRepositoryProvider);
  return repository.fetchCustomerById(id);
});

/// Фасад для обратной совместимости с существующим UI кодом
class CustomerService {
  static final _repo = CustomersRepository(apiClient);

  static Future<Customer> createCustomer({
    required String name,
    required String phone,
    String city = '',
    String preferredTk = '',
    String tkDetails = '',
    String notes = '',
    String category = 'regular',
    double discountPercent = 0.0,
  }) {
    return _repo.createCustomer(
      name: name,
      phone: phone,
      city: city,
      preferredTk: preferredTk,
      tkDetails: tkDetails,
      notes: notes,
      category: category,
      discountPercent: discountPercent,
    );
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
  }) {
    return _repo.updateCustomer(
      id: id,
      name: name,
      phone: phone,
      city: city,
      preferredTk: preferredTk,
      tkDetails: tkDetails,
      notes: notes,
      category: category,
      discountPercent: discountPercent,
    );
  }

  static Future<void> deleteCustomer(int id) {
    return _repo.deleteCustomer(id);
  }
}
