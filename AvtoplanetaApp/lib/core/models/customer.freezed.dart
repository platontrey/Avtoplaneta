// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'customer.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Customer {

 int get id; String get name; String get phone; String get city;@JsonKey(readValue: _readPreferredTk) String get preferredTk;@JsonKey(readValue: _readTkDetails) String get tkDetails; String get notes; String get category;@JsonKey(readValue: _readDiscountPercent) double get discountPercent;@JsonKey(readValue: _readTotalOrders) int get totalOrders;@JsonKey(readValue: _readTotalSpent) double get totalSpent;@JsonKey(readValue: _readCreatedAt) DateTime? get createdAt;@JsonKey(readValue: _readUpdatedAt) DateTime? get updatedAt;
/// Create a copy of Customer
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$CustomerCopyWith<Customer> get copyWith => _$CustomerCopyWithImpl<Customer>(this as Customer, _$identity);

  /// Serializes this Customer to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Customer&&(identical(other.id, id) || other.id == id)&&(identical(other.name, name) || other.name == name)&&(identical(other.phone, phone) || other.phone == phone)&&(identical(other.city, city) || other.city == city)&&(identical(other.preferredTk, preferredTk) || other.preferredTk == preferredTk)&&(identical(other.tkDetails, tkDetails) || other.tkDetails == tkDetails)&&(identical(other.notes, notes) || other.notes == notes)&&(identical(other.category, category) || other.category == category)&&(identical(other.discountPercent, discountPercent) || other.discountPercent == discountPercent)&&(identical(other.totalOrders, totalOrders) || other.totalOrders == totalOrders)&&(identical(other.totalSpent, totalSpent) || other.totalSpent == totalSpent)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,name,phone,city,preferredTk,tkDetails,notes,category,discountPercent,totalOrders,totalSpent,createdAt,updatedAt);

