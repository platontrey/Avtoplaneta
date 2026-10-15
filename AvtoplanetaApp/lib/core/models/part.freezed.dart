// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'part.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Part {

 int get id; String get name; String? get description; String get category; double get price; int get quantity; bool? get status; String? get brand; String? get model;@JsonKey(readValue: _readBodyBrand) String? get bodyBrand;@JsonKey(readValue: _readEngineBrand) String? get engineBrand;@JsonKey(readValue: _readCarReleaseDate) String? get carReleaseDate;@JsonKey(readValue: _readFrontRear) String? get frontRear;@JsonKey(readValue: _readLeftRight) String? get leftRight;@JsonKey(readValue: _readTopBottom) String? get topBottom; String? get number; String? get manufacturer;@JsonKey(readValue: _readManufacturerCode) String? get manufacturerCode; String? get color; String? get condition;@JsonKey(readValue: _readOemCode) String? get oemCode;@JsonKey(readValue: _readSupplierCode) String? get supplierCode;@JsonKey(readValue: _readTransmissionModel) String? get transmissionModel; String? get defect; String? get transmission; String? get drive;@JsonKey(readValue: _readWearPercentage) String? get wearPercentage; String? get season; String? get diameter; String? get width; String? get profile;@JsonKey(readValue: _readTireQuantity) String? get tireQuantity; String? get drilling; String? get offset;@JsonKey(readValue: _readCenterHoleDiameter) String? get centerHoleDiameter;@JsonKey(readValue: _readTireModel) String? get tireModel; String? get vin;@JsonKey(readValue: _readCarReleasePeriod) String? get carReleasePeriod; String get location; String get address; String? get salesman; List<String> get photos;@JsonKey(readValue: _readCreatedAt) DateTime? get createdAt;@JsonKey(readValue: _readMarkedForDeletion) bool get markedForDeletion;
/// Create a copy of Part
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PartCopyWith<Part> get copyWith => _$PartCopyWithImpl<Part>(this as Part, _$identity);

  /// Serializes this Part to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Part&&(identical(other.id, id) || other.id == id)&&(identical(other.name, name) || other.name == name)&&(identical(other.description, description) || other.description == description)&&(identical(other.category, category) || other.category == category)&&(identical(other.price, price) || other.price == price)&&(identical(other.quantity, quantity) || other.quantity == quantity)&&(identical(other.status, status) || other.status == status)&&(identical(other.brand, brand) || other.brand == brand)&&(identical(other.model, model) || other.model == model)&&(identical(other.bodyBrand, bodyBrand) || other.bodyBrand == bodyBrand)&&(identical(other.engineBrand, engineBrand) || other.engineBrand == engineBrand)&&(identical(other.carReleaseDate, carReleaseDate) || other.carReleaseDate == carReleaseDate)&&(identical(other.frontRear, frontRear) || other.frontRear == frontRear)&&(identical(other.leftRight, leftRight) || other.leftRight == leftRight)&&(identical(other.topBottom, topBottom) || other.topBottom == topBottom)&&(identical(other.number, number) || other.number == number)&&(identical(other.manufacturer, manufacturer) || other.manufacturer == manufacturer)&&(identical(other.manufacturerCode, manufacturerCode) || other.manufacturerCode == manufacturerCode)&&(identical(other.color, color) || other.color == color)&&(identical(other.condition, condition) || other.condition == condition)&&(identical(other.oemCode, oemCode) || other.oemCode == oemCode)&&(identical(other.supplierCode, supplierCode) || other.supplierCode == supplierCode)&&(identical(other.transmissionModel, transmissionModel) || other.transmissionModel == transmissionModel)&&(identical(other.defect, defect) || other.defect == defect)&&(identical(other.transmission, transmission) || other.transmission == transmission)&&(identical(other.drive, drive) || other.drive == drive)&&(identical(other.wearPercentage, wearPercentage) || other.wearPercentage == wearPercentage)&&(identical(other.season, season) || other.season == season)&&(identical(other.diameter, diameter) || other.diameter == diameter)&&(identical(other.width, width) || other.width == width)&&(identical(other.profile, profile) || other.profile == profile)&&(identical(other.tireQuantity, tireQuantity) || other.tireQuantity == tireQuantity)&&(identical(other.drilling, drilling) || other.drilling == drilling)&&(identical(other.offset, offset) || other.offset == offset)&&(identical(other.centerHoleDiameter, centerHoleDiameter) || other.centerHoleDiameter == centerHoleDiameter)&&(identical(other.tireModel, tireModel) || other.tireModel == tireModel)&&(identical(other.vin, vin) || other.vin == vin)&&(identical(other.carReleasePeriod, carReleasePeriod) || other.carReleasePeriod == carReleasePeriod)&&(identical(other.location, location) || other.location == location)&&(identical(other.address, address) || other.address == address)&&(identical(other.salesman, salesman) || other.salesman == salesman)&&const DeepCollectionEquality().equals(other.photos, photos)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.markedForDeletion, markedForDeletion) || other.markedForDeletion == markedForDeletion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hashAll([runtimeType,id,name,description,category,price,quantity,status,brand,model,bodyBrand,engineBrand,carReleaseDate,frontRear,leftRight,topBottom,number,manufacturer,manufacturerCode,color,condition,oemCode,supplierCode,transmissionModel,defect,transmission,drive,wearPercentage,season,diameter,width,profile,tireQuantity,drilling,offset,centerHoleDiameter,tireModel,vin,carReleasePeriod,location,address,salesman,const DeepCollectionEquality().hash(photos),createdAt,markedForDeletion]);

