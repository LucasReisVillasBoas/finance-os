import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/recurrences_provider.dart';
import '../models/recurrence_model.dart';

class RecurrencesScreen extends ConsumerStatefulWidget {
  const RecurrencesScreen({super.key});

  @override
  ConsumerState<RecurrencesScreen> createState() => _RecurrencesScreenState();
}

class _RecurrencesScreenState extends ConsumerState<RecurrencesScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(recurrencesProvider.notifier).load());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(recurrencesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Recorrências'),
        centerTitle: true,
      ),
      body: state.isLoading
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
                            ref.read(recurrencesProvider.notifier).load(),
                        child: const Text('Tentar novamente'),
                      ),
                    ],
                  ),
                )
              : state.items.isEmpty
                  ? const Center(
                      child: Text(
                        'Nenhuma recorrência cadastrada.\nAdicione sua primeira recorrência!',
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey),
                      ),
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.all(16),
                      itemCount: state.items.length,
                      itemBuilder: (context, index) {
                        return _RecurrenceCard(
                          recurrence: state.items[index],
                          onEdit: () => context.push(
                            '/recurrences/${state.items[index].id}/edit',
                          ),
                          onDelete: () =>
                              _confirmDelete(context, state.items[index]),
                          onToggleAutoLaunch: (value) =>
                              ref.read(recurrencesProvider.notifier).update(
                                    state.items[index].id,
                                    {'auto_launch': value},
                                  ),
                        );
                      },
                    ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/recurrences/new'),
        tooltip: 'Nova recorrência',
        child: const Icon(Icons.add),
      ),
    );
  }

  Future<void> _confirmDelete(BuildContext context, RecurrenceModel rec) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Excluir recorrência'),
        content: Text(
          'Deseja excluir "${rec.description ?? rec.frequencyLabel}"?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Cancelar'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child:
                const Text('Excluir', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
    if (confirmed == true && mounted) {
      await ref.read(recurrencesProvider.notifier).delete(rec.id);
    }
  }
}

class _RecurrenceCard extends StatelessWidget {
  final RecurrenceModel recurrence;
  final VoidCallback onEdit;
  final VoidCallback onDelete;
  final ValueChanged<bool> onToggleAutoLaunch;

  const _RecurrenceCard({
    required this.recurrence,
    required this.onEdit,
    required this.onDelete,
    required this.onToggleAutoLaunch,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final isIncome = recurrence.type == 'income';
    final amountColor = isIncome ? Colors.green : Colors.red;
    final formatter = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final dateFormatter = DateFormat('dd/MM/yyyy', 'pt_BR');

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
                    recurrence.description ??
                        recurrence.accountName ??
                        'Sem descrição',
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                ),
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
            Row(
              children: [
                Text(
                  formatter.format(recurrence.amount),
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: amountColor,
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: colorScheme.primaryContainer,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    recurrence.frequencyLabel,
                    style: TextStyle(
                      fontSize: 12,
                      color: colorScheme.onPrimaryContainer,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: isIncome
                        ? Colors.green.withValues(alpha: 0.1)
                        : Colors.red.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    recurrence.typeLabel,
                    style: TextStyle(
                      fontSize: 12,
                      color: amountColor,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (recurrence.categoryName != null)
              Text(
                'Categoria: ${recurrence.categoryName}',
                style: TextStyle(color: Colors.grey[600], fontSize: 13),
              ),
            if (recurrence.accountName != null)
              Text(
                'Conta: ${recurrence.accountName}',
                style: TextStyle(color: Colors.grey[600], fontSize: 13),
              ),
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.calendar_today,
                    size: 14, color: Colors.grey[600]),
                const SizedBox(width: 4),
                Text(
                  'Próximo: ${dateFormatter.format(recurrence.nextDueDate)}',
                  style: TextStyle(color: Colors.grey[600], fontSize: 13),
                ),
                const Spacer(),
                Text(
                  'Auto-lançar',
                  style:
                      TextStyle(fontSize: 13, color: Colors.grey[700]),
                ),
                Switch(
                  value: recurrence.autoLaunch,
                  onChanged: onToggleAutoLaunch,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
