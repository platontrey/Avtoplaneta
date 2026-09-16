import '../../../core/api/api_client.dart';

class DefectReportPreview {
  final String catalogVersion;
  final int total;
  final List<Map<String, dynamic>> parts;

  const DefectReportPreview({
    required this.catalogVersion,
    required this.total,
    required this.parts,
  });

  factory DefectReportPreview.fromJson(Map<String, dynamic> json) =>
      DefectReportPreview(
        catalogVersion: json['catalog_version'] as String? ?? '',
        total: json['total'] as int? ?? 0,
        parts: (json['parts'] as List<dynamic>? ?? const [])
            .map((part) => Map<String, dynamic>.from(part as Map))
            .toList(),
      );
}

class DefectReportApi {
  const DefectReportApi._();

  static Future<DefectReportPreview> preview(
    Map<String, dynamic> payload,
  ) async {
    final response = await apiClient.dio.post<Map<String, dynamic>>(
      '/api/defect-reports/preview',
      data: payload,
    );
    return DefectReportPreview.fromJson(response.data ?? const {});
  }

  static Future<void> create(Map<String, dynamic> payload) async {
    await apiClient.dio.post('/api/defect-reports', data: payload);
  }
}