@override
String toString() {
  return 'Part(id: $id, name: $name, description: $description, category: $category, price: $price, quantity: $quantity, status: $status, brand: $brand, model: $model, bodyBrand: $bodyBrand, engineBrand: $engineBrand, carReleaseDate: $carReleaseDate, frontRear: $frontRear, leftRight: $leftRight, topBottom: $topBottom, number: $number, manufacturer: $manufacturer, manufacturerCode: $manufacturerCode, color: $color, condition: $condition, oemCode: $oemCode, supplierCode: $supplierCode, transmissionModel: $transmissionModel, defect: $defect, transmission: $transmission, drive: $drive, wearPercentage: $wearPercentage, season: $season, diameter: $diameter, width: $width, profile: $profile, tireQuantity: $tireQuantity, drilling: $drilling, offset: $offset, centerHoleDiameter: $centerHoleDiameter, tireModel: $tireModel, vin: $vin, carReleasePeriod: $carReleasePeriod, location: $location, address: $address, salesman: $salesman, photos: $photos, createdAt: $createdAt, markedForDeletion: $markedForDeletion)';
}


}

/// @nodoc
abstract mixin class $PartCopyWith<$Res>  {
  factory $PartCopyWith(Part value, $Res Function(Part) _then) = _$PartCopyWithImpl;
@useResult
$Res call({
 int id, String name, String? description, String category, double price, int quantity, bool? status, String? brand, String? model,@JsonKey(readValue: _readBodyBrand) String? bodyBrand,@JsonKey(readValue: _readEngineBrand) String? engineBrand,@JsonKey(readValue: _readCarReleaseDate) String? carReleaseDate,@JsonKey(readValue: _readFrontRear) String? frontRear,@JsonKey(readValue: _readLeftRight) String? leftRight,@JsonKey(readValue: _readTopBottom) String? topBottom, String? number, String? manufacturer,@JsonKey(readValue: _readManufacturerCode) String? manufacturerCode, String? color, String? condition,@JsonKey(readValue: _readOemCode) String? oemCode,@JsonKey(readValue: _readSupplierCode) String? supplierCode,@JsonKey(readValue: _readTransmissionModel) String? transmissionModel, String? defect, String? transmission, String? drive,@JsonKey(readValue: _readWearPercentage) String? wearPercentage, String? season, String? diameter, String? width, String? profile,@JsonKey(readValue: _readTireQuantity) String? tireQuantity, String? drilling, String? offset,@JsonKey(readValue: _readCenterHoleDiameter) String? centerHoleDiameter,@JsonKey(readValue: _readTireModel) String? tireModel, String? vin,@JsonKey(readValue: _readCarReleasePeriod) String? carReleasePeriod, String location, String address, String? salesman, List<String> photos,@JsonKey(readValue: _readCreatedAt) DateTime? createdAt,@JsonKey(readValue: _readMarkedForDeletion) bool markedForDeletion
});




}
/// @nodoc
class _$PartCopyWithImpl<$Res>
    implements $PartCopyWith<$Res> {
  _$PartCopyWithImpl(this._self, this._then);

  final Part _self;
  final $Res Function(Part) _then;

/// Create a copy of Part
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? name = null,Object? description = freezed,Object? category = null,Object? price = null,Object? quantity = null,Object? status = freezed,Object? brand = freezed,Object? model = freezed,Object? bodyBrand = freezed,Object? engineBrand = freezed,Object? carReleaseDate = freezed,Object? frontRear = freezed,Object? leftRight = freezed,Object? topBottom = freezed,Object? number = freezed,Object? manufacturer = freezed,Object? manufacturerCode = freezed,Object? color = freezed,Object? condition = freezed,Object? oemCode = freezed,Object? supplierCode = freezed,Object? transmissionModel = freezed,Object? defect = freezed,Object? transmission = freezed,Object? drive = freezed,Object? wearPercentage = freezed,Object? season = freezed,Object? diameter = freezed,Object? width = freezed,Object? profile = freezed,Object? tireQuantity = freezed,Object? drilling = freezed,Object? offset = freezed,Object? centerHoleDiameter = freezed,Object? tireModel = freezed,Object? vin = freezed,Object? carReleasePeriod = freezed,Object? location = null,Object? address = null,Object? salesman = freezed,Object? photos = null,Object? createdAt = freezed,Object? markedForDeletion = null,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as int,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String?,category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as String,price: null == price ? _self.price : price // ignore: cast_nullable_to_non_nullable
as double,quantity: null == quantity ? _self.quantity : quantity // ignore: cast_nullable_to_non_nullable
as int,status: freezed == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as bool?,brand: freezed == brand ? _self.brand : brand // ignore: cast_nullable_to_non_nullable
as String?,model: freezed == model ? _self.model : model // ignore: cast_nullable_to_non_nullable
as String?,bodyBrand: freezed == bodyBrand ? _self.bodyBrand : bodyBrand // ignore: cast_nullable_to_non_nullable
as String?,engineBrand: freezed == engineBrand ? _self.engineBrand : engineBrand // ignore: cast_nullable_to_non_nullable
as String?,carReleaseDate: freezed == carReleaseDate ? _self.carReleaseDate : carReleaseDate // ignore: cast_nullable_to_non_nullable
as String?,frontRear: freezed == frontRear ? _self.frontRear : frontRear // ignore: cast_nullable_to_non_nullable
as String?,leftRight: freezed == leftRight ? _self.leftRight : leftRight // ignore: cast_nullable_to_non_nullable
as String?,topBottom: freezed == topBottom ? _self.topBottom : topBottom // ignore: cast_nullable_to_non_nullable
as String?,number: freezed == number ? _self.number : number // ignore: cast_nullable_to_non_nullable
as String?,manufacturer: freezed == manufacturer ? _self.manufacturer : manufacturer // ignore: cast_nullable_to_non_nullable
as String?,manufacturerCode: freezed == manufacturerCode ? _self.manufacturerCode : manufacturerCode // ignore: cast_nullable_to_non_nullable
as String?,color: freezed == color ? _self.color : color // ignore: cast_nullable_to_non_nullable
as String?,condition: freezed == condition ? _self.condition : condition // ignore: cast_nullable_to_non_nullable
as String?,oemCode: freezed == oemCode ? _self.oemCode : oemCode // ignore: cast_nullable_to_non_nullable
as String?,supplierCode: freezed == supplierCode ? _self.supplierCode : supplierCode // ignore: cast_nullable_to_non_nullable
as String?,transmissionModel: freezed == transmissionModel ? _self.transmissionModel : transmissionModel // ignore: cast_nullable_to_non_nullable
as String?,defect: freezed == defect ? _self.defect : defect // ignore: cast_nullable_to_non_nullable
as String?,transmission: freezed == transmission ? _self.transmission : transmission // ignore: cast_nullable_to_non_nullable
as String?,drive: freezed == drive ? _self.drive : drive // ignore: cast_nullable_to_non_nullable
as String?,wearPercentage: freezed == wearPercentage ? _self.wearPercentage : wearPercentage // ignore: cast_nullable_to_non_nullable
as String?,season: freezed == season ? _self.season : season // ignore: cast_nullable_to_non_nullable
as String?,diameter: freezed == diameter ? _self.diameter : diameter // ignore: cast_nullable_to_non_nullable
as String?,width: freezed == width ? _self.width : width // ignore: cast_nullable_to_non_nullable
as String?,profile: freezed == profile ? _self.profile : profile // ignore: cast_nullable_to_non_nullable
as String?,tireQuantity: freezed == tireQuantity ? _self.tireQuantity : tireQuantity // ignore: cast_nullable_to_non_nullable
as String?,drilling: freezed == drilling ? _self.drilling : drilling // ignore: cast_nullable_to_non_nullable
as String?,offset: freezed == offset ? _self.offset : offset // ignore: cast_nullable_to_non_nullable
as String?,centerHoleDiameter: freezed == centerHoleDiameter ? _self.centerHoleDiameter : centerHoleDiameter // ignore: cast_nullable_to_non_nullable
as String?,tireModel: freezed == tireModel ? _self.tireModel : tireModel // ignore: cast_nullable_to_non_nullable
as String?,vin: freezed == vin ? _self.vin : vin // ignore: cast_nullable_to_non_nullable
as String?,carReleasePeriod: freezed == carReleasePeriod ? _self.carReleasePeriod : carReleasePeriod // ignore: cast_nullable_to_non_nullable
as String?,location: null == location ? _self.location : location // ignore: cast_nullable_to_non_nullable
as String,address: null == address ? _self.address : address // ignore: cast_nullable_to_non_nullable
as String,salesman: freezed == salesman ? _self.salesman : salesman // ignore: cast_nullable_to_non_nullable
as String?,photos: null == photos ? _self.photos : photos // ignore: cast_nullable_to_non_nullable
as List<String>,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,markedForDeletion: null == markedForDeletion ? _self.markedForDeletion : markedForDeletion // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [Part].
extension PartPatterns on Part {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Part value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Part() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Part value)  $default,){
final _that = this;
switch (_that) {
case _Part():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Part value)?  $default,){
final _that = this;
switch (_that) {
case _Part() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int id,  String name,  String? description,  String category,  double price,  int quantity,  bool? status,  String? brand,  String? model, @JsonKey(readValue: _readBodyBrand)  String? bodyBrand, @JsonKey(readValue: _readEngineBrand)  String? engineBrand, @JsonKey(readValue: _readCarReleaseDate)  String? carReleaseDate, @JsonKey(readValue: _readFrontRear)  String? frontRear, @JsonKey(readValue: _readLeftRight)  String? leftRight, @JsonKey(readValue: _readTopBottom)  String? topBottom,  String? number,  String? manufacturer, @JsonKey(readValue: _readManufacturerCode)  String? manufacturerCode,  String? color,  String? condition, @JsonKey(readValue: _readOemCode)  String? oemCode, @JsonKey(readValue: _readSupplierCode)  String? supplierCode, @JsonKey(readValue: _readTransmissionModel)  String? transmissionModel,  String? defect,  String? transmission,  String? drive, @JsonKey(readValue: _readWearPercentage)  String? wearPercentage,  String? season,  String? diameter,  String? width,  String? profile, @JsonKey(readValue: _readTireQuantity)  String? tireQuantity,  String? drilling,  String? offset, @JsonKey(readValue: _readCenterHoleDiameter)  String? centerHoleDiameter, @JsonKey(readValue: _readTireModel)  String? tireModel,  String? vin, @JsonKey(readValue: _readCarReleasePeriod)  String? carReleasePeriod,  String location,  String address,  String? salesman,  List<String> photos, @JsonKey(readValue: _readCreatedAt)  DateTime? createdAt, @JsonKey(readValue: _readMarkedForDeletion)  bool markedForDeletion)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Part() when $default != null:
return $default(_that.id,_that.name,_that.description,_that.category,_that.price,_that.quantity,_that.status,_that.brand,_that.model,_that.bodyBrand,_that.engineBrand,_that.carReleaseDate,_that.frontRear,_that.leftRight,_that.topBottom,_that.number,_that.manufacturer,_that.manufacturerCode,_that.color,_that.condition,_that.oemCode,_that.supplierCode,_that.transmissionModel,_that.defect,_that.transmission,_that.drive,_that.wearPercentage,_that.season,_that.diameter,_that.width,_that.profile,_that.tireQuantity,_that.drilling,_that.offset,_that.centerHoleDiameter,_that.tireModel,_that.vin,_that.carReleasePeriod,_that.location,_that.address,_that.salesman,_that.photos,_that.createdAt,_that.markedForDeletion);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int id,  String name,  String? description,  String category,  double price,  int quantity,  bool? status,  String? brand,  String? model, @JsonKey(readValue: _readBodyBrand)  String? bodyBrand, @JsonKey(readValue: _readEngineBrand)  String? engineBrand, @JsonKey(readValue: _readCarReleaseDate)  String? carReleaseDate, @JsonKey(readValue: _readFrontRear)  String? frontRear, @JsonKey(readValue: _readLeftRight)  String? leftRight, @JsonKey(readValue: _readTopBottom)  String? topBottom,  String? number,  String? manufacturer, @JsonKey(readValue: _readManufacturerCode)  String? manufacturerCode,  String? color,  String? condition, @JsonKey(readValue: _readOemCode)  String? oemCode, @JsonKey(readValue: _readSupplierCode)  String? supplierCode, @JsonKey(readValue: _readTransmissionModel)  String? transmissionModel,  String? defect,  String? transmission,  String? drive, @JsonKey(readValue: _readWearPercentage)  String? wearPercentage,  String? season,  String? diameter,  String? width,  String? profile, @JsonKey(readValue: _readTireQuantity)  String? tireQuantity,  String? drilling,  String? offset, @JsonKey(readValue: _readCenterHoleDiameter)  String? centerHoleDiameter, @JsonKey(readValue: _readTireModel)  String? tireModel,  String? vin, @JsonKey(readValue: _readCarReleasePeriod)  String? carReleasePeriod,  String location,  String address,  String? salesman,  List<String> photos, @JsonKey(readValue: _readCreatedAt)  DateTime? createdAt, @JsonKey(readValue: _readMarkedForDeletion)  bool markedForDeletion)  $default,) {final _that = this;
switch (_that) {
case _Part():
return $default(_that.id,_that.name,_that.description,_that.category,_that.price,_that.quantity,_that.status,_that.brand,_that.model,_that.bodyBrand,_that.engineBrand,_that.carReleaseDate,_that.frontRear,_that.leftRight,_that.topBottom,_that.number,_that.manufacturer,_that.manufacturerCode,_that.color,_that.condition,_that.oemCode,_that.supplierCode,_that.transmissionModel,_that.defect,_that.transmission,_that.drive,_that.wearPercentage,_that.season,_that.diameter,_that.width,_that.profile,_that.tireQuantity,_that.drilling,_that.offset,_that.centerHoleDiameter,_that.tireModel,_that.vin,_that.carReleasePeriod,_that.location,_that.address,_that.salesman,_that.photos,_that.createdAt,_that.markedForDeletion);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int id,  String name,  String? description,  String category,  double price,  int quantity,  bool? status,  String? brand,  String? model, @JsonKey(readValue: _readBodyBrand)  String? bodyBrand, @JsonKey(readValue: _readEngineBrand)  String? engineBrand, @JsonKey(readValue: _readCarReleaseDate)  String? carReleaseDate, @JsonKey(readValue: _readFrontRear)  String? frontRear, @JsonKey(readValue: _readLeftRight)  String? leftRight, @JsonKey(readValue: _readTopBottom)  String? topBottom,  String? number,  String? manufacturer, @JsonKey(readValue: _readManufacturerCode)  String? manufacturerCode,  String? color,  String? condition, @JsonKey(readValue: _readOemCode)  String? oemCode, @JsonKey(readValue: _readSupplierCode)  String? supplierCode, @JsonKey(readValue: _readTransmissionModel)  String? transmissionModel,  String? defect,  String? transmission,  String? drive, @JsonKey(readValue: _readWearPercentage)  String? wearPercentage,  String? season,  String? diameter,  String? width,  String? profile, @JsonKey(readValue: _readTireQuantity)  String? tireQuantity,  String? drilling,  String? offset, @JsonKey(readValue: _readCenterHoleDiameter)  String? centerHoleDiameter, @JsonKey(readValue: _readTireModel)  String? tireModel,  String? vin, @JsonKey(readValue: _readCarReleasePeriod)  String? carReleasePeriod,  String location,  String address,  String? salesman,  List<String> photos, @JsonKey(readValue: _readCreatedAt)  DateTime? createdAt, @JsonKey(readValue: _readMarkedForDeletion)  bool markedForDeletion)?  $default,) {final _that = this;
switch (_that) {
case _Part() when $default != null:
return $default(_that.id,_that.name,_that.description,_that.category,_that.price,_that.quantity,_that.status,_that.brand,_that.model,_that.bodyBrand,_that.engineBrand,_that.carReleaseDate,_that.frontRear,_that.leftRight,_that.topBottom,_that.number,_that.manufacturer,_that.manufacturerCode,_that.color,_that.condition,_that.oemCode,_that.supplierCode,_that.transmissionModel,_that.defect,_that.transmission,_that.drive,_that.wearPercentage,_that.season,_that.diameter,_that.width,_that.profile,_that.tireQuantity,_that.drilling,_that.offset,_that.centerHoleDiameter,_that.tireModel,_that.vin,_that.carReleasePeriod,_that.location,_that.address,_that.salesman,_that.photos,_that.createdAt,_that.markedForDeletion);case _:
  return null;

}
}

}

