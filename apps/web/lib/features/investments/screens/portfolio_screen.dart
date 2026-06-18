import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:intl/intl.dart';
import '../providers/investments_provider.dart';
import '../models/holding_model.dart';
import '../models/currency_quote_model.dart';

class PortfolioScreen extends ConsumerStatefulWidget {
  const PortfolioScreen({super.key});

  @override
  ConsumerState<PortfolioScreen> createState() => _PortfolioScreenState();
}

class _PortfolioScreenState extends ConsumerState<PortfolioScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      final notifier = ref.read(investmentsProvider.notifier);
      notifier.loadPortfolios();
      notifier.loadCurrencyQuotes();
    });
  }

  Future<void> _showCreatePortfolioDialog(BuildContext context) async {
    final nameController = TextEditingController();
    final formKey = GlobalKey<FormState>();

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Novo portfólio'),
        content: Form(
          key: formKey,
          child: TextFormField(
            controller: nameController,
            autofocus: true,
            decoration: const InputDecoration(
              labelText: 'Nome do portfólio',
              hintText: 'Ex: Renda Variável',
              border: OutlineInputBorder(),
            ),
            validator: (v) =>
                (v == null || v.trim().isEmpty) ? 'Informe um nome' : null,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Cancelar'),
          ),
          ElevatedButton(
            onPressed: () {
              if (formKey.currentState!.validate()) {
                Navigator.of(ctx).pop(true);
              }
            },
            child: const Text('Criar'),
          ),
        ],
      ),
    );

    if (confirmed == true && mounted) {
      await ref.read(investmentsProvider.notifier).createPortfolio({
        'name': nameController.text.trim(),
      });
      nameController.dispose();
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(investmentsProvider);
    final currency = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Carteira'),
        actions: [
          IconButton(
            icon: const Icon(Icons.create_new_folder_outlined),
            tooltip: 'Novo portfólio',
            onPressed: () => _showCreatePortfolioDialog(context),
          ),
          IconButton(
            icon: const Icon(Icons.add),
            tooltip: 'Nova operação',
            onPressed: () => context.push('/investments/new'),
          ),
          IconButton(
            icon: const Icon(Icons.pie_chart_outline),
            tooltip: 'Análise',
            onPressed: () => context.push('/investments/analysis'),
          ),
          IconButton(
            icon: const Icon(Icons.receipt_long_outlined),
            tooltip: 'Relatório de IR',
            onPressed: () => context.push('/investments/tax-report'),
          ),
        ],
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : state.error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(state.error!,
                          style: const TextStyle(color: Colors.red)),
                      TextButton(
                        onPressed: () => ref
                            .read(investmentsProvider.notifier)
                            .loadPortfolios(),
                        child: const Text('Tentar novamente'),
                      ),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: () async {
                    final notifier =
                        ref.read(investmentsProvider.notifier);
                    await Future.wait([
                      notifier.loadPortfolios(),
                      notifier.loadCurrencyQuotes(),
                    ]);
                  },
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Summary card
                        _SummaryCard(
                          totalInvested: state.totalInvested,
                          totalCurrentValue: state.totalCurrentValue,
                          totalPnl: state.totalUnrealizedPnl,
                          currency: currency,
                        ),
                        const SizedBox(height: 16),
                        // Currency quotes (USD, EUR)
                        if (state.currencyQuotes.isNotEmpty) ...[
                          _CurrencyQuotesCard(quotes: state.currencyQuotes),
                          const SizedBox(height: 16),
                        ],
                        // Allocation pie chart
                        if (state.holdings.isNotEmpty) ...[
                          _AllocationChart(holdings: state.holdings),
                          const SizedBox(height: 16),
                        ],
                        // Holdings list
                        const Text(
                          'Posições',
                          style: TextStyle(
                              fontSize: 18, fontWeight: FontWeight.bold),
                        ),
                        const SizedBox(height: 8),
                        if (state.portfolios.isEmpty)
                          Center(
                            child: Padding(
                              padding: const EdgeInsets.all(32),
                              child: Column(
                                children: [
                                  const Icon(Icons.account_balance_wallet_outlined,
                                      size: 48, color: Colors.grey),
                                  const SizedBox(height: 12),
                                  const Text(
                                    'Nenhum portfólio encontrado.\nCrie um portfólio para começar.',
                                    textAlign: TextAlign.center,
                                    style: TextStyle(color: Colors.grey),
                                  ),
                                  const SizedBox(height: 16),
                                  ElevatedButton.icon(
                                    onPressed: () => _showCreatePortfolioDialog(context),
                                    icon: const Icon(Icons.add),
                                    label: const Text('Criar portfólio'),
                                  ),
                                ],
                              ),
                            ),
                          )
                        else if (state.holdings.isEmpty)
                          const Center(
                            child: Padding(
                              padding: EdgeInsets.all(32),
                              child: Text('Nenhuma posição encontrada.\nAdicione sua primeira operação.', textAlign: TextAlign.center),
                            ),
                          )
                        else
                          ...state.holdings.map(
                            (h) => _HoldingCard(
                              holding: h,
                              currency: currency,
                              onTap: () => context.push(
                                  '/investments/holdings/${h.id}'),
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/investments/custom/new'),
        icon: const Icon(Icons.add_business),
        label: const Text('Ativo personalizado'),
      ),
    );
  }
}

