import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:go_router/go_router.dart';
import 'package:qr_flutter/qr_flutter.dart';
import '../../../core/api/api_client.dart';
import '../../../core/utils/qr_signer.dart';
import '../providers/inventory_provider.dart';
import '../providers/part_catalog_provider.dart';
import '../../auth/providers/auth_provider.dart';
import '../../orders/widgets/part_order_sheet.dart';

class PartDetailScreen extends ConsumerWidget {
  final int id;
  const PartDetailScreen({super.key, required this.id});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final partAsync = ref.watch(partProvider(id));
    final user = ref.watch(authProvider).valueOrNull;

    // Читаем текущий оффлайн статус из фильтрованного списка
    final filter = ref.watch(inventoryFilterProvider);
    final inventoryAsync = ref.watch(inventoryProvider(filter));
    final isOffline = inventoryAsync.valueOrNull?.isOffline ?? false;

    final catalog = ref.watch(partCatalogProvider).valueOrNull;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Запчасть'),
        actions: [
          if (user?.isOperator == true)
            IconButton(
              icon: const Icon(Icons.edit_outlined),
              onPressed: isOffline
                  ? () => ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(
                        content: Text(
                          'В оффлайн-режиме редактирование недоступно',
                        ),
                      ),
                    )
                  : () => context.go('/inventory/edit/$id'),
            ),
        ],
      ),
      body: partAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) {
          String msg = 'Не удалось загрузить данные запчасти';
          if (e is DioException && e.response?.statusCode == 401) {
            msg = 'Сессия устарела. Пожалуйста, выполните вход заново.';
          } else if (e is DioException &&
              (e.type == DioExceptionType.connectionError ||
                  e.type == DioExceptionType.connectionTimeout)) {
            msg = 'Нет соединения с сервером';
          }
          return Center(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline,
                      size: 48, color: Colors.orangeAccent),
                  const SizedBox(height: 12),
                  Text(
                    msg,
                    textAlign: TextAlign.center,
                    style:
                        const TextStyle(color: Colors.white70, fontSize: 15),
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton.icon(
                    onPressed: () => ref.invalidate(partProvider(id)),
                    icon: const Icon(Icons.refresh),
                    label: const Text('Повторить'),
                  ),
                ],
              ),
            ),
          );
        },
        data: (part) {
          final categoryAttrs = catalog?.attributesForCategory(part.category);
          bool shows(String field) {
            if (categoryAttrs == null || categoryAttrs.isEmpty) {
              return true;
            }
            return categoryAttrs.contains(field);
          }
          String val(String? v) => _notEmpty(v) ? v! : '—';

          return RefreshIndicator(
            onRefresh: () async {
              ref.invalidate(partProvider(id));
              await ref.read(partProvider(id).future);
            },
            child: SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 100),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Галерея фото
                  if (part.photos.isNotEmpty)
                    SizedBox(
                      height: 220,
                      child: PageView.builder(
                        itemCount: part.photos.length,
                        itemBuilder: (_, i) => ClipRRect(
                          borderRadius: BorderRadius.circular(12),
                          child: CachedNetworkImage(
                            imageUrl: apiClient.resolveUrl(part.photos[i]),
                            fit: BoxFit.cover,
                            placeholder: (ctx, url) => Container(
                              color: const Color(0xFF16213E),
                              child: const Icon(
                                Icons.image_outlined,
                                size: 64,
                                color: Colors.white24,
                              ),
                            ),
                            errorWidget: (ctx, url, err) => Container(
                              color: const Color(0xFF16213E),
                              child: const Icon(
                                Icons.broken_image_outlined,
                                size: 64,
                                color: Colors.white24,
                              ),
                            ),
                          ),
                        ),
                      ),
                    )
                  else
                    Container(
                      height: 160,
                      decoration: BoxDecoration(
                        color: const Color(0xFF16213E),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Center(
                        child: Icon(
                          Icons.directions_car_outlined,
                          size: 64,
                          color: Colors.white24,
                        ),
                      ),
                    ),
                  const SizedBox(height: 16),
                  Text(
                    part.name,
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Text(
                        '${part.price.toStringAsFixed(0)} ₽',
                        style: const TextStyle(
                          color: Color(0xFF4F8EF7),
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const Spacer(),
                      _statusChip(part.quantity),
                    ],
                  ),
                  const SizedBox(height: 16),
                  _section('Основная информация', [
                    if (part.category.isNotEmpty) _row('Категория', part.category),
                    if (part.brand != null) _row('Бренд', part.brand!),
                    if (part.model != null) _row('Модель', part.model!),
                    if (part.bodyBrand != null) _row('Марка кузова', part.bodyBrand!),
                    if (part.condition != null) _row('Состояние', part.condition!),
                  ]),
                  _section('Характеристики', [
                    if (shows('engine_brand'))
                      _row('Марка двигателя', val(part.engineBrand)),
                    if (shows('car_release_date'))
                      _row('Год выпуска', val(part.carReleaseDate)),
                    if (part.vin != null) _row('VIN / Номер кузова', part.vin!),
                    if (shows('transmission'))
                      _row('Тип трансмиссии', val(part.transmission)),
                    if (shows('transmission_model'))
                      _row('Модель трансмиссии', val(part.transmissionModel)),
                    if (shows('drive')) _row('Привод', val(part.drive)),
                    if (part.color != null) _row('Цвет кузовных деталей', part.color!),
                    if (shows('front_rear'))
                      _row('Перед / зад', val(part.frontRear)),
                    if (shows('left_right'))
                      _row('Лево / право', val(part.leftRight)),
                    if (shows('top_bottom'))
                      _row('Верх / низ', val(part.topBottom)),
                    if (shows('number'))
                      _row('Номер детали', val(part.number)),
                    if (shows('manufacturer'))
                      _row('Производитель', val(part.manufacturer)),
                    if (shows('manufacturer_code'))
                      _row('Код производителя', val(part.manufacturerCode)),
                    if (shows('color'))
                      _row('Цвет', val(part.color)),
                    if (shows('condition'))
                      _row('Состояние', val(part.condition)),
                    if (shows('defect'))
                      _row('Дефект', val(part.defect)),
                    if (shows('supplier_code'))
                      _row('Код поставки', val(part.supplierCode)),
                    if (shows('wear_percentage'))
                      _row('Процент износа', _notEmpty(part.wearPercentage) ? '${part.wearPercentage}%' : '—'),
                    if (shows('season'))
                      _row('Сезон', val(part.season)),
                    if (shows('diameter'))
                      _row('Диаметр', val(part.diameter)),
                    if (shows('width'))
                      _row('Ширина', val(part.width)),
                    if (shows('profile'))
                      _row('Профиль', val(part.profile)),
                    if (shows('tire_quantity'))
                      _row('Количество шин', val(part.tireQuantity)),
                    if (shows('drilling'))
                      _row('Сверловка', val(part.drilling)),
                    if (shows('offset'))
                      _row('Вылет', val(part.offset)),
                    if (shows('center_hole_diameter'))
                      _row('Центральное отверстие', val(part.centerHoleDiameter)),
                    if (shows('tire_model'))
                      _row('Модель шины', val(part.tireModel)),
                  ]),
                if (part.oemCode != null || part.supplierCode != null)
                  _section('Коды', [
                    if (part.oemCode != null) _row('OEM код', part.oemCode!),
                    if (part.supplierCode != null)
                      _row('Код поставщика', part.supplierCode!),
                  ]),
                _section('Расположение', [
                  if (part.location.isNotEmpty) _row('Место', part.location),
                  if (part.address.isNotEmpty) _row('Адрес склада', part.address),
                  if (part.salesman != null) _row('Продавец', part.salesman!),
                ]),
                if (part.description != null && part.description!.isNotEmpty)
                  _section('Описание', [
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Text(
                        part.description!,
                        style: const TextStyle(color: Colors.white70),
                      ),
                    ),
                  ]),
                _section('QR-код запчасти', [
                  Center(
                    child: Container(
                      margin: const EdgeInsets.symmetric(vertical: 12),
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: QrImageView(
                        data: QrSigner.generateQrData(
                          partId: part.id,
                          createdAt: part.createdAt,
                        ),
                        version: QrVersions.auto,
                        size: 160.0,
                      ),
                    ),
                  ),
                  const Center(
                    child: Text(
                      'Отсканируйте код для быстрого поиска',
                      style: TextStyle(color: Colors.white38, fontSize: 11),
                    ),
                  ),
                ]),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton.icon(
                    onPressed: part.quantity > 0 && !isOffline
                        ? () async {
                            final created = await showPartOrderSheet(
                              context,
                              ref,
                              part,
                            );
                            if (created && context.mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(content: Text('Заказ оформлен')),
                              );
                            }
                          }
                        : null,
                    icon: const Icon(Icons.shopping_cart_checkout_rounded),
                    label: Text(
                      isOffline
                          ? 'Заказ недоступен оффлайн'
                          : part.quantity > 0
                          ? 'Оформить заказ'
                          : 'Нет в наличии',
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
        },
      ),
    );
  }

  Widget _statusChip(int qty) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
    decoration: BoxDecoration(
      color: qty > 0
          ? const Color(0xFF43A047).withValues(alpha: 0.2)
          : const Color(0xFFE53935).withValues(alpha: 0.2),
      borderRadius: BorderRadius.circular(8),
    ),
    child: Text(
      qty > 0 ? '$qty шт. в наличии' : 'Нет в наличии',
      style: TextStyle(
        color: qty > 0 ? const Color(0xFF43A047) : const Color(0xFFE53935),
        fontWeight: FontWeight.w600,
      ),
    ),
  );

  bool _notEmpty(String? value) => value != null && value.isNotEmpty;

  Widget _section(String title, List<Widget> children) => children.isEmpty
      ? const SizedBox.shrink()
      : Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Divider(color: Colors.white12),
            const SizedBox(height: 4),
            Text(
              title,
              style: const TextStyle(
                color: Colors.white54,
                fontSize: 12,
                fontWeight: FontWeight.w600,
                letterSpacing: 0.5,
              ),
            ),
            const SizedBox(height: 8),
            ...children,
            const SizedBox(height: 8),
          ],
        );

  Widget _row(String label, String value) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 4),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 130,
          child: Text(
            label,
            style: const TextStyle(color: Colors.white38, fontSize: 13),
          ),
        ),
        Expanded(
          child: Text(
            value,
            style: const TextStyle(color: Colors.white, fontSize: 13),
          ),
        ),
      ],
    ),
  );
}
