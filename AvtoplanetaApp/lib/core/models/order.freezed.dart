// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'order.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Order {

 int get id; int? get customerId;@JsonKey(readValue: _readPartName) String get partName; int? get partId; String get location; String get buyerNumber; String get orderNumber; String get status; String get statusText; String get source; String get paymentStatus; String get warehouseStatus; String get deliveryMethod; String get transportCompany; String get trackingNumber; String get notes; double get discount; double get totalAmount;@JsonKey(readValue: _readSellerName) String get sellerName; String get timeAgo; String get createdAtFormatted; String get completedAtFormatted; DateTime? get createdAt; bool get autoDeleted; List<OrderItem> get items;
/// Create a copy of Order
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$OrderCopyWith<Order> get copyWith => _$OrderCopyWithImpl<Order>(this as Order, _$identity);

  /// Serializes this Order to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Order&&(identical(other.id, id) || other.id == id)&&(identical(other.customerId, customerId) || other.customerId == customerId)&&(identical(other.partName, partName) || other.partName == partName)&&(identical(other.partId, partId) || other.partId == partId)&&(identical(other.location, location) || other.location == location)&&(identical(other.buyerNumber, buyerNumber) || other.buyerNumber == buyerNumber)&&(identical(other.orderNumber, orderNumber) || other.orderNumber == orderNumber)&&(identical(other.status, status) || other.status == status)&&(identical(other.statusText, statusText) || other.statusText == statusText)&&(identical(other.source, source) || other.source == source)&&(identical(other.paymentStatus, paymentStatus) || other.paymentStatus == paymentStatus)&&(identical(other.warehouseStatus, warehouseStatus) || other.warehouseStatus == warehouseStatus)&&(identical(other.deliveryMethod, deliveryMethod) || other.deliveryMethod == deliveryMethod)&&(identical(other.transportCompany, transportCompany) || other.transportCompany == transportCompany)&&(identical(other.trackingNumber, trackingNumber) || other.trackingNumber == trackingNumber)&&(identical(other.notes, notes) || other.notes == notes)&&(identical(other.discount, discount) || other.discount == discount)&&(identical(other.totalAmount, totalAmount) || other.totalAmount == totalAmount)&&(identical(other.sellerName, sellerName) || other.sellerName == sellerName)&&(identical(other.timeAgo, timeAgo) || other.timeAgo == timeAgo)&&(identical(other.createdAtFormatted, createdAtFormatted) || other.createdAtFormatted == createdAtFormatted)&&(identical(other.completedAtFormatted, completedAtFormatted) || other.completedAtFormatted == completedAtFormatted)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.autoDeleted, autoDeleted) || other.autoDeleted == autoDeleted)&&const DeepCollectionEquality().equals(other.items, items));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hashAll([runtimeType,id,customerId,partName,partId,location,buyerNumber,orderNumber,status,statusText,source,paymentStatus,warehouseStatus,deliveryMethod,transportCompany,trackingNumber,notes,discount,totalAmount,sellerName,timeAgo,createdAtFormatted,completedAtFormatted,createdAt,autoDeleted,const DeepCollectionEquality().hash(items)]);

@override
String toString() {
  return 'Order(id: $id, customerId: $customerId, partName: $partName, partId: $partId, location: $location, buyerNumber: $buyerNumber, orderNumber: $orderNumber, status: $status, statusText: $statusText, source: $source, paymentStatus: $paymentStatus, warehouseStatus: $warehouseStatus, deliveryMethod: $deliveryMethod, transportCompany: $transportCompany, trackingNumber: $trackingNumber, notes: $notes, discount: $discount, totalAmount: $totalAmount, sellerName: $sellerName, timeAgo: $timeAgo, createdAtFormatted: $createdAtFormatted, completedAtFormatted: $completedAtFormatted, createdAt: $createdAt, autoDeleted: $autoDeleted, items: $items)';
}


}

