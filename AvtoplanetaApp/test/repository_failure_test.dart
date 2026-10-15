import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/errors/failure.dart';
import 'package:avtoplaneta_app/features/orders/repositories/customers_repository.dart';
import 'package:avtoplaneta_app/features/orders/repositories/orders_repository.dart';

void main() {
  group('AppFailure Tests', () {
    test('converts SocketException to NetworkFailure', () {
      const socketEx = SocketException('Failed host lookup');
      final failure = AppFailure.from(socketEx);

      expect(failure, isA<NetworkFailure>());
      expect(failure.message, contains('Отсутствует подключение'));
    });

    test('converts FormatException to ValidationFailure', () {
      const formatEx = FormatException('Invalid JSON payload');
      final failure = AppFailure.from(formatEx);

      expect(failure, isA<ValidationFailure>());
      expect(failure.message, contains('Ошибка формата данных'));
    });

    test('converts DioException timeout to NetworkFailure', () {
      final dioEx = DioException(
        requestOptions: RequestOptions(path: '/api/v1/orders'),
        type: DioExceptionType.connectionTimeout,
      );
      final failure = AppFailure.from(dioEx);

      expect(failure, isA<NetworkFailure>());
      expect(failure.message, contains('Превышено время ожидания'));
    });

    test('converts 401/403 badResponse to AuthFailure', () {
      final dioEx = DioException(
        requestOptions: RequestOptions(path: '/api/v1/orders'),
        type: DioExceptionType.badResponse,
        response: Response(
          requestOptions: RequestOptions(path: '/api/v1/orders'),
          statusCode: 401,
          data: {'error': 'Требуется авторизация'},
        ),
      );
      final failure = AppFailure.from(dioEx);

      expect(failure, isA<AuthFailure>());
      expect(failure.statusCode, 401);
      expect(failure.message, 'Требуется авторизация');
    });

    test('converts 409 Conflict to ValidationFailure with message', () {
      final dioEx = DioException(
        requestOptions: RequestOptions(path: '/api/v1/customers'),
        type: DioExceptionType.badResponse,
        response: Response(
          requestOptions: RequestOptions(path: '/api/v1/customers'),
          statusCode: 409,
          data: {'error': 'Клиент с таким телефоном уже существует'},
        ),
      );
      final failure = AppFailure.from(dioEx);

      expect(failure, isA<ValidationFailure>());
      expect(failure.statusCode, 409);
      expect(failure.message, 'Клиент с таким телефоном уже существует');
    });

    test('converts 500 badResponse to ServerFailure', () {
      final dioEx = DioException(
        requestOptions: RequestOptions(path: '/api/v1/orders'),
        type: DioExceptionType.badResponse,
        response: Response(
          requestOptions: RequestOptions(path: '/api/v1/orders'),
          statusCode: 500,
        ),
      );
      final failure = AppFailure.from(dioEx);

      expect(failure, isA<ServerFailure>());
      expect(failure.statusCode, 500);
      expect(failure.message, contains('Внутренняя ошибка сервера'));
    });

    test('converts HTML SPA error response to descriptive failure', () {
      final dioEx = DioException(
        requestOptions: RequestOptions(path: '/orders'),
        type: DioExceptionType.badResponse,
        response: Response(
          requestOptions: RequestOptions(path: '/orders'),
          statusCode: 200,
          data: '<!doctype html><html><body>Avtoplaneta Web</body></html>',
        ),
      );
      final failure = AppFailure.from(dioEx);

      expect(failure.message, contains('Сервер вернул HTML-страницу'));
    });
  });

  group('OrdersRepository and CustomersRepository parsing tests', () {
    test('OrdersRepository parses list response into OrdersResponse', () async {
      final dio = Dio();
      dio.interceptors.add(InterceptorsWrapper(
        onRequest: (options, handler) {
          handler.resolve(Response(
            requestOptions: options,
            statusCode: 200,
            data: [
              {
                'id': 101,
                'part_name': 'Бампер передний Toyota Corolla',
                'buyer_number': '+79991234567',
                'status': 'yellow',
                'price': 15000.0,
              }
            ],
          ));
        },
      ));

      final repo = OrdersRepository(null, dio);
      final res = await repo.fetchOrders();

      expect(res.orders.length, 1);
      expect(res.orders.first.id, 101);
      expect(res.orders.first.partName, 'Бампер передний Toyota Corolla');
      expect(res.orders.first.buyerNumber, '+79991234567');
      expect(res.total, 1);
    });

    test('CustomersRepository parses list response into List<Customer>', () async {
      final dio = Dio();
      dio.interceptors.add(InterceptorsWrapper(
        onRequest: (options, handler) {
          handler.resolve(Response(
            requestOptions: options,
            statusCode: 200,
            data: [
              {
                'id': 12,
                'name': 'Иван Иванов',
                'phone': '+79990001122',
                'category': 'regular',
                'orders_count': 3,
                'total_spent': 45000.0,
              }
            ],
          ));
        },
      ));

      final repo = CustomersRepository(null, dio);
      final list = await repo.fetchCustomers();

      expect(list.length, 1);
      expect(list.first.id, 12);
      expect(list.first.name, 'Иван Иванов');
      expect(list.first.phone, '+79990001122');
      expect(list.first.totalSpent, 45000.0);
    });
  });
}
