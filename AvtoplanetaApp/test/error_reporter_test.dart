import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/services/error_reporter.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late ErrorReporter reporter;

  setUp(() async {
    reporter = ErrorReporter.instance;
    await reporter.clearLogs();
  });

  tearDown(() async {
    await reporter.clearLogs();
  });

  group('ErrorReporter Core Tests', () {
    test('запись сетевой ошибки с маскированием конфиденциальных данных', () {
      final dioException = DioException(
        requestOptions: RequestOptions(path: '/auth/login', method: 'POST'),
        response: Response(
          requestOptions: RequestOptions(path: '/auth/login', method: 'POST'),
          statusCode: 401,
          data: {
            'error': 'Неверный пароль',
            'password': 'super_secret_password',
            'refresh_token': 'jwt_refresh_token_123',
            'status': 'unauthorized',
          },
        ),
      );

      reporter.recordNetworkError(
        method: 'POST',
        endpoint: '/auth/login',
        statusCode: 401,
        error: dioException,
        responseData: dioException.response?.data,
      );

      expect(reporter.logs.length, 1);
      final log = reporter.logs.first;
      expect(log.type, 'network');
      expect(log.statusCode, 401);
      expect(log.method, 'POST');
      expect(log.endpoint, '/auth/login');
      expect(log.message, contains('401 Unauthorized'));

      // Проверяем, что пароли и токены замаскированы
      expect(log.responseBody, isNotNull);
      expect(log.responseBody, contains('***MASKED***'));
      expect(log.responseBody, isNot(contains('super_secret_password')));
      expect(log.responseBody, isNot(contains('jwt_refresh_token_123')));
      expect(log.responseBody, contains('Неверный пароль'));
    });

    test('запись ошибки рендеринга Flutter', () {
      final details = FlutterErrorDetails(
        exception: Exception('RenderFlex overflowed by 42 pixels'),
        stack: StackTrace.current,
        library: 'rendering library',
      );

      reporter.recordFlutterError(details);

      expect(reporter.logs.length, 1);
      final log = reporter.logs.first;
      expect(log.type, 'flutter');
      expect(log.message, contains('RenderFlex overflowed'));
      expect(log.stackTrace, isNotNull);
    });

    test('запись фатального асинхронного сбоя Dart', () {
      reporter.recordError(
        'RangeError (index): Invalid value: Valid value range is empty: 0',
        StackTrace.current,
        type: 'fatal',
        customMessage: 'Критический сбой списка',
      );

      expect(reporter.logs.length, 1);
      final log = reporter.logs.first;
      expect(log.type, 'fatal');
      expect(log.message, 'Критический сбой списка');
      expect(log.stackTrace, isNotNull);
    });

    test('экспорт форматированного текстового отчета', () {
      reporter.recordError('Тестовая ошибка 1', null, type: 'custom');
      reporter.recordNetworkError(
        method: 'GET',
        endpoint: '/admin/users',
        statusCode: 401,
        error: 'Unauthorized',
      );

      final report = reporter.exportReportText();

      expect(report, contains('ОТЧЕТ ОБ ОШИБКАХ ПРИЛОЖЕНИЯ «АВТОПЛАНЕТА»'));
      expect(report, contains('Устройство:'));
      expect(report, contains('Версия:'));
      expect(report, contains('/admin/users'));
      expect(report, contains('401'));
      expect(report, contains('Тестовая ошибка 1'));
    });

    test('очистка журнала логов', () async {
      reporter.recordError('Ошибка 1', null);
      reporter.recordError('Ошибка 2', null);
      expect(reporter.logs.length, 2);

      await reporter.clearLogs();
      expect(reporter.logs.isEmpty, true);
    });

    test('ограничение максимального количества логов', () {
      for (int i = 0; i < 150; i++) {
        reporter.recordError('Ошибка #$i', null);
      }

      expect(reporter.logs.length, ErrorReporter.maxLogsCount);
      // Самая свежая ошибка должна быть в начале
      expect(reporter.logs.first.message, 'Ошибка #149');
    });
  });
}
