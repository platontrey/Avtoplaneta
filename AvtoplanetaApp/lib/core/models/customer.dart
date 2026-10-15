import 'package:freezed_annotation/freezed_annotation.dart';
import 'order.dart';

part 'customer.freezed.dart';
part 'customer.g.dart';

Object? _readField(Map json, String snakeKey, String camelKey) {
  return json[snakeKey] ?? json[camelKey];
}

@freezed
abstract class Customer with _$Customer {
  const Customer._();

  const factory Customer({
    @Default(0) int id,
    @Default('') String name,
    @Default('') String phone,
    @Default('') String city,
    @JsonKey(readValue: _readPreferredTk) @Default('') String preferredTk,
    @JsonKey(readValue: _readTkDetails) @Default('') String tkDetails,
    @Default('') String notes,
    @Default('regular') String category,
    @JsonKey(readValue: _readDiscountPercent) @Default(0.0) double discountPercent,
    @JsonKey(readValue: _readTotalOrders) @Default(0) int totalOrders,
    @JsonKey(readValue: _readTotalSpent) @Default(0.0) double totalSpent,
    @JsonKey(readValue: _readCreatedAt) DateTime? createdAt,
    @JsonKey(readValue: _readUpdatedAt) DateTime? updatedAt,
  }) = _Customer;

  factory Customer.fromJson(Map<String, dynamic> json) =>
      _$CustomerFromJson(json);

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
}

Object? _readPreferredTk(Map json, String key) => _readField(json, 'preferred_tk', 'preferredTk');
Object? _readTkDetails(Map json, String key) => _readField(json, 'tk_details', 'tkDetails');
Object? _readDiscountPercent(Map json, String key) => _readField(json, 'discount_percent', 'discountPercent');
Object? _readTotalOrders(Map json, String key) => _readField(json, 'total_orders', 'totalOrders');
Object? _readTotalSpent(Map json, String key) => _readField(json, 'total_spent', 'totalSpent');
Object? _readCreatedAt(Map json, String key) => _readField(json, 'created_at', 'createdAt');
Object? _readUpdatedAt(Map json, String key) => _readField(json, 'updated_at', 'updatedAt');

@freezed
abstract class CustomerDetails with _$CustomerDetails {
  const factory CustomerDetails({
    required Customer customer,
    @Default([]) List<Order> orders,
  }) = _CustomerDetails;

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

@freezed
abstract class CustomersResponse with _$CustomersResponse {
  const factory CustomersResponse({
    @Default([]) List<Customer> customers,
    @Default(0) int total,
  }) = _CustomersResponse;

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
