import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../app/theme.dart';
import 'ai_assistant_sheet.dart';

class MainScaffold extends StatelessWidget {
  final Widget child;
  const MainScaffold({super.key, required this.child});

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
                  Icons.add_box_outlined,
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
                  Icons.fact_check_outlined,
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

    if (location.startsWith('/admin')) return Scaffold(body: child);

    const destinations = [
      NavigationDestination(
        icon: Icon(Icons.inventory_2_outlined),
        selectedIcon: Icon(Icons.inventory_2_rounded),
        label: 'Инвентарь',
      ),
      NavigationDestination(
        icon: Icon(Icons.bar_chart_outlined),
        selectedIcon: Icon(Icons.bar_chart_rounded),
        label: 'Статистика',
      ),
      NavigationDestination(
        icon: Icon(Icons.chat_outlined),
        selectedIcon: Icon(Icons.chat_rounded),
        label: 'Сообщения',
      ),
      NavigationDestination(
        icon: Icon(Icons.add_circle_outline_rounded),
        selectedIcon: Icon(Icons.add_circle_rounded),
        label: 'Добавить',
      ),
      NavigationDestination(
        icon: Icon(Icons.receipt_long_outlined),
        selectedIcon: Icon(Icons.receipt_long_rounded),
        label: 'Заказы',
      ),
      NavigationDestination(
        icon: Icon(Icons.smart_toy_outlined),
        selectedIcon: Icon(Icons.smart_toy_rounded),
        label: 'ИИ',
      ),
    ];

    return Scaffold(
      body: child,
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