@override
String toString() {
  return 'Customer(id: $id, name: $name, phone: $phone, city: $city, preferredTk: $preferredTk, tkDetails: $tkDetails, notes: $notes, category: $category, discountPercent: $discountPercent, totalOrders: $totalOrders, totalSpent: $totalSpent, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class $CustomerCopyWith<$Res>  {
  factory $CustomerCopyWith(Customer value, $Res Function(Customer) _then) = _$CustomerCopyWithImpl;
@useResult
$Res call({
 int id, String name, String phone, String city,@JsonKey(readValue: _readPreferredTk) String preferredTk,@JsonKey(readValue: _readTkDetails) String tkDetails, String notes, String category,@JsonKey(readValue: _readDiscountPercent) double discountPercent,@JsonKey(readValue: _readTotalOrders) int totalOrders,@JsonKey(readValue: _readTotalSpent) double totalSpent,@JsonKey(readValue: _readCreatedAt) DateTime? createdAt,@JsonKey(readValue: _readUpdatedAt) DateTime? updatedAt
});




}
/// @nodoc
class _$CustomerCopyWithImpl<$Res>
    implements $CustomerCopyWith<$Res> {
  _$CustomerCopyWithImpl(this._self, this._then);

  final Customer _self;
  final $Res Function(Customer) _then;

/// Create a copy of Customer
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? name = null,Object? phone = null,Object? city = null,Object? preferredTk = null,Object? tkDetails = null,Object? notes = null,Object? category = null,Object? discountPercent = null,Object? totalOrders = null,Object? totalSpent = null,Object? createdAt = freezed,Object? updatedAt = freezed,}) {
  return _then(_self.copyWith(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as int,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,phone: null == phone ? _self.phone : phone // ignore: cast_nullable_to_non_nullable
as String,city: null == city ? _self.city : city // ignore: cast_nullable_to_non_nullable
as String,preferredTk: null == preferredTk ? _self.preferredTk : preferredTk // ignore: cast_nullable_to_non_nullable
as String,tkDetails: null == tkDetails ? _self.tkDetails : tkDetails // ignore: cast_nullable_to_non_nullable
as String,notes: null == notes ? _self.notes : notes // ignore: cast_nullable_to_non_nullable
as String,category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as String,discountPercent: null == discountPercent ? _self.discountPercent : discountPercent // ignore: cast_nullable_to_non_nullable
as double,totalOrders: null == totalOrders ? _self.totalOrders : totalOrders // ignore: cast_nullable_to_non_nullable
as int,totalSpent: null == totalSpent ? _self.totalSpent : totalSpent // ignore: cast_nullable_to_non_nullable
as double,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

}


/// Adds pattern-matching-related methods to [Customer].
extension CustomerPatterns on Customer {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Customer value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Customer() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Customer value)  $default,){
final _that = this;
switch (_that) {
case _Customer():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Customer value)?  $default,){
final _that = this;
switch (_that) {
case _Customer() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int id,  String name,  String phone,  String city, @JsonKey(readValue: _readPreferredTk)  String preferredTk, @JsonKey(readValue: _readTkDetails)  String tkDetails,  String notes,  String category, @JsonKey(readValue: _readDiscountPercent)  double discountPercent, @JsonKey(readValue: _readTotalOrders)  int totalOrders, @JsonKey(readValue: _readTotalSpent)  double totalSpent, @JsonKey(readValue: _readCreatedAt)  DateTime? createdAt, @JsonKey(readValue: _readUpdatedAt)  DateTime? updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Customer() when $default != null:
return $default(_that.id,_that.name,_that.phone,_that.city,_that.preferredTk,_that.tkDetails,_that.notes,_that.category,_that.discountPercent,_that.totalOrders,_that.totalSpent,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int id,  String name,  String phone,  String city, @JsonKey(readValue: _readPreferredTk)  String preferredTk, @JsonKey(readValue: _readTkDetails)  String tkDetails,  String notes,  String category, @JsonKey(readValue: _readDiscountPercent)  double discountPercent, @JsonKey(readValue: _readTotalOrders)  int totalOrders, @JsonKey(readValue: _readTotalSpent)  double totalSpent, @JsonKey(readValue: _readCreatedAt)  DateTime? createdAt, @JsonKey(readValue: _readUpdatedAt)  DateTime? updatedAt)  $default,) {final _that = this;
switch (_that) {
case _Customer():
return $default(_that.id,_that.name,_that.phone,_that.city,_that.preferredTk,_that.tkDetails,_that.notes,_that.category,_that.discountPercent,_that.totalOrders,_that.totalSpent,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int id,  String name,  String phone,  String city, @JsonKey(readValue: _readPreferredTk)  String preferredTk, @JsonKey(readValue: _readTkDetails)  String tkDetails,  String notes,  String category, @JsonKey(readValue: _readDiscountPercent)  double discountPercent, @JsonKey(readValue: _readTotalOrders)  int totalOrders, @JsonKey(readValue: _readTotalSpent)  double totalSpent, @JsonKey(readValue: _readCreatedAt)  DateTime? createdAt, @JsonKey(readValue: _readUpdatedAt)  DateTime? updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _Customer() when $default != null:
return $default(_that.id,_that.name,_that.phone,_that.city,_that.preferredTk,_that.tkDetails,_that.notes,_that.category,_that.discountPercent,_that.totalOrders,_that.totalSpent,_that.createdAt,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Customer extends Customer {
  const _Customer({this.id = 0, this.name = '', this.phone = '', this.city = '', @JsonKey(readValue: _readPreferredTk) this.preferredTk = '', @JsonKey(readValue: _readTkDetails) this.tkDetails = '', this.notes = '', this.category = 'regular', @JsonKey(readValue: _readDiscountPercent) this.discountPercent = 0.0, @JsonKey(readValue: _readTotalOrders) this.totalOrders = 0, @JsonKey(readValue: _readTotalSpent) this.totalSpent = 0.0, @JsonKey(readValue: _readCreatedAt) this.createdAt, @JsonKey(readValue: _readUpdatedAt) this.updatedAt}): super._();
  factory _Customer.fromJson(Map<String, dynamic> json) => _$CustomerFromJson(json);

@override@JsonKey() final  int id;
@override@JsonKey() final  String name;
@override@JsonKey() final  String phone;
@override@JsonKey() final  String city;
@override@JsonKey(readValue: _readPreferredTk) final  String preferredTk;
@override@JsonKey(readValue: _readTkDetails) final  String tkDetails;
@override@JsonKey() final  String notes;
@override@JsonKey() final  String category;
@override@JsonKey(readValue: _readDiscountPercent) final  double discountPercent;
@override@JsonKey(readValue: _readTotalOrders) final  int totalOrders;
@override@JsonKey(readValue: _readTotalSpent) final  double totalSpent;
@override@JsonKey(readValue: _readCreatedAt) final  DateTime? createdAt;
@override@JsonKey(readValue: _readUpdatedAt) final  DateTime? updatedAt;

/// Create a copy of Customer
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$CustomerCopyWith<_Customer> get copyWith => __$CustomerCopyWithImpl<_Customer>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$CustomerToJson(this, );
}

@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _Customer&&(identical(other.id, id) || other.id == id)&&(identical(other.name, name) || other.name == name)&&(identical(other.phone, phone) || other.phone == phone)&&(identical(other.city, city) || other.city == city)&&(identical(other.preferredTk, preferredTk) || other.preferredTk == preferredTk)&&(identical(other.tkDetails, tkDetails) || other.tkDetails == tkDetails)&&(identical(other.notes, notes) || other.notes == notes)&&(identical(other.category, category) || other.category == category)&&(identical(other.discountPercent, discountPercent) || other.discountPercent == discountPercent)&&(identical(other.totalOrders, totalOrders) || other.totalOrders == totalOrders)&&(identical(other.totalSpent, totalSpent) || other.totalSpent == totalSpent)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode => Object.hash(runtimeType,id,name,phone,city,preferredTk,tkDetails,notes,category,discountPercent,totalOrders,totalSpent,createdAt,updatedAt);

@override
String toString() {
  return 'Customer(id: $id, name: $name, phone: $phone, city: $city, preferredTk: $preferredTk, tkDetails: $tkDetails, notes: $notes, category: $category, discountPercent: $discountPercent, totalOrders: $totalOrders, totalSpent: $totalSpent, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$CustomerCopyWith<$Res> implements $CustomerCopyWith<$Res> {
  factory _$CustomerCopyWith(_Customer value, $Res Function(_Customer) _then) = __$CustomerCopyWithImpl;
@override @useResult
$Res call({
 int id, String name, String phone, String city,@JsonKey(readValue: _readPreferredTk) String preferredTk,@JsonKey(readValue: _readTkDetails) String tkDetails, String notes, String category,@JsonKey(readValue: _readDiscountPercent) double discountPercent,@JsonKey(readValue: _readTotalOrders) int totalOrders,@JsonKey(readValue: _readTotalSpent) double totalSpent,@JsonKey(readValue: _readCreatedAt) DateTime? createdAt,@JsonKey(readValue: _readUpdatedAt) DateTime? updatedAt
});




}
/// @nodoc
class __$CustomerCopyWithImpl<$Res>
    implements _$CustomerCopyWith<$Res> {
  __$CustomerCopyWithImpl(this._self, this._then);

  final _Customer _self;
  final $Res Function(_Customer) _then;

/// Create a copy of Customer
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? name = null,Object? phone = null,Object? city = null,Object? preferredTk = null,Object? tkDetails = null,Object? notes = null,Object? category = null,Object? discountPercent = null,Object? totalOrders = null,Object? totalSpent = null,Object? createdAt = freezed,Object? updatedAt = freezed,}) {
  return _then(_Customer(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as int,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,phone: null == phone ? _self.phone : phone // ignore: cast_nullable_to_non_nullable
as String,city: null == city ? _self.city : city // ignore: cast_nullable_to_non_nullable
as String,preferredTk: null == preferredTk ? _self.preferredTk : preferredTk // ignore: cast_nullable_to_non_nullable
as String,tkDetails: null == tkDetails ? _self.tkDetails : tkDetails // ignore: cast_nullable_to_non_nullable
as String,notes: null == notes ? _self.notes : notes // ignore: cast_nullable_to_non_nullable
as String,category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as String,discountPercent: null == discountPercent ? _self.discountPercent : discountPercent // ignore: cast_nullable_to_non_nullable
as double,totalOrders: null == totalOrders ? _self.totalOrders : totalOrders // ignore: cast_nullable_to_non_nullable
as int,totalSpent: null == totalSpent ? _self.totalSpent : totalSpent // ignore: cast_nullable_to_non_nullable
as double,createdAt: freezed == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime?,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}


}

/// @nodoc
mixin _$CustomerDetails {

 Customer get customer; List<Order> get orders;
/// Create a copy of CustomerDetails
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$CustomerDetailsCopyWith<CustomerDetails> get copyWith => _$CustomerDetailsCopyWithImpl<CustomerDetails>(this as CustomerDetails, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is CustomerDetails&&(identical(other.customer, customer) || other.customer == customer)&&const DeepCollectionEquality().equals(other.orders, orders));
}


@override
int get hashCode => Object.hash(runtimeType,customer,const DeepCollectionEquality().hash(orders));

@override
String toString() {
  return 'CustomerDetails(customer: $customer, orders: $orders)';
}


}

/// @nodoc
abstract mixin class $CustomerDetailsCopyWith<$Res>  {
  factory $CustomerDetailsCopyWith(CustomerDetails value, $Res Function(CustomerDetails) _then) = _$CustomerDetailsCopyWithImpl;
@useResult
$Res call({
 Customer customer, List<Order> orders
});


$CustomerCopyWith<$Res> get customer;

}
/// @nodoc
class _$CustomerDetailsCopyWithImpl<$Res>
    implements $CustomerDetailsCopyWith<$Res> {
  _$CustomerDetailsCopyWithImpl(this._self, this._then);

  final CustomerDetails _self;
  final $Res Function(CustomerDetails) _then;

/// Create a copy of CustomerDetails
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? customer = null,Object? orders = null,}) {
  return _then(_self.copyWith(
customer: null == customer ? _self.customer : customer // ignore: cast_nullable_to_non_nullable
as Customer,orders: null == orders ? _self.orders : orders // ignore: cast_nullable_to_non_nullable
as List<Order>,
  ));
}
/// Create a copy of CustomerDetails
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$CustomerCopyWith<$Res> get customer {
  
  return $CustomerCopyWith<$Res>(_self.customer, (value) {
    return _then(_self.copyWith(customer: value));
  });
}
}


/// Adds pattern-matching-related methods to [CustomerDetails].
extension CustomerDetailsPatterns on CustomerDetails {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _CustomerDetails value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _CustomerDetails() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _CustomerDetails value)  $default,){
final _that = this;
switch (_that) {
case _CustomerDetails():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _CustomerDetails value)?  $default,){
final _that = this;
switch (_that) {
case _CustomerDetails() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Customer customer,  List<Order> orders)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _CustomerDetails() when $default != null:
return $default(_that.customer,_that.orders);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Customer customer,  List<Order> orders)  $default,) {final _that = this;
switch (_that) {
case _CustomerDetails():
return $default(_that.customer,_that.orders);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Customer customer,  List<Order> orders)?  $default,) {final _that = this;
switch (_that) {
case _CustomerDetails() when $default != null:
return $default(_that.customer,_that.orders);case _:
  return null;

}
}

}

