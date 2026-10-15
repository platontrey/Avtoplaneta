import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/api_endpoints.dart';
import '../../../core/errors/failure.dart';
import '../../../core/models/order.dart';

final ordersRepositoryProvider = Provider<OrdersRepository>((ref) {
  return OrdersRepository(apiClient);
});

/// Репозиторий для работы с заказами, изолирующий вызовы Dio и возвращающий типизированный [AppFailure].
class OrdersRepository {
  final Dio _dio;

  OrdersRepository([ApiClient? client, Dio? dio])
      : _dio = dio ?? (client ?? apiClient).dio;

  /// Получение списка заказов с опциональной фильтрацией по статусу/состоянию.
  Future<OrdersResponse> fetchOrders({String? state}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (state != null && state.isNotEmpty) {
        queryParams['state'] = state;
      }

      final response = await _dio.get(
        ApiEndpoints.orders,
        queryParameters: queryParams.isEmpty ? null : queryParams,
      );

      final data = response.data;
      if (data is List) {
        final orders = data
            .whereType<Map>()
            .map((order) => Order.fromJson(Map<String, dynamic>.from(order)))
            .toList();
        return OrdersResponse(orders: orders, total: orders.length);
      }
      if (data is Map) {
        return OrdersResponse.fromJson(Map<String, dynamic>.from(data));
      }

      throw const ValidationFailure('Неверный формат ответа сервера при загрузке заказов');
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Получение детальной информации по конкретному заказу.
  Future<Order> fetchOrderById(int id) async {
    try {
      final response = await _dio.get(ApiEndpoints.orderById(id));
      final data = response.data;
      if (data is Map) {
        return Order.fromJson(Map<String, dynamic>.from(data));
      }
      throw const ValidationFailure('Неверный формат данных заказа');
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Создание нового заказа.
  Future<Order> createOrder(Map<String, dynamic> data) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.orders,
        data: data,
      );
      final responseData = response.data;
      if (responseData is Map) {
        return Order.fromJson(Map<String, dynamic>.from(responseData));
      }
      return const Order();
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Добавление детали к существующему заказу.
  Future<void> addOrderItem(int orderId, Map<String, dynamic> itemData) async {
    try {
      await _dio.post(
        ApiEndpoints.orderItems(orderId),
        data: itemData,
      );
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Частичное обновление полей заказа (статусы, трек-номер, скидка и т.д.).
  Future<void> updateOrder(int orderId, Map<String, dynamic> patchData) async {
    try {
      await _dio.patch(
        ApiEndpoints.orderById(orderId),
        data: patchData,
      );
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Обновление конкретной позиции внутри заказа (количество, цена).
  Future<void> updateOrderItem(
    int orderId,
    int itemId,
    Map<String, dynamic> patchData,
  ) async {
    try {
      await _dio.patch(
        ApiEndpoints.orderItem(orderId, itemId),
        data: patchData,
      );
    } catch (e) {
      throw AppFailure.from(e);
    }
  }

  /// Выполнение действия жизненного цикла заказа (PUT или DELETE).
  Future<void> executeOrderAction(String path, {String method = 'PUT'}) async {
    try {
      if (method.toUpperCase() == 'DELETE') {
        await _dio.delete(path);
      } else {
        await _dio.put(path);
      }
    } catch (e) {
      throw AppFailure.from(e);
    }
  }
}
