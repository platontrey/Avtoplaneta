import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
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
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.add_box_outlined),
              title: const Text('Добавить запчасть'),
              onTap: () {
                Navigator.pop(context);
                context.go('/inventory/add');
              },
            ),
            ListTile(
              leading: const Icon(Icons.fact_check_outlined),
              title: const Text('Создать дефектную ведомость'),
              onTap: () {
                Navigator.pop(context);
                context.go('/inventory/defect-report');
              },
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final location = GoRouterState.of(context).matchedLocation;
    final currentIndex = _locationToIndex(location);

    if (location.startsWith('/admin')) return Scaffold(body: child);

    final items = [
      const BottomNavigationBarItem(
        icon: Icon(Icons.inventory_2_outlined),
        activeIcon: Icon(Icons.inventory_2),
        label: 'Инвентарь',
      ),
      const BottomNavigationBarItem(
        icon: Icon(Icons.bar_chart_outlined),
        activeIcon: Icon(Icons.bar_chart),
        label: 'Статистика',
      ),
      const BottomNavigationBarItem(
        icon: Icon(Icons.chat_outlined),
        activeIcon: Icon(Icons.chat),
        label: 'Сообщения',
      ),
      const BottomNavigationBarItem(
        icon: Icon(Icons.add_circle_outline),
        activeIcon: Icon(Icons.add_circle),
        label: 'Добавить',
      ),
      const BottomNavigationBarItem(
        icon: Icon(Icons.receipt_long_outlined),
        activeIcon: Icon(Icons.receipt_long),
        label: 'Заказы',
      ),
      const BottomNavigationBarItem(
        icon: Icon(Icons.smart_toy_outlined),
        activeIcon: Icon(Icons.smart_toy),
        label: 'ИИ',
      ),
    ];

    return Scaffold(
      body: child,
      bottomNavigationBar: BottomNavigationBar(
        type: BottomNavigationBarType.fixed,
        currentIndex: currentIndex,
        items: items,
        onTap: (i) {
          switch (i) {
            case 0: context.go('/inventory');
            case 1: context.go('/statistics');
            case 2: context.go('/messages');
            case 3: _showAddMenu(context);
            case 4: context.go('/orders');
            case 5:
              showModalBottomSheet<void>(
                context: context,
                isScrollControlled: true,
                builder: (_) => const AIAssistantSheet(),
              );
          }
        },
      ),
    );
  }
}