/// @nodoc


class _CustomerDetails implements CustomerDetails {
  const _CustomerDetails({required this.customer, final  List<Order> orders = const []}): _orders = orders;
  

@override final  Customer customer;
 final  List<Order> _orders;
@override@JsonKey() List<Order> get orders {
  if (_orders is EqualUnmodifiableListView) return _orders;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_orders);
}


/// Create a copy of CustomerDetails
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$CustomerDetailsCopyWith<_CustomerDetails> get copyWith => __$CustomerDetailsCopyWithImpl<_CustomerDetails>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _CustomerDetails&&(identical(other.customer, customer) || other.customer == customer)&&const DeepCollectionEquality().equals(other._orders, _orders));
}


@override
int get hashCode => Object.hash(runtimeType,customer,const DeepCollectionEquality().hash(_orders));

@override
String toString() {
  return 'CustomerDetails(customer: $customer, orders: $orders)';
}


}

/// @nodoc
abstract mixin class _$CustomerDetailsCopyWith<$Res> implements $CustomerDetailsCopyWith<$Res> {
  factory _$CustomerDetailsCopyWith(_CustomerDetails value, $Res Function(_CustomerDetails) _then) = __$CustomerDetailsCopyWithImpl;
@override @useResult
$Res call({
 Customer customer, List<Order> orders
});


@override $CustomerCopyWith<$Res> get customer;

}
/// @nodoc
class __$CustomerDetailsCopyWithImpl<$Res>
    implements _$CustomerDetailsCopyWith<$Res> {
  __$CustomerDetailsCopyWithImpl(this._self, this._then);

  final _CustomerDetails _self;
  final $Res Function(_CustomerDetails) _then;

/// Create a copy of CustomerDetails
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? customer = null,Object? orders = null,}) {
  return _then(_CustomerDetails(
customer: null == customer ? _self.customer : customer // ignore: cast_nullable_to_non_nullable
as Customer,orders: null == orders ? _self._orders : orders // ignore: cast_nullable_to_non_nullable
as List<Order>,
  ));
}

