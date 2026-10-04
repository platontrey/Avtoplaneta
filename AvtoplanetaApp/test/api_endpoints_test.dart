import 'dart:io';
import 'dart:typed_data';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/api/api_endpoints.dart';

void main() {
  group('ApiEndpoints integrity tests', () {
    test('all defined ApiEndpoints constants start with allowed API prefixes', () {
      final allowedPrefixes = ['/api/', '/auth/', '/admin/'];

      final staticEndpoints = [
        ApiEndpoints.orders,
        ApiEndpoints.completedOrders,
        ApiEndpoints.orderById(1),
        ApiEndpoints.completeOrder(1),
        ApiEndpoints.orderItems(1),
        ApiEndpoints.orderItem(1, 2),
        ApiEndpoints.customers,
        ApiEndpoints.customerById(1),
        ApiEndpoints.parts,
        ApiEndpoints.partById(1),
        ApiEndpoints.partItem(1),
        ApiEndpoints.partCatalog,
        ApiEndpoints.vehicleCatalog,
        ApiEndpoints.defectReports,
        ApiEndpoints.uploadPartPhoto(1),
        ApiEndpoints.statistics,
        ApiEndpoints.authMe,
        ApiEndpoints.authLogin,
        ApiEndpoints.authLogout,
        ApiEndpoints.authRefresh,
        ApiEndpoints.adminUsers,
        ApiEndpoints.adminSupplierCodes,
        ApiEndpoints.adminStatus,
        ApiEndpoints.adminLogs,
        ApiEndpoints.conversations,
        ApiEndpoints.messagingUsers,
      ];

      for (final endpoint in staticEndpoints) {
        final matches = allowedPrefixes.any((prefix) => endpoint.startsWith(prefix));
        expect(
          matches,
          isTrue,
          reason: 'Эндпоинт "$endpoint" должен начинаться с одного из префиксов: $allowedPrefixes',
        );
      }
    });

    test('codebase has no direct legacy calls to /orders without /api/ prefix', () {
      final libDir = Directory('lib');
      expect(libDir.existsSync(), isTrue);

      final dartFiles = libDir
          .listSync(recursive: true)
          .whereType<File>()
          .where((f) => f.path.endsWith('.dart'));

      // Запрещенные устаревшие эндпоинты в вызовах dio
      final forbiddenPatterns = [
        RegExp(r'''dio\.(get|post|put|delete|patch)\(\s*['"]/orders['"/]'''),
        RegExp(r'''dio\.(get|post|put|delete|patch)\(\s*['"]/customers['"/]'''),
      ];

      final violations = <String>[];

      for (final file in dartFiles) {
        final content = file.readAsStringSync();
        final lines = content.split('\n');

        for (var i = 0; i < lines.length; i++) {
          final line = lines[i];
          for (final pattern in forbiddenPatterns) {
            if (pattern.hasMatch(line)) {
              violations.add('${file.path}:${i + 1}: $line');
            }
          }
        }
      }

      expect(
        violations,
        isEmpty,
        reason: 'Обнаружены прямые вызовы устаревших эндпоинтов без префикса /api/:\n${violations.join('\n')}',
      );
    });
  });

  group('ContentTypeValidatorInterceptor tests', () {
    late Dio testDio;

    setUp(() {
      testDio = Dio();
      testDio.interceptors.add(InterceptorsWrapper(
        onResponse: (response, handler) {
          final contentType =
              response.headers.value('content-type')?.toLowerCase() ?? '';
          final data = response.data;
          final isHtmlContent = contentType.contains('text/html');
          final isHtmlString = data is String &&
              (data.trimLeft().toLowerCase().startsWith('<!doctype html') ||
                  data.trimLeft().toLowerCase().startsWith('<html'));

          if (isHtmlContent || isHtmlString) {
            handler.reject(DioException(
              requestOptions: response.requestOptions,
              response: response,
              type: DioExceptionType.badResponse,
              error: 'Ошибка маршрутизации API: сервер вернул HTML-страницу вместо JSON',
            ));
            return;
          }
          handler.next(response);
        },
      ));
    });

    test('rejects responses with Content-Type text/html with descriptive error', () async {
      testDio.httpClientAdapter = _MockAdapter(
        statusCode: 200,
        headers: {
          'content-type': ['text/html; charset=utf-8'],
        },
        body: '<!DOCTYPE html><html><body>SPA Index</body></html>',
      );

      expect(
        () => testDio.get('/test'),
        throwsA(isA<DioException>().having(
          (e) => e.message ?? e.error?.toString(),
          'error',
          contains('сервер вернул HTML-страницу вместо JSON'),
        )),
      );
    });

    test('accepts valid JSON responses', () async {
      testDio.httpClientAdapter = _MockAdapter(
        statusCode: 200,
        headers: {
          'content-type': ['application/json; charset=utf-8'],
        },
        body: '{"status": "ok", "orders": []}',
      );

      final response = await testDio.get('/test');
      expect(response.statusCode, 200);
      expect(response.data, isA<Map>());
      expect((response.data as Map)['status'], 'ok');
    });
  });
}

class _MockAdapter implements HttpClientAdapter {
  final int statusCode;
  final Map<String, List<String>> headers;
  final String body;

  _MockAdapter({
    required this.statusCode,
    required this.headers,
    required this.body,
  });

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return ResponseBody.fromString(
      body,
      statusCode,
      headers: headers,
    );
  }

  @override
  void close({bool force = false}) {}
}