/// @nodoc
abstract mixin class $OrderCopyWith<$Res>  {
  factory $OrderCopyWith(Order value, $Res Function(Order) _then) = _$OrderCopyWithImpl;
@useResult
$Res call({
 int id, int? customerId,@JsonKey(readValue: _readPartName) String partName, int? partId, String location, String buyerNumber, String orderNumber, String status, String statusText, String source, String paymentStatus, String warehouseStatus, String deliveryMethod, String transportCompany, String trackingNumber, String notes, double discount, double totalAmount,@JsonKey(readValue: _readSellerName) String sellerName, String timeAgo, String createdAtFormatted, String completedAtFormatted, DateTime? createdAt, bool autoDeleted, List<OrderItem> items
});




}
/// @nodoc
class _$OrderCopyWithImpl<$Res>
    implements $OrderCopyWith<$Res> {
  _$OrderCopyWithImpl(this._self, this._then);

  final Order _self;
  final $Res Function(Order) _then;

/// Create a copy of Order
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? customerId = freezed,Object? partName = null,Object? partId = freezed,Object? location = null,Object? buyerNumber = null,Object? orderNumber = null,Object? status = null,Object? statusText = null,Object? source = null,Object? paymentStatus = null,Object? warehouseStatus = null,Object? deliveryMethod = null,Object? transportCompany = null,Object? trackingNumber = null,Object? notes = null,Object? discount = null,Object? totalAmount = null,Object? sellerName = null,Object? timeAgo = null,Object? createdAtFormatted = null,Object? completedAtFormatted = null,Object? createdAt = freezed,Object? autoDeleted = null,Object? items = null,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as int,customerId: freezed == customerId ? _self.customerId : customerId // ignore: cast_nullable_to_non_nullable
as int?,partName: null == partName ? _self.partName : partName // ignore: cast_nullable_to_non_nullable
as String,partId: freezed == partId ? _self.partId : partId // ignore: cast_nullable_to_non_nullable
as int?,location: null == location ? _self.location : location // ignore: cast_nullable_to_non_nullable
as String,buyerNumber: null == buyerNumber ? _self.buyerNumber : buyerNumber // ignore: cast_nullable_to_non_nullable
as String,orderNumber: null == orderNumber ? _self.orderNumber : orderNumber // ignore: cast_nullable_to_non_nullable
as String,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,statusText: null == statusText ? _self.statusText : statusText // ignore: cast_nullable_to_non_nullable
as String,source: null == source ? _self.source : source // ignore: cast_nullable_to_non_nullable
as String,paymentStatus: null == paymentStatus ? _self.paymentStatus : paymentStatus // ignore: cast_nullable_to_non_nullable
as String,warehouseStatus: null == warehouseStatus ? _self.warehouseStatus : warehouseStatus // ignore: cast_nullable_to_non_nullable
as String,deliveryMethod: null == deliveryMethod ? _self.deliveryMethod : deliveryMethod // ignore: cast_nullable_to_non_nullable
as String,transportCompany: null == transportCompany ? _self.transportCompany : transportCompany // ignore: cast_nullable_to_non_nullable
as String,trackingNumber: null == trackingNumber ? _self.trackingNumber : trackingNumber // ignore: cast_nullable_to_non_nullable
as String,notes: null == notes ? _self.notes : notes // ignore: cast_nullable_to_non_nullable
as String,discount: null == discount ? _self.discount : discount // ignore: cast_nullable_to_non_nullable
as double,totalAmount: null == totalAmount ? _self.totalAmount : totalAmount // ignore: cast_nullable_to_non_nullable
as double,sellerName: null == sellerName ? _self.sellerName : sellerName // ignore: cast_nullable_to_non_nullable
as String,timeAgo: null == timeAgo ? _self.timeAgo : timeAgo // ignore: cast_nullable_to_non_nullable
as String,createdAtFormatted: null == createdAtFormatted ? _self.createdAtFormatted : createdAtFormatted // ignore: cast_nullable_to_non_nullable
as String,completedAtFormatted: null == completedAtFormatted ? _self.completedAtFormatted : completedAtFormatted // ignore: cast_nullable_to_non_nullable
as String,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,autoDeleted: null == autoDeleted ? _self.autoDeleted : autoDeleted // ignore: cast_nullable_to_non_nullable
as bool,items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<OrderItem>,
  ));
}

}