/// @nodoc

@JsonSerializable(fieldRename: FieldRename.snake)
class _Part extends Part {
  const _Part({this.id = 0, this.name = '', this.description, this.category = '', this.price = 0.0, this.quantity = 0, this.status, this.brand, this.model, @JsonKey(readValue: _readBodyBrand) this.bodyBrand, @JsonKey(readValue: _readEngineBrand) this.engineBrand, @JsonKey(readValue: _readCarReleaseDate) this.carReleaseDate, @JsonKey(readValue: _readFrontRear) this.frontRear, @JsonKey(readValue: _readLeftRight) this.leftRight, @JsonKey(readValue: _readTopBottom) this.topBottom, this.number, this.manufacturer, @JsonKey(readValue: _readManufacturerCode) this.manufacturerCode, this.color, this.condition, @JsonKey(readValue: _readOemCode) this.oemCode, @JsonKey(readValue: _readSupplierCode) this.supplierCode, @JsonKey(readValue: _readTransmissionModel) this.transmissionModel, this.defect, this.transmission, this.drive, @JsonKey(readValue: _readWearPercentage) this.wearPercentage, this.season, this.diameter, this.width, this.profile, @JsonKey(readValue: _readTireQuantity) this.tireQuantity, this.drilling, this.offset, @JsonKey(readValue: _readCenterHoleDiameter) this.centerHoleDiameter, @JsonKey(readValue: _readTireModel) this.tireModel, this.vin, @JsonKey(readValue: _readCarReleasePeriod) this.carReleasePeriod, this.location = '', this.address = '', this.salesman, final  List<String> photos = const [], @JsonKey(readValue: _readCreatedAt) this.createdAt, @JsonKey(readValue: _readMarkedForDeletion) this.markedForDeletion = false}): _photos = photos,super._();
  factory _Part.fromJson(Map<String, dynamic> json) => _$PartFromJson(json);

@override@JsonKey() final  int id;
@override@JsonKey() final  String name;
@override final  String? description;
@override@JsonKey() final  String category;
@override@JsonKey() final  double price;
@override@JsonKey() final  int quantity;
@override final  bool? status;
@override final  String? brand;
@override final  String? model;
@override@JsonKey(readValue: _readBodyBrand) final  String? bodyBrand;
@override@JsonKey(readValue: _readEngineBrand) final  String? engineBrand;
@override@JsonKey(readValue: _readCarReleaseDate) final  String? carReleaseDate;
@override@JsonKey(readValue: _readFrontRear) final  String? frontRear;
@override@JsonKey(readValue: _readLeftRight) final  String? leftRight;
@override@JsonKey(readValue: _readTopBottom) final  String? topBottom;
@override final  String? number;
@override final  String? manufacturer;
@override@JsonKey(readValue: _readManufacturerCode) final  String? manufacturerCode;
@override final  String? color;
@override final  String? condition;
@override@JsonKey(readValue: _readOemCode) final  String? oemCode;
@override@JsonKey(readValue: _readSupplierCode) final  String? supplierCode;
@override@JsonKey(readValue: _readTransmissionModel) final  String? transmissionModel;
@override final  String? defect;
@override final  String? transmission;
@override final  String? drive;
@override@JsonKey(readValue: _readWearPercentage) final  String? wearPercentage;
@override final  String? season;
@override final  String? diameter;
@override final  String? width;
@override final  String? profile;
@override@JsonKey(readValue: _readTireQuantity) final  String? tireQuantity;
@override final  String? drilling;
@override final  String? offset;
@override@JsonKey(readValue: _readCenterHoleDiameter) final  String? centerHoleDiameter;
@override@JsonKey(readValue: _readTireModel) final  String? tireModel;
@override final  String? vin;
@override@JsonKey(readValue: _readCarReleasePeriod) final  String? carReleasePeriod;
@override@JsonKey() final  String location;
@override@JsonKey() final  String address;
@override final  String? salesman;
 final  List<String> _photos;
@override@JsonKey() List<String> get photos {
  if (_photos is EqualUnmodifiableListView) return _photos;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_photos);
}