class _SummaryCard extends StatelessWidget {
  final double totalInvested;
  final double totalCurrentValue;
  final double totalPnl;
  final NumberFormat currency;

  const _SummaryCard({
    required this.totalInvested,
    required this.totalCurrentValue,
    required this.totalPnl,
    required this.currency,
  });

  @override
  Widget build(BuildContext context) {
    final pnlPct = totalInvested > 0 ? (totalPnl / totalInvested) * 100 : 0.0;
    final pnlColor = totalPnl >= 0 ? const Color(0xFF22C55E) : const Color(0xFFEF4444);

    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Resumo da Carteira',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                _MetricItem(label: 'Investido', value: currency.format(totalInvested)),
                _MetricItem(label: 'Atual', value: currency.format(totalCurrentValue)),
                _MetricItem(
                  label: 'P&L',
                  value: '${totalPnl >= 0 ? '+' : ''}${currency.format(totalPnl)}',
                  color: pnlColor,
                  subtitle: '${pnlPct >= 0 ? '+' : ''}${pnlPct.toStringAsFixed(2)}%',
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _MetricItem extends StatelessWidget {
  final String label;
  final String value;
  final Color? color;
  final String? subtitle;

  const _MetricItem({
    required this.label,
    required this.value,
    this.color,
    this.subtitle,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label,
            style: const TextStyle(fontSize: 12, color: Colors.grey)),
        Text(value,
            style: TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.bold,
              color: color,
            )),
        if (subtitle != null)
          Text(subtitle!,
              style: TextStyle(fontSize: 12, color: color ?? Colors.grey)),
      ],
    );
  }
}

class _CurrencyQuotesCard extends StatelessWidget {
  final List<CurrencyQuoteModel> quotes;

  const _CurrencyQuotesCard({required this.quotes});

  @override
  Widget build(BuildContext context) {
    final fmt = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final theme = Theme.of(context);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest.withOpacity(0.5),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.max,
        children: [
          const Icon(Icons.currency_exchange, size: 14, color: Colors.grey),
          const SizedBox(width: 6),
          ...quotes.map((q) => Padding(
                padding: const EdgeInsets.only(right: 12),
                child: Text(
                  '${q.code}: ${fmt.format(q.bid)}',
                  style: theme.textTheme.bodySmall?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
                ),
              )),
        ],
      ),
    );
  }
}

class _AllocationChart extends StatelessWidget {
  final List<HoldingModel> holdings;

  const _AllocationChart({required this.holdings});

  @override
  Widget build(BuildContext context) {
    // Group by type
    final Map<String, double> byType = {};
    for (final h in holdings) {
      byType[h.type] = (byType[h.type] ?? 0) + h.currentValue;
    }

    final total = byType.values.fold(0.0, (a, b) => a + b);
    if (total == 0) return const SizedBox.shrink();

    final colors = [
      const Color(0xFF6366F1),
      const Color(0xFF22C55E),
      const Color(0xFFF59E0B),
      const Color(0xFFEF4444),
      const Color(0xFF3B82F6),
      const Color(0xFF8B5CF6),
    ];

    int colorIdx = 0;
    final sections = byType.entries.map((e) {
      final pct = (e.value / total) * 100;
      final color = colors[colorIdx % colors.length];
      colorIdx++;
      return PieChartSectionData(
        value: e.value,
        title: '${pct.toStringAsFixed(1)}%',
        color: color,
        radius: 80,
        titleStyle: const TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.bold,
          color: Colors.white,
        ),
      );
    }).toList();

    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Alocação por Tipo',
                style:
                    TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 12),
            SizedBox(
              height: 200,
              child: Row(
                children: [
                  Expanded(
                    child: PieChart(
                      PieChartData(
                        sections: sections,
                        centerSpaceRadius: 40,
                        sectionsSpace: 2,
                      ),
                    ),
                  ),
                  const SizedBox(width: 16),
                  Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: byType.entries.mapIndexed((i, e) {
                      return Padding(
                        padding: const EdgeInsets.symmetric(vertical: 2),
                        child: Row(
                          children: [
                            Container(
                              width: 12,
                              height: 12,
                              decoration: BoxDecoration(
                                color: colors[i % colors.length],
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 6),
                            Text(e.key,
                                style: const TextStyle(fontSize: 12)),
                          ],
                        ),
                      );
                    }).toList(),
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

class _HoldingCard extends StatelessWidget {
  final HoldingModel holding;
  final NumberFormat currency;
  final VoidCallback onTap;

  const _HoldingCard({
    required this.holding,
    required this.currency,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        onTap: onTap,
        title: Text(
          holding.displayTicker,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Text(
          '${holding.quantity.toStringAsFixed(2)} un. @ ${currency.format(holding.avgPrice)}',
          style: const TextStyle(fontSize: 12),
        ),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(
              currency.format(holding.currentValue),
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
            ),
            Text(
              '${holding.unrealizedPnl >= 0 ? '+' : ''}${holding.unrealizedPnlPct.toStringAsFixed(2)}%',
              style: TextStyle(
                color: holding.pnlColor,
                fontSize: 12,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// Helper extension for mapIndexed
extension IterableIndexed<T> on Iterable<T> {
  Iterable<R> mapIndexed<R>(R Function(int index, T item) f) sync* {
    var index = 0;
    for (final item in this) {
      yield f(index, item);
      index++;
    }
  }
}
