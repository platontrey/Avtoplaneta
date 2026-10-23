import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../app/theme.dart';
import '../../core/services/update_service.dart';
import '../../features/inventory/providers/inventory_provider.dart';
import '../../features/updater/widgets/update_dialog.dart';

class MainScaffold extends ConsumerStatefulWidget {
  final Widget child;
  const MainScaffold({super.key, required this.child});

  @override
  ConsumerState<MainScaffold> createState() => _MainScaffoldState();
}

class _MainScaffoldState extends ConsumerState<MainScaffold> with WidgetsBindingObserver {
  static DateTime? _lastUpdateCheck;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _checkUpdatesIfNeeded();
    });
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _recoverFromOfflineIfNeeded();
      _checkUpdatesIfNeeded();
    }
  }

  void _recoverFromOfflineIfNeeded() {
    final filter = ref.read(inventoryFilterProvider);
    final currentInventory = ref.read(inventoryProvider(filter));
    if (currentInventory.hasError ||
        currentInventory.valueOrNull?.isOffline == true) {
      ref.invalidate(inventoryProvider(filter));
    }
  }

  Future<void> _checkUpdatesIfNeeded() async {
    final now = DateTime.now();
    if (_lastUpdateCheck != null && now.difference(_lastUpdateCheck!) < const Duration(minutes: 15)) {
      return;
    }
    _lastUpdateCheck = now;

    try {
      final updateService = ref.read(updateServiceProvider);
      final result = await updateService.checkForUpdates();
      if (result.hasUpdate && result.updateInfo != null && mounted) {
        final info = result.updateInfo!;
        // Запускаем фоновую предзагрузку APK для ускорения установки
        updateService.preDownloadUpdate(info);
        UpdateDialog.show(
          context: context,
          info: info,
          updateService: updateService,
        );
      }
    } catch (e) {
      debugPrint('Ошибка автопроверки обновлений: $e');
    }
  }

  int _locationToIndex(String location) {
    if (location.startsWith('/inventory')) return 0;
    if (location.startsWith('/orders')) return 1;
    if (location.startsWith('/messages')) return 2;
    if (location.startsWith('/statistics')) return 3;
    return 0;
  }

  @override
  Widget build(BuildContext context) {
    final location = GoRouterState.of(context).matchedLocation;
    final currentIndex = _locationToIndex(location);

    if (location.startsWith('/admin')) return Scaffold(body: widget.child);

    const destinations = [
      NavigationDestination(
        icon: Icon(LucideIcons.package),
        selectedIcon: Icon(LucideIcons.package),
        label: 'Инвентарь',
      ),
      NavigationDestination(
        icon: Icon(LucideIcons.receipt),
        selectedIcon: Icon(LucideIcons.receipt),
        label: 'Заказы',
      ),
      NavigationDestination(
        icon: Icon(LucideIcons.message_circle),
        selectedIcon: Icon(LucideIcons.message_circle),
        label: 'Сообщения',
      ),
      NavigationDestination(
        icon: Icon(LucideIcons.chart_column_increasing),
        selectedIcon: Icon(LucideIcons.chart_column_increasing),
        label: 'Статистика',
      ),
    ];

    return Scaffold(
      body: widget.child,
      bottomNavigationBar: Container(
        decoration: const BoxDecoration(
          border: Border(top: BorderSide(color: AppTheme.borderColor)),
        ),
        child: NavigationBar(
          height: 68,
          backgroundColor: AppTheme.surfaceColor,
          surfaceTintColor: Colors.transparent,
          indicatorColor: AppTheme.primaryColor.withValues(alpha: 0.16),
          labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
          selectedIndex: currentIndex,
          destinations: destinations,
          onDestinationSelected: (index) {
            switch (index) {
              case 0:
                context.go('/inventory');
              case 1:
                context.go('/orders');
              case 2:
                context.go('/messages');
              case 3:
                context.go('/statistics');
            }
          },
        ),
      ),
    );
  }
}
