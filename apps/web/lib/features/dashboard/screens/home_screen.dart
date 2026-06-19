import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

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
      body: _DashboardBody(
        dashState: dashState,
        accountsState: accountsState,
      ),
    );
  }
}

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
        SliverAppBar(
          floating: true,
          snap: true,
          title: const Text('FinanceOS'),
          centerTitle: false,
          actions: [
            IconButton(
              icon: const Icon(Icons.notifications_none),
              tooltip: 'Notificações',
              onPressed: () => context.push('/notifications'),
            ),
            IconButton(
              icon: const Icon(Icons.person_outline),
              tooltip: 'Perfil',
              onPressed: () => context.push('/settings/profile'),
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
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    'Erro ao carregar dados',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  ElevatedButton(
                    onPressed: () =>
                        ref.read(dashboardProvider.notifier).load(),
                    child: const Text('Tentar novamente'),
                  ),
                ],
              ),
            ),
          )
        else ...[
          // Net Balance Card
          SliverToBoxAdapter(
            child: _NetBalanceCard(overview: overview),
          ),
          // Net Worth + Investment Capacity cards
          SliverToBoxAdapter(
            child: _NetWorthAndCapacityRow(overview: overview, patrimonyHistory: dashState.patrimonyHistory),
          ),
          // Month selector
          SliverToBoxAdapter(
            child: _MonthSelector(
              month: dashState.month,
              year: dashState.year,
            ),
          ),
          // Accounts mini cards
          if (accountsState.accounts.isNotEmpty) ...[
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.fromLTRB(16, 16, 16, 4),
                child: Text(
                  'Contas',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: SizedBox(
                height: 110,
                child: ListView.separated(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  itemCount: accountsState.accounts.length,
                  separatorBuilder: (context, index) => const SizedBox(width: 10),
                  itemBuilder: (context, index) {
                    return _AccountMiniCard(
                        account: accountsState.accounts[index]);
                  },
                ),
              ),
            ),
          ],
          // Bar chart cashflow
          if (dashState.cashflow.isNotEmpty) ...[
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.fromLTRB(16, 20, 16, 4),
                child: Text(
                  'Receitas vs Despesas',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: _CashflowChart(cashflow: dashState.cashflow),
            ),
          ],
          // Top categories
          if (overview != null && overview.topCategories.isNotEmpty) ...[
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.fromLTRB(16, 20, 16, 4),
                child: Text(
                  'Maiores Gastos',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: _TopCategoriesList(categories: overview.topCategories),
            ),
          ],
          // Alert budgets
          if (overview != null && overview.alertBudgets.isNotEmpty) ...[
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.fromLTRB(16, 20, 16, 4),
                child: Text(
                  'Orçamentos em Alerta',
                  style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Colors.red),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: _AlertBudgetsList(budgets: overview.alertBudgets),
            ),
          ],
          // Quick actions
          SliverToBoxAdapter(
            child: _QuickActions(),
          ),
          // Recent transactions
          if (overview != null && overview.recentTransactions.isNotEmpty) ...[
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.fromLTRB(16, 20, 16, 4),
                child: Text(
                  'Transações Recentes',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            SliverList(
              delegate: SliverChildBuilderDelegate(
                (context, index) {
                  return _RecentTransactionTile(
                    transaction: overview.recentTransactions[index],
                  );
                },
                childCount: overview.recentTransactions.length,
              ),
            ),
          ],
          const SliverToBoxAdapter(child: SizedBox(height: 24)),
        ],
      ],
    );
  }
}

// ------ Net Balance Card ------

class _NetBalanceCard extends StatelessWidget {
  const _NetBalanceCard({this.overview});
  final DashboardOverview? overview;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final netBalance = overview?.netBalance ?? 0.0;
    final totalIncome = overview?.totalIncome ?? 0.0;
    final totalExpense = overview?.totalExpense ?? 0.0;
    final isPositive = netBalance >= 0;

