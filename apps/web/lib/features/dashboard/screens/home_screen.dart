import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_theme.dart';
import '../../../shared/widgets/app_shell.dart';
import '../../accounts/models/account_model.dart';
import '../../accounts/providers/accounts_provider.dart';
import '../models/dashboard_model.dart';
import '../providers/dashboard_provider.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(dashboardProvider.notifier).load();
      ref.read(accountsProvider.notifier).loadAccounts();
    });
  }

  @override
  Widget build(BuildContext context) {
    final dashState = ref.watch(dashboardProvider);
    final accountsState = ref.watch(accountsProvider);

    return Scaffold(
      backgroundColor: Theme.of(context).scaffoldBackgroundColor,
      body: _DashboardBody(
        dashState: dashState,
        accountsState: accountsState,
      ),
    );
  }
}

// ────────────────────────────────────────────────────────────────────────────

class _DashboardBody extends ConsumerWidget {
  const _DashboardBody({
    required this.dashState,
    required this.accountsState,
  });

  final DashboardState dashState;
  final AccountsState accountsState;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final overview = dashState.overview;

    return CustomScrollView(
      slivers: [
        // ── App Bar ──
        SliverAppBar(
          floating: true,
          snap: true,
          backgroundColor: Theme.of(context).scaffoldBackgroundColor,
          elevation: 0,
          scrolledUnderElevation: 0,
          title: Row(
            children: [
              Container(
                width: 32,
                height: 32,
                decoration: BoxDecoration(
                  color: AppColors.primary,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(Icons.bolt_rounded,
                    color: Colors.white, size: 18),
              ),
              const SizedBox(width: 8),
              Text(
                'FinanceOS',
                style: Theme.of(context).textTheme.titleLarge,
              ),
            ],
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.notifications_outlined),
              tooltip: 'Notificações',
              onPressed: () => context.push('/notifications'),
            ),
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: GestureDetector(
                onTap: () => context.push('/settings/profile'),
                child: CircleAvatar(
                  radius: 16,
                  backgroundColor: AppColors.primaryContainer,
                  child: const Icon(Icons.person_rounded,
                      size: 18, color: AppColors.primary),
                ),
              ),
            ),
          ],
        ),