/// Adds pattern-matching-related methods to [Order].
extension OrderPatterns on Order {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Order value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Order() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Order value)  $default,){
final _that = this;
switch (_that) {
case _Order():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Order value)?  $default,){
final _that = this;
switch (_that) {
case _Order() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int id,  int? customerId, @JsonKey(readValue: _readPartName)  String partName,  int? partId,  String location,  String buyerNumber,  String orderNumber,  String status,  String statusText,  String source,  String paymentStatus,  String warehouseStatus,  String deliveryMethod,  String transportCompany,  String trackingNumber,  String notes,  double discount,  double totalAmount, @JsonKey(readValue: _readSellerName)  String sellerName,  String timeAgo,  String createdAtFormatted,  String completedAtFormatted,  DateTime? createdAt,  bool autoDeleted,  List<OrderItem> items)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Order() when $default != null:
return $default(_that.id,_that.customerId,_that.partName,_that.partId,_that.location,_that.buyerNumber,_that.orderNumber,_that.status,_that.statusText,_that.source,_that.paymentStatus,_that.warehouseStatus,_that.deliveryMethod,_that.transportCompany,_that.trackingNumber,_that.notes,_that.discount,_that.totalAmount,_that.sellerName,_that.timeAgo,_that.createdAtFormatted,_that.completedAtFormatted,_that.createdAt,_that.autoDeleted,_that.items);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int id,  int? customerId, @JsonKey(readValue: _readPartName)  String partName,  int? partId,  String location,  String buyerNumber,  String orderNumber,  String status,  String statusText,  String source,  String paymentStatus,  String warehouseStatus,  String deliveryMethod,  String transportCompany,  String trackingNumber,  String notes,  double discount,  double totalAmount, @JsonKey(readValue: _readSellerName)  String sellerName,  String timeAgo,  String createdAtFormatted,  String completedAtFormatted,  DateTime? createdAt,  bool autoDeleted,  List<OrderItem> items)  $default,) {final _that = this;
switch (_that) {
case _Order():
return $default(_that.id,_that.customerId,_that.partName,_that.partId,_that.location,_that.buyerNumber,_that.orderNumber,_that.status,_that.statusText,_that.source,_that.paymentStatus,_that.warehouseStatus,_that.deliveryMethod,_that.transportCompany,_that.trackingNumber,_that.notes,_that.discount,_that.totalAmount,_that.sellerName,_that.timeAgo,_that.createdAtFormatted,_that.completedAtFormatted,_that.createdAt,_that.autoDeleted,_that.items);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int id,  int? customerId, @JsonKey(readValue: _readPartName)  String partName,  int? partId,  String location,  String buyerNumber,  String orderNumber,  String status,  String statusText,  String source,  String paymentStatus,  String warehouseStatus,  String deliveryMethod,  String transportCompany,  String trackingNumber,  String notes,  double discount,  double totalAmount, @JsonKey(readValue: _readSellerName)  String sellerName,  String timeAgo,  String createdAtFormatted,  String completedAtFormatted,  DateTime? createdAt,  bool autoDeleted,  List<OrderItem> items)?  $default,) {final _that = this;
switch (_that) {
case _Order() when $default != null:
return $default(_that.id,_that.customerId,_that.partName,_that.partId,_that.location,_that.buyerNumber,_that.orderNumber,_that.status,_that.statusText,_that.source,_that.paymentStatus,_that.warehouseStatus,_that.deliveryMethod,_that.transportCompany,_that.trackingNumber,_that.notes,_that.discount,_that.totalAmount,_that.sellerName,_that.timeAgo,_that.createdAtFormatted,_that.completedAtFormatted,_that.createdAt,_that.autoDeleted,_that.items);case _:
  return null;

}
}

}

/// @nodoc