@override@JsonKey(readValue: _readCreatedAt) final  DateTime? createdAt;
@override@JsonKey(readValue: _readMarkedForDeletion) final  bool markedForDeletion;

/// Create a copy of Part
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PartCopyWith<_Part> get copyWith => __$PartCopyWithImpl<_Part>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$PartToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _Part&&(identical(other.id, id) || other.id == id)&&(identical(other.name, name) || other.name == name)&&(identical(other.description, description) || other.description == description)&&(identical(other.category, category) || other.category == category)&&(identical(other.price, price) || other.price == price)&&(identical(other.quantity, quantity) || other.quantity == quantity)&&(identical(other.status, status) || other.status == status)&&(identical(other.brand, brand) || other.brand == brand)&&(identical(other.model, model) || other.model == model)&&(identical(other.bodyBrand, bodyBrand) || other.bodyBrand == bodyBrand)&&(identical(other.engineBrand, engineBrand) || other.engineBrand == engineBrand)&&(identical(other.carReleaseDate, carReleaseDate) || other.carReleaseDate == carReleaseDate)&&(identical(other.frontRear, frontRear) || other.frontRear == frontRear)&&(identical(other.leftRight, leftRight) || other.leftRight == leftRight)&&(identical(other.topBottom, topBottom) || other.topBottom == topBottom)&&(identical(other.number, number) || other.number == number)&&(identical(other.manufacturer, manufacturer) || other.manufacturer == manufacturer)&&(identical(other.manufacturerCode, manufacturerCode) || other.manufacturerCode == manufacturerCode)&&(identical(other.color, color) || other.color == color)&&(identical(other.condition, condition) || other.condition == condition)&&(identical(other.oemCode, oemCode) || other.oemCode == oemCode)&&(identical(other.supplierCode, supplierCode) || other.supplierCode == supplierCode)&&(identical(other.transmissionModel, transmissionModel) || other.transmissionModel == transmissionModel)&&(identical(other.defect, defect) || other.defect == defect)&&(identical(other.transmission, transmission) || other.transmission == transmission)&&(identical(other.drive, drive) || other.drive == drive)&&(identical(other.wearPercentage, wearPercentage) || other.wearPercentage == wearPercentage)&&(identical(other.season, season) || other.season == season)&&(identical(other.diameter, diameter) || other.diameter == diameter)&&(identical(other.width, width) || other.width == width)&&(identical(other.profile, profile) || other.profile == profile)&&(identical(other.tireQuantity, tireQuantity) || other.tireQuantity == tireQuantity)&&(identical(other.drilling, drilling) || other.drilling == drilling)&&(identical(other.offset, offset) || other.offset == offset)&&(identical(other.centerHoleDiameter, centerHoleDiameter) || other.centerHoleDiameter == centerHoleDiameter)&&(identical(other.tireModel, tireModel) || other.tireModel == tireModel)&&(identical(other.vin, vin) || other.vin == vin)&&(identical(other.carReleasePeriod, carReleasePeriod) || other.carReleasePeriod == carReleasePeriod)&&(identical(other.location, location) || other.location == location)&&(identical(other.address, address) || other.address == address)&&(identical(other.salesman, salesman) || other.salesman == salesman)&&const DeepCollectionEquality().equals(other._photos, _photos)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.markedForDeletion, markedForDeletion) || other.markedForDeletion == markedForDeletion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hashAll([runtimeType,id,name,description,category,price,quantity,status,brand,model,bodyBrand,engineBrand,carReleaseDate,frontRear,leftRight,topBottom,number,manufacturer,manufacturerCode,color,condition,oemCode,supplierCode,transmissionModel,defect,transmission,drive,wearPercentage,season,diameter,width,profile,tireQuantity,drilling,offset,centerHoleDiameter,tireModel,vin,carReleasePeriod,location,address,salesman,const DeepCollectionEquality().hash(_photos),createdAt,markedForDeletion]);