/// Create a copy of CustomerDetails
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$CustomerCopyWith<$Res> get customer {
  
  return $CustomerCopyWith<$Res>(_self.customer, (value) {
    return _then(_self.copyWith(customer: value));
  });
}
}

/// @nodoc
mixin _$CustomersResponse {

 List<Customer> get customers; int get total;
/// Create a copy of CustomersResponse
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$CustomersResponseCopyWith<CustomersResponse> get copyWith => _$CustomersResponseCopyWithImpl<CustomersResponse>(this as CustomersResponse, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is CustomersResponse&&const DeepCollectionEquality().equals(other.customers, customers)&&(identical(other.total, total) || other.total == total));
}


@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(customers),total);

@override
String toString() {
  return 'CustomersResponse(customers: $customers, total: $total)';
}


}

/// @nodoc
abstract mixin class $CustomersResponseCopyWith<$Res>  {
  factory $CustomersResponseCopyWith(CustomersResponse value, $Res Function(CustomersResponse) _then) = _$CustomersResponseCopyWithImpl;
@useResult
$Res call({
 List<Customer> customers, int total
});




}
/// @nodoc
class _$CustomersResponseCopyWithImpl<$Res>
    implements $CustomersResponseCopyWith<$Res> {
  _$CustomersResponseCopyWithImpl(this._self, this._then);

  final CustomersResponse _self;
  final $Res Function(CustomersResponse) _then;

/// Create a copy of CustomersResponse
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? customers = null,Object? total = null,}) {
  return _then(_self.copyWith(
customers: null == customers ? _self.customers : customers // ignore: cast_nullable_to_non_nullable
as List<Customer>,total: null == total ? _self.total : total // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

}


/// Adds pattern-matching-related methods to [CustomersResponse].
extension CustomersResponsePatterns on CustomersResponse {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _CustomersResponse value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _CustomersResponse() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _CustomersResponse value)  $default,){
final _that = this;
switch (_that) {
case _CustomersResponse():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _CustomersResponse value)?  $default,){
final _that = this;
switch (_that) {
case _CustomersResponse() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<Customer> customers,  int total)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _CustomersResponse() when $default != null:
return $default(_that.customers,_that.total);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<Customer> customers,  int total)  $default,) {final _that = this;
switch (_that) {
case _CustomersResponse():
return $default(_that.customers,_that.total);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<Customer> customers,  int total)?  $default,) {final _that = this;
switch (_that) {
case _CustomersResponse() when $default != null:
return $default(_that.customers,_that.total);case _:
  return null;

}
}

}