@JsonSerializable(fieldRename: FieldRename.snake, explicitToJson: true)
class _Order extends Order {
  const _Order({this.id = 0, this.customerId, @JsonKey(readValue: _readPartName) this.partName = '', this.partId, this.location = '', this.buyerNumber = '', this.orderNumber = '', this.status = 'Принят в обработку', this.statusText = '', this.source = 'drom', this.paymentStatus = 'unpaid', this.warehouseStatus = 'inspecting', this.deliveryMethod = 'tk', this.transportCompany = '', this.trackingNumber = '', this.notes = '', this.discount = 0.0, this.totalAmount = 0.0, @JsonKey(readValue: _readSellerName) this.sellerName = '', this.timeAgo = '', this.createdAtFormatted = '', this.completedAtFormatted = '', this.createdAt, this.autoDeleted = false, final  List<OrderItem> items = const []}): _items = items,super._();
  factory _Order.fromJson(Map<String, dynamic> json) => _$OrderFromJson(json);

@override@JsonKey() final  int id;
@override final  int? customerId;
@override@JsonKey(readValue: _readPartName) final  String partName;
@override final  int? partId;
@override@JsonKey() final  String location;
@override@JsonKey() final  String buyerNumber;
@override@JsonKey() final  String orderNumber;
@override@JsonKey() final  String status;
@override@JsonKey() final  String statusText;
@override@JsonKey() final  String source;
@override@JsonKey() final  String paymentStatus;
@override@JsonKey() final  String warehouseStatus;
@override@JsonKey() final  String deliveryMethod;
@override@JsonKey() final  String transportCompany;
@override@JsonKey() final  String trackingNumber;
@override@JsonKey() final  String notes;
@override@JsonKey() final  double discount;
@override@JsonKey() final  double totalAmount;
@override@JsonKey(readValue: _readSellerName) final  String sellerName;
@override@JsonKey() final  String timeAgo;
@override@JsonKey() final  String createdAtFormatted;
@override@JsonKey() final  String completedAtFormatted;
@override final  DateTime? createdAt;
@override@JsonKey() final  bool autoDeleted;
 final  List<OrderItem> _items;
@override@JsonKey() List<OrderItem> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}


/// Create a copy of Order
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$OrderCopyWith<_Order> get copyWith => __$OrderCopyWithImpl<_Order>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$OrderToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _Order&&(identical(other.id, id) || other.id == id)&&(identical(other.customerId, customerId) || other.customerId == customerId)&&(identical(other.partName, partName) || other.partName == partName)&&(identical(other.partId, partId) || other.partId == partId)&&(identical(other.location, location) || other.location == location)&&(identical(other.buyerNumber, buyerNumber) || other.buyerNumber == buyerNumber)&&(identical(other.orderNumber, orderNumber) || other.orderNumber == orderNumber)&&(identical(other.status, status) || other.status == status)&&(identical(other.statusText, statusText) || other.statusText == statusText)&&(identical(other.source, source) || other.source == source)&&(identical(other.paymentStatus, paymentStatus) || other.paymentStatus == paymentStatus)&&(identical(other.warehouseStatus, warehouseStatus) || other.warehouseStatus == warehouseStatus)&&(identical(other.deliveryMethod, deliveryMethod) || other.deliveryMethod == deliveryMethod)&&(identical(other.transportCompany, transportCompany) || other.transportCompany == transportCompany)&&(identical(other.trackingNumber, trackingNumber) || other.trackingNumber == trackingNumber)&&(identical(other.notes, notes) || other.notes == notes)&&(identical(other.discount, discount) || other.discount == discount)&&(identical(other.totalAmount, totalAmount) || other.totalAmount == totalAmount)&&(identical(other.sellerName, sellerName) || other.sellerName == sellerName)&&(identical(other.timeAgo, timeAgo) || other.timeAgo == timeAgo)&&(identical(other.createdAtFormatted, createdAtFormatted) || other.createdAtFormatted == createdAtFormatted)&&(identical(other.completedAtFormatted, completedAtFormatted) || other.completedAtFormatted == completedAtFormatted)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.autoDeleted, autoDeleted) || other.autoDeleted == autoDeleted)&&const DeepCollectionEquality().equals(other._items, _items));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hashAll([runtimeType,id,customerId,partName,partId,location,buyerNumber,orderNumber,status,statusText,source,paymentStatus,warehouseStatus,deliveryMethod,transportCompany,trackingNumber,notes,discount,totalAmount,sellerName,timeAgo,createdAtFormatted,completedAtFormatted,createdAt,autoDeleted,const DeepCollectionEquality().hash(_items)]);

