class PartCatalogAttribute {
  final String code;
  final String label;
  final String inputType;
  final List<String> options;

  const PartCatalogAttribute({
    required this.code,
    required this.label,
    required this.inputType,
    this.options = const [],
  });

  factory PartCatalogAttribute.fromJson(Map<String, dynamic> json) =>
      PartCatalogAttribute(
        code: json['code'] as String? ?? '',
        label: json['label'] as String? ?? '',
        inputType: json['input_type'] as String? ?? 'text',
        options: (json['options'] as List<dynamic>? ?? const [])
            .map((value) => value.toString())
            .toList(),
      );
}

class PartCatalogCategory {
  final String code;
  final String name;
  final Set<String> attributes;

  const PartCatalogCategory({
    required this.code,
    required this.name,
    required this.attributes,
  });

  factory PartCatalogCategory.fromJson(Map<String, dynamic> json) =>
      PartCatalogCategory(
        code: json['code'] as String? ?? '',
        name: json['name'] as String? ?? '',
        attributes: (json['attributes'] as List<dynamic>? ?? const [])
            .map((value) => value.toString())
            .toSet(),
      );
}

class PartCatalog {
  final String version;
  final List<PartCatalogAttribute> attributes;
  final List<PartCatalogCategory> partFormCategories;

  const PartCatalog({
    required this.version,
    required this.attributes,
    required this.partFormCategories,
  });

  factory PartCatalog.fromJson(Map<String, dynamic> json) => PartCatalog(
    version: json['version'] as String? ?? '',
    attributes: (json['attributes'] as List<dynamic>? ?? const [])
        .map(
          (value) => PartCatalogAttribute.fromJson(
            Map<String, dynamic>.from(value as Map),
          ),
        )
        .toList(),
    partFormCategories:
        (json['part_form_categories'] as List<dynamic>? ?? const [])
            .map(
              (value) => PartCatalogCategory.fromJson(
                Map<String, dynamic>.from(value as Map),
              ),
            )
            .toList(),
  );

  Set<String> attributesForCategory(String category) {
    for (final item in partFormCategories) {
      if (item.name == category) {
        return item.attributes;
      }
    }
    return const {};
  }

  List<String> optionsForAttribute(String code) {
    for (final attribute in attributes) {
      if (attribute.code == code) {
        return attribute.options;
      }
    }
    return const [];
  }
}
