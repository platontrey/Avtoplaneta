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

class PartCatalogBinding {
  final String source;
  final String target;
  final Set<String> categories;
  final Set<String> excludedCategories;
  final String? defaultValue;

  const PartCatalogBinding({
    required this.source,
    required this.target,
    this.categories = const {},
    this.excludedCategories = const {},
    this.defaultValue,
  });

  factory PartCatalogBinding.fromJson(Map<String, dynamic> json) =>
      PartCatalogBinding(
        source: json['source'] as String? ?? '',
        target: json['target'] as String? ?? '',
        categories: (json['categories'] as List<dynamic>? ?? const [])
            .map((value) => value.toString())
            .toSet(),
        excludedCategories:
            (json['excluded_categories'] as List<dynamic>? ?? const [])
                .map((value) => value.toString())
                .toSet(),
        defaultValue: json['default_value'] as String?,
      );
}

class PartCatalogTemplate {
  final String id;
  final String name;
  final String category;
  final int quantity;
  final double price;
  final Map<String, String> defaults;

  const PartCatalogTemplate({
    required this.id,
    required this.name,
    required this.category,
    required this.quantity,
    required this.price,
    this.defaults = const {},
  });

  factory PartCatalogTemplate.fromJson(Map<String, dynamic> json) =>
      PartCatalogTemplate(
        id: json['id'] as String? ?? '',
        name: json['name'] as String? ?? '',
        category: json['category'] as String? ?? '',
        quantity: (json['quantity'] as num?)?.toInt() ?? 0,
        price: (json['price'] as num?)?.toDouble() ?? 0,
        defaults: Map<String, dynamic>.from(
          json['defaults'] as Map? ?? const {},
        ).map((key, value) => MapEntry(key, value.toString())),
      );

  Map<String, dynamic> toPreviewMap() => {
    'id': id,
    'name': name,
    'category': category,
    'quantity': quantity,
    'price': price,
    ...defaults,
  };
}

class PartCatalog {
  final String version;
  final List<PartCatalogAttribute> attributes;
  final List<PartCatalogCategory> partFormCategories;
  final List<PartCatalogBinding> reportBindings;
  final List<PartCatalogTemplate> parts;

  const PartCatalog({
    required this.version,
    required this.attributes,
    required this.partFormCategories,
    required this.reportBindings,
    required this.parts,
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
    reportBindings: (json['report_bindings'] as List<dynamic>? ?? const [])
        .map(
          (value) => PartCatalogBinding.fromJson(
            Map<String, dynamic>.from(value as Map),
          ),
        )
        .toList(),
    parts: (json['parts'] as List<dynamic>? ?? const [])
        .map(
          (value) => PartCatalogTemplate.fromJson(
            Map<String, dynamic>.from(value as Map),
          ),
        )
        .toList(),
  );

  Set<String> categoriesForBinding(String source) {
    for (final binding in reportBindings) {
      if (binding.source == source) {
        return binding.categories;
      }
    }
    return const {};
  }

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
