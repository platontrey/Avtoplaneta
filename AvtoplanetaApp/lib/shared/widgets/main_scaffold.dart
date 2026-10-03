import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../app/theme.dart';
import '../../core/services/update_service.dart';
import '../../features/updater/widgets/update_dialog.dart';

class MainScaffold extends ConsumerStatefulWidget {
  final Widget child;
  const MainScaffold({super.key, required this.child});

  @override
  ConsumerState<MainScaffold> createState() => _MainScaffoldState();
}

class _MainScaffoldState extends ConsumerState<MainScaffold> {
  static bool _hasCheckedForUpdates = false;

  @override
  void initState() {
    super.initState();
    if (!_hasCheckedForUpdates) {
      _hasCheckedForUpdates = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _checkUpdates();
      });
    }
  }

  Future<void> _checkUpdates() async {
    final updateService = ref.read(updateServiceProvider);
    final result = await updateService.checkForUpdates();
    if (result.hasUpdate && result.updateInfo != null && mounted) {
      UpdateDialog.show(
        context: context,
        info: result.updateInfo!,
        updateService: updateService,
      );
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
