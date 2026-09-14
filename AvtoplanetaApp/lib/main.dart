import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'app/router.dart';
import 'app/theme.dart';
import 'core/services/error_reporter.dart';
import 'core/storage/cache_storage.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await CacheStorage.init();
  await ErrorReporter.instance.init();

  // Глобальный перехват ошибок рендеринга и виджетов Flutter
  FlutterError.onError = (details) {
    FlutterError.presentError(details);
    ErrorReporter.instance.recordFlutterError(details);
  };

  // Глобальный перехват необработанных асинхронных сбоев Dart
  PlatformDispatcher.instance.onError = (error, stack) {
    ErrorReporter.instance.recordError(error, stack, type: 'fatal');
    return true;
  };

  runApp(const ProviderScope(child: AvtoplanetaApp()));
}

class AvtoplanetaApp extends ConsumerWidget {
  const AvtoplanetaApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    return MaterialApp.router(
      title: 'Автопланета',
      theme: AppTheme.dark,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
