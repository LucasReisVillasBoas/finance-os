import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/budgets_provider.dart';
import '../models/budget_model.dart';

class BudgetsScreen extends ConsumerStatefulWidget {
  const BudgetsScreen({super.key});

  @override
  ConsumerState<BudgetsScreen> createState() => _BudgetsScreenState();
}

class _BudgetsScreenState extends ConsumerState<BudgetsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(budgetsProvider.notifier).load());
  }

  static const _monthNames = [
    'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
    'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro',
  ];

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(budgetsProvider);
    final monthName = _monthNames[state.month - 1];

    return Scaffold(
      appBar: AppBar(
        title: const Text('Orçamentos'),
        centerTitle: true,
      ),
      body: Column(
        children: [
          // Month/Year navigation
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                IconButton(
                  icon: const Icon(Icons.chevron_left),
                  onPressed: () =>
                      ref.read(budgetsProvider.notifier).changeMonth(-1),
                ),
                Text(
                  '$monthName ${state.year}',
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.chevron_right),
                  onPressed: () =>
                      ref.read(budgetsProvider.notifier).changeMonth(1),
                ),
              ],
            ),
          ),
          // Content
          Expanded(
            child: state.isLoading
                ? const Center(child: CircularProgressIndicator())
                : state.error != null
                    ? Center(
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              state.error!,
                              textAlign: TextAlign.center,
                              style: const TextStyle(color: Colors.red),
                            ),
                            const SizedBox(height: 12),
                            ElevatedButton(
                              onPressed: () =>
                                  ref.read(budgetsProvider.notifier).load(),
                              child: const Text('Tentar novamente'),
                            ),
                          ],
                        ),
                      )
                    : state.progress.isEmpty
                        ? const Center(
                            child: Text(
                              'Nenhum orçamento cadastrado.\nAdicione seu primeiro orçamento!',
                              textAlign: TextAlign.center,
                              style: TextStyle(color: Colors.grey),
                            ),
                          )
                        : ListView.builder(
                            padding: const EdgeInsets.all(16),
                            itemCount: state.progress.length,
                            itemBuilder: (context, index) {
                              return _BudgetProgressCard(
                                progress: state.progress[index],
                                onEdit: () => context.push(
                                  '/budgets/${state.progress[index].budgetId}/edit',
                                ),
                                onDelete: () => _confirmDelete(
                                  context,
                                  state.progress[index],
                                ),
                              );
                            },
                          ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/budgets/new'),
        tooltip: 'Novo orçamento',
        child: const Icon(Icons.add),
      ),
    );
  }

  Future<void> _confirmDelete(
    BuildContext context,
    BudgetProgressModel progress,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Excluir orçamento'),
        content: Text('Deseja excluir o orçamento de "${progress.categoryName}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Cancelar'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('Excluir',
                style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
    if (confirmed == true && mounted) {
      await ref.read(budgetsProvider.notifier).delete(progress.budgetId);
    }
  }
}

class _BudgetProgressCard extends StatelessWidget {
  final BudgetProgressModel progress;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const _BudgetProgressCard({
    required this.progress,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final formatter = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final pct = progress.percentage.clamp(0.0, 999.0);

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    progress.categoryName,
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                ),
                if (progress.isAlert)
                  const Icon(Icons.warning_amber, color: Colors.orange),
                PopupMenuButton<String>(
                  onSelected: (value) {
                    if (value == 'edit') onEdit();
                    if (value == 'delete') onDelete();
                  },
                  itemBuilder: (ctx) => [
                    const PopupMenuItem(value: 'edit', child: Text('Editar')),
                    const PopupMenuItem(
                      value: 'delete',
                      child: Text('Excluir',
                          style: TextStyle(color: Colors.red)),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 8),
            // Progress bar
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: progress.progressValue,
                backgroundColor: Colors.grey[200],
                valueColor:
                    AlwaysStoppedAnimation<Color>(progress.progressColor),
                minHeight: 8,
              ),
            ),
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '${formatter.format(progress.actual)} / ${formatter.format(progress.planned)}',
                  style: const TextStyle(fontSize: 14),
                ),
                Text(
                  '${pct.toStringAsFixed(1)}%',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: progress.progressColor,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