@override
String toString() {
  return 'Order(id: $id, customerId: $customerId, partName: $partName, partId: $partId, location: $location, buyerNumber: $buyerNumber, orderNumber: $orderNumber, status: $status, statusText: $statusText, source: $source, paymentStatus: $paymentStatus, warehouseStatus: $warehouseStatus, deliveryMethod: $deliveryMethod, transportCompany: $transportCompany, trackingNumber: $trackingNumber, notes: $notes, discount: $discount, totalAmount: $totalAmount, sellerName: $sellerName, timeAgo: $timeAgo, createdAtFormatted: $createdAtFormatted, completedAtFormatted: $completedAtFormatted, createdAt: $createdAt, autoDeleted: $autoDeleted, items: $items)';
}


}

/// @nodoc
abstract mixin class _$OrderCopyWith<$Res> implements $OrderCopyWith<$Res> {
  factory _$OrderCopyWith(_Order value, $Res Function(_Order) _then) = __$OrderCopyWithImpl;
@override @useResult
$Res call({
 int id, int? customerId,@JsonKey(readValue: _readPartName) String partName, int? partId, String location, String buyerNumber, String orderNumber, String status, String statusText, String source, String paymentStatus, String warehouseStatus, String deliveryMethod, String transportCompany, String trackingNumber, String notes, double discount, double totalAmount,@JsonKey(readValue: _readSellerName) String sellerName, String timeAgo, String createdAtFormatted, String completedAtFormatted, DateTime? createdAt, bool autoDeleted, List<OrderItem> items
});




}
/// @nodoc
class __$OrderCopyWithImpl<$Res>
    implements _$OrderCopyWith<$Res> {
  __$OrderCopyWithImpl(this._self, this._then);

  final _Order _self;
  final $Res Function(_Order) _then;

/// Create a copy of Order
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? customerId = freezed,Object? partName = null,Object? partId = freezed,Object? location = null,Object? buyerNumber = null,Object? orderNumber = null,Object? status = null,Object? statusText = null,Object? source = null,Object? paymentStatus = null,Object? warehouseStatus = null,Object? deliveryMethod = null,Object? transportCompany = null,Object? trackingNumber = null,Object? notes = null,Object? discount = null,Object? totalAmount = null,Object? sellerName = null,Object? timeAgo = null,Object? createdAtFormatted = null,Object? completedAtFormatted = null,Object? createdAt = freezed,Object? autoDeleted = null,Object? items = null,}) {
  return _then(_Order(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as int,customerId: freezed == customerId ? _self.customerId : customerId // ignore: cast_nullable_to_non_nullable
as int?,partName: null == partName ? _self.partName : partName // ignore: cast_nullable_to_non_nullable
as String,partId: freezed == partId ? _self.partId : partId // ignore: cast_nullable_to_non_nullable
as int?,location: null == location ? _self.location : location // ignore: cast_nullable_to_non_nullable
as String,buyerNumber: null == buyerNumber ? _self.buyerNumber : buyerNumber // ignore: cast_nullable_to_non_nullable
as String,orderNumber: null == orderNumber ? _self.orderNumber : orderNumber // ignore: cast_nullable_to_non_nullable
as String,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,statusText: null == statusText ? _self.statusText : statusText // ignore: cast_nullable_to_non_nullable
as String,source: null == source ? _self.source : source // ignore: cast_nullable_to_non_nullable
as String,paymentStatus: null == paymentStatus ? _self.paymentStatus : paymentStatus // ignore: cast_nullable_to_non_nullable
as String,warehouseStatus: null == warehouseStatus ? _self.warehouseStatus : warehouseStatus // ignore: cast_nullable_to_non_nullable
as String,deliveryMethod: null == deliveryMethod ? _self.deliveryMethod : deliveryMethod // ignore: cast_nullable_to_non_nullable
as String,transportCompany: null == transportCompany ? _self.transportCompany : transportCompany // ignore: cast_nullable_to_non_nullable
as String,trackingNumber: null == trackingNumber ? _self.trackingNumber : trackingNumber // ignore: cast_nullable_to_non_nullable
as String,notes: null == notes ? _self.notes : notes // ignore: cast_nullable_to_non_nullable
as String,discount: null == discount ? _self.discount : discount // ignore: cast_nullable_to_non_nullable
as double,totalAmount: null == totalAmount ? _self.totalAmount : totalAmount // ignore: cast_nullable_to_non_nullable
as double,sellerName: null == sellerName ? _self.sellerName : sellerName // ignore: cast_nullable_to_non_nullable
as String,timeAgo: null == timeAgo ? _self.timeAgo : timeAgo // ignore: cast_nullable_to_non_nullable
as String,createdAtFormatted: null == createdAtFormatted ? _self.createdAtFormatted : createdAtFormatted // ignore: cast_nullable_to_non_nullable
as String,completedAtFormatted: null == completedAtFormatted ? _self.completedAtFormatted : completedAtFormatted // ignore: cast_nullable_to_non_nullable
as String,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,autoDeleted: null == autoDeleted ? _self.autoDeleted : autoDeleted // ignore: cast_nullable_to_non_nullable
as bool,items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<OrderItem>,
  ));
}


}


