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
              reportBindings: [],
              parts: [],
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
}