    return Container(
      margin: const EdgeInsets.fromLTRB(16, 16, 16, 0),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: isPositive
              ? [const Color(0xFF1565C0), const Color(0xFF42A5F5)]
              : [const Color(0xFFB71C1C), const Color(0xFFEF5350)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.15),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Saldo Líquido',
            style: TextStyle(color: Colors.white70, fontSize: 14),
          ),
          const SizedBox(height: 8),
          Text(
            fmt.format(netBalance),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 32,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _BalanceStat(
                  label: 'Receitas',
                  value: fmt.format(totalIncome),
                  icon: Icons.arrow_upward,
                  color: Colors.greenAccent,
                ),
              ),
              Expanded(
                child: _BalanceStat(
                  label: 'Despesas',
                  value: fmt.format(totalExpense),
                  icon: Icons.arrow_downward,
                  color: Colors.redAccent,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _BalanceStat extends StatelessWidget {
  const _BalanceStat({
    required this.label,
    required this.value,
    required this.icon,
    required this.color,
  });

  final String label;
  final String value;
  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(icon, color: color, size: 16),
        const SizedBox(width: 4),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label,
                style: const TextStyle(color: Colors.white60, fontSize: 11)),
            Text(value,
                style: const TextStyle(
                    color: Colors.white,
                    fontSize: 13,
                    fontWeight: FontWeight.w600)),
          ],
        ),
      ],
    );
  }
}

// ------ Net Worth + Investment Capacity Row ------

class _NetWorthAndCapacityRow extends StatelessWidget {
  const _NetWorthAndCapacityRow({required this.overview, required this.patrimonyHistory});
  final DashboardOverview? overview;
  final List<PatrimonySnapshotModel> patrimonyHistory;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final netWorth = overview?.totalNetWorth ?? 0.0;
    final investValue = overview?.investmentValue ?? 0.0;
    final customValue = overview?.customAssetValue ?? 0.0;
    final capacity = overview?.investmentCapacity ?? 0.0;
    final capacityPct = overview?.investmentCapacityPct ?? 0.0;

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Patrimônio total card
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Patrimônio Total',
                    style: TextStyle(fontSize: 13, fontWeight: FontWeight.w500)),
                const SizedBox(height: 4),
                Text(fmt.format(netWorth),
                    style: const TextStyle(
                        fontSize: 24, fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                Row(
                  children: [
                    _PatrimonyChip(
                        label: 'Investimentos', value: fmt.format(investValue),
                        color: Colors.blue),
                    const SizedBox(width: 8),
                    _PatrimonyChip(
                        label: 'Outros ativos', value: fmt.format(customValue),
                        color: Colors.purple),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 8),
          // Investment capacity card
          if (capacity > 0)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: Colors.green.shade50,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.green.shade200),
              ),
              child: Row(
                children: [
                  Icon(Icons.trending_up, color: Colors.green.shade700, size: 28),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Capacidade de Investir este mês',
                            style: TextStyle(
                                fontSize: 12, color: Colors.green.shade700)),
                        Text(fmt.format(capacity),
                            style: TextStyle(
                                fontSize: 20,
                                fontWeight: FontWeight.bold,
                                color: Colors.green.shade800)),
                        Text(
                            '${capacityPct.toStringAsFixed(1)}% da sua renda mensal',
                            style: TextStyle(
                                fontSize: 11, color: Colors.green.shade600)),
                      ],
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class _PatrimonyChip extends StatelessWidget {
  const _PatrimonyChip(
      {required this.label, required this.value, required this.color});
  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label,
              style: TextStyle(fontSize: 10, color: color)),
          Text(value,
              style: TextStyle(
                  fontSize: 12, fontWeight: FontWeight.bold, color: color)),
        ],
      ),
    );
  }
}

// ------ Month Selector ------

class _MonthSelector extends ConsumerWidget {
  const _MonthSelector({required this.month, required this.year});
  final int month;
  final int year;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final monthNames = [
      'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
      'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro',
    ];

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          IconButton(
            icon: const Icon(Icons.chevron_left),
            onPressed: () =>
                ref.read(dashboardProvider.notifier).changeMonth(-1),
          ),
          Text(
            '${monthNames[month - 1]} $year',
            style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
          ),
          IconButton(
            icon: const Icon(Icons.chevron_right),
            onPressed: () =>
                ref.read(dashboardProvider.notifier).changeMonth(1),
          ),
        ],
      ),
    );
  }
}

// ------ Account Mini Card ------

class _AccountMiniCard extends StatelessWidget {
  const _AccountMiniCard({required this.account});

