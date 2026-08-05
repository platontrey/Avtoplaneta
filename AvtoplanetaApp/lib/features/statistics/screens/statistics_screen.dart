import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:intl/intl.dart';
import '../../../app/theme.dart';
import '../../../core/api/api_client.dart';
import '../../../core/models/statistics.dart';
import '../../../shared/widgets/app_states.dart';

final statisticsProvider = FutureProvider<StatisticsData>((ref) async {
  final response = await apiClient.dio.get('/api/v1/statistics');
  return StatisticsData.fromJson(
    Map<String, dynamic>.from(response.data as Map),
  );
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
            tooltip: 'Обновить',
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(statisticsProvider),
          ),
        ],
      ),
      body: statsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => AppEmptyState(
          icon: Icons.cloud_off_rounded,
          title: 'Не удалось загрузить статистику',
          message: 'Проверьте соединение и повторите попытку.',
          actionLabel: 'Повторить',
          onAction: () => ref.invalidate(statisticsProvider),
        ),
        data: (data) {
          final categories = {
            for (final category in data.categories)
              category.name: category.count,
          };
          final currency = NumberFormat.currency(
            locale: 'ru_RU',
            symbol: '₽',
            decimalDigits: 2,
          );

          return SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const AppSectionHeader(
                  title: 'Обзор склада',
                  caption: 'Актуальные данные',
                ),
                const SizedBox(height: 14),
                GridView.count(
                  crossAxisCount: MediaQuery.sizeOf(context).width >= 600 ? 4 : 2,
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  crossAxisSpacing: 10,
                  mainAxisSpacing: 10,
                  childAspectRatio: MediaQuery.sizeOf(context).width >= 600
                      ? 1.15
                      : 1.08,
                  children: [
                    _StatCard(
                      label: 'Всего позиций',
                      value: data.totalParts.toString(),
                      icon: Icons.inventory_2_outlined,
                      color: AppTheme.primaryColor,
                    ),
                    _StatCard(
                      label: 'Общее количество',
                      value: data.totalQuantity.toString(),
                      icon: Icons.inventory_outlined,
                      color: const Color(0xFF7C3AED),
                    ),
                    _StatCard(
                      label: 'Общая стоимость',
                      value: currency.format(data.totalValue),
                      icon: Icons.monetization_on_outlined,
                      color: const Color(0xFFEA580C),
                    ),
                    _StatCard(
                      label: 'Общий заработок',
                      value: currency.format(data.totalEarnings),
                      icon: Icons.trending_up,
                      color: AppTheme.successColor,
                    ),
                  ],
                ),
                const SizedBox(height: 24),

                if (categories.isNotEmpty) ...[
                  const AppSectionHeader(
                    title: 'По категориям',
                    caption: 'Доля позиций',
                  ),
                  const SizedBox(height: 14),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        children: [
                          SizedBox(
                            height: 210,
                            child: _CategoryPieChart(categories: categories),
                          ),
                          const SizedBox(height: 12),
                          _CategoryLegend(categories: categories),
                        ],
                      ),
                    ),
                  ),
                ],
                if (data.monthlySales.isNotEmpty) ...[
                  const SizedBox(height: 24),
                  const AppSectionHeader(
                    title: 'Продажи по месяцам',
                    caption: 'Динамика',
                  ),
                  const SizedBox(height: 14),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(10, 20, 18, 12),
                      child: SizedBox(
                        height: 230,
                        child: _MonthlySalesChart(data: data.monthlySales),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          );
        },
      ),
    );
  }
}

class _MonthlySalesChart extends StatelessWidget {
  final List<MonthlySales> data;
  const _MonthlySalesChart({required this.data});

  @override
  Widget build(BuildContext context) {
    final maxSales = data.fold<double>(
      0,
      (max, item) => item.sales > max ? item.sales : max,
    );

    return BarChart(
      BarChartData(
        maxY: maxSales > 0 ? maxSales * 1.15 : 1,
        borderData: FlBorderData(show: false),
        gridData: const FlGridData(show: false),
        barTouchData: BarTouchData(enabled: true),
        titlesData: FlTitlesData(
          topTitles: const AxisTitles(
            sideTitles: SideTitles(showTitles: false),
          ),
          rightTitles: const AxisTitles(
            sideTitles: SideTitles(showTitles: false),
          ),
          leftTitles: const AxisTitles(
            sideTitles: SideTitles(showTitles: true, reservedSize: 44),
          ),
          bottomTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              getTitlesWidget: (value, meta) {
                final index = value.toInt();
                if (index < 0 || index >= data.length) {
                  return const SizedBox.shrink();
                }
                return Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: Text(
                    data[index].month,
                    style: const TextStyle(
                      color: Colors.white54,
                      fontSize: 10,
                    ),
                  ),
                );
              },
            ),
          ),
        ),
        barGroups: List.generate(
          data.length,
          (index) => BarChartGroupData(
            x: index,
            barRods: [
              BarChartRodData(
                toY: data[index].sales,
                color: AppTheme.successColor,
                width: 16,
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(4),
                ),
              ),
            ],
          ),
        ),
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
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 38,
              height: 38,
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: color, size: 21),
            ),
            const Spacer(),
            FittedBox(
              fit: BoxFit.scaleDown,
              alignment: Alignment.centerLeft,
              child: Text(
                value,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 20,
                  fontWeight: FontWeight.w800,
                  letterSpacing: -0.4,
                ),
              ),
            ),
            const SizedBox(height: 5),
            Text(
              label,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(color: AppTheme.mutedColor, fontSize: 11),
            ),
          ],
        ),
      ),
    );
  }
}

const _chartColors = [
  AppTheme.primaryColor,
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
                  style: const TextStyle(
                    color: Color(0xFFCBD5E1),
                    fontSize: 13,
                  ),
                ),
              ),
              Text(
                '${entries[i].value} шт.',
                style: const TextStyle(
                    color: AppTheme.mutedColor, fontSize: 13),
              ),
            ],
          ),
        );
      }),
    );
  }
}