/// @nodoc
mixin _$OrderItem {

 int get id; int get partId;@JsonKey(readValue: _readItemPartName) String get partName; String get partNameSnapshot; int get quantity; double get price;
/// Create a copy of OrderItem
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$OrderItemCopyWith<OrderItem> get copyWith => _$OrderItemCopyWithImpl<OrderItem>(this as OrderItem, _$identity);

  /// Serializes this OrderItem to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is OrderItem&&(identical(other.id, id) || other.id == id)&&(identical(other.partId, partId) || other.partId == partId)&&(identical(other.partName, partName) || other.partName == partName)&&(identical(other.partNameSnapshot, partNameSnapshot) || other.partNameSnapshot == partNameSnapshot)&&(identical(other.quantity, quantity) || other.quantity == quantity)&&(identical(other.price, price) || other.price == price));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,partId,partName,partNameSnapshot,quantity,price);

@override
String toString() {
  return 'OrderItem(id: $id, partId: $partId, partName: $partName, partNameSnapshot: $partNameSnapshot, quantity: $quantity, price: $price)';
}


}

/// @nodoc
abstract mixin class $OrderItemCopyWith<$Res>  {
  factory $OrderItemCopyWith(OrderItem value, $Res Function(OrderItem) _then) = _$OrderItemCopyWithImpl;
@useResult
$Res call({
 int id, int partId,@JsonKey(readValue: _readItemPartName) String partName, String partNameSnapshot, int quantity, double price
});




}
/// @nodoc
class _$OrderItemCopyWithImpl<$Res>
    implements $OrderItemCopyWith<$Res> {
  _$OrderItemCopyWithImpl(this._self, this._then);

  final OrderItem _self;
  final $Res Function(OrderItem) _then;

/// Create a copy of OrderItem
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? partId = null,Object? partName = null,Object? partNameSnapshot = null,Object? quantity = null,Object? price = null,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as int,partId: null == partId ? _self.partId : partId // ignore: cast_nullable_to_non_nullable
as int,partName: null == partName ? _self.partName : partName // ignore: cast_nullable_to_non_nullable
as String,partNameSnapshot: null == partNameSnapshot ? _self.partNameSnapshot : partNameSnapshot // ignore: cast_nullable_to_non_nullable
as String,quantity: null == quantity ? _self.quantity : quantity // ignore: cast_nullable_to_non_nullable
as int,price: null == price ? _self.price : price // ignore: cast_nullable_to_non_nullable
as double,
  ));
}

}


