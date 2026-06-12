class Part {
  final int id;
  final String name;
  final String? description;
  final String category;
  final double price;
  final int quantity;
  final String? brand;
  final String? model;
  final String? color;
  final String? condition;
  final String? oemCode;
  final String? supplierCode;
  final String? vin;
  final String location;
  final String? salesman;
  final List<String> photos;
  final DateTime? createdAt;
  final bool markedForDeletion;

  const Part({
    required this.id,
    required this.name,
    this.description,
    required this.category,
    required this.price,
    required this.quantity,
    this.brand,
    this.model,
    this.color,
    this.condition,
    this.oemCode,
    this.supplierCode,
    this.vin,
    this.location = '',
    this.salesman,
    this.photos = const [],
    this.createdAt,
    this.markedForDeletion = false,
  });

  factory Part.fromJson(Map<String, dynamic> json) {
    List<String> photosList = [];
    final photosRaw = json['photos'];
    if (photosRaw is List) {
      photosList = photosRaw.map((e) => e.toString()).toList();
    }

    return Part(
      id: json['id'] as int,
      name: json['name'] as String? ?? '',
      description: json['description'] as String?,
      category: json['category'] as String? ?? '',
      price: (json['price'] as num?)?.toDouble() ?? 0.0,
      quantity: json['quantity'] as int? ?? 0,
      brand: json['brand'] as String?,
      model: json['model'] as String?,
      color: json['color'] as String?,
      condition: json['condition'] as String?,
      oemCode: json['oem_code'] as String?,
      supplierCode: json['supplier_code'] as String?,
      vin: json['vin'] as String?,
      location: json['location'] as String? ?? '',
      salesman: json['salesman'] as String?,
      photos: photosList,
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'] as String)
          : null,
      markedForDeletion: json['marked_for_deletion'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() => {
        'name': name,
        'description': description,
        'category': category,
        'price': price,
        'quantity': quantity,
        'brand': brand,
        'model': model,
        'color': color,
        'condition': condition,
        'oem_code': oemCode,
        'supplier_code': supplierCode,
        'vin': vin,
        'location': location,
        'salesman': salesman,
      };
}

class InventoryResponse {
  final List<Part> parts;
  final int total;
  final int page;
  final int limit;
  final bool isOffline;

  const InventoryResponse({
    required this.parts,
    required this.total,
    required this.page,
    required this.limit,
    this.isOffline = false,
  });

  factory InventoryResponse.fromJson(Map<String, dynamic> json) {
    final partsRaw = json['parts'] as List<dynamic>? ?? [];
    return InventoryResponse(
      parts: partsRaw
          .map((e) => Part.fromJson(e as Map<String, dynamic>))
          .toList(),
      total: json['total'] as int? ?? 0,
      page: json['page'] as int? ?? 1,
      limit: json['limit'] as int? ?? 20,
      isOffline: json['is_offline'] as bool? ?? false,
    );
  }
}
