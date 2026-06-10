import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/goals_provider.dart';
import '../models/goal_model.dart';
import 'goal_contribute_dialog.dart';

class GoalsScreen extends ConsumerStatefulWidget {
  const GoalsScreen({super.key});

  @override
  ConsumerState<GoalsScreen> createState() => _GoalsScreenState();
}

class _GoalsScreenState extends ConsumerState<GoalsScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(goalsProvider.notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(goalsProvider);
    final currencyFormat = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Metas'),
        actions: [
          if (state.goals.isNotEmpty)
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () => ref.read(goalsProvider.notifier).load(),
            ),
        ],
      ),
      body: Builder(
        builder: (context) {
          if (state.isLoading && state.goals.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state.error != null && state.goals.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, size: 48, color: Colors.red),
                  const SizedBox(height: 16),
                  Text(state.error!),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => ref.read(goalsProvider.notifier).load(),
                    child: const Text('Tentar novamente'),
                  ),
                ],
              ),
            );
          }

          if (state.goals.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.flag_outlined, size: 64, color: Colors.grey),
                  const SizedBox(height: 16),
                  const Text(
                    'Nenhuma meta criada',
                    style: TextStyle(fontSize: 18, color: Colors.grey),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Defina objetivos financeiros e acompanhe seu progresso.',
                    textAlign: TextAlign.center,
                    style: TextStyle(color: Colors.grey),
                  ),
                  const SizedBox(height: 24),
                  ElevatedButton.icon(
                    onPressed: () => context.push('/goals/new'),
                    icon: const Icon(Icons.add),
                    label: const Text('Criar primeira meta'),
                  ),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () => ref.read(goalsProvider.notifier).load(),
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: state.goals.length,
              itemBuilder: (context, index) {
                final goal = state.goals[index];
                final projection = state.projections
                    .where((p) => p.goalId == goal.id)
                    .firstOrNull;
                return _GoalCard(
                  goal: goal,
                  projection: projection,
                  currencyFormat: currencyFormat,
                  onContribute: () => _showContributeDialog(context, goal),
                  onEdit: () => context.push('/goals/${goal.id}/edit'),
                  onDelete: () => _confirmDelete(context, goal),
                );
              },
            ),
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/goals/new'),
        tooltip: 'Nova meta',
        child: const Icon(Icons.add),
      ),
    );
  }

  void _showContributeDialog(BuildContext context, GoalModel goal) {
    showDialog(
      context: context,
      builder: (_) => GoalContributeDialog(
        goalId: goal.id,
        goalName: goal.name,
        onConfirm: (amount, date, notes) async {
          final success = await ref
              .read(goalsProvider.notifier)
              .contribute(goal.id, amount, date, notes: notes);
          if (success && context.mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Aporte registrado com sucesso!')),
            );
          }
        },
      ),
    );
  }

  void _confirmDelete(BuildContext context, GoalModel goal) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Excluir Meta'),
        content: Text('Deseja excluir a meta "${goal.name}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Cancelar'),
          ),
          TextButton(
            onPressed: () async {
              Navigator.of(ctx).pop();
              await ref.read(goalsProvider.notifier).delete(goal.id);
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Excluir'),
          ),
        ],
      ),
    );
  }
}

class _GoalCard extends StatelessWidget {
  final GoalModel goal;
  final GoalProjectionModel? projection;
  final NumberFormat currencyFormat;
  final VoidCallback onContribute;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const _GoalCard({
    required this.goal,
    this.projection,
    required this.currencyFormat,
    required this.onContribute,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final progressPct = goal.progressPct;
    final pctLabel = '${(progressPct * 100).toStringAsFixed(1)}%';

    Color progressColor = Colors.blue;
    if (goal.isAchieved) {
      progressColor = Colors.green;
    } else if (progressPct > 0.75) {
      progressColor = Colors.orange;
    }

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                if (goal.icon != null) ...[
                  Text(goal.icon!, style: const TextStyle(fontSize: 24)),
                  const SizedBox(width: 8),
                ] else ...[
                  const Icon(Icons.flag, color: Colors.blue),
                  const SizedBox(width: 8),
                ],
                Expanded(
                  child: Text(
                    goal.name,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                if (goal.isAchieved)
                  const Chip(
                    label: Text('Concluída'),
                    backgroundColor: Colors.green,
                    labelStyle: TextStyle(color: Colors.white),
                    padding: EdgeInsets.zero,
                  ),
                PopupMenuButton<String>(
                  onSelected: (value) {
                    if (value == 'edit') onEdit();
                    if (value == 'delete') onDelete();
                  },
                  itemBuilder: (_) => [
                    const PopupMenuItem(value: 'edit', child: Text('Editar')),
                    const PopupMenuItem(
                      value: 'delete',
                      child: Text('Excluir', style: TextStyle(color: Colors.red)),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  currencyFormat.format(goal.currentAmount),
                  style: theme.textTheme.bodyLarge?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
                ),
                Text(
                  '${currencyFormat.format(goal.targetAmount)} ($pctLabel)',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: progressPct,
                minHeight: 8,
                backgroundColor: Colors.grey[200],
                valueColor: AlwaysStoppedAnimation<Color>(progressColor),
              ),
            ),
            Builder(
              builder: (context) {
                final proj = projection;
                final estimatedDate = proj?.estimatedDate;
                if (proj != null && estimatedDate != null) {
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          const Icon(Icons.calendar_today, size: 14, color: Colors.grey),
                          const SizedBox(width: 4),
                          Text(
                            'Estimativa: ${DateFormat('MMM yyyy', 'pt_BR').format(estimatedDate)} '
                            '(${proj.monthsToGoal ?? 0} meses)',
                            style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey[600]),
                          ),
                        ],
                      ),
                    ],
                  );
                }
                return const SizedBox.shrink();
              },
            ),
            if (projection?.estimatedDate == null && goal.targetDate != null) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  const Icon(Icons.calendar_today, size: 14, color: Colors.grey),
                  const SizedBox(width: 4),
                  Text(
                    'Meta: ${DateFormat('dd/MM/yyyy', 'pt_BR').format(goal.targetDate!)}',
                    style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey[600]),
                  ),
                ],
              ),
            ],
            if (!goal.isAchieved) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: onContribute,
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Aportar'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
