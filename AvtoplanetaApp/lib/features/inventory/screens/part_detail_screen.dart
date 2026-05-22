import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:go_router/go_router.dart';
import '../providers/inventory_provider.dart';
import '../../auth/providers/auth_provider.dart';

class PartDetailScreen extends ConsumerWidget {
  final int id;
  const PartDetailScreen({super.key, required this.id});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final partAsync = ref.watch(partProvider(id));
    final user = ref.watch(authProvider).valueOrNull;
    const baseUrl = 'https://avtoplaneta.avtoplaneta.crazedns.ru';

    return Scaffold(
      appBar: AppBar(
        title: const Text('Запчасть'),
        actions: [
          if (user?.isOperator == true)
            IconButton(
              icon: const Icon(Icons.edit_outlined),
              onPressed: () => context.go('/inventory/edit/$id'),
            ),
        ],
      ),
      body: partAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (part) => SingleChildScrollView(
          padding: const EdgeInsets.all(16),
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
                        imageUrl: '$baseUrl${part.photos[i]}',
                        fit: BoxFit.cover,
                        placeholder: (ctx, url) => Container(
                          color: const Color(0xFF16213E),
                          child: const Icon(Icons.image_outlined,
                              size: 64, color: Colors.white24),
                        ),
                        errorWidget: (ctx, url, err) => Container(
                          color: const Color(0xFF16213E),
                          child: const Icon(Icons.broken_image_outlined,
                              size: 64, color: Colors.white24),
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
                    child: Icon(Icons.directions_car_outlined,
                        size: 64, color: Colors.white24),
                  ),
                ),
              const SizedBox(height: 16),
              Text(
                part.name,
                style: Theme.of(context)
                    .textTheme
                    .titleLarge
                    ?.copyWith(color: Colors.white, fontWeight: FontWeight.bold),
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
                if (part.color != null) _row('Цвет', part.color!),
                if (part.condition != null) _row('Состояние', part.condition!),
              ]),
              if (part.oemCode != null || part.supplierCode != null)
                _section('Коды', [
                  if (part.oemCode != null) _row('OEM код', part.oemCode!),
                  if (part.supplierCode != null)
                    _row('Код поставщика', part.supplierCode!),
                  if (part.vin != null) _row('VIN', part.vin!),
                ]),
              _section('Расположение', [
                if (part.location.isNotEmpty) _row('Место', part.location),
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
            ],
          ),
        ),
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

  Widget _section(String title, List<Widget> children) => Column(
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
                letterSpacing: 0.5),
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
              child: Text(label,
                  style: const TextStyle(color: Colors.white38, fontSize: 13)),
            ),
            Expanded(
              child: Text(value,
                  style: const TextStyle(color: Colors.white, fontSize: 13)),
            ),
          ],
        ),
      );
}