/// @nodoc


class _CustomersResponse implements CustomersResponse {
  const _CustomersResponse({final  List<Customer> customers = const [], this.total = 0}): _customers = customers;
  

 final  List<Customer> _customers;
@override@JsonKey() List<Customer> get customers {
  if (_customers is EqualUnmodifiableListView) return _customers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_customers);
}

@override@JsonKey() final  int total;

/// Create a copy of CustomersResponse
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$CustomersResponseCopyWith<_CustomersResponse> get copyWith => __$CustomersResponseCopyWithImpl<_CustomersResponse>(this, _$identity);



@override
bool operator ==(Object other) {
  return identical(this, other) || (other.runtimeType == runtimeType&&other is _CustomersResponse&&const DeepCollectionEquality().equals(other._customers, _customers)&&(identical(other.total, total) || other.total == total));
}


@override
int get hashCode => Object.hash(runtimeType,const DeepCollectionEquality().hash(_customers),total);

@override
String toString() {
  return 'CustomersResponse(customers: $customers, total: $total)';
}


}

/// @nodoc
abstract mixin class _$CustomersResponseCopyWith<$Res> implements $CustomersResponseCopyWith<$Res> {
  factory _$CustomersResponseCopyWith(_CustomersResponse value, $Res Function(_CustomersResponse) _then) = __$CustomersResponseCopyWithImpl;
@override @useResult
$Res call({
 List<Customer> customers, int total
});




}
/// @nodoc
class __$CustomersResponseCopyWithImpl<$Res>
    implements _$CustomersResponseCopyWith<$Res> {
  __$CustomersResponseCopyWithImpl(this._self, this._then);

  final _CustomersResponse _self;
  final $Res Function(_CustomersResponse) _then;

/// Create a copy of CustomersResponse
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? customers = null,Object? total = null,}) {
  return _then(_CustomersResponse(
customers: null == customers ? _self._customers : customers // ignore: cast_nullable_to_non_nullable
as List<Customer>,total: null == total ? _self.total : total // ignore: cast_nullable_to_non_nullable
as int,
  ));
}


}

// dart format on
