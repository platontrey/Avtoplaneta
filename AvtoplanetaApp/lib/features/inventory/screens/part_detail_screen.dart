import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:go_router/go_router.dart';
import 'package:qr_flutter/qr_flutter.dart';
import '../../../core/api/api_client.dart';
import '../../../core/utils/qr_signer.dart';
import '../widgets/photo_viewer_dialog.dart';
import 'package:share_plus/share_plus.dart';
import '../../../app/theme.dart';
import '../providers/inventory_provider.dart';
import '../../auth/providers/auth_provider.dart';
import '../../orders/widgets/part_order_sheet.dart';
import '../widgets/photo_editor_screen.dart';

class PartDetailScreen extends ConsumerStatefulWidget {
  final int id;
  const PartDetailScreen({super.key, required this.id});

  @override
  ConsumerState<PartDetailScreen> createState() => _PartDetailScreenState();
}

class _PartDetailScreenState extends ConsumerState<PartDetailScreen> {
  int _carouselIndex = 0;

  @override
  Widget build(BuildContext context) {
    final partAsync = ref.watch(partProvider(widget.id));
    final user = ref.watch(authProvider).valueOrNull;

    // Читаем текущий оффлайн статус из фильтрованного списка
    final filter = ref.watch(inventoryFilterProvider);
    final inventoryAsync = ref.watch(inventoryProvider(filter));
    final isOffline = inventoryAsync.valueOrNull?.isOffline ?? false;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Запчасть'),
        actions: [
          partAsync.maybeWhen(
            data: (part) => part.photos.isNotEmpty
                ? Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (user?.isOperator == true)
                        IconButton(
                          tooltip: 'Редактировать фото (маркер, размытие, обрезка)',
                          icon: const Icon(LucideIcons.wand_sparkles, color: AppTheme.accentPurple),
                          onPressed: isOffline
                              ? () => ScaffoldMessenger.of(context).showSnackBar(
                                    const SnackBar(content: Text('В оффлайн-режиме редактирование недоступно')),
                                  )
                              : () async {
                                  final idx = _carouselIndex.clamp(0, part.photos.length - 1);
                                  await PhotoEditorScreen.show(
                                    context,
                                    imageUrl: apiClient.resolveUrl(part.photos[idx]),
                                    partId: part.id,
                                    photoPathToReplace: part.photos[idx],
                                  );
                                },
                        ),
                      IconButton(
                        tooltip: 'Поделиться фото',
                        icon: const Icon(LucideIcons.share_2),
                        onPressed: () {
                          final idx = _carouselIndex.clamp(0, part.photos.length - 1);
                          final url = apiClient.resolveUrl(part.photos[idx]);
                          final uri = Uri.tryParse(url);
                          if (uri != null) {
                            SharePlus.instance.share(ShareParams(uri: uri, subject: part.name));
                          }
                        },
                      ),
                    ],
                  )
                : const SizedBox.shrink(),
            orElse: () => const SizedBox.shrink(),
          ),
          if (user?.isOperator == true)
            IconButton(
              icon: const Icon(LucideIcons.pencil),
              tooltip: 'Редактировать параметры запчасти',
              onPressed: isOffline
                  ? () => ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(
                        content: Text(
                          'В оффлайн-режиме редактирование недоступно',
                        ),
                      ),
                    )
                  : () => context.go('/inventory/edit/${widget.id}'),
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
                  const Icon(LucideIcons.circle_alert,
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
                    onPressed: () => ref.invalidate(partProvider(widget.id)),
                    icon: const Icon(LucideIcons.refresh_cw),
                    label: const Text('Повторить'),
                  ),
                ],
              ),
            ),
          );
        },
        data: (part) {
          String formatFrontRear(String? v) {
            if (!_notEmpty(v)) return '—';
            final upper = v!.trim().toUpperCase();
            if (upper == 'F') return 'F (Перед)';
            if (upper == 'R') return 'R (Зад)';
            return v;
          }
          String formatLeftRight(String? v) {
            if (!_notEmpty(v)) return '—';
            final upper = v!.trim().toUpperCase();
            if (upper == 'L') return 'L (Лево)';
            if (upper == 'R') return 'R (Право)';
            return v;
          }

          return RefreshIndicator(
            onRefresh: () async {
              ref.invalidate(partProvider(widget.id));
              await ref.read(partProvider(widget.id).future);
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
                      child: Stack(
                        children: [
                          PageView.builder(
                            itemCount: part.photos.length,
                            onPageChanged: (idx) => setState(() => _carouselIndex = idx),
                            itemBuilder: (_, i) => GestureDetector(
                              onTap: () => PhotoViewerDialog.show(
                                context,
                                photos: part.photos,
                                initialIndex: i,
                                title: part.name,
                                partId: part.id,
                              ),
                              child: ClipRRect(
                                borderRadius: BorderRadius.circular(12),
                                child: CachedNetworkImage(
                                  imageUrl: apiClient.resolveUrl(part.photos[i]),
                                  fit: BoxFit.cover,
                                  placeholder: (ctx, url) => Container(
                                    color: AppTheme.cardColor,
                                    child: const Icon(
                                      LucideIcons.image,
                                      size: 64,
                                      color: Colors.white24,
                                    ),
                                  ),
                                  errorWidget: (ctx, url, err) => Container(
                                    color: AppTheme.cardColor,
                                    child: const Icon(
                                      LucideIcons.image_off,
                                      size: 64,
                                      color: Colors.white24,
                                    ),
                                  ),
                                ),
                              ),
                            ),
                          ),
                          // Индикатор индекса фото
                          if (part.photos.length > 1)
                            Positioned(
                              bottom: 8,
                              left: 8,
                              child: Container(
                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                decoration: BoxDecoration(
                                  color: Colors.black54,
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Text(
                                  '${_carouselIndex + 1} / ${part.photos.length}',
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 11,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ),
                          // Кнопки управления фото (Редактировать и Увеличить)
                          Positioned(
                            bottom: 8,
                            right: 8,
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                if (user?.isOperator == true) ...[
                                  GestureDetector(
                                    onTap: isOffline
                                        ? () => ScaffoldMessenger.of(context).showSnackBar(
                                              const SnackBar(content: Text('В оффлайн-режиме редактирование недоступно')),
                                            )
                                        : () async {
                                            final idx = _carouselIndex.clamp(0, part.photos.length - 1);
                                            await PhotoEditorScreen.show(
                                              context,
                                              imageUrl: apiClient.resolveUrl(part.photos[idx]),
                                              partId: part.id,
                                              photoPathToReplace: part.photos[idx],
                                            );
                                          },
                                    child: Container(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 10,
                                        vertical: 5,
                                      ),
                                      decoration: BoxDecoration(
                                        color: AppTheme.primaryColor.withValues(alpha: 0.92),
                                        borderRadius: BorderRadius.circular(16),
                                        boxShadow: const [BoxShadow(color: Colors.black45, blurRadius: 4)],
                                      ),
                                      child: const Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          Icon(LucideIcons.wand_sparkles, size: 14, color: Colors.white),
                                          SizedBox(width: 4),
                                          Text(
                                            'Редактировать',
                                            style: TextStyle(
                                              color: Colors.white,
                                              fontSize: 11,
                                              fontWeight: FontWeight.w600,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 6),
                                ],
                                GestureDetector(
                                  onTap: () => PhotoViewerDialog.show(
                                    context,
                                    photos: part.photos,
                                    initialIndex: _carouselIndex,
                                    title: part.name,
                                    partId: part.id,
                                  ),
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 10,
                                      vertical: 5,
                                    ),
                                    decoration: BoxDecoration(
                                      color: Colors.black54,
                                      borderRadius: BorderRadius.circular(16),
                                      boxShadow: const [BoxShadow(color: Colors.black45, blurRadius: 4)],
                                    ),
                                    child: const Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(LucideIcons.zoom_in, size: 14, color: Colors.white),
                                        SizedBox(width: 4),
                                        Text(
                                          'Увеличить',
                                          style: TextStyle(
                                            color: Colors.white,
                                            fontSize: 11,
                                            fontWeight: FontWeight.w600,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    )
                  else
                    Container(
                      height: 160,
                      decoration: BoxDecoration(
                        color: AppTheme.cardColor,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: AppTheme.borderColor),
                      ),
                      child: const Center(
                        child: Icon(
                          LucideIcons.car,
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
                  const SizedBox(height: 10),
                  Row(
                    children: [
                      Text(
                        part.price > 0
                            ? '${part.price.toStringAsFixed(0)} ₽'
                            : 'отсутствует',
                        style: TextStyle(
                          color: part.price > 0 ? AppTheme.primaryColor : AppTheme.warningColor,
                          fontSize: part.price > 0 ? 26 : 18,
                          fontWeight: FontWeight.w800,
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
                  ]),
                  _section('Автомобиль', [
                    if (_notEmpty(part.engineBrand)) _row('Марка двигателя', part.engineBrand!),
                    if (_notEmpty(part.carReleaseDate)) _row('Год выпуска', part.carReleaseDate!),
                    if (_notEmpty(part.vin)) _row('VIN / Кузов', part.vin!),
                    if (_notEmpty(part.carReleasePeriod)) _row('Период выпуска', part.carReleasePeriod!),
                    if (_notEmpty(part.transmission)) _row('Тип трансмиссии', part.transmission!),
                    if (_notEmpty(part.transmissionModel)) _row('Модель трансмиссии', part.transmissionModel!),
                    if (_notEmpty(part.drive)) _row('Привод', part.drive!),
                  ]),
                  _section('Параметры детали', [
                    if (_notEmpty(part.frontRear)) _row('Перед / зад', formatFrontRear(part.frontRear)),
                    if (_notEmpty(part.leftRight)) _row('Лево / право', formatLeftRight(part.leftRight)),
                    if (_notEmpty(part.topBottom)) _row('Верх / низ', part.topBottom!),
                    if (_notEmpty(part.number)) _row('Номер детали', part.number!),
                    if (_notEmpty(part.manufacturer)) _row('Производитель', part.manufacturer!),
                    if (_notEmpty(part.manufacturerCode)) _row('Код производителя', part.manufacturerCode!),
                    if (_notEmpty(part.color)) _row('Цвет', part.color!),
                    if (_notEmpty(part.condition)) _row('Состояние', part.condition!),
                    if (_notEmpty(part.defect)) _row('Дефект', part.defect!),
                  ]),
                  if (_notEmpty(part.wearPercentage) ||
                      _notEmpty(part.season) ||
                      _notEmpty(part.diameter) ||
                      _notEmpty(part.width) ||
                      _notEmpty(part.profile) ||
                      _notEmpty(part.tireQuantity) ||
                      _notEmpty(part.drilling) ||
                      _notEmpty(part.offset) ||
                      _notEmpty(part.centerHoleDiameter) ||
                      _notEmpty(part.tireModel))
                    _section('Колеса и шины', [
                      if (_notEmpty(part.wearPercentage)) _row('Износ', '${part.wearPercentage}%'),
                      if (_notEmpty(part.season)) _row('Сезон', part.season!),
                      if (_notEmpty(part.diameter)) _row('Диаметр', part.diameter!),
                      if (_notEmpty(part.width)) _row('Ширина', part.width!),
                      if (_notEmpty(part.profile)) _row('Профиль', part.profile!),
                      if (_notEmpty(part.tireQuantity)) _row('Количество шин', part.tireQuantity!),
                      if (_notEmpty(part.drilling)) _row('Сверловка', part.drilling!),
                      if (_notEmpty(part.offset)) _row('Вылет', part.offset!),
                      if (_notEmpty(part.centerHoleDiameter)) _row('ЦО', part.centerHoleDiameter!),
                      if (_notEmpty(part.tireModel)) _row('Модель шины', part.tireModel!),
                    ]),
                if (_notEmpty(part.oemCode) || _notEmpty(part.supplierCode))
                  _section('Коды', [
                    if (_notEmpty(part.oemCode)) _row('OEM код', part.oemCode!),
                    if (_notEmpty(part.supplierCode))
                      _row('Код поставки', part.supplierCode!),
                  ]),
                _section('Расположение', [
                  if (part.location.isNotEmpty) _row('Место', part.location),
                  if (part.address.isNotEmpty) _row('Адрес склада', part.address),
                  if (part.salesman != null) _row('Продавец', part.salesman!),
                ]),
                if (part.description != null && part.description!.isNotEmpty)
                  _section('Описание', [
                    Padding(
                      padding: const EdgeInsets.only(top: 2),
                      child: Text(
                        part.description!,
                        style: const TextStyle(color: Colors.white, height: 1.45, fontSize: 13),
                      ),
                    ),
                  ]),
                _section('QR-код запчасти', [
                  Center(
                    child: Container(
                      margin: const EdgeInsets.symmetric(vertical: 8),
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
                      'Отсканируйте код для быстрого поиска на складе',
                      style: TextStyle(color: AppTheme.mutedColor, fontSize: 11),
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
                    icon: const Icon(LucideIcons.shopping_cart),
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
          ? AppTheme.successColor.withValues(alpha: 0.15)
          : AppTheme.dangerColor.withValues(alpha: 0.15),
      borderRadius: BorderRadius.circular(10),
      border: Border.all(
        color: qty > 0
            ? AppTheme.successColor.withValues(alpha: 0.35)
            : AppTheme.dangerColor.withValues(alpha: 0.35),
      ),
    ),
    child: Text(
      qty > 0 ? '$qty шт. в наличии' : 'Нет в наличии',
      style: TextStyle(
        color: qty > 0 ? AppTheme.successColor : AppTheme.dangerColor,
        fontWeight: FontWeight.w700,
        fontSize: 12,
      ),
    ),
  );

  bool _notEmpty(String? value) => value != null && value.trim().isNotEmpty && value.trim() != '—';

  Widget _section(String title, List<Widget> children) => children.isEmpty
      ? const SizedBox.shrink()
      : Container(
          margin: const EdgeInsets.only(bottom: 12),
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: AppTheme.cardColor,
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: AppTheme.borderColor),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title.toUpperCase(),
                style: const TextStyle(
                  color: AppTheme.mutedColor,
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                  letterSpacing: 0.8,
                ),
              ),
              const SizedBox(height: 10),
              ...children,
            ],
          ),
        );

  Widget _row(String label, String value) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 4),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 140,
          child: Text(
            label,
            style: const TextStyle(color: AppTheme.mutedColor, fontSize: 13),
          ),
        ),
        Expanded(
          child: Text(
            value,
            style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w500),
          ),
        ),
      ],
    ),
  );
}