        if (dashState.isLoading)
          const SliverFillRemaining(
            child: Center(child: CircularProgressIndicator()),
          )
        else if (dashState.error != null)
          SliverFillRemaining(
            child: Center(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.wifi_off_rounded,
                        size: 48, color: AppColors.textLow),
                    const SizedBox(height: 12),
                    Text('Erro ao carregar dados',
                        style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 8),
                    Text(
                      dashState.error!,
                      textAlign: TextAlign.center,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.textLow,
                          ),
                    ),
                    const SizedBox(height: 16),
                    FilledButton(
                      onPressed: () =>
                          ref.read(dashboardProvider.notifier).load(),
                      child: const Text('Tentar novamente'),
                    ),
                  ],
                ),
              ),
            ),
          )
        else ...[
          // ── Hero: Saldo + Patrimônio ──
          SliverToBoxAdapter(
            child: _HeroCard(
              overview: overview,
              month: dashState.month,
              year: dashState.year,
            ),
          ),

          // ── Capacidade de Investir ──
          if (overview != null && overview.investmentCapacity > 0)
            SliverToBoxAdapter(
              child: _CapacityBanner(overview: overview),
            ),

          // ── Contas ──
          if (accountsState.accounts.isNotEmpty) ...[
            SliverToBoxAdapter(
              child: SectionHeader(
                title: 'Minhas Contas',
                actionLabel: 'Ver todas',
                onAction: () => context.go('/accounts'),
              ),
            ),
            SliverToBoxAdapter(
              child: SizedBox(
                height: 108,
                child: ListView.separated(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  itemCount: accountsState.accounts.length,
                  separatorBuilder: (_, __) => const SizedBox(width: 12),
                  itemBuilder: (context, i) =>
                      _AccountCard(account: accountsState.accounts[i]),
                ),
              ),
            ),
          ],

          // ── Atalhos ──
          SliverToBoxAdapter(child: _QuickActions()),

          // ── Fluxo de Caixa ──
          if (dashState.cashflow.isNotEmpty) ...[
            SliverToBoxAdapter(
              child: SectionHeader(
                title: 'Receitas vs Despesas',
                actionLabel: 'Detalhar',
                onAction: () => context.go('/transactions'),
              ),
            ),
            SliverToBoxAdapter(
              child: _CashflowChart(cashflow: dashState.cashflow),
            ),
          ],

          // ── Maiores gastos ──
          if (overview != null && overview.topCategories.isNotEmpty) ...[
            SliverToBoxAdapter(
              child: SectionHeader(
                title: 'Maiores Gastos',
                actionLabel: 'Ver orçamentos',
                onAction: () => context.go('/budgets'),
              ),
            ),
            SliverToBoxAdapter(
              child: _TopCategoriesList(
                  categories: overview.topCategories),
            ),
          ],

          // ── Alertas de orçamento ──
          if (overview != null && overview.alertBudgets.isNotEmpty) ...[
            SliverToBoxAdapter(
              child: SectionHeader(
                title: 'Orçamentos em Alerta',
                color: AppColors.expense,
              ),
            ),
            SliverToBoxAdapter(
              child: _AlertBudgetsList(budgets: overview.alertBudgets),
            ),
          ],

          // ── Transações recentes ──
          if (overview != null &&
              overview.recentTransactions.isNotEmpty) ...[
            SliverToBoxAdapter(
              child: SectionHeader(
                title: 'Últimas Transações',
                actionLabel: 'Ver todas',
                onAction: () => context.go('/transactions'),
              ),
            ),
            SliverPadding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              sliver: SliverList.separated(
                itemCount: overview.recentTransactions.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) => _TransactionTile(
                  transaction: overview.recentTransactions[i],
                ),
              ),
            ),
          ],

          const SliverToBoxAdapter(child: SizedBox(height: 100)),
        ],
      ],
    );
  }
}

// ── Hero Card ────────────────────────────────────────────────────────────────

class _HeroCard extends ConsumerWidget {
  const _HeroCard({
    required this.overview,
    required this.month,
    required this.year,
  });

  final DashboardOverview? overview;
  final int month;
  final int year;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final netBalance = overview?.netBalance ?? 0.0;
    final totalIncome = overview?.totalIncome ?? 0.0;
    final totalExpense = overview?.totalExpense ?? 0.0;
    final netWorth = overview?.totalNetWorth ?? 0.0;
    final investValue = overview?.investmentValue ?? 0.0;
    final customValue = overview?.customAssetValue ?? 0.0;
    final isPositive = netBalance >= 0;

    final monthNames = [
      'Jan', 'Fev', 'Mar', 'Abr', 'Mai', 'Jun',
      'Jul', 'Ago', 'Set', 'Out', 'Nov', 'Dez',
    ];