/// Adds pattern-matching-related methods to [OrderItem].
extension OrderItemPatterns on OrderItem {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _OrderItem value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _OrderItem() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _OrderItem value)  $default,){
final _that = this;
switch (_that) {
case _OrderItem():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _OrderItem value)?  $default,){
final _that = this;
switch (_that) {
case _OrderItem() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int id,  int partId, @JsonKey(readValue: _readItemPartName)  String partName,  String partNameSnapshot,  int quantity,  double price)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _OrderItem() when $default != null:
return $default(_that.id,_that.partId,_that.partName,_that.partNameSnapshot,_that.quantity,_that.price);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int id,  int partId, @JsonKey(readValue: _readItemPartName)  String partName,  String partNameSnapshot,  int quantity,  double price)  $default,) {final _that = this;
switch (_that) {
case _OrderItem():
return $default(_that.id,_that.partId,_that.partName,_that.partNameSnapshot,_that.quantity,_that.price);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int id,  int partId, @JsonKey(readValue: _readItemPartName)  String partName,  String partNameSnapshot,  int quantity,  double price)?  $default,) {final _that = this;
switch (_that) {
case _OrderItem() when $default != null:
return $default(_that.id,_that.partId,_that.partName,_that.partNameSnapshot,_that.quantity,_that.price);case _:
  return null;

}
}

}

/// @nodoc

@JsonSerializable(fieldRename: FieldRename.snake)
class _OrderItem extends OrderItem {
  const _OrderItem({this.id = 0, this.partId = 0, @JsonKey(readValue: _readItemPartName) this.partName = '', this.partNameSnapshot = '', this.quantity = 0, this.price = 0.0}): super._();
  factory _OrderItem.fromJson(Map<String, dynamic> json) => _$OrderItemFromJson(json);

@override@JsonKey() final  int id;
@override@JsonKey() final  int partId;
@override@JsonKey(readValue: _readItemPartName) final  String partName;
@override@JsonKey() final  String partNameSnapshot;
@override@JsonKey() final  int quantity;
@override@JsonKey() final  double price;

/// Create a copy of OrderItem
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$OrderItemCopyWith<_OrderItem> get copyWith => __$OrderItemCopyWithImpl<_OrderItem>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$OrderItemToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _OrderItem&&(identical(other.id, id) || other.id == id)&&(identical(other.partId, partId) || other.partId == partId)&&(identical(other.partName, partName) || other.partName == partName)&&(identical(other.partNameSnapshot, partNameSnapshot) || other.partNameSnapshot == partNameSnapshot)&&(identical(other.quantity, quantity) || other.quantity == quantity)&&(identical(other.price, price) || other.price == price));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,partId,partName,partNameSnapshot,quantity,price);

@override
String toString() {
  return 'OrderItem(id: $id, partId: $partId, partName: $partName, partNameSnapshot: $partNameSnapshot, quantity: $quantity, price: $price)';
}


}

/// @nodoc
abstract mixin class _$OrderItemCopyWith<$Res> implements $OrderItemCopyWith<$Res> {
  factory _$OrderItemCopyWith(_OrderItem value, $Res Function(_OrderItem) _then) = __$OrderItemCopyWithImpl;
@override @useResult
$Res call({
 int id, int partId,@JsonKey(readValue: _readItemPartName) String partName, String partNameSnapshot, int quantity, double price
});




}
/// @nodoc
class __$OrderItemCopyWithImpl<$Res>
    implements _$OrderItemCopyWith<$Res> {
  __$OrderItemCopyWithImpl(this._self, this._then);

  final _OrderItem _self;
  final $Res Function(_OrderItem) _then;

/// Create a copy of OrderItem
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? partId = null,Object? partName = null,Object? partNameSnapshot = null,Object? quantity = null,Object? price = null,}) {
  return _then(_OrderItem(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as int,partId: null == partId ? _self.partId : partId // ignore: cast_nullable_to_non_nullable
as int,partName: null == partName ? _self.partName : partName // ignore: cast_nullable_to_non_nullable
as String,partNameSnapshot: null == partNameSnapshot ? _self.partNameSnapshot : partNameSnapshot // ignore: cast_nullable_to_non_nullable
as String,quantity: null == quantity ? _self.quantity : quantity // ignore: cast_nullable_to_non_nullable
as int,price: null == price ? _self.price : price // ignore: cast_nullable_to_non_nullable
as double,
  ));
}


}

