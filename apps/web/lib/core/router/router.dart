import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/providers/auth_provider.dart';
import '../../features/auth/screens/splash_screen.dart';
import '../../features/auth/screens/login_screen.dart';
import '../../features/auth/screens/register_screen.dart';
import '../../features/auth/screens/onboarding_screen.dart';
import '../../features/dashboard/screens/home_screen.dart';
import '../../features/accounts/screens/accounts_screen.dart';
import '../../features/accounts/screens/account_form_screen.dart';
import '../../features/accounts/screens/account_detail_screen.dart';
import '../../features/settings/screens/categories_screen.dart';
import '../../features/transactions/screens/transactions_screen.dart';
import '../../features/transactions/screens/transaction_form_screen.dart';
import '../../features/transactions/screens/transaction_detail_screen.dart';
import '../../features/transactions/screens/transaction_filters_screen.dart';
import '../../features/transactions/screens/recurrences_screen.dart';
import '../../features/transactions/screens/recurrence_form_screen.dart';
import '../../features/budgets/screens/budgets_screen.dart';
import '../../features/budgets/screens/budget_form_screen.dart';
import '../../features/dashboard/screens/dashboard_config_screen.dart';
import '../../features/investments/screens/portfolio_screen.dart';
import '../../features/investments/screens/holding_detail_screen.dart';
import '../../features/investments/screens/investment_form_screen.dart';
import '../../features/investments/screens/custom_asset_form_screen.dart';
import '../../features/investments/screens/portfolio_analysis_screen.dart';
import '../../features/goals/screens/goals_screen.dart';
import '../../features/goals/screens/goal_form_screen.dart';
import '../../features/settings/screens/import_screen.dart';
import '../../features/settings/screens/whatsapp_settings_screen.dart';
import '../../features/settings/screens/notifications_screen.dart';
import '../../features/settings/screens/notification_preferences_screen.dart';
import '../../features/settings/screens/settings_screen.dart';
import '../../features/settings/screens/profile_screen.dart';
import '../../features/settings/screens/family_screen.dart';
import '../../features/settings/screens/plans_screen.dart';
import '../../features/settings/screens/ai_assistant_screen.dart';
import '../../shared/widgets/app_shell.dart';

/// RouterNotifier bridges Riverpod auth state with GoRouter's refreshListenable.
/// The GoRouter instance is created ONCE and notified on auth changes —
/// avoids the bug of recreating GoRouter on every auth state change.
class RouterNotifier extends ChangeNotifier {
  RouterNotifier(this._ref);

  final Ref _ref;

  /// Called by the provider when auth state changes.
  void onAuthStateChanged() => notifyListeners();

  String? redirect(BuildContext context, GoRouterState state) {
    final authState = _ref.read(authProvider);
    final isAuthenticated = authState.user != null;
    final isLoading = authState.isLoading;

    // Stay on splash while auth check is in progress
    if (isLoading && state.matchedLocation == '/') return null;

    final publicRoutes = {'/login', '/register', '/onboarding', '/'};

    if (!isAuthenticated && !publicRoutes.contains(state.matchedLocation)) {
      return '/login';
    }
    if (isAuthenticated &&
        (state.matchedLocation == '/login' ||
            state.matchedLocation == '/register')) {
      return '/home';
    }
    return null;
  }
}

final _routerNotifierProvider = Provider<RouterNotifier>((ref) {
  final notifier = RouterNotifier(ref);
  // ref.listen is called on the Provider's ref — correct Riverpod 2.x pattern.
  // Calling it inside the class constructor causes silent crash on web.
  ref.listen<AuthState>(authProvider, (prev, next) => notifier.onAuthStateChanged());
  ref.onDispose(notifier.dispose);
  return notifier;
});

