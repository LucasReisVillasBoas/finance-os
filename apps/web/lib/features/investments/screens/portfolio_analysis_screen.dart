import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:intl/intl.dart';
import '../providers/investments_provider.dart';
import '../models/holding_model.dart';

class PortfolioAnalysisScreen extends ConsumerStatefulWidget {
  const PortfolioAnalysisScreen({super.key});

  @override
  ConsumerState<PortfolioAnalysisScreen> createState() =>
      _PortfolioAnalysisScreenState();
}

class _PortfolioAnalysisScreenState
    extends ConsumerState<PortfolioAnalysisScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
        () => ref.read(investmentsProvider.notifier).loadPortfolios());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(investmentsProvider);
    final currency = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Análise da Carteira'),
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : state.holdings.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.pie_chart_outline,
                          size: 64, color: Colors.grey),
                      const SizedBox(height: 16),
                      const Text(
                        'Nenhuma posição encontrada.\nAdicione ativos à sua carteira.',
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey),
                      ),
                      TextButton(
                        onPressed: () => ref
                            .read(investmentsProvider.notifier)
                            .loadPortfolios(),
                        child: const Text('Recarregar'),
                      ),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: () =>
                      ref.read(investmentsProvider.notifier).loadPortfolios(),
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Benchmark comparison (real portfolio return vs estimates)
                        _BenchmarkCard(state: state),
                        const SizedBox(height: 16),

                        // Diversification pie chart
                        _DiversificationChart(holdings: state.holdings),
                        const SizedBox(height: 16),

                        // Concentration list
                        _ConcentrationList(
                          holdings: state.holdings,
                          currency: currency,
                        ),
                      ],
                    ),
                  ),
                ),
    );
  }
}

class _BenchmarkCard extends ConsumerStatefulWidget {
  final InvestmentsState state;
  const _BenchmarkCard({required this.state});

  @override
  ConsumerState<_BenchmarkCard> createState() => _BenchmarkCardState();
}

class _BenchmarkCardState extends ConsumerState<_BenchmarkCard> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
        () => ref.read(investmentsProvider.notifier).loadPortfolioPerformance());
  }

  @override
  Widget build(BuildContext context) {
    final perf = widget.state.portfolioPerformance;
    final pct = NumberFormat('+0.##%;-0.##%', 'pt_BR');

    final portfolioReturn = perf != null
        ? pct.format((perf['return_pct'] as num? ?? 0) / 100)
        : '—';
    final cdiRate = perf != null
        ? '+${(perf['cdi_estimate_pct'] as num? ?? 0).toStringAsFixed(2)}%'
        : '—';
    final ibovRate = perf != null
        ? '+${(perf['ibov_estimate_pct'] as num? ?? 0).toStringAsFixed(2)}%'
        : '—';

    final returnPct = (perf?['return_pct'] as num?)?.toDouble() ?? 0;
    final cdiEstimate = (perf?['cdi_estimate_pct'] as num?)?.toDouble() ?? 0;

    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Rentabilidade vs Benchmarks',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: _BenchmarkItem(
                    name: 'Sua Carteira',
                    value: portfolioReturn,
                    period: 'total',
                    color: returnPct >= 0
                        ? const Color(0xFF22C55E)
                        : Colors.red,
                  ),
                ),
                Expanded(
                  child: _BenchmarkItem(
                    name: 'CDI',
                    value: cdiRate,
                    period: 'a.a. est.',
                    color: const Color(0xFF3B82F6),
                  ),
                ),
                Expanded(
                  child: _BenchmarkItem(
                    name: 'IBOV',
                    value: ibovRate,
                    period: 'a.a. est.',
                    color: const Color(0xFFF59E0B),
                  ),
                ),
              ],
            ),
            if (perf != null && returnPct > 0) ...[
              const SizedBox(height: 8),
              LinearProgressIndicator(
                value: (returnPct / (cdiEstimate > 0 ? cdiEstimate * 2 : 20))
                    .clamp(0.0, 1.0),
                color: returnPct >= cdiEstimate ? Colors.green : Colors.orange,
                backgroundColor: Colors.grey.shade200,
              ),
              const SizedBox(height: 4),
              Text(
                returnPct >= cdiEstimate
                    ? '✓ Sua carteira está batendo o CDI'
                    : '⚠ Sua carteira está abaixo do CDI',
                style: TextStyle(
                    fontSize: 12,
                    color: returnPct >= cdiEstimate ? Colors.green : Colors.orange),
              ),
            ],
            const SizedBox(height: 4),
            const Text(
              'CDI e IBOV são estimativas. Rentabilidade da carteira calculada sobre custo médio atual.',
              style: TextStyle(fontSize: 11, color: Colors.grey),
            ),
          ],
        ),
      ),
    );
  }
}