  final AccountModel account;

  String get _typeIcon {
    switch (account.type) {
      case 'checking':
        return '🏦';
      case 'savings':
        return '💰';
      case 'credit_card':
        return '💳';
      case 'investment':
        return '📈';
      case 'wallet':
        return '👛';
      default:
        return '💵';
    }
  }

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final balance = account.balance;
    final isNegative = balance < 0;

    return Container(
      width: 140,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: Theme.of(context).colorScheme.outline.withValues(alpha: 0.3),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(_typeIcon, style: const TextStyle(fontSize: 20)),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                account.name,
                style:
                    const TextStyle(fontSize: 12, fontWeight: FontWeight.w600),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 2),
              Text(
                fmt.format(balance),
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.bold,
                  color: isNegative ? Colors.red : Colors.green.shade700,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

// ------ Cashflow BarChart ------

class _CashflowChart extends StatelessWidget {
  const _CashflowChart({required this.cashflow});
  final List<MonthlyCashflowModel> cashflow;

  @override
  Widget build(BuildContext context) {
    final last6 = cashflow.length > 6
        ? cashflow.sublist(cashflow.length - 6)
        : cashflow;

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16),
      padding: const EdgeInsets.all(16),
      height: 220,
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: Theme.of(context).colorScheme.outline.withValues(alpha: 0.2),
        ),
      ),
      child: BarChart(
        BarChartData(
          alignment: BarChartAlignment.spaceAround,
          barTouchData: BarTouchData(enabled: false),
          titlesData: FlTitlesData(
            leftTitles: const AxisTitles(
              sideTitles: SideTitles(showTitles: false),
            ),
            rightTitles: const AxisTitles(
              sideTitles: SideTitles(showTitles: false),
            ),
            topTitles: const AxisTitles(
              sideTitles: SideTitles(showTitles: false),
            ),
            bottomTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                getTitlesWidget: (value, meta) {
                  final idx = value.toInt();
                  if (idx < 0 || idx >= last6.length) {
                    return const SizedBox.shrink();
                  }
                  final label = last6[idx].label.split('/')[0];
                  return Text(
                    label,
                    style: const TextStyle(fontSize: 10),
                  );
                },
              ),
            ),
          ),
          gridData: const FlGridData(show: false),
          borderData: FlBorderData(show: false),
          barGroups: List.generate(last6.length, (i) {
            final cf = last6[i];
            return BarChartGroupData(
              x: i,
              barRods: [
                BarChartRodData(
                  toY: cf.income,
                  color: Colors.blue.shade400,
                  width: 10,
                  borderRadius: const BorderRadius.vertical(
                      top: Radius.circular(4)),
                ),
                BarChartRodData(
                  toY: cf.expense,
                  color: Colors.red.shade400,
                  width: 10,
                  borderRadius: const BorderRadius.vertical(
                      top: Radius.circular(4)),
                ),
              ],
              barsSpace: 4,
            );
          }),
        ),
      ),
    );
  }
}

// ------ Top Categories ------