    final gradientColors = isPositive
        ? [const Color(0xFF2D42C8), const Color(0xFF4B6EF6)]
        : [const Color(0xFF991B1B), const Color(0xFFDC2626)];

    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 0),
      child: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: gradientColors,
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(24),
        ),
        child: Stack(
          children: [
            // Decorative circles
            Positioned(
              top: -30,
              right: -20,
              child: Container(
                width: 140,
                height: 140,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: Colors.white.withValues(alpha: 0.06),
                ),
              ),
            ),
            Positioned(
              bottom: -20,
              right: 60,
              child: Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: Colors.white.withValues(alpha: 0.06),
                ),
              ),
            ),

            Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Month selector
                  Row(
                    children: [
                      const Text('Saldo do mês',
                          style: TextStyle(
                              color: Colors.white70,
                              fontSize: 13,
                              fontWeight: FontWeight.w500)),
                      const Spacer(),
                      _MonthButton(
                        onPrev: () => ref
                            .read(dashboardProvider.notifier)
                            .changeMonth(-1),
                        onNext: () => ref
                            .read(dashboardProvider.notifier)
                            .changeMonth(1),
                        label: '${monthNames[month - 1]} $year',
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),

                  // Net balance
                  Text(
                    fmt.format(netBalance),
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 34,
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.5,
                    ),
                  ),
                  const SizedBox(height: 20),

                  // Income / Expense row
                  Row(
                    children: [
                      Expanded(
                        child: StatBadge(
                          label: 'Receitas',
                          value: fmt.format(totalIncome),
                          icon: Icons.arrow_upward_rounded,
                          color: const Color(0xFF6EE7B7),
                        ),
                      ),
                      Expanded(
                        child: StatBadge(
                          label: 'Despesas',
                          value: fmt.format(totalExpense),
                          icon: Icons.arrow_downward_rounded,
                          color: const Color(0xFFFCA5A5),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),

                  // Divider
                  Divider(color: Colors.white.withValues(alpha: 0.15), height: 1),
                  const SizedBox(height: 16),

                  // Patrimônio total
                  const Text('Patrimônio total',
                      style: TextStyle(
                          color: Colors.white60,
                          fontSize: 11,
                          fontWeight: FontWeight.w500,
                          letterSpacing: 0.5)),
                  const SizedBox(height: 4),
                  Text(
                    fmt.format(netWorth),
                    style: const TextStyle(
                        color: Colors.white,
                        fontSize: 20,
                        fontWeight: FontWeight.w700),
                  ),
                  const SizedBox(height: 10),
                  Row(
                    children: [
                      InfoPill(
                        label: 'Investido',
                        value: fmt.format(investValue),
                        color: Colors.white,
                      ),
                      const SizedBox(width: 8),
                      InfoPill(
                        label: 'Outros ativos',
                        value: fmt.format(customValue),
                        color: Colors.white,
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _MonthButton extends StatelessWidget {
  const _MonthButton({
    required this.label,
    required this.onPrev,
    required this.onNext,
  });

  final String label;
  final VoidCallback onPrev;
  final VoidCallback onNext;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          GestureDetector(
            onTap: onPrev,
            child: const Icon(Icons.chevron_left,
                color: Colors.white, size: 18),
          ),
          const SizedBox(width: 4),
          Text(label,
              style: const TextStyle(
                  color: Colors.white,
                  fontSize: 12,
                  fontWeight: FontWeight.w600)),
          const SizedBox(width: 4),
          GestureDetector(
            onTap: onNext,
            child: const Icon(Icons.chevron_right,
                color: Colors.white, size: 18),
          ),
        ],
      ),
    );
  }
}

// ── Capacity Banner ──────────────────────────────────────────────────────────

class _CapacityBanner extends StatelessWidget {
  const _CapacityBanner({required this.overview});
  final DashboardOverview overview;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final capacity = overview.investmentCapacity;
    final pct = overview.investmentCapacityPct;

    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 12, 20, 0),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: AppColors.incomeLight,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
              color: AppColors.income.withValues(alpha: 0.3)),
        ),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: AppColors.income.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(10),
              ),
              child: const Icon(Icons.trending_up_rounded,
                  color: AppColors.income, size: 20),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Capacidade de investir este mês',
                      style: TextStyle(
                          fontSize: 11,
                          color: AppColors.income,
                          fontWeight: FontWeight.w500)),
                  const SizedBox(height: 1),
                  Text(fmt.format(capacity),
                      style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w800,
                          color: AppColors.income)),
                ],
              ),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: AppColors.income,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                '${pct.toStringAsFixed(0)}%',
                style: const TextStyle(
                    color: Colors.white,
                    fontSize: 12,
                    fontWeight: FontWeight.w700),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ── Account Card ─────────────────────────────────────────────────────────────

