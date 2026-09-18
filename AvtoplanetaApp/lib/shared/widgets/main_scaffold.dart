import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../app/theme.dart';
import '../../core/services/update_service.dart';
import '../../features/updater/widgets/update_dialog.dart';
import 'ai_assistant_sheet.dart';

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
    if (location.startsWith('/inventory/add') ||
        location.startsWith('/inventory/defect-report')) {
      return 3;
    }
    if (location.startsWith('/inventory')) return 0;
    if (location.startsWith('/statistics')) return 1;
    if (location.startsWith('/messages')) return 2;
    if (location.startsWith('/orders')) return 4;
    return 0;
  }

  void _showAddMenu(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Что добавить?',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 6),
              const Text(
                'Выберите подходящий сценарий',
                style: TextStyle(color: AppTheme.mutedColor),
              ),
              const SizedBox(height: 16),
              ListTile(
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                tileColor: AppTheme.cardColor,
                leading: const Icon(
                  LucideIcons.package_plus,
                  color: AppTheme.primaryColor,
                ),
                title: const Text('Добавить запчасть'),
                subtitle: const Text('Одна позиция в инвентарь'),
                onTap: () {
                  Navigator.pop(context);
                  context.go('/inventory/add');
                },
              ),
              const SizedBox(height: 10),
              ListTile(
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                tileColor: AppTheme.cardColor,
                leading: const Icon(
                  LucideIcons.clipboard_check,
                  color: AppTheme.secondaryColor,
                ),
                title: const Text('Создать дефектную ведомость'),
                subtitle: const Text('Добавить сразу несколько позиций'),
                onTap: () {
                  Navigator.pop(context);
                  context.go('/inventory/defect-report');
                },
              ),
            ],
          ),
        ),
      ),
    );
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
        icon: Icon(LucideIcons.chart_column_increasing),
        selectedIcon: Icon(LucideIcons.chart_column_increasing),
        label: 'Статистика',
      ),
      NavigationDestination(
        icon: Icon(LucideIcons.message_circle),
        selectedIcon: Icon(LucideIcons.message_circle),
        label: 'Сообщения',
      ),
      NavigationDestination(
        icon: Icon(LucideIcons.circle_plus),
        selectedIcon: Icon(LucideIcons.circle_plus),
        label: 'Добавить',
      ),
      NavigationDestination(
        icon: Icon(LucideIcons.receipt),
        selectedIcon: Icon(LucideIcons.receipt),
        label: 'Заказы',
      ),
      NavigationDestination(
        icon: Icon(LucideIcons.bot),
        selectedIcon: Icon(LucideIcons.bot),
        label: 'ИИ',
      ),
    ];

    return Scaffold(
      body: widget.child,
      bottomNavigationBar: Container(
        decoration: const BoxDecoration(
          border: Border(top: BorderSide(color: AppTheme.borderColor)),
        ),
        child: NavigationBar(
          height: 70,
          backgroundColor: AppTheme.surfaceColor,
          surfaceTintColor: Colors.transparent,
          indicatorColor: AppTheme.primaryColor.withValues(alpha: 0.16),
          labelBehavior: NavigationDestinationLabelBehavior.onlyShowSelected,
          selectedIndex: currentIndex,
          destinations: destinations,
          onDestinationSelected: (index) {
            switch (index) {
              case 0:
                context.go('/inventory');
              case 1:
                context.go('/statistics');
              case 2:
                context.go('/messages');
              case 3:
                _showAddMenu(context);
              case 4:
                context.go('/orders');
              case 5:
                showModalBottomSheet<void>(
                  context: context,
                  isScrollControlled: true,
                  builder: (_) => const AIAssistantSheet(),
                );
            }
          },
        ),
      ),
    );
  }
}