@override
String toString() {
  return 'Part(id: $id, name: $name, description: $description, category: $category, price: $price, quantity: $quantity, status: $status, brand: $brand, model: $model, bodyBrand: $bodyBrand, engineBrand: $engineBrand, carReleaseDate: $carReleaseDate, frontRear: $frontRear, leftRight: $leftRight, topBottom: $topBottom, number: $number, manufacturer: $manufacturer, manufacturerCode: $manufacturerCode, color: $color, condition: $condition, oemCode: $oemCode, supplierCode: $supplierCode, transmissionModel: $transmissionModel, defect: $defect, transmission: $transmission, drive: $drive, wearPercentage: $wearPercentage, season: $season, diameter: $diameter, width: $width, profile: $profile, tireQuantity: $tireQuantity, drilling: $drilling, offset: $offset, centerHoleDiameter: $centerHoleDiameter, tireModel: $tireModel, vin: $vin, carReleasePeriod: $carReleasePeriod, location: $location, address: $address, salesman: $salesman, photos: $photos, createdAt: $createdAt, markedForDeletion: $markedForDeletion)';
}


}

/// @nodoc
abstract mixin class _$PartCopyWith<$Res> implements $PartCopyWith<$Res> {
  factory _$PartCopyWith(_Part value, $Res Function(_Part) _then) = __$PartCopyWithImpl;
@override @useResult
$Res call({
 int id, String name, String? description, String category, double price, int quantity, bool? status, String? brand, String? model,@JsonKey(readValue: _readBodyBrand) String? bodyBrand,@JsonKey(readValue: _readEngineBrand) String? engineBrand,@JsonKey(readValue: _readCarReleaseDate) String? carReleaseDate,@JsonKey(readValue: _readFrontRear) String? frontRear,@JsonKey(readValue: _readLeftRight) String? leftRight,@JsonKey(readValue: _readTopBottom) String? topBottom, String? number, String? manufacturer,@JsonKey(readValue: _readManufacturerCode) String? manufacturerCode, String? color, String? condition,@JsonKey(readValue: _readOemCode) String? oemCode,@JsonKey(readValue: _readSupplierCode) String? supplierCode,@JsonKey(readValue: _readTransmissionModel) String? transmissionModel, String? defect, String? transmission, String? drive,@JsonKey(readValue: _readWearPercentage) String? wearPercentage, String? season, String? diameter, String? width, String? profile,@JsonKey(readValue: _readTireQuantity) String? tireQuantity, String? drilling, String? offset,@JsonKey(readValue: _readCenterHoleDiameter) String? centerHoleDiameter,@JsonKey(readValue: _readTireModel) String? tireModel, String? vin,@JsonKey(readValue: _readCarReleasePeriod) String? carReleasePeriod, String location, String address, String? salesman, List<String> photos,@JsonKey(readValue: _readCreatedAt) DateTime? createdAt,@JsonKey(readValue: _readMarkedForDeletion) bool markedForDeletion
});




}
/// @nodoc
class __$PartCopyWithImpl<$Res>
    implements _$PartCopyWith<$Res> {
  __$PartCopyWithImpl(this._self, this._then);

  final _Part _self;
  final $Res Function(_Part) _then;

/// Create a copy of Part
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? name = null,Object? description = freezed,Object? category = null,Object? price = null,Object? quantity = null,Object? status = freezed,Object? brand = freezed,Object? model = freezed,Object? bodyBrand = freezed,Object? engineBrand = freezed,Object? carReleaseDate = freezed,Object? frontRear = freezed,Object? leftRight = freezed,Object? topBottom = freezed,Object? number = freezed,Object? manufacturer = freezed,Object? manufacturerCode = freezed,Object? color = freezed,Object? condition = freezed,Object? oemCode = freezed,Object? supplierCode = freezed,Object? transmissionModel = freezed,Object? defect = freezed,Object? transmission = freezed,Object? drive = freezed,Object? wearPercentage = freezed,Object? season = freezed,Object? diameter = freezed,Object? width = freezed,Object? profile = freezed,Object? tireQuantity = freezed,Object? drilling = freezed,Object? offset = freezed,Object? centerHoleDiameter = freezed,Object? tireModel = freezed,Object? vin = freezed,Object? carReleasePeriod = freezed,Object? location = null,Object? address = null,Object? salesman = freezed,Object? photos = null,Object? createdAt = freezed,Object? markedForDeletion = null,}) {
  return _then(_Part(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as int,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String?,category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as String,price: null == price ? _self.price : price // ignore: cast_nullable_to_non_nullable
as double,quantity: null == quantity ? _self.quantity : quantity // ignore: cast_nullable_to_non_nullable
as int,status: freezed == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as bool?,brand: freezed == brand ? _self.brand : brand // ignore: cast_nullable_to_non_nullable
as String?,model: freezed == model ? _self.model : model // ignore: cast_nullable_to_non_nullable
as String?,bodyBrand: freezed == bodyBrand ? _self.bodyBrand : bodyBrand // ignore: cast_nullable_to_non_nullable
as String?,engineBrand: freezed == engineBrand ? _self.engineBrand : engineBrand // ignore: cast_nullable_to_non_nullable
as String?,carReleaseDate: freezed == carReleaseDate ? _self.carReleaseDate : carReleaseDate // ignore: cast_nullable_to_non_nullable
as String?,frontRear: freezed == frontRear ? _self.frontRear : frontRear // ignore: cast_nullable_to_non_nullable
as String?,leftRight: freezed == leftRight ? _self.leftRight : leftRight // ignore: cast_nullable_to_non_nullable
as String?,topBottom: freezed == topBottom ? _self.topBottom : topBottom // ignore: cast_nullable_to_non_nullable
as String?,number: freezed == number ? _self.number : number // ignore: cast_nullable_to_non_nullable
as String?,manufacturer: freezed == manufacturer ? _self.manufacturer : manufacturer // ignore: cast_nullable_to_non_nullable
as String?,manufacturerCode: freezed == manufacturerCode ? _self.manufacturerCode : manufacturerCode // ignore: cast_nullable_to_non_nullable
as String?,color: freezed == color ? _self.color : color // ignore: cast_nullable_to_non_nullable
as String?,condition: freezed == condition ? _self.condition : condition // ignore: cast_nullable_to_non_nullable
as String?,oemCode: freezed == oemCode ? _self.oemCode : oemCode // ignore: cast_nullable_to_non_nullable
as String?,supplierCode: freezed == supplierCode ? _self.supplierCode : supplierCode // ignore: cast_nullable_to_non_nullable
as String?,transmissionModel: freezed == transmissionModel ? _self.transmissionModel : transmissionModel // ignore: cast_nullable_to_non_nullable
as String?,defect: freezed == defect ? _self.defect : defect // ignore: cast_nullable_to_non_nullable
as String?,transmission: freezed == transmission ? _self.transmission : transmission // ignore: cast_nullable_to_non_nullable
as String?,drive: freezed == drive ? _self.drive : drive // ignore: cast_nullable_to_non_nullable
as String?,wearPercentage: freezed == wearPercentage ? _self.wearPercentage : wearPercentage // ignore: cast_nullable_to_non_nullable
as String?,season: freezed == season ? _self.season : season // ignore: cast_nullable_to_non_nullable
as String?,diameter: freezed == diameter ? _self.diameter : diameter // ignore: cast_nullable_to_non_nullable
as String?,width: freezed == width ? _self.width : width // ignore: cast_nullable_to_non_nullable
as String?,profile: freezed == profile ? _self.profile : profile // ignore: cast_nullable_to_non_nullable
as String?,tireQuantity: freezed == tireQuantity ? _self.tireQuantity : tireQuantity // ignore: cast_nullable_to_non_nullable
as String?,drilling: freezed == drilling ? _self.drilling : drilling // ignore: cast_nullable_to_non_nullable
as String?,offset: freezed == offset ? _self.offset : offset // ignore: cast_nullable_to_non_nullable
as String?,centerHoleDiameter: freezed == centerHoleDiameter ? _self.centerHoleDiameter : centerHoleDiameter // ignore: cast_nullable_to_non_nullable
as String?,tireModel: freezed == tireModel ? _self.tireModel : tireModel // ignore: cast_nullable_to_non_nullable
as String?,vin: freezed == vin ? _self.vin : vin // ignore: cast_nullable_to_non_nullable
as String?,carReleasePeriod: freezed == carReleasePeriod ? _self.carReleasePeriod : carReleasePeriod // ignore: cast_nullable_to_non_nullable
as String?,location: null == location ? _self.location : location // ignore: cast_nullable_to_non_nullable
as String,address: null == address ? _self.address : address // ignore: cast_nullable_to_non_nullable
as String,salesman: freezed == salesman ? _self.salesman : salesman // ignore: cast_nullable_to_non_nullable
as String?,photos: null == photos ? _self._photos : photos // ignore: cast_nullable_to_non_nullable
as List<String>,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,markedForDeletion: null == markedForDeletion ? _self.markedForDeletion : markedForDeletion // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

/// @nodoc
mixin _$InventoryResponse {

 List<Part> get parts; int get total; int get page; int get limit; bool get isOffline;
/// Create a copy of InventoryResponse
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$InventoryResponseCopyWith<InventoryResponse> get copyWith => _$InventoryResponseCopyWithImpl<InventoryResponse>(this as InventoryResponse, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is InventoryResponse&&const DeepCollectionEquality().equals(other.parts, parts)&&(identical(other.total, total) || other.total == total)&&(identical(other.page, page) || other.page == page)&&(identical(other.limit, limit) || other.limit == limit)&&(identical(other.isOffline, isOffline) || other.isOffline == isOffline));
}


@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(parts),total,page,limit,isOffline);

@override
String toString() {
  return 'InventoryResponse(parts: $parts, total: $total, page: $page, limit: $limit, isOffline: $isOffline)';
}


}

/// @nodoc
abstract mixin class $InventoryResponseCopyWith<$Res>  {
  factory $InventoryResponseCopyWith(InventoryResponse value, $Res Function(InventoryResponse) _then) = _$InventoryResponseCopyWithImpl;
@useResult
$Res call({
 List<Part> parts, int total, int page, int limit, bool isOffline
});




}
/// @nodoc
class _$InventoryResponseCopyWithImpl<$Res>
    implements $InventoryResponseCopyWith<$Res> {
  _$InventoryResponseCopyWithImpl(this._self, this._then);

  final InventoryResponse _self;
  final $Res Function(InventoryResponse) _then;

/// Create a copy of InventoryResponse
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? parts = null,Object? total = null,Object? page = null,Object? limit = null,Object? isOffline = null,}) {
  return _then(_self.copyWith(
parts: null == parts ? _self.parts : parts // ignore: cast_nullable_to_non_nullable
as List<Part>,total: null == total ? _self.total : total // ignore: cast_nullable_to_non_nullable
as int,page: null == page ? _self.page : page // ignore: cast_nullable_to_non_nullable
as int,limit: null == limit ? _self.limit : limit // ignore: cast_nullable_to_non_nullable
as int,isOffline: null == isOffline ? _self.isOffline : isOffline // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [InventoryResponse].
extension InventoryResponsePatterns on InventoryResponse {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _InventoryResponse value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _InventoryResponse() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _InventoryResponse value)  $default,){
final _that = this;
switch (_that) {
case _InventoryResponse():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _InventoryResponse value)?  $default,){
final _that = this;
switch (_that) {
case _InventoryResponse() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<Part> parts,  int total,  int page,  int limit,  bool isOffline)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _InventoryResponse() when $default != null:
return $default(_that.parts,_that.total,_that.page,_that.limit,_that.isOffline);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<Part> parts,  int total,  int page,  int limit,  bool isOffline)  $default,) {final _that = this;
switch (_that) {
case _InventoryResponse():
return $default(_that.parts,_that.total,_that.page,_that.limit,_that.isOffline);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<Part> parts,  int total,  int page,  int limit,  bool isOffline)?  $default,) {final _that = this;
switch (_that) {
case _InventoryResponse() when $default != null:
return $default(_that.parts,_that.total,_that.page,_that.limit,_that.isOffline);case _:
  return null;

}
}

}

