import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:open_filex/open_filex.dart';
import 'package:avtoplaneta_app/core/services/update_service.dart';
import 'package:avtoplaneta_app/features/updater/widgets/update_dialog.dart';

/// Мок-сервис обновления с подсчетом вызовов и симуляцией ошибки сокета
class _SimulatedFailingUpdateService extends UpdateService {
  int attempts = 0;
  final bool alwaysFail;

  _SimulatedFailingUpdateService({this.alwaysFail = false});

  @override
  Future<OpenResult> downloadAndInstall({
    required AppUpdateInfo info,
    required void Function(double progress, int receivedBytes, int totalBytes) onProgress,
    CancelToken? cancelToken,
  }) async {
    attempts++;
    if (alwaysFail || attempts == 1) {
      throw DioException(
        requestOptions: RequestOptions(path: info.downloadUrl),
        error: const HttpException('Connection closed while receiving data, uri = https://release-assets.githubusercontent.com/test.apk'),
        type: DioExceptionType.unknown,
      );
    }
    onProgress(1.0, 40000000, 40000000);
    return OpenResult(type: ResultType.done, message: 'Успешно');
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUpAll(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger.setMockMethodCallHandler(
      const MethodChannel('plugins.flutter.io/path_provider'),
      (methodCall) async {
        if (methodCall.method == 'getTemporaryDirectory') {
          return Directory.systemTemp.path;
        }
        return null;
      },
    );
  });

  tearDownAll(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger.setMockMethodCallHandler(
      const MethodChannel('plugins.flutter.io/path_provider'),
      null,
    );
  });

  group('Network Resilience Download Tests', () {
    testWidgets('диалог обновления отображает понятное сообщение при Connection closed', (tester) async {
      const info = AppUpdateInfo(
        version: '1.2.0',
        buildNumber: 15,
        downloadUrl: 'https://release-assets.githubusercontent.com/test.apk',
        forceUpdate: false,
        minSupportedBuild: 1,
        changelog: 'Тест разрыва соединения',
      );

      final failingService = _SimulatedFailingUpdateService(alwaysFail: true);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => UpdateDialog.show(
                  context: context,
                  info: info,
                  updateService: failingService,
                ),
                child: const Text('Открыть'),
              ),
            ),
          ),
        ),
      );

      // Открываем диалог
      await tester.tap(find.text('Открыть'));
      await tester.pumpAndSettle();

      // Нажимаем «Обновить»
      await tester.tap(find.text('Обновить'));
      await tester.pumpAndSettle();

      // Проверяем, что отображается понятное русское сообщение без сырого CDN URL
      expect(
        find.textContaining('Ошибка загрузки: связь с сервером прервана. Нажмите «Повторить загрузку».'),
        findsOneWidget,
      );

      // Проверяем, что кнопка переключилась в «Повторить загрузку»
      expect(find.text('Повторить загрузку'), findsOneWidget);
    });

    testWidgets('кнопка Повторить загрузку повторно вызывает downloadAndInstall', (tester) async {
      const info = AppUpdateInfo(
        version: '1.2.0',
        buildNumber: 15,
        downloadUrl: 'https://release-assets.githubusercontent.com/test.apk',
        forceUpdate: false,
        minSupportedBuild: 1,
        changelog: 'Тест повтора',
      );

      // В первый раз сбоит, во второй раз проходит успешно
      final recoveringService = _SimulatedFailingUpdateService(alwaysFail: false);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => UpdateDialog.show(
                  context: context,
                  info: info,
                  updateService: recoveringService,
                ),
                child: const Text('Открыть'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Открыть'));
      await tester.pumpAndSettle();

      // Попытка 1 (сбой)
      await tester.tap(find.text('Обновить'));
      await tester.pumpAndSettle();

      expect(recoveringService.attempts, 1);
      expect(find.text('Повторить загрузку'), findsOneWidget);

      // Попытка 2 (нажимаем «Повторить загрузку»)
      await tester.tap(find.text('Повторить загрузку'));
      await tester.pumpAndSettle();

      expect(recoveringService.attempts, 2);
    });

    test('отмена загрузки через CancelToken прерывает процесс', () async {
      final service = UpdateService();
      final cancelToken = CancelToken();

      const info = AppUpdateInfo(
        version: '2.0.0',
        buildNumber: 20,
        downloadUrl: 'https://example.com/cancelled.apk',
        forceUpdate: false,
        minSupportedBuild: 1,
        changelog: 'Отмена',
      );

      cancelToken.cancel('Пользователь закрыл диалог');

      expect(
        () async => await service.downloadAndInstall(
          info: info,
          onProgress: (progress, received, total) {},
          cancelToken: cancelToken,
        ),
        throwsA(isA<DioException>()),
      );
    });
  });
}
