import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fl_chart/fl_chart.dart';
import '../../../core/api/api_client.dart';

final statisticsProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  final response = await apiClient.dio.get('/api/statistics');
  return response.data as Map<String, dynamic>;
});

class StatisticsScreen extends ConsumerWidget {
  const StatisticsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statsAsync = ref.watch(statisticsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Статистика'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(statisticsProvider),
          ),
        ],
      ),
      body: statsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (data) {
          final totalParts = data['total_parts'] as int? ?? 0;
          final totalValue = (data['total_value'] as num?)?.toDouble() ?? 0;
          final categoriesRaw = data['categories'];
          final Map<String, dynamic> categories = {};
          if (categoriesRaw is List) {
            for (final item in categoriesRaw) {
              if (item is Map) {
                final name = item['name']?.toString() ?? 'Неизвестно';
                final count = item['count'] ?? 0;
                categories[name] = count;
              }
            }
          } else if (categoriesRaw is Map) {
            categories.addAll(Map<String, dynamic>.from(categoriesRaw));
          }

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Карточки итогов
                Row(
                  children: [
                    Expanded(
                      child: _StatCard(
                        label: 'Всего запчастей',
                        value: totalParts.toString(),
                        icon: Icons.inventory_2_outlined,
                        color: const Color(0xFF4F8EF7),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: _StatCard(
                        label: 'Общая стоимость',
                        value: '${(totalValue / 1000).toStringAsFixed(0)}K ₽',
                        icon: Icons.monetization_on_outlined,
                        color: const Color(0xFF43A047),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),

                if (categories.isNotEmpty) ...[
                  Text(
                    'По категориям',
                    style: Theme.of(context)
                        .textTheme
                        .titleMedium
                        ?.copyWith(color: Colors.white),
                  ),
                  const SizedBox(height: 16),
                  SizedBox(
                    height: 220,
                    child: _CategoryPieChart(categories: categories),
                  ),
                  const SizedBox(height: 16),
                  _CategoryLegend(categories: categories),
                ],
              ],
            ),
          );
        },
      ),
    );
  }
}

class _StatCard extends StatelessWidget {
  final String label;
  final String value;
  final IconData icon;
  final Color color;

  const _StatCard({
    required this.label,
    required this.value,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: color, size: 28),
            const SizedBox(height: 8),
            Text(
              value,
              style: TextStyle(
                  color: color,
                  fontSize: 22,
                  fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 4),
            Text(
              label,
              style: const TextStyle(color: Colors.white54, fontSize: 12),
            ),
          ],
        ),
      ),
    );
  }
}

const _chartColors = [
  Color(0xFF4F8EF7),
  Color(0xFF43A047),
  Color(0xFFFDD835),
  Color(0xFFE53935),
  Color(0xFF795548),
  Color(0xFF9C27B0),
  Color(0xFF00BCD4),
  Color(0xFFFF9800),
];

class _CategoryPieChart extends StatelessWidget {
  final Map<String, dynamic> categories;
  const _CategoryPieChart({required this.categories});

  @override
  Widget build(BuildContext context) {
    final entries = categories.entries.toList();
    final total = entries.fold<double>(
        0, (sum, e) => sum + ((e.value as num?)?.toDouble() ?? 0));

    return PieChart(
      PieChartData(
        sections: List.generate(entries.length, (i) {
          final value = (entries[i].value as num?)?.toDouble() ?? 0;
          return PieChartSectionData(
            color: _chartColors[i % _chartColors.length],
            value: value,
            title: total > 0 ? '${(value / total * 100).toStringAsFixed(0)}%' : '',
            radius: 80,
            titleStyle: const TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.white),
          );
        }),
        sectionsSpace: 2,
        centerSpaceRadius: 30,
      ),
    );
  }
}

class _CategoryLegend extends StatelessWidget {
  final Map<String, dynamic> categories;
  const _CategoryLegend({required this.categories});

  @override
  Widget build(BuildContext context) {
    final entries = categories.entries.toList();
    return Column(
      children: List.generate(entries.length, (i) {
        return Padding(
          padding: const EdgeInsets.symmetric(vertical: 3),
          child: Row(
            children: [
              Container(
                width: 12,
                height: 12,
                decoration: BoxDecoration(
                  color: _chartColors[i % _chartColors.length],
                  shape: BoxShape.circle,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  entries[i].key,
                  style: const TextStyle(color: Colors.white70, fontSize: 13),
                ),
              ),
              Text(
                '${entries[i].value} шт.',
                style: const TextStyle(
                    color: Colors.white54, fontSize: 13),
              ),
            ],
          ),
        );
      }),
    );
  }
}