/// @nodoc


class _InventoryResponse implements InventoryResponse {
  const _InventoryResponse({final  List<Part> parts = const [], this.total = 0, this.page = 1, this.limit = 20, this.isOffline = false}): _parts = parts;
  

 final  List<Part> _parts;
@override@JsonKey() List<Part> get parts {
  if (_parts is EqualUnmodifiableListView) return _parts;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_parts);
}

@override@JsonKey() final  int total;
@override@JsonKey() final  int page;
@override@JsonKey() final  int limit;
@override@JsonKey() final  bool isOffline;

/// Create a copy of InventoryResponse
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$InventoryResponseCopyWith<_InventoryResponse> get copyWith => __$InventoryResponseCopyWithImpl<_InventoryResponse>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _InventoryResponse&&const DeepCollectionEquality().equals(other._parts, _parts)&&(identical(other.total, total) || other.total == total)&&(identical(other.page, page) || other.page == page)&&(identical(other.limit, limit) || other.limit == limit)&&(identical(other.isOffline, isOffline) || other.isOffline == isOffline));
}


@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(_parts),total,page,limit,isOffline);

@override
String toString() {
  return 'InventoryResponse(parts: $parts, total: $total, page: $page, limit: $limit, isOffline: $isOffline)';
}


}

/// @nodoc
abstract mixin class _$InventoryResponseCopyWith<$Res> implements $InventoryResponseCopyWith<$Res> {
  factory _$InventoryResponseCopyWith(_InventoryResponse value, $Res Function(_InventoryResponse) _then) = __$InventoryResponseCopyWithImpl;
@override @useResult
$Res call({
 List<Part> parts, int total, int page, int limit, bool isOffline
});




}
/// @nodoc
class __$InventoryResponseCopyWithImpl<$Res>
    implements _$InventoryResponseCopyWith<$Res> {
  __$InventoryResponseCopyWithImpl(this._self, this._then);

  final _InventoryResponse _self;
  final $Res Function(_InventoryResponse) _then;

/// Create a copy of InventoryResponse
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? parts = null,Object? total = null,Object? page = null,Object? limit = null,Object? isOffline = null,}) {
  return _then(_InventoryResponse(
parts: null == parts ? _self._parts : parts // ignore: cast_nullable_to_non_nullable
as List<Part>,total: null == total ? _self.total : total // ignore: cast_nullable_to_non_nullable
as int,page: null == page ? _self.page : page // ignore: cast_nullable_to_non_nullable
as int,limit: null == limit ? _self.limit : limit // ignore: cast_nullable_to_non_nullable
as int,isOffline: null == isOffline ? _self.isOffline : isOffline // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

// dart format on
