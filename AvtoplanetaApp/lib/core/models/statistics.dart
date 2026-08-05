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
  final List<CategoryCount> categories;
  final List<MonthlySales> monthlySales;

  const StatisticsData({
    required this.totalParts,
    required this.totalQuantity,
    required this.totalValue,
    required this.totalEarnings,
    required this.categories,
    required this.monthlySales,
  });

  factory StatisticsData.fromJson(Map<String, dynamic> json) {
    final categoriesRaw = json['categories'];
    final monthlySalesRaw = json['monthlySales'] ?? json['monthly_sales'];

    return StatisticsData(
      totalParts: _toInt(json['totalParts'] ?? json['total_parts']),
      totalQuantity: _toInt(json['totalQuantity'] ?? json['total_quantity']),
      totalValue: _toDouble(json['totalValue'] ?? json['total_value']),
      totalEarnings:
          _toDouble(json['totalEarnings'] ?? json['total_earnings']),
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