class _AccountCard extends StatelessWidget {
  const _AccountCard({required this.account});
  final AccountModel account;

  static const _icons = {
    'checking': Icons.account_balance_outlined,
    'savings': Icons.savings_outlined,
    'credit_card': Icons.credit_card_outlined,
    'investment': Icons.candlestick_chart_outlined,
    'wallet': Icons.account_balance_wallet_outlined,
  };

  static const _colors = {
    'checking': AppColors.primary,
    'savings': AppColors.income,
    'credit_card': AppColors.expense,
    'investment': AppColors.violet,
    'wallet': AppColors.warning,
  };

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final balance = account.balance;
    final isNegative = balance < 0;
    final color =
        _colors[account.type] ?? AppColors.primary;
    final iconData =
        _icons[account.type] ?? Icons.account_balance_outlined;

    return Container(
      width: 148,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: Theme.of(context)
              .colorScheme
              .outline
              .withValues(alpha: 0.4),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Container(
            width: 32,
            height: 32,
            decoration: BoxDecoration(
              color: color.withValues(alpha: 0.12),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(iconData, color: color, size: 17),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(account.name,
                  style: const TextStyle(
                      fontSize: 12, fontWeight: FontWeight.w600),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis),
              const SizedBox(height: 2),
              Text(fmt.format(balance),
                  style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w700,
                      color: isNegative
                          ? AppColors.expense
                          : AppColors.income)),
            ],
          ),
        ],
      ),
    );
  }
}

// ── Quick Actions ─────────────────────────────────────────────────────────────

class _QuickActions extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
      child: IntrinsicHeight(
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _ActionTile(
              icon: Icons.add_rounded,
              label: 'Adicionar\ntransação',
              gradient: const [Color(0xFF3B5BDB), Color(0xFF5E7AF8)],
              onTap: () => context.push('/transactions/new'),
            ),
            const SizedBox(width: 10),
            _ActionTile(
              icon: Icons.candlestick_chart_rounded,
              label: 'Investimentos',
              gradient: const [Color(0xFF6D28D9), Color(0xFF9B5CF6)],
              onTap: () => context.go('/investments'),
            ),
            const SizedBox(width: 10),
            _ActionTile(
              icon: Icons.flag_rounded,
              label: 'Metas\nfinanceiras',
              gradient: const [Color(0xFF047857), Color(0xFF059669)],
              onTap: () => context.push('/goals'),
            ),
          ],
        ),
      ),
    );
  }
}

