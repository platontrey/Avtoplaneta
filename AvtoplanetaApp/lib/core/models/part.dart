class Part {
  final int id;
  final String name;
  final String? description;
  final String category;
  final double price;
  final int quantity;
  final bool? status;
  final String? brand;
  final String? model;
  final String? bodyBrand;
  final String? engineBrand;
  final String? carReleaseDate;
  final String? frontRear;
  final String? leftRight;
  final String? topBottom;
  final String? number;
  final String? manufacturer;
  final String? manufacturerCode;
  final String? color;
  final String? condition;
  final String? oemCode;
  final String? supplierCode;
  final String? transmissionModel;
  final String? defect;
  final String? transmission;
  final String? drive;
  final String? wearPercentage;
  final String? season;
  final String? diameter;
  final String? width;
  final String? profile;
  final String? tireQuantity;
  final String? drilling;
  final String? offset;
  final String? centerHoleDiameter;
  final String? tireModel;
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
    this.status,
    this.brand,
    this.model,
    this.bodyBrand,
    this.engineBrand,
    this.carReleaseDate,
    this.frontRear,
    this.leftRight,
    this.topBottom,
    this.number,
    this.manufacturer,
    this.manufacturerCode,
    this.color,
    this.condition,
    this.oemCode,
    this.supplierCode,
    this.transmissionModel,
    this.defect,
    this.transmission,
    this.drive,
    this.wearPercentage,
    this.season,
    this.diameter,
    this.width,
    this.profile,
    this.tireQuantity,
    this.drilling,
    this.offset,
    this.centerHoleDiameter,
    this.tireModel,
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
      photosList = photosRaw
          .map((e) => e.toString())
          .where((value) => value.isNotEmpty)
          .toList();
    }
    final legacyPhoto = json['photo']?.toString();
    if (photosList.isEmpty && legacyPhoto != null && legacyPhoto.isNotEmpty) {
      photosList = [legacyPhoto];
    }

    return Part(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: json['name'] as String? ?? '',
      description: json['description'] as String?,
      category: json['category'] as String? ?? '',
      price: (json['price'] as num?)?.toDouble() ?? 0.0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      status: json['status'] as bool?,
      brand: json['brand'] as String?,
      model: json['model'] as String?,
      bodyBrand: json['body_brand'] as String?,
      engineBrand: json['engine_brand'] as String?,
      carReleaseDate: json['car_release_date'] as String?,
      frontRear: json['front_rear'] as String?,
      leftRight: json['left_right'] as String?,
      topBottom: json['top_bottom'] as String?,
      number: json['number'] as String?,
      manufacturer: json['manufacturer'] as String?,
      manufacturerCode: json['manufacturer_code'] as String?,
      color: json['color'] as String?,
      condition: json['condition'] as String?,
      oemCode: (json['oem_code'] ?? json['oemCode']) as String?,
      supplierCode: (json['supplier_code'] ?? json['supplierCode']) as String?,
      transmissionModel:
          (json['transmission_model'] ?? json['transmissionModel']) as String?,
      defect: json['defect'] as String?,
      transmission: json['transmission'] as String?,
      drive: json['drive'] as String?,
      wearPercentage: json['wear_percentage'] as String?,
      season: json['season'] as String?,
      diameter: json['diameter'] as String?,
      width: json['width'] as String?,
      profile: json['profile'] as String?,
      tireQuantity: json['tire_quantity'] as String?,
      drilling: json['drilling'] as String?,
      offset: json['offset'] as String?,
      centerHoleDiameter: json['center_hole_diameter'] as String?,
      tireModel: json['tire_model'] as String?,
      vin: json['vin'] as String?,
      location: json['location'] as String? ?? '',
      salesman: json['salesman'] as String?,
      photos: photosList,
      createdAt: (json['created_at'] ?? json['createdAt']) != null
          ? DateTime.tryParse(
              (json['created_at'] ?? json['createdAt']).toString(),
            )
          : null,
      markedForDeletion:
          (json['marked_for_deletion'] ?? json['markedForDeletion']) as bool? ??
          false,
    );
  }

  Map<String, dynamic> toJson() => {
    'name': name,
    'description': description,
    'category': category,
    'price': price,
    'quantity': quantity,
    'status': status,
    'brand': brand,
    'model': model,
    'body_brand': bodyBrand,
    'engine_brand': engineBrand,
    'car_release_date': carReleaseDate,
    'front_rear': frontRear,
    'left_right': leftRight,
    'top_bottom': topBottom,
    'number': number,
    'manufacturer': manufacturer,
    'manufacturer_code': manufacturerCode,
    'color': color,
    'condition': condition,
    'oem_code': oemCode,
    'supplier_code': supplierCode,
    'transmission_model': transmissionModel,
    'defect': defect,
    'transmission': transmission,
    'drive': drive,
    'wear_percentage': wearPercentage,
    'season': season,
    'diameter': diameter,
    'width': width,
    'profile': profile,
    'tire_quantity': tireQuantity,
    'drilling': drilling,
    'offset': offset,
    'center_hole_diameter': centerHoleDiameter,
    'tire_model': tireModel,
    'vin': vin,
    'location': location,
    'salesman': salesman,
  };

  bool get isAvailable => status ?? quantity > 0;
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
      total: (json['total'] as num?)?.toInt() ?? 0,
      page: (json['page'] as num?)?.toInt() ?? 1,
      limit: (json['limit'] as num?)?.toInt() ?? 20,
      isOffline: json['is_offline'] as bool? ?? false,
    );
  }
}
