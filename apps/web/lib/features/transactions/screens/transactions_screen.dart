import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/transactions_provider.dart';
import '../models/transaction_model.dart';

class TransactionsScreen extends ConsumerStatefulWidget {
  const TransactionsScreen({super.key});

  @override
  ConsumerState<TransactionsScreen> createState() => _TransactionsScreenState();
}

class _TransactionsScreenState extends ConsumerState<TransactionsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
      () => ref.read(transactionsProvider.notifier).loadTransactions(reset: true),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(transactionsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Transações'),
        centerTitle: true,
        actions: [
          IconButton(
            icon: const Icon(Icons.filter_list),
            tooltip: 'Filtrar',
            onPressed: () => context.push('/transactions/filters'),
          ),
        ],
      ),
      body: state.isLoading && state.transactions.isEmpty
          ? const Center(child: CircularProgressIndicator())
          : state.error != null
              ? _ErrorView(
                  message: state.error!,
                  onRetry: () => ref
                      .read(transactionsProvider.notifier)
                      .loadTransactions(reset: true),
                )
              : state.transactions.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.receipt_long,
                              size: 64, color: Colors.grey),
                          const SizedBox(height: 16),
                          const Text(
                            'Nenhuma transação encontrada.\nAdicione sua primeira transação!',
                            textAlign: TextAlign.center,
                            style: TextStyle(color: Colors.grey),
                          ),
                          const SizedBox(height: 16),
                          ElevatedButton.icon(
                            onPressed: () => context.push('/transactions/new'),
                            icon: const Icon(Icons.add),
                            label: const Text('Nova Transação'),
                          ),
                        ],
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: () => ref
                          .read(transactionsProvider.notifier)
                          .loadTransactions(reset: true),
                      child: _TransactionGroupedList(
                        transactions: state.transactions,
                      ),
                    ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/transactions/new'),
        tooltip: 'Nova Transação',
        child: const Icon(Icons.add),
      ),
    );
  }
}

class _TransactionGroupedList extends StatelessWidget {
  final List<TransactionModel> transactions;

  const _TransactionGroupedList({required this.transactions});

  Map<String, List<TransactionModel>> _groupByDate() {
    final groups = <String, List<TransactionModel>>{};
    for (final tx in transactions) {
      final key = DateFormat('yyyy-MM-dd').format(tx.date);
      groups.putIfAbsent(key, () => []).add(tx);
    }
    return groups;
  }

  @override
  Widget build(BuildContext context) {
    final groups = _groupByDate();
    final sortedKeys = groups.keys.toList()
      ..sort((a, b) => b.compareTo(a));

    final items = <Widget>[];
    for (final key in sortedKeys) {
      final date = DateTime.parse(key);
      items.add(_DateHeader(date: date));
      for (final tx in groups[key]!) {
        items.add(_TransactionTile(transaction: tx));
      }
    }

    return ListView.builder(
      padding: const EdgeInsets.only(bottom: 80),
      itemCount: items.length,
      itemBuilder: (_, index) => items[index],
    );
  }
}

class _DateHeader extends StatelessWidget {
  final DateTime date;

  const _DateHeader({required this.date});

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final yesterday = today.subtract(const Duration(days: 1));
    final d = DateTime(date.year, date.month, date.day);

    if (d == today) return 'Hoje';
    if (d == yesterday) return 'Ontem';
    return DateFormat("d 'de' MMMM", 'pt_BR').format(date);
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
      child: Text(
        _formatDate(date),
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              color: Colors.grey[600],
              fontWeight: FontWeight.w600,
            ),
      ),
    );
  }
}

class _TransactionTile extends StatelessWidget {
  final TransactionModel transaction;

  const _TransactionTile({required this.transaction});

  Color _typeColor() {
    if (transaction.isIncome) return Colors.green;
    if (transaction.isExpense) return Colors.red;
    return Colors.blue;
  }

  IconData _typeIcon() {
    if (transaction.isIncome) return Icons.arrow_downward;
    if (transaction.isExpense) return Icons.arrow_upward;
    return Icons.swap_horiz;
  }

  String _amountPrefix() {
    if (transaction.isIncome) return '+';
    if (transaction.isExpense) return '-';
    return '';
  }

  @override
  Widget build(BuildContext context) {
    final color = _typeColor();
    final numberFormat =
        NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: color.withValues(alpha: 0.15),
          child: Icon(_typeIcon(), color: color, size: 20),
        ),
        title: Text(
          transaction.description ?? transaction.typeLabel,
          style: const TextStyle(fontWeight: FontWeight.w500),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Text(
          transaction.categoryName ??
              transaction.accountName ??
              transaction.typeLabel,
          style: TextStyle(color: Colors.grey[600], fontSize: 12),
        ),
        trailing: Text(
          '${_amountPrefix()}${numberFormat.format(transaction.amount)}',
          style: TextStyle(
            color: color,
            fontWeight: FontWeight.bold,
            fontSize: 14,
          ),
        ),
        onTap: () => context.push('/transactions/${transaction.id}'),
      ),
    );
  }
}

class _ErrorView extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _ErrorView({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.error_outline, size: 48, color: Colors.red),
          const SizedBox(height: 16),
          Text(message, textAlign: TextAlign.center),
          const SizedBox(height: 16),
          ElevatedButton(
              onPressed: onRetry, child: const Text('Tentar novamente')),
        ],
      ),
    );
  }
}
