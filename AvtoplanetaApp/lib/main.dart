import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'app/router.dart';
import 'app/theme.dart';
import 'core/storage/cache_storage.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await CacheStorage.init();
  runApp(const ProviderScope(child: AvtoplanetaApp()));
}

class AvtoplanetaApp extends ConsumerWidget {
  const AvtoplanetaApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    return MaterialApp.router(
      title: 'Avtoplaneta',
      theme: AppTheme.dark,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