class _TopCategoriesList extends StatelessWidget {
  const _TopCategoriesList({required this.categories});
  final List<CategorySummaryModel> categories;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final maxTotal =
        categories.isEmpty ? 1.0 : categories.first.total;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        children: categories.map((cat) {
          final proportion = maxTotal > 0 ? cat.total / maxTotal : 0.0;
          return Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: Row(
              children: [
                Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(
                    color: _parseColor(cat.color).withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(
                    Icons.label,
                    color: _parseColor(cat.color),
                    size: 18,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            cat.categoryName,
                            style: const TextStyle(
                                fontSize: 13, fontWeight: FontWeight.w500),
                          ),
                          Text(
                            fmt.format(cat.total),
                            style: const TextStyle(
                                fontSize: 13, fontWeight: FontWeight.bold),
                          ),
                        ],
                      ),
                      const SizedBox(height: 4),
                      LinearProgressIndicator(
                        value: proportion.toDouble().clamp(0.0, 1.0),
                        backgroundColor: Colors.grey.shade200,
                        color: _parseColor(cat.color),
                        minHeight: 4,
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }

  Color _parseColor(String? hex) {
    if (hex == null || hex.isEmpty) return Colors.blueGrey;
    try {
      final clean = hex.replaceAll('#', '');
      return Color(int.parse('FF$clean', radix: 16));
    } catch (_) {
      return Colors.blueGrey;
    }
  }
}

// ------ Alert Budgets ------

class _AlertBudgetsList extends StatelessWidget {
  const _AlertBudgetsList({required this.budgets});
  final List<BudgetAlertModel> budgets;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        children: budgets.map((b) {
          return Card(
            margin: const EdgeInsets.only(bottom: 8),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(10),
              side: const BorderSide(color: Colors.red, width: 0.5),
            ),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        b.categoryName,
                        style: const TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 14),
                      ),
                      Text(
                        '${b.percentage.toStringAsFixed(0)}%',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: b.percentage >= 100
                              ? Colors.red
                              : Colors.orange,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  LinearProgressIndicator(
                    value: b.progressValue,
                    backgroundColor: Colors.grey.shade200,
                    color: b.percentage >= 100 ? Colors.red : Colors.orange,
                    minHeight: 6,
                    borderRadius: BorderRadius.circular(3),
                  ),
                  const SizedBox(height: 4),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Gasto: ${fmt.format(b.actual)}',
                        style:
                            const TextStyle(fontSize: 12, color: Colors.grey),
                      ),
                      Text(
                        'Orçado: ${fmt.format(b.planned)}',
                        style:
                            const TextStyle(fontSize: 12, color: Colors.grey),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          );
        }).toList(),
      ),
    );
  }
}

// ------ Quick Actions ------

class _QuickActions extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 20, 16, 0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Atalhos Rápidos',
            style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: _QuickActionButton(
                  icon: Icons.add_circle_outline,
                  label: 'Nova Transação',
                  color: Colors.green,
                  onTap: () => context.go('/transactions/new'),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: _QuickActionButton(
                  icon: Icons.account_balance_wallet_outlined,
                  label: 'Ver Contas',
                  color: Colors.blue,
                  onTap: () => context.go('/accounts'),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: _QuickActionButton(
                  icon: Icons.trending_up,
                  label: 'Investimentos',
                  color: Colors.purple,
                  onTap: () => context.go('/investments'),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _QuickActionButton extends StatelessWidget {
  const _QuickActionButton({
    required this.icon,
    required this.label,
    required this.color,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: color.withValues(alpha: 0.3)),
        ),
        child: Column(
          children: [
            Icon(icon, color: color, size: 24),
            const SizedBox(height: 6),
            Text(
              label,
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 11,
                color: color,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ------ Recent Transactions ------

class _RecentTransactionTile extends StatelessWidget {
  const _RecentTransactionTile({required this.transaction});
  final RecentTransactionModel transaction;

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final isExpense = transaction.isExpense;
    final isTransfer = transaction.isTransfer;
    final amount = transaction.amount;

    Color amountColor;
    String amountPrefix;
    IconData typeIcon;

    if (isTransfer) {
      amountColor = Colors.blue;
      amountPrefix = '';
      typeIcon = Icons.swap_horiz;
    } else if (isExpense) {
      amountColor = Colors.red;
      amountPrefix = '- ';
      typeIcon = Icons.arrow_downward;
    } else {
      amountColor = Colors.green;
      amountPrefix = '+ ';
      typeIcon = Icons.arrow_upward;
    }

    return ListTile(
      leading: Container(
        width: 40,
        height: 40,
        decoration: BoxDecoration(
          color: amountColor.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(typeIcon, color: amountColor, size: 18),
      ),
      title: Text(
        transaction.description ??
            transaction.categoryName ??
            'Sem descrição',
        style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: Text(
        '${transaction.accountName ?? ''} • ${_formatDate(transaction.date)}',
        style: const TextStyle(fontSize: 11, color: Colors.grey),
      ),
      trailing: Text(
        '$amountPrefix${fmt.format(amount)}',
        style: TextStyle(
          color: amountColor,
          fontWeight: FontWeight.bold,
          fontSize: 14,
        ),
      ),
    );
  }

  String _formatDate(String dateStr) {
    try {
      final dt = DateTime.parse(dateStr);
      return DateFormat('dd/MM/yy', 'pt_BR').format(dt);
    } catch (_) {
      return dateStr;
    }
  }
}
