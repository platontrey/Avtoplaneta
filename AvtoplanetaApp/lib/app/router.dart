import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/auth/providers/auth_provider.dart';
import '../features/auth/screens/login_screen.dart';
import '../features/inventory/screens/inventory_screen.dart';
import '../features/inventory/screens/part_detail_screen.dart';
import '../features/inventory/screens/add_part_screen.dart';
import '../features/inventory/screens/defect_report_screen.dart';
import '../features/orders/screens/orders_screen.dart';
import '../features/statistics/screens/statistics_screen.dart';
import '../features/messaging/screens/messaging_screen.dart';
import '../features/admin/screens/admin_screen.dart';
import '../shared/widgets/main_scaffold.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/inventory',
    redirect: (context, state) {
      final isLoggedIn = authState.valueOrNull != null;
      final isLoading = authState.isLoading;
      final isLoginPage = state.matchedLocation == '/login';
      final isAdminPage = state.matchedLocation.startsWith('/admin');
      final user = authState.valueOrNull;

      if (isLoading) return null;
      if (!isLoggedIn && !isLoginPage) return '/login';
      if (isLoggedIn && isLoginPage) return '/inventory';
      if (isAdminPage && user?.isAdmin != true) return '/inventory';
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
      ShellRoute(
        builder: (context, state, child) => MainScaffold(child: child),
        routes: [
          GoRoute(
            path: '/inventory',
            builder: (context, state) => const InventoryScreen(),
            routes: [
              GoRoute(
                path: 'part/:id',
                builder: (context, state) =>
                    PartDetailScreen(id: int.parse(state.pathParameters['id']!)),
              ),
              GoRoute(
                path: 'add',
                builder: (context, state) => const AddPartScreen(),
              ),
              GoRoute(
                path: 'defect-report',
                builder: (context, state) => const DefectReportScreen(),
              ),
              GoRoute(
                path: 'edit/:id',
                builder: (context, state) =>
                    AddPartScreen(editId: int.parse(state.pathParameters['id']!)),
              ),
            ],
          ),
          GoRoute(
            path: '/orders',
            builder: (context, state) => const OrdersScreen(),
          ),
          GoRoute(
            path: '/statistics',
            builder: (context, state) => const StatisticsScreen(),
          ),
          GoRoute(
            path: '/messages',
            builder: (context, state) => const MessagingScreen(),
          ),
          GoRoute(
            path: '/admin',
            builder: (context, state) => const AdminScreen(),
          ),
        ],
      ),
    ],
  );
});
