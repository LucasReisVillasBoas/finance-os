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

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/',
    redirect: (BuildContext context, GoRouterState state) {
      final isAuthenticated = authState.user != null;
      final isLoading = authState.isLoading;

      // While checking auth, stay on splash
      if (isLoading && state.matchedLocation == '/') {
        return null;
      }

      final publicRoutes = ['/login', '/register', '/onboarding', '/'];

      if (!isAuthenticated && !publicRoutes.contains(state.matchedLocation)) {
        return '/login';
      }

      if (isAuthenticated && (state.matchedLocation == '/login' || state.matchedLocation == '/register')) {
        return '/home';
      }

      return null;
    },
    routes: [
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
      GoRoute(
        path: '/home',
        builder: (context, state) => const HomeScreen(),
      ),
      // Accounts
      GoRoute(
        path: '/accounts',
        builder: (context, state) => const AccountsScreen(),
      ),
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
      // Settings
      GoRoute(
        path: '/settings/categories',
        builder: (context, state) => const CategoriesScreen(),
      ),
      // Transactions
      GoRoute(
        path: '/transactions',
        builder: (context, state) => const TransactionsScreen(),
      ),
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
        builder: (context, state) =>
            TransactionDetailScreen(transactionId: state.pathParameters['id']!),
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
        path: '/budgets',
        builder: (context, state) => const BudgetsScreen(),
      ),
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
        path: '/investments',
        builder: (context, state) => const PortfolioScreen(),
      ),
      GoRoute(
        path: '/investments/holdings/:id',
        builder: (context, state) =>
            HoldingDetailScreen(holdingId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/investments/new',
        builder: (context, state) => const InvestmentFormScreen(),
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
    ],
  );
});