class _ActionTile extends StatelessWidget {
  const _ActionTile({
    required this.icon,
    required this.label,
    required this.gradient,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final List<Color> gradient;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 10),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: gradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(icon, color: Colors.white, size: 22),
              const SizedBox(height: 10),
              Text(
                label,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 11,
                  fontWeight: FontWeight.w600,
                  height: 1.3,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ── Cashflow Chart ────────────────────────────────────────────────────────────

class _CashflowChart extends StatelessWidget {
  const _CashflowChart({required this.cashflow});
  final List<MonthlyCashflowModel> cashflow;

  @override
  Widget build(BuildContext context) {
    final last6 = cashflow.length > 6
        ? cashflow.sublist(cashflow.length - 6)
        : cashflow;

    return SurfaceCard(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
      child: Column(
        children: [
          // Legend
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              _LegendDot(color: AppColors.primary, label: 'Receitas'),
              const SizedBox(width: 14),
              _LegendDot(
                  color: AppColors.expense.withValues(alpha: 0.7),
                  label: 'Despesas'),
            ],
          ),
          const SizedBox(height: 8),
          SizedBox(
            height: 180,
            child: BarChart(
              BarChartData(
                alignment: BarChartAlignment.spaceAround,
                barTouchData: BarTouchData(enabled: false),
                titlesData: FlTitlesData(
                  leftTitles: const AxisTitles(
                      sideTitles: SideTitles(showTitles: false)),
                  rightTitles: const AxisTitles(
                      sideTitles: SideTitles(showTitles: false)),
                  topTitles: const AxisTitles(
                      sideTitles: SideTitles(showTitles: false)),
                  bottomTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      reservedSize: 24,
                      getTitlesWidget: (value, _) {
                        final idx = value.toInt();
                        if (idx < 0 || idx >= last6.length) {
                          return const SizedBox.shrink();
                        }
                        return Padding(
                          padding: const EdgeInsets.only(top: 6),
                          child: Text(
                            last6[idx].label.split('/')[0],
                            style: const TextStyle(
                                fontSize: 11,
                                color: AppColors.textLow,
                                fontWeight: FontWeight.w500),
                          ),
                        );
                      },
                    ),
                  ),
                ),
                gridData: FlGridData(
                  show: true,
                  drawVerticalLine: false,
                  getDrawingHorizontalLine: (_) => FlLine(
                    color: AppColors.borderLight,
                    strokeWidth: 1,
                  ),
                ),
                borderData: FlBorderData(show: false),
                barGroups: List.generate(last6.length, (i) {
                  final cf = last6[i];
                  return BarChartGroupData(
                    x: i,
                    barRods: [
                      BarChartRodData(
                        toY: cf.income,
                        gradient: const LinearGradient(
                          colors: [Color(0xFF3B5BDB), Color(0xFF7395F5)],
                          begin: Alignment.bottomCenter,
                          end: Alignment.topCenter,
                        ),
                        width: 12,
                        borderRadius: const BorderRadius.vertical(
                            top: Radius.circular(5)),
                      ),
                      BarChartRodData(
                        toY: cf.expense,
                        gradient: LinearGradient(
                          colors: [
                            AppColors.expense.withValues(alpha: 0.8),
                            AppColors.expense.withValues(alpha: 0.4),
                          ],
                          begin: Alignment.bottomCenter,
                          end: Alignment.topCenter,
                        ),
                        width: 12,
                        borderRadius: const BorderRadius.vertical(
                            top: Radius.circular(5)),
                      ),
                    ],
                    barsSpace: 4,
                  );
                }),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _LegendDot extends StatelessWidget {
  const _LegendDot({required this.color, required this.label});
  final Color color;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          width: 10,
          height: 10,
          decoration:
              BoxDecoration(color: color, shape: BoxShape.circle),
        ),
        const SizedBox(width: 5),
        Text(label,
            style: const TextStyle(
                fontSize: 11,
                color: AppColors.textMedium,
                fontWeight: FontWeight.w500)),
      ],
    );
  }
}

// ── Top Categories ────────────────────────────────────────────────────────────

class _TopCategoriesList extends StatelessWidget {
  const _TopCategoriesList({required this.categories});
  final List<CategorySummaryModel> categories;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final maxTotal =
        categories.isEmpty ? 1.0 : categories.first.total;

    return SurfaceCard(
      child: Column(
        children: [
          ...categories.map((cat) {
            final proportion =
                maxTotal > 0 ? cat.total / maxTotal : 0.0;
            final color = _parseColor(cat.color);
            return Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Row(
                children: [
                  Container(
                    width: 38,
                    height: 38,
                    decoration: BoxDecoration(
                      color: color.withValues(alpha: 0.12),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child:
                        Icon(Icons.label_rounded, color: color, size: 18),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment:
                              MainAxisAlignment.spaceBetween,
                          children: [
                            Text(cat.categoryName,
                                style: const TextStyle(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w600)),
                            Text(fmt.format(cat.total),
                                style: const TextStyle(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w700)),
                          ],
                        ),
                        const SizedBox(height: 6),
                        ClipRRect(
                          borderRadius: BorderRadius.circular(4),
                          child: LinearProgressIndicator(
                            value: proportion.clamp(0.0, 1.0),
                            backgroundColor: AppColors.surfaceElevatedLight,
                            valueColor:
                                AlwaysStoppedAnimation<Color>(color),
                            minHeight: 5,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  Color _parseColor(String? hex) {
    if (hex == null || hex.isEmpty) return AppColors.primary;
    try {
      return Color(int.parse('FF${hex.replaceAll('#', '')}', radix: 16));
    } catch (_) {
      return AppColors.primary;
    }
  }
}

// ── Alert Budgets ─────────────────────────────────────────────────────────────

class _AlertBudgetsList extends StatelessWidget {
  const _AlertBudgetsList({required this.budgets});
  final List<BudgetAlertModel> budgets;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Column(
        children: budgets.map((b) {
          final isOver = b.percentage >= 100;
          final accentColor =
              isOver ? AppColors.expense : AppColors.warning;

          return Container(
            margin: const EdgeInsets.only(bottom: 10),
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: accentColor.withValues(alpha: 0.06),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                  color: accentColor.withValues(alpha: 0.3)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(b.categoryName,
                        style: const TextStyle(
                            fontWeight: FontWeight.w700,
                            fontSize: 14)),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 8, vertical: 3),
                      decoration: BoxDecoration(
                        color: accentColor,
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        '${b.percentage.toStringAsFixed(0)}%',
                        style: const TextStyle(
                            color: Colors.white,
                            fontSize: 11,
                            fontWeight: FontWeight.w700),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: b.progressValue,
                    backgroundColor: AppColors.borderLight,
                    valueColor:
                        AlwaysStoppedAnimation<Color>(accentColor),
                    minHeight: 6,
                  ),
                ),
                const SizedBox(height: 6),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Gasto: ${fmt.format(b.actual)}',
                        style: const TextStyle(
                            fontSize: 11,
                            color: AppColors.textMedium)),
                    Text('Limite: ${fmt.format(b.planned)}',
                        style: const TextStyle(
                            fontSize: 11,
                            color: AppColors.textMedium)),
                  ],
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }
}

// ── Transaction Tile ──────────────────────────────────────────────────────────

class _TransactionTile extends StatelessWidget {
  const _TransactionTile({required this.transaction});
  final RecentTransactionModel transaction;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final isExpense = transaction.isExpense;
    final isTransfer = transaction.isTransfer;

    final Color color;
    final String prefix;
    final IconData icon;

    if (isTransfer) {
      color = AppColors.primary;
      prefix = '';
      icon = Icons.swap_horiz_rounded;
    } else if (isExpense) {
      color = AppColors.expense;
      prefix = '- ';
      icon = Icons.arrow_downward_rounded;
    } else {
      color = AppColors.income;
      prefix = '+ ';
      icon = Icons.arrow_upward_rounded;
    }

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 10),
      child: Row(
        children: [
          Container(
            width: 42,
            height: 42,
            decoration: BoxDecoration(
              color: color.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(icon, color: color, size: 18),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  transaction.description ??
                      transaction.categoryName ??
                      'Sem descrição',
                  style: const TextStyle(
                      fontSize: 14, fontWeight: FontWeight.w600),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),
                Text(
                  '${transaction.accountName ?? ''} • ${_formatDate(transaction.date)}',
                  style: const TextStyle(
                      fontSize: 11, color: AppColors.textLow),
                ),
              ],
            ),
          ),
          Text(
            '$prefix${fmt.format(transaction.amount)}',
            style: TextStyle(
                color: color,
                fontWeight: FontWeight.w700,
                fontSize: 14),
          ),
        ],
      ),
    );
  }

  String _formatDate(String dateStr) {
    try {
      return DateFormat('dd/MM/yy', 'pt_BR').format(DateTime.parse(dateStr));
    } catch (_) {
      return dateStr;
    }
  }
}