/// @nodoc
mixin _$OrdersResponse {

 List<Order> get orders; int get total;
/// Create a copy of OrdersResponse
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$OrdersResponseCopyWith<OrdersResponse> get copyWith => _$OrdersResponseCopyWithImpl<OrdersResponse>(this as OrdersResponse, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is OrdersResponse&&const DeepCollectionEquality().equals(other.orders, orders)&&(identical(other.total, total) || other.total == total));
}


@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(orders),total);

@override
String toString() {
  return 'OrdersResponse(orders: $orders, total: $total)';
}


}

/// @nodoc
abstract mixin class $OrdersResponseCopyWith<$Res>  {
  factory $OrdersResponseCopyWith(OrdersResponse value, $Res Function(OrdersResponse) _then) = _$OrdersResponseCopyWithImpl;
@useResult
$Res call({
 List<Order> orders, int total
});




}
/// @nodoc
class _$OrdersResponseCopyWithImpl<$Res>
    implements $OrdersResponseCopyWith<$Res> {
  _$OrdersResponseCopyWithImpl(this._self, this._then);

  final OrdersResponse _self;
  final $Res Function(OrdersResponse) _then;

/// Create a copy of OrdersResponse
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? orders = null,Object? total = null,}) {
  return _then(_self.copyWith(
orders: null == orders ? _self.orders : orders // ignore: cast_nullable_to_non_nullable
as List<Order>,total: null == total ? _self.total : total // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

}


/// Adds pattern-matching-related methods to [OrdersResponse].
extension OrdersResponsePatterns on OrdersResponse {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _OrdersResponse value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _OrdersResponse() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _OrdersResponse value)  $default,){
final _that = this;
switch (_that) {
case _OrdersResponse():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _OrdersResponse value)?  $default,){
final _that = this;
switch (_that) {
case _OrdersResponse() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<Order> orders,  int total)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _OrdersResponse() when $default != null:
return $default(_that.orders,_that.total);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<Order> orders,  int total)  $default,) {final _that = this;
switch (_that) {
case _OrdersResponse():
return $default(_that.orders,_that.total);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<Order> orders,  int total)?  $default,) {final _that = this;
switch (_that) {
case _OrdersResponse() when $default != null:
return $default(_that.orders,_that.total);case _:
  return null;

}
}

}

/// @nodoc


class _OrdersResponse implements OrdersResponse {
  const _OrdersResponse({final  List<Order> orders = const [], this.total = 0}): _orders = orders;
  

 final  List<Order> _orders;
@override@JsonKey() List<Order> get orders {
  if (_orders is EqualUnmodifiableListView) return _orders;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_orders);
}

@override@JsonKey() final  int total;

/// Create a copy of OrdersResponse
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$OrdersResponseCopyWith<_OrdersResponse> get copyWith => __$OrdersResponseCopyWithImpl<_OrdersResponse>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _OrdersResponse&&const DeepCollectionEquality().equals(other._orders, _orders)&&(identical(other.total, total) || other.total == total));
}


@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(_orders),total);

@override
String toString() {
  return 'OrdersResponse(orders: $orders, total: $total)';
}


}

/// @nodoc
abstract mixin class _$OrdersResponseCopyWith<$Res> implements $OrdersResponseCopyWith<$Res> {
  factory _$OrdersResponseCopyWith(_OrdersResponse value, $Res Function(_OrdersResponse) _then) = __$OrdersResponseCopyWithImpl;
@override @useResult
$Res call({
 List<Order> orders, int total
});




}
/// @nodoc
class __$OrdersResponseCopyWithImpl<$Res>
    implements _$OrdersResponseCopyWith<$Res> {
  __$OrdersResponseCopyWithImpl(this._self, this._then);

  final _OrdersResponse _self;
  final $Res Function(_OrdersResponse) _then;

/// Create a copy of OrdersResponse
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? orders = null,Object? total = null,}) {
  return _then(_OrdersResponse(
orders: null == orders ? _self._orders : orders // ignore: cast_nullable_to_non_nullable
as List<Order>,total: null == total ? _self.total : total // ignore: cast_nullable_to_non_nullable
as int,
  ));
}


}

// dart format on
