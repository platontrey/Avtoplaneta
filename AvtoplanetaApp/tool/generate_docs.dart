import 'dart:io';

void main() {
  stdout.writeln('🚀 Генерация архитектурной документации мобильного приложения Avtoplaneta...');

  final projectDir = Directory.current;
  final endpointsFile = File('${projectDir.path}/lib/core/api/api_endpoints.dart');
  final routerFile = File('${projectDir.path}/lib/app/router.dart');

  final endpoints = _extractEndpoints(endpointsFile);
  final routes = _extractRoutes(routerFile);

  final markdown = _buildMarkdown(endpoints: endpoints, routes: routes);

  final readmeFile = File('${projectDir.path}/README.md');
  readmeFile.writeAsStringSync(markdown);

  stdout.writeln('✅ Документация успешно сгенерирована и сохранена в ${readmeFile.path}');
}

List<Map<String, String>> _extractEndpoints(File file) {
  if (!file.existsSync()) return [];
  final content = file.readAsStringSync();
  final lines = content.split('\n');

  final endpoints = <Map<String, String>>[];
  final staticEndpointRegex = RegExp(r"static const String (\w+)\s*=\s*'([^']+)';");
  final dynamicEndpointRegex = RegExp(r"static String (\w+)\([^)]*\)\s*=>\s*['`$]([^'`]+)['`$];");

  for (final line in lines) {
    final staticMatch = staticEndpointRegex.firstMatch(line);
    if (staticMatch != null) {
      endpoints.add({
        'name': staticMatch.group(1)!,
        'path': staticMatch.group(2)!,
        'type': 'Статический',
      });
      continue;
    }

    final dynamicMatch = dynamicEndpointRegex.firstMatch(line);
    if (dynamicMatch != null) {
      endpoints.add({
        'name': dynamicMatch.group(1)!,
        'path': dynamicMatch.group(2)!,
        'type': 'Параметризованный',
      });
    }
  }

  return endpoints;
}

List<Map<String, String>> _extractRoutes(File file) {
  if (!file.existsSync()) return [];
  final content = file.readAsStringSync();
  final lines = content.split('\n');

  final routes = <Map<String, String>>[];
  final routeRegex = RegExp(r"path:\s*'([^']+)',");
  final builderRegex = RegExp(r"builder:\s*\(context,\s*state\)\s*=>\s*([^,\n]+)");

  String? currentPath;
  for (final line in lines) {
    final routeMatch = routeRegex.firstMatch(line);
    if (routeMatch != null) {
      currentPath = routeMatch.group(1)!;
    }
    final builderMatch = builderRegex.firstMatch(line);
    if (builderMatch != null && currentPath != null) {
      var screen = builderMatch.group(1)!.trim();
      screen = screen.replaceAll('const ', '').replaceAll('()', '');
      if (screen.contains('(')) {
        screen = screen.substring(0, screen.indexOf('('));
      }
      routes.add({
        'path': currentPath,
        'screen': screen,
      });
      currentPath = null;
    }
  }

  return routes;
}

