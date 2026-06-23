import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import '../providers/investments_provider.dart';

class TaxReportScreen extends ConsumerStatefulWidget {
  const TaxReportScreen({super.key});

  @override
  ConsumerState<TaxReportScreen> createState() => _TaxReportScreenState();
}

class _TaxReportScreenState extends ConsumerState<TaxReportScreen> {
  int _selectedYear = DateTime.now().year;

  @override
  void initState() {
    super.initState();
    Future.microtask(() =>
        ref.read(investmentsProvider.notifier).loadTaxReport(_selectedYear));
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(investmentsProvider);
    final currency = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final pct = NumberFormat.decimalPercentPattern(locale: 'pt_BR', decimalDigits: 1);
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Relatório de IR'),
        actions: [
          DropdownButtonHideUnderline(
            child: DropdownButton<int>(
              value: _selectedYear,
              items: List.generate(5, (i) => DateTime.now().year - i)
                  .map((y) => DropdownMenuItem(value: y, child: Text('$y')))
                  .toList(),
              onChanged: (y) {
                if (y == null) return;
                setState(() => _selectedYear = y);
                ref.read(investmentsProvider.notifier).loadTaxReport(y);
              },
            ),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : state.taxReport == null
              ? const Center(child: Text('Nenhum dado disponível'))
              : SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Summary card
                      Card(
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text('Resumo $_selectedYear',
                                  style: theme.textTheme.titleMedium
                                      ?.copyWith(fontWeight: FontWeight.bold)),
                              const SizedBox(height: 12),
                              _SummaryRow(
                                label: 'Lucro bruto total',
                                value: currency
                                    .format(state.taxReport!['total_profit'] ?? 0),
                                color: Colors.green,
                              ),
                              _SummaryRow(
                                label: 'IR devido (estimado)',
                                value: currency
                                    .format(state.taxReport!['total_tax_due'] ?? 0),
                                color: Colors.red,
                              ),
                              Container(
                                margin: const EdgeInsets.only(top: 12),
                                padding: const EdgeInsets.all(8),
                                decoration: BoxDecoration(
                                  color: theme.colorScheme.primary
                                      .withValues(alpha: 0.08),
                                  borderRadius: BorderRadius.circular(8),
                                  border: Border.all(
                                    color: theme.colorScheme.primary
                                        .withValues(alpha: 0.25),
                                  ),
                                ),
                                child: Text(
                                  '⚠️ Valores estimados com base no preço médio atual. '
                                  'Consulte um contador para declaração oficial.',
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: theme.colorScheme.onSurface
                                        .withValues(alpha: 0.85),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                      const SizedBox(height: 16),
                      Text('Detalhamento por mês e classe',
                          style: theme.textTheme.titleSmall
                              ?.copyWith(fontWeight: FontWeight.bold)),
                      const SizedBox(height: 8),
                      // Entries list
                      ...((state.taxReport!['entries'] as List<dynamic>?) ?? [])
                          .map((e) {
                        final entry = e as Map<String, dynamic>;
                        final isExempt = entry['is_exempt'] as bool? ?? false;
                        final profit = (entry['gross_profit'] as num?)?.toDouble() ?? 0;
                        final taxDue = (entry['tax_due'] as num?)?.toDouble() ?? 0;
                        final taxRate = (entry['tax_rate'] as num?)?.toDouble() ?? 0;

                        return Card(
                          margin: const EdgeInsets.only(bottom: 8),
                          child: ListTile(
                            leading: CircleAvatar(
                              backgroundColor: isExempt
                                  ? Colors.green.shade100
                                  : Colors.orange.shade100,
                              child: Icon(
                                isExempt ? Icons.check : Icons.receipt_long,
                                color: isExempt ? Colors.green : Colors.orange,
                                size: 20,
                              ),
                            ),
                            title: Text(
                              '${entry['label']} · ${_assetTypeLabel(entry['asset_type'] as String? ?? '')}',
                              style: const TextStyle(fontWeight: FontWeight.w600),
                            ),
                            subtitle: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text('Lucro: ${currency.format(profit)}'),
                                if (isExempt)
                                  Text(
                                    entry['exemption_reason'] as String? ?? 'Isento',
                                    style: const TextStyle(
                                        color: Colors.green, fontSize: 12),
                                  )
                                else
                                  Text(
                                    'Alíquota: ${pct.format(taxRate / 100)} → IR: ${currency.format(taxDue)}',
                                    style: const TextStyle(
                                        color: Colors.red, fontSize: 12),
                                  ),
                              ],
                            ),
                            trailing: isExempt
                                ? const Chip(
                                    label: Text('Isento',
                                        style: TextStyle(fontSize: 11)),
                                    backgroundColor: Color(0xFFE8F5E9),
                                  )
                                : Text(
                                    currency.format(taxDue),
                                    style: TextStyle(
                                        color: Colors.red.shade700,
                                        fontWeight: FontWeight.bold),
                                  ),
                          ),
                        );
                      }),
                    ],
                  ),
                ),
    );
  }

  String _assetTypeLabel(String type) {
    const labels = {
      'stock': 'Ações',
      'fii': 'FIIs',
      'etf': 'ETFs',
      'crypto': 'Cripto',
      'fixed_income': 'Renda Fixa',
      'fund': 'Fundos',
    };
    return labels[type] ?? type;
  }
}

class _SummaryRow extends StatelessWidget {
  const _SummaryRow({
    required this.label,
    required this.value,
    required this.color,
  });

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label),
          Text(value,
              style: TextStyle(fontWeight: FontWeight.bold, color: color)),
        ],
      ),
    );
  }
}
