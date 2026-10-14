class CategoryCount {
  final String name;
  final int count;

  const CategoryCount({required this.name, required this.count});

  factory CategoryCount.fromJson(Map<String, dynamic> json) => CategoryCount(
        name: json['name']?.toString() ?? 'Неизвестно',
        count: _toInt(json['count']),
      );
}

class MonthlySales {
  final String month;
  final double sales;

  const MonthlySales({required this.month, required this.sales});

  factory MonthlySales.fromJson(Map<String, dynamic> json) => MonthlySales(
        month: json['month']?.toString() ?? '',
        sales: _toDouble(json['sales']),
      );
}

class StatisticsData {
  final int totalParts;
  final int totalQuantity;
  final double totalValue;
  final double totalEarnings;
  final double? partsGrowth;
  final double? quantityGrowth;
  final double? valueGrowth;
  final double? earningsGrowth;
  final List<CategoryCount> categories;
  final List<MonthlySales> monthlySales;

  const StatisticsData({
    required this.totalParts,
    required this.totalQuantity,
    required this.totalValue,
    required this.totalEarnings,
    this.partsGrowth,
    this.quantityGrowth,
    this.valueGrowth,
    this.earningsGrowth,
    required this.categories,
    required this.monthlySales,
  });

  double get effectiveEarningsGrowth {
    if (earningsGrowth != null && earningsGrowth!.abs() >= 0.05) {
      return earningsGrowth!;
    }
    if (monthlySales.length >= 2) {
      final sorted = List<MonthlySales>.from(monthlySales)
        ..sort((a, b) => a.month.compareTo(b.month));
      final last = sorted[sorted.length - 1].sales;
      final prev = sorted[sorted.length - 2].sales;
      if (prev > 0) {
        return double.parse((((last - prev) / prev) * 100).toStringAsFixed(1));
      } else if (last > 0) {
        return 100.0;
      }
    }
    return 0.0;
  }

  double get effectivePartsGrowth =>
      (partsGrowth != null && partsGrowth!.abs() >= 0.05)
          ? partsGrowth!
          : effectiveEarningsGrowth;

  double get effectiveQuantityGrowth =>
      (quantityGrowth != null && quantityGrowth!.abs() >= 0.05)
          ? quantityGrowth!
          : effectiveEarningsGrowth;

  double get effectiveValueGrowth =>
      (valueGrowth != null && valueGrowth!.abs() >= 0.05)
          ? valueGrowth!
          : effectiveEarningsGrowth;

  factory StatisticsData.fromJson(Map<String, dynamic> json) {
    final categoriesRaw = json['categories'];
    final monthlySalesRaw = json['monthlySales'] ?? json['monthly_sales'];

    return StatisticsData(
      totalParts: _toInt(json['totalParts'] ?? json['total_parts']),
      totalQuantity: _toInt(json['totalQuantity'] ?? json['total_quantity']),
      totalValue: _toDouble(json['totalValue'] ?? json['total_value']),
      totalEarnings:
          _toDouble(json['totalEarnings'] ?? json['total_earnings']),
      partsGrowth:
          _toDoubleNullable(json['partsGrowth'] ?? json['parts_growth']),
      quantityGrowth:
          _toDoubleNullable(json['quantityGrowth'] ?? json['quantity_growth']),
      valueGrowth:
          _toDoubleNullable(json['valueGrowth'] ?? json['value_growth']),
      earningsGrowth:
          _toDoubleNullable(json['earningsGrowth'] ?? json['earnings_growth']),
      categories: categoriesRaw is List
          ? categoriesRaw
              .whereType<Map>()
              .map((item) => CategoryCount.fromJson(
                    Map<String, dynamic>.from(item),
                  ))
              .toList()
          : const [],
      monthlySales: monthlySalesRaw is List
          ? monthlySalesRaw
              .whereType<Map>()
              .map((item) => MonthlySales.fromJson(
                    Map<String, dynamic>.from(item),
                  ))
              .toList()
          : const [],
    );
  }
}

int _toInt(dynamic value) => value is num
    ? value.toInt()
    : int.tryParse(value?.toString() ?? '') ?? 0;

double _toDouble(dynamic value) => value is num
    ? value.toDouble()
    : double.tryParse(value?.toString() ?? '') ?? 0;

double? _toDoubleNullable(dynamic value) {
  if (value == null) return null;
  if (value is num) return value.toDouble();
  return double.tryParse(value.toString());
}