class _BenchmarkItem extends StatelessWidget {
  final String name;
  final String value;
  final String period;
  final Color color;

  const _BenchmarkItem({
    required this.name,
    required this.value,
    required this.period,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(name,
            style: const TextStyle(
                fontSize: 12, fontWeight: FontWeight.w600)),
        const SizedBox(height: 4),
        Text(value,
            style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: color)),
        Text(period,
            style: const TextStyle(fontSize: 11, color: Colors.grey)),
      ],
    );
  }
}

class _DiversificationChart extends StatelessWidget {
  final List<HoldingModel> holdings;

  const _DiversificationChart({required this.holdings});

  @override
  Widget build(BuildContext context) {
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
        radius: 90,
        titleStyle: const TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.bold,
          color: Colors.white,
        ),
      );
    }).toList();

    final legendColors = <String, Color>{};
    int lIdx = 0;
    for (final key in byType.keys) {
      legendColors[key] = colors[lIdx % colors.length];
      lIdx++;
    }

    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Diversificação por Tipo',
                style:
                    TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 12),
            SizedBox(
              height: 220,
              child: Row(
                children: [
                  Expanded(
                    child: PieChart(
                      PieChartData(
                        sections: sections,
                        centerSpaceRadius: 45,
                        sectionsSpace: 2,
                      ),
                    ),
                  ),
                  const SizedBox(width: 16),
                  Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: legendColors.entries.map((e) {
                      final pct = ((byType[e.key] ?? 0) / total * 100);
                      return Padding(
                        padding: const EdgeInsets.symmetric(vertical: 3),
                        child: Row(
                          children: [
                            Container(
                              width: 12,
                              height: 12,
                              decoration: BoxDecoration(
                                color: e.value,
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 6),
                            Text(
                              '${e.key} (${pct.toStringAsFixed(1)}%)',
                              style: const TextStyle(fontSize: 12),
                            ),
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

class _ConcentrationList extends StatelessWidget {
  final List<HoldingModel> holdings;
  final NumberFormat currency;

  const _ConcentrationList({
    required this.holdings,
    required this.currency,
  });

  @override
  Widget build(BuildContext context) {
    final total =
        holdings.fold(0.0, (sum, h) => sum + h.currentValue);
    if (total == 0) return const SizedBox.shrink();

    // Sort by current value descending
    final sorted = [...holdings]
      ..sort((a, b) => b.currentValue.compareTo(a.currentValue));

    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Concentração por Ativo',
                style:
                    TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 4),
            const Text(
              'Ativos acima de 30% da carteira estão em destaque.',
              style: TextStyle(fontSize: 12, color: Colors.grey),
            ),
            const SizedBox(height: 12),
            ...sorted.map((h) {
              final pct = (h.currentValue / total) * 100;
              final isConcentrated = pct > 30;
              return Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Expanded(
                          child: Row(
                            children: [
                              if (isConcentrated)
                                const Padding(
                                  padding: EdgeInsets.only(right: 4),
                                  child: Icon(Icons.warning_amber_rounded,
                                      color: Colors.red, size: 16),
                                ),
                              Flexible(
                                child: Text(
                                  h.displayTicker,
                                  style: TextStyle(
                                    fontWeight: FontWeight.w600,
                                    fontSize: 14,
                                    color: isConcentrated
                                        ? Colors.red
                                        : null,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          '${pct.toStringAsFixed(1)}% · ${currency.format(h.currentValue)}',
                          style: TextStyle(
                            fontSize: 12,
                            color: isConcentrated
                                ? Colors.red
                                : Colors.grey,
                            fontWeight: isConcentrated
                                ? FontWeight.w600
                                : FontWeight.normal,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: LinearProgressIndicator(
                        value: pct / 100,
                        minHeight: 6,
                        backgroundColor: Colors.grey.shade200,
                        valueColor: AlwaysStoppedAnimation<Color>(
                          isConcentrated
                              ? Colors.red
                              : const Color(0xFF6366F1),
                        ),
                      ),
                    ),
                  ],
                ),
              );
            }),
          ],
        ),
      ),
    );
  }
}