String _buildMarkdown({
  required List<Map<String, String>> endpoints,
  required List<Map<String, String>> routes,
}) {
  final buffer = StringBuffer();

  buffer.writeln('# Мобильное приложение Avtoplaneta (Flutter Client)');
  buffer.writeln();
  buffer.writeln('> 🤖 **Автоматически сгенерированная документация архитектуры и API.**');
  buffer.writeln('> Синхронизируется с веб-порталом документации `/readme` через `frontend/scripts/sync-docs.js`.');
  buffer.writeln();
  buffer.writeln('---');
  buffer.writeln();

  buffer.writeln('## 1. Архитектурная топология');
  buffer.writeln();
  buffer.writeln('Архитектура клиента построена по принципам **Clean Architecture** с разделением ответственности на независимые слои:');
  buffer.writeln();
  buffer.writeln('```mermaid');
  buffer.writeln('flowchart TD');
  buffer.writeln('    subgraph UI["1. Слой отображения (Presentation Layer)"]');
  buffer.writeln('        Screens["Экраны: InventoryScreen, OrdersScreen, AdminScreen"]');
  buffer.writeln('        Widgets["Компоненты: PartOrderSheet, CustomerFormSheet"]');
  buffer.writeln('        AsyncExt["AsyncValueUI (.toWidget(), .showSnackbarOnError())"]');
  buffer.writeln('    end');
  buffer.writeln();
  buffer.writeln('    subgraph State["2. Слой состояния (State Management - Riverpod)"]');
  buffer.writeln('        OrdersProv["ordersProvider / completedOrdersProvider"]');
  buffer.writeln('        CustProv["customersProvider / customerDetailsProvider"]');
  buffer.writeln('        InvProv["inventoryProvider / inventoryFilterProvider"]');
  buffer.writeln('        AuthProv["authProvider (User session)"]');
  buffer.writeln('    end');
  buffer.writeln();
  buffer.writeln('    subgraph Domain["3. Слой домена и репозиториев (Data Layer)"]');
  buffer.writeln('        OrdersRepo["OrdersRepository"]');
  buffer.writeln('        CustRepo["CustomersRepository"]');
  buffer.writeln('        Failure["Единая типизация ошибок AppFailure"]');
  buffer.writeln('        FreezedModels["Модели Freezed 3: Part, Order, Customer"]');
  buffer.writeln('    end');
  buffer.writeln();
  buffer.writeln('    subgraph Infra["4. Сетевой и локальный транспорт"]');
  buffer.writeln('        DioClient["Dio HTTP Client + ContentType & Auth Interceptors"]');
  buffer.writeln('        LocalStorage["CacheStorage (Hive) + SecureStorage"]');
  buffer.writeln('    end');
  buffer.writeln();
  buffer.writeln('    subgraph Backend["5. Бэкенд-инфраструктура"]');
  buffer.writeln('        Traefik["Traefik Ingress (ForwardAuth & SSL)"]');
  buffer.writeln('        Services["Микросервисы: parts, orders, auth, messaging"]');
  buffer.writeln('    end');
  buffer.writeln();
  buffer.writeln('    Screens --> State');
  buffer.writeln('    Widgets --> State');
  buffer.writeln('    State --> Domain');
  buffer.writeln('    Domain --> Infra');
  buffer.writeln('    Infra --> Traefik --> Services');
  buffer.writeln('```');
  buffer.writeln();

  buffer.writeln('## 2. Модели данных (Freezed 3 + json_serializable)');
  buffer.writeln();
  buffer.writeln('Все основные сущности переведены на кодогенерацию **Freezed 3**, что гарантирует:');
  buffer.writeln('- Неизменяемость (Value Equality & HashCode).');
  buffer.writeln('- Автоматическую генерацию `copyWith`, `toJson`, `fromJson`.');
  buffer.writeln('- Отсутствие сотен строк ручного бойлерплейта.');
  buffer.writeln('- Изоляцию от UI (UI-хелперы и цвета вынесены в extensions).');
  buffer.writeln();
  buffer.writeln('| Модель | Назначение | Ключевые поля | Особенности |');
  buffer.writeln('|---|---|---|---|');
  buffer.writeln('| `Part` | Карточка детали склада | `id, name, price, quantity, photos, status` | Поддержка `quantity == 0` (дефектовка) и `quantity == -1` (продано) |');
  buffer.writeln('| `Order` | Заказ покупателя | `id, items, buyerNumber, status, paymentStatus, discount` | Цветовые статусы, вложенные `OrderItem` |');
  buffer.writeln('| `Customer` | Клиент (CRM) | `id, name, phone, category, discountPercent, ordersCount, totalSpent` | Автоматическая нормализация телефона к `+7XXXXXXXXXX` |');
  buffer.writeln('| `CustomerDetails` | Детальная карточка клиента | `customer, orders` | Расчет LTV и история всех покупок |');
  buffer.writeln('| `InventoryFilter` | Фильтр каталога склада | `search, category, brand, model, location, page, pageSize...` | Свыше 35 параметров фильтрации, `activeFilterCount` |');
  buffer.writeln();

  buffer.writeln('## 3. Критический доменный инвариант: учет количества (quantity)');
  buffer.writeln();
  buffer.writeln('В соответствии с [AGENTS.md](../AGENTS.md), поле `quantity` имеет строгую семантику:');
  buffer.writeln();
  buffer.writeln('```mermaid');
  buffer.writeln('stateDiagram-v2');
  buffer.writeln('    direction LR');
  buffer.writeln('    [*] --> InStock: Приемка детали (quantity > 0)');
  buffer.writeln('    [*] --> DefectAct: Дефектовка (quantity = 0)');
  buffer.writeln('    InStock --> Sold: Продажа через заказ (quantity = -1)');
  buffer.writeln('    Sold --> InStock: Отмена заказа (возврат остатка)');
  buffer.writeln('    DefectAct --> [*]: Отображается в поиске');
  buffer.writeln('    Sold --> [*]: Скрыто из выдачи (история сохранена)');
  buffer.writeln('```');
  buffer.writeln();
  buffer.writeln('- `quantity > 0` — деталь в наличии на складе.');
  buffer.writeln('- `quantity == 0` — дефектная ведомость / спецдеталь (**ОБЯЗАНА ОСТАВАТЬСЯ ВИДИМОЙ** в каталоге и поиске).');
  buffer.writeln('- `quantity == -1` — деталь полностью продана через `orders-service` (маркер сокрытия из выдачи).');
  buffer.writeln('- ⚠️ **ПРАВИЛО ЗАПРОСОВ:** Для фильтрации склада всегда используется условие `quantity >= 0`. Запрещено использовать `> 0`!');
  buffer.writeln();

  buffer.writeln('## 4. Единая система обработки ошибок (AppFailure)');
  buffer.writeln();
  buffer.writeln('Все сетевые и логические ошибки трансформируются в типизированный класс `AppFailure`:');
  buffer.writeln();
  buffer.writeln('| Тип ошибки | Источник | Поведение в приложении |');
  buffer.writeln('|---|---|---|');
  buffer.writeln('| `NetworkFailure` | Таймаут, отсутствие интернета, обрыв соединения | Сообщение о проверке сети, кнопка "Повторить попытку" |');
  buffer.writeln('| `AuthFailure` | 401 / 403 Forbidden | Предложение войти заново, автоматический переход на `/login` |');
  buffer.writeln('| `ValidationFailure` | 400 Bad Request, 409 Conflict | Понятный текст валидации (например: "Клиент с таким телефоном уже существует") |');
  buffer.writeln('| `NotFoundFailure` | 404 Not Found | "Запрашиваемый ресурс или запись не найдены" |');
  buffer.writeln('| `ServerFailure` | 500, 502, 503 Service Unavailable | Уведомление о сбое на стороне сервера |');
  buffer.writeln();

  buffer.writeln('## 5. Маршрутизация (GoRouter)');
  buffer.writeln();
  buffer.writeln('Навигация приложения управляется централизованно через `lib/app/router.dart`:');
  buffer.writeln();
  buffer.writeln('| Маршрут | Экран / Компонент | Доступ |');
  buffer.writeln('|---|---|---|');
  if (routes.isEmpty) {
    buffer.writeln('| `/login` | `LoginScreen` | Публичный |');
    buffer.writeln('| `/inventory` | `InventoryScreen` | Авторизованный |');
    buffer.writeln('| `/inventory/part/:id` | `PartDetailScreen` | Авторизованный |');
    buffer.writeln('| `/inventory/add` | `AddPartScreen` | Авторизованный |');
    buffer.writeln('| `/orders` | `OrdersScreen` | Авторизованный |');
    buffer.writeln('| `/statistics` | `StatisticsScreen` | Авторизованный |');
    buffer.writeln('| `/messages` | `MessagingScreen` | Авторизованный |');
    buffer.writeln('| `/admin` | `AdminScreen` | Администратор (`isAdmin == true`) |');
    buffer.writeln('| `/admin/logs` | `AppLogsScreen` | Администратор (`isAdmin == true`) |');
  } else {
    for (final r in routes) {
      final isAdmin = r['path']!.startsWith('/admin') || r['path'] == 'logs';
      final isLogin = r['path'] == '/login';
      final access = isLogin ? 'Публичный' : (isAdmin ? 'Администратор' : 'Авторизованный');
      buffer.writeln('| `${r['path']}` | `${r['screen']}` | $access |');
    }
  }
  buffer.writeln();

  buffer.writeln('## 6. Каталог API Эндпоинтов (`ApiEndpoints`)');
  buffer.writeln();
  buffer.writeln('Все обращения к бэкенду используют строго централизованный каталог маршрутов в `lib/core/api/api_endpoints.dart`:');
  buffer.writeln();
  buffer.writeln('| Название константы / метода | URL Эндпоинта | Тип вызова |');
  buffer.writeln('|---|---|---|');
  for (final ep in endpoints) {
    buffer.writeln('| `ApiEndpoints.${ep['name']}` | `${ep['path']}` | ${ep['type']} |');
  }
  buffer.writeln();

  buffer.writeln('## 7. Справочник команд для разработчика');
  buffer.writeln();
  buffer.writeln('```shell');
  buffer.writeln('# 1. Установка зависимостей:');
  buffer.writeln('flutter pub get');
  buffer.writeln();
  buffer.writeln('# 2. Генерация моделей Freezed и json_serializable:');
  buffer.writeln('dart run build_runner build --delete-conflicting-outputs');
  buffer.writeln();
  buffer.writeln('# 3. Генерация актуальной документации архитектуры:');
  buffer.writeln('dart run tool/generate_docs.dart');
  buffer.writeln();
  buffer.writeln('# 4. Проверка статического анализатора (чистый линт):');
  buffer.writeln('flutter analyze');
  buffer.writeln();
  buffer.writeln('# 5. Запуск всех тестов приложения:');
  buffer.writeln('flutter test');
  buffer.writeln();
  buffer.writeln('# 6. Сборка debug APK:');
  buffer.writeln('flutter build apk --debug');
  buffer.writeln('```');
  buffer.writeln();

  return buffer.toString();
}
