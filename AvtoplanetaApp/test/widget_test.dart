// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:avtoplaneta_app/core/models/part.dart';
import 'package:avtoplaneta_app/core/models/user.dart';
import 'package:avtoplaneta_app/features/auth/providers/auth_provider.dart';
import 'package:avtoplaneta_app/features/inventory/data/part_catalog.dart';
import 'package:avtoplaneta_app/features/inventory/providers/inventory_provider.dart';
import 'package:avtoplaneta_app/features/inventory/providers/part_catalog_provider.dart';
import 'package:avtoplaneta_app/main.dart';

void main() {
  testWidgets('App smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authProvider.overrideWith(
            (ref) => AuthNotifier(
              initialUser: const User(
                id: 1,
                email: 'test@avtoplaneta.local',
                name: 'Test User',
                role: 'operator',
              ),
              initialize: false,
            ),
          ),
          inventoryProvider.overrideWith(
            (ref, filter) async => const InventoryResponse(
              parts: [],
              total: 0,
              page: 1,
              limit: 20,
            ),
          ),
          partCatalogProvider.overrideWith(
            (ref) async => const PartCatalog(
              version: 'test',
              attributes: [],
              partFormCategories: [],
            ),
          ),
        ],
        child: const AvtoplanetaApp(),
      ),
    );
    await tester.pump();

    expect(find.text('Инвентарь'), findsWidgets);
    expect(find.byTooltip('Фильтры'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('inventory supports multi-select bulk actions', (
    WidgetTester tester,
  ) async {
    const parts = [
      Part(
        id: 101,
        name: 'Фара левая',
        category: 'Кузов',
        price: 5000,
        quantity: 1,
      ),
      Part(
        id: 102,
        name: 'Фара правая',
        category: 'Кузов',
        price: 5500,
        quantity: 2,
      ),
    ];
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authProvider.overrideWith(
            (ref) => AuthNotifier(
              initialUser: const User(
                id: 1,
                email: 'test@avtoplaneta.local',
                name: 'Test User',
                role: 'operator',
              ),
              initialize: false,
            ),
          ),
          inventoryProvider.overrideWith(
            (ref, filter) async => const InventoryResponse(
              parts: parts,
              total: 2,
              page: 1,
              limit: 20,
            ),
          ),
          partCatalogProvider.overrideWith(
            (ref) async => const PartCatalog(
              version: 'test',
              attributes: [],
              partFormCategories: [],
            ),
          ),
        ],
        child: const AvtoplanetaApp(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.longPress(find.text('Фара левая'));
    await tester.pump();

    expect(find.text('Выбрано: 1'), findsOneWidget);
    expect(find.text('Заказать (1)'), findsOneWidget);
    expect(find.text('Изменить'), findsOneWidget);
    expect(find.text('Удалить'), findsOneWidget);

    await tester.tap(find.text('Фара правая'));
    await tester.pump();
    expect(find.text('Выбрано: 2'), findsOneWidget);
    expect(find.text('Заказать (2)'), findsOneWidget);
  });
}
