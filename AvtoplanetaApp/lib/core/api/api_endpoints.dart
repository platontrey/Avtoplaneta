/// Централизованные эндпоинты бэкенд API для предотвращения опечаток и рассинхронизации маршрутов.
abstract class ApiEndpoints {
  // Orders
  static const String orders = '/api/v1/orders';
  static const String completedOrders = '/api/v1/orders?state=completed';
  static String orderById(int id) => '/api/v1/orders/$id';
  static String completeOrder(int id) => '/api/v1/orders/$id/complete';
  static String orderItems(int orderId) => '/api/v1/orders/$orderId/items';
  static String orderItem(int orderId, int itemId) =>
      '/api/v1/orders/$orderId/items/$itemId';

  // Customers
  static const String customers = '/api/v1/orders/customers';
  static String customerById(int id) => '/api/v1/orders/customers/$id';

  // Inventory & Parts
  static const String parts = '/api/v1/parts';
  static String partById(int id) => '/api/v1/parts/$id';
  static String partItem(int id) => '/api/v1/parts/item/$id';
  static const String partCatalog = '/api/part-catalog';
  static const String vehicleCatalog = '/api/vehicle-catalog';
  static const String defectReports = '/api/defect-reports';
  static String uploadPartPhoto(int partId) => '/api/uploadpartphoto/$partId';

  // Statistics
  static const String statistics = '/api/v1/statistics';

  // Auth
  static const String authMe = '/auth/me';
  static const String authLogin = '/auth/login';
  static const String authLogout = '/auth/logout';
  static const String authRefresh = '/auth/refresh';

  // Admin
  static const String adminUsers = '/admin/users';
  static const String adminSupplierCodes = '/api/v1/admin/supplier-codes';
  static const String adminStatus = '/admin/status';
  static const String adminLogs = '/admin/logs';

  // Messaging
  static const String conversations = '/api/messaging/conversations';
  static const String messagingUsers = '/api/messaging/users';
}