final routerProvider = Provider<GoRouter>((ref) {
  final notifier = ref.watch(_routerNotifierProvider);

  final router = GoRouter(
    initialLocation: '/',
    refreshListenable: notifier,
    redirect: notifier.redirect,
    routes: [
      // ── Public routes (no shell) ──────────────────────────────────────────
      GoRoute(
        path: '/',
        builder: (context, state) => const SplashScreen(),
      ),
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/register',
        builder: (context, state) => const RegisterScreen(),
      ),
      GoRoute(
        path: '/onboarding',
        builder: (context, state) => const OnboardingScreen(),
      ),

      // ── Main shell (persistent BottomNavigationBar + directional animation) ─
      ShellRoute(
        builder: (context, state, child) => AppShell(
          location: state.matchedLocation,
          child: child,
        ),
        routes: [
          GoRoute(
            path: '/home',
            builder: (context, state) => const HomeScreen(),
          ),
          GoRoute(
            path: '/transactions',
            builder: (context, state) => const TransactionsScreen(),
          ),
          GoRoute(
            path: '/accounts',
            builder: (context, state) => const AccountsScreen(),
          ),
          GoRoute(
            path: '/budgets',
            builder: (context, state) => const BudgetsScreen(),
          ),
          GoRoute(
            path: '/settings',
            builder: (context, state) => const SettingsScreen(),
          ),
          GoRoute(
            path: '/investments',
            builder: (context, state) => const PortfolioScreen(),
          ),
          GoRoute(
            path: '/goals',
            builder: (context, state) => const GoalsScreen(),
          ),
          GoRoute(
            path: '/notifications',
            builder: (context, state) => const NotificationsScreen(),
          ),
        ],
      ),

      // ── Secondary routes (no shell) ───────────────────────────────────────
      // Accounts
      GoRoute(
        path: '/accounts/new',
        builder: (context, state) => const AccountFormScreen(),
      ),
      GoRoute(
        path: '/accounts/:id',
        builder: (context, state) =>
            AccountDetailScreen(accountId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/accounts/:id/edit',
        builder: (context, state) =>
            AccountFormScreen(accountId: state.pathParameters['id']),
      ),
      // Settings sub-pages
      GoRoute(
        path: '/settings/categories',
        builder: (context, state) => const CategoriesScreen(),
      ),
      GoRoute(
        path: '/settings/profile',
        builder: (context, state) => const ProfileScreen(),
      ),
      GoRoute(
        path: '/settings/import',
        builder: (context, state) => const ImportScreen(),
      ),
      GoRoute(
        path: '/settings/whatsapp',
        builder: (context, state) => const WhatsAppSettingsScreen(),
      ),
      GoRoute(
        path: '/settings/notifications',
        builder: (context, state) => const NotificationPreferencesScreen(),
      ),
      GoRoute(
        path: '/settings/family',
        builder: (context, state) => const FamilyScreen(),
      ),
      GoRoute(
        path: '/settings/plans',
        builder: (context, state) => const PlansScreen(),
      ),
      GoRoute(
        path: '/settings/ai',
        builder: (context, state) => const AIAssistantScreen(),
      ),
      // Transactions
      GoRoute(
        path: '/transactions/new',
        builder: (context, state) => const TransactionFormScreen(),
      ),
      GoRoute(
        path: '/transactions/filters',
        builder: (context, state) => const TransactionFiltersScreen(),
      ),
      GoRoute(
        path: '/transactions/:id',
        builder: (context, state) => TransactionDetailScreen(
            transactionId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/transactions/:id/edit',
        builder: (context, state) =>
            TransactionFormScreen(transactionId: state.pathParameters['id']),
      ),
      // Recurrences
      GoRoute(
        path: '/recurrences',
        builder: (context, state) => const RecurrencesScreen(),
      ),
      GoRoute(
        path: '/recurrences/new',
        builder: (context, state) => const RecurrenceFormScreen(),
      ),
      GoRoute(
        path: '/recurrences/:id/edit',
        builder: (context, state) =>
            RecurrenceFormScreen(recurrenceId: state.pathParameters['id']),
      ),
      // Budgets
      GoRoute(
        path: '/budgets/new',
        builder: (context, state) => const BudgetFormScreen(),
      ),
      GoRoute(
        path: '/budgets/:id/edit',
        builder: (context, state) =>
            BudgetFormScreen(budgetId: state.pathParameters['id']),
      ),
      // Dashboard config
      GoRoute(
        path: '/dashboard/config',
        builder: (context, state) => const DashboardConfigScreen(),
      ),
      // Investments
      GoRoute(
        path: '/investments/new',
        builder: (context, state) => const InvestmentFormScreen(),
      ),
      GoRoute(
        path: '/investments/holdings/:id',
        builder: (context, state) =>
            HoldingDetailScreen(holdingId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/investments/custom/new',
        builder: (context, state) => const CustomAssetFormScreen(),
      ),
      GoRoute(
        path: '/investments/custom/:id/edit',
        builder: (context, state) =>
            CustomAssetFormScreen(customAssetId: state.pathParameters['id']),
      ),
      GoRoute(
        path: '/investments/analysis',
        builder: (context, state) => const PortfolioAnalysisScreen(),
      ),
      // Goals
      GoRoute(
        path: '/goals/new',
        builder: (context, state) => const GoalFormScreen(),
      ),
      GoRoute(
        path: '/goals/:id/edit',
        builder: (context, state) =>
            GoalFormScreen(goalId: state.pathParameters['id']),
      ),
    ],
  );
  ref.onDispose(router.dispose);
  return router;
});
