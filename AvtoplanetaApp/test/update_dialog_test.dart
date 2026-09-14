import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:open_filex/open_filex.dart';
import 'package:avtoplaneta_app/core/services/update_service.dart';
import 'package:avtoplaneta_app/features/updater/widgets/update_dialog.dart';

class _FakeUpdateService extends UpdateService {
  bool downloadCalled = false;
  final bool shouldFailDownload;

  _FakeUpdateService({this.shouldFailDownload = false}) : super();

  @override
  Future<OpenResult> downloadAndInstall({
    required AppUpdateInfo info,
    required void Function(double progress, int receivedBytes, int totalBytes) onProgress,
    CancelToken? cancelToken,
  }) async {
    downloadCalled = true;
    if (shouldFailDownload) {
      throw Exception('Сбой загрузки файла');
    }
    // Имитируем прогресс загрузки
    onProgress(0.5, 5000000, 10000000);
    onProgress(1.0, 10000000, 10000000);
    return OpenResult(type: ResultType.done, message: 'Успешно');
  }
}

void main() {
  group('UpdateDialog Widget Tests', () {
    testWidgets('отображение информации об обычном обновлении', (tester) async {
      const info = AppUpdateInfo(
        version: '1.2.0',
        buildNumber: 15,
        downloadUrl: '/api/v1/app/download',
        forceUpdate: false,
        minSupportedBuild: 1,
        changelog: 'Тестовый список изменений',
      );

      final fakeService = _FakeUpdateService();

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => UpdateDialog.show(
                  context: context,
                  info: info,
                  updateService: fakeService,
                ),
                child: const Text('Открыть'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Открыть'));
      await tester.pumpAndSettle();

      expect(find.text('Доступно обновление'), findsOneWidget);
      expect(find.textContaining('1.2.0'), findsOneWidget);
      expect(find.textContaining('сборка 15'), findsOneWidget);
      expect(find.text('Тестовый список изменений'), findsOneWidget);
      expect(find.text('Позже'), findsOneWidget);
      expect(find.text('Обновить'), findsOneWidget);
      expect(find.text('Это обязательное обновление для продолжения работы.'), findsNothing);

      // Клик на кнопку Позже закрывает диалог
      await tester.tap(find.text('Позже'));
      await tester.pumpAndSettle();
      expect(find.text('Доступно обновление'), findsNothing);
    });

    testWidgets('отображение обязательного обновления (force_update)', (tester) async {
      const info = AppUpdateInfo(
        version: '2.0.0',
        buildNumber: 50,
        downloadUrl: '/api/v1/app/download',
        forceUpdate: true,
        minSupportedBuild: 40,
        changelog: 'Критическое обновление',
      );

      final fakeService = _FakeUpdateService();

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => UpdateDialog.show(
                  context: context,
                  info: info,
                  updateService: fakeService,
                ),
                child: const Text('Открыть'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Открыть'));
      await tester.pumpAndSettle();

      expect(find.text('Доступно обновление'), findsOneWidget);
      expect(find.text('Это обязательное обновление для продолжения работы.'), findsOneWidget);
      // Кнопка Позже должна отсутствовать
      expect(find.text('Позже'), findsNothing);
      expect(find.text('Обновить'), findsOneWidget);
    });

    testWidgets('нажатие кнопки Обновить запускает скачивание', (tester) async {
      const info = AppUpdateInfo(
        version: '1.2.0',
        buildNumber: 15,
        downloadUrl: '/api/v1/app/download',
        forceUpdate: false,
        minSupportedBuild: 1,
        changelog: 'Чейнджлог',
      );

      final fakeService = _FakeUpdateService();

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => UpdateDialog.show(
                  context: context,
                  info: info,
                  updateService: fakeService,
                ),
                child: const Text('Открыть'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Открыть'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Обновить'));
      await tester.pump();

      expect(fakeService.downloadCalled, isTrue);
    });

    testWidgets('обработка ошибки при скачивании файла', (tester) async {
      const info = AppUpdateInfo(
        version: '1.2.0',
        buildNumber: 15,
        downloadUrl: '/api/v1/app/download',
        forceUpdate: false,
        minSupportedBuild: 1,
        changelog: 'Чейнджлог',
      );

      final fakeService = _FakeUpdateService(shouldFailDownload: true);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => UpdateDialog.show(
                  context: context,
                  info: info,
                  updateService: fakeService,
                ),
                child: const Text('Открыть'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Открыть'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Обновить'));
      await tester.pumpAndSettle();

      expect(find.textContaining('Ошибка загрузки:'), findsOneWidget);
    });
  });
}
