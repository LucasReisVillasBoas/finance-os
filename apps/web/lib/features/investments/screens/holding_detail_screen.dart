import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/investments_provider.dart';
import '../models/holding_model.dart';
import '../models/investment_transaction_model.dart';

class HoldingDetailScreen extends ConsumerStatefulWidget {
  final String holdingId;

  const HoldingDetailScreen({super.key, required this.holdingId});

  @override
  ConsumerState<HoldingDetailScreen> createState() =>
      _HoldingDetailScreenState();
}

class _HoldingDetailScreenState extends ConsumerState<HoldingDetailScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() =>
        ref.read(investmentsProvider.notifier).loadTransactions(widget.holdingId));
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(investmentsProvider);
    final currency = NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$');
    final dateFormat = DateFormat('dd/MM/yyyy');

    // Find the holding from state
    final holding = state.holdings.cast<HoldingModel?>().firstWhere(
          (h) => h?.id == widget.holdingId,
          orElse: () => null,
        );

    return Scaffold(
      appBar: AppBar(
        title: Text(holding?.displayTicker ?? 'Posição'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            tooltip: 'Nova operação',
            onPressed: () => context.push('/investments/new'),
          ),
        ],
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : holding == null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Text('Posição não encontrada.'),
                      TextButton(
                        onPressed: () => context.pop(),
                        child: const Text('Voltar'),
                      ),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: () => ref
                      .read(investmentsProvider.notifier)
                      .loadTransactions(widget.holdingId),
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _HoldingHeader(holding: holding, currency: currency),
                        const SizedBox(height: 16),
                        _HoldingDetails(holding: holding, currency: currency),
                        const SizedBox(height: 16),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text(
                              'Operações',
                              style: TextStyle(
                                  fontSize: 18, fontWeight: FontWeight.bold),
                            ),
                            if (state.transactions.isNotEmpty)
                              Text(
                                '${state.transactions.length} registros',
                                style: const TextStyle(
                                    color: Colors.grey, fontSize: 12),
                              ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        if (state.transactions.isEmpty)
                          const Center(
                            child: Padding(
                              padding: EdgeInsets.all(32),
                              child: Text(
                                'Nenhuma operação registrada.\nAdicione uma compra, venda ou dividendo.',
                                textAlign: TextAlign.center,
                                style: TextStyle(color: Colors.grey),
                              ),
                            ),
                          )
                        else
                          ...state.transactions.map(
                            (tx) => _TransactionTile(
                              transaction: tx,
                              currency: currency,
                              dateFormat: dateFormat,
                              onDelete: () => _confirmDelete(context, tx),
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
    );
  }

  Future<void> _confirmDelete(
      BuildContext context, InvestmentTransactionModel tx) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Excluir operação?'),
        content: Text(
            'Deseja excluir a operação de ${tx.typeLabel}? Esta ação é irreversível.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: const Text('Cancelar')),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child:
                const Text('Excluir', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
    if (confirmed == true && mounted) {
      await ref
          .read(investmentsProvider.notifier)
          .deleteTransaction(tx.id, widget.holdingId);
    }
  }
}

class _HoldingHeader extends StatelessWidget {
  final HoldingModel holding;
  final NumberFormat currency;

  const _HoldingHeader({required this.holding, required this.currency});

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        holding.displayTicker,
                        style: const TextStyle(
                            fontSize: 22, fontWeight: FontWeight.bold),
                      ),
                      Text(
                        holding.name,
                        style: const TextStyle(
                            color: Colors.grey, fontSize: 14),
                      ),
                    ],
                  ),
                ),
                if (holding.assetCurrentPrice != null)
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        currency.format(holding.assetCurrentPrice),
                        style: const TextStyle(
                            fontSize: 18, fontWeight: FontWeight.bold),
                      ),
                      const Text('Preço atual',
                          style: TextStyle(
                              color: Colors.grey, fontSize: 12)),
                    ],
                  ),
              ],
            ),
            const Divider(height: 24),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                _PnlChip(
                  label: 'P&L Não Realizado',
                  value: holding.unrealizedPnl,
                  pct: holding.unrealizedPnlPct,
                  currency: currency,
                ),
                _PnlChip(
                  label: 'P&L Realizado',
                  value: holding.realizedPnl,
                  currency: currency,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _PnlChip extends StatelessWidget {
  final String label;
  final double value;
  final double? pct;
  final NumberFormat currency;

  const _PnlChip({
    required this.label,
    required this.value,
    this.pct,
    required this.currency,
  });

  @override
  Widget build(BuildContext context) {
    final color = value >= 0 ? const Color(0xFF22C55E) : const Color(0xFFEF4444);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label,
            style: const TextStyle(fontSize: 12, color: Colors.grey)),
        Text(
          '${value >= 0 ? '+' : ''}${currency.format(value)}',
          style: TextStyle(
              fontSize: 15, fontWeight: FontWeight.bold, color: color),
        ),
        if (pct != null)
          Text(
            '${pct! >= 0 ? '+' : ''}${pct!.toStringAsFixed(2)}%',
            style: TextStyle(fontSize: 12, color: color),
          ),
      ],
    );
  }
}

class _HoldingDetails extends StatelessWidget {
  final HoldingModel holding;
  final NumberFormat currency;

  const _HoldingDetails({required this.holding, required this.currency});

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Detalhes da Posição',
                style:
                    TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 12),
            _DetailRow(
                label: 'Quantidade',
                value: holding.quantity.toStringAsFixed(4)),
            _DetailRow(
                label: 'Preço médio',
                value: currency.format(holding.avgPrice)),
            _DetailRow(
                label: 'Total investido',
                value: currency.format(holding.totalInvested)),
            _DetailRow(
                label: 'Valor atual',
                value: currency.format(holding.currentValue)),
            _DetailRow(label: 'Tipo', value: holding.type),
          ],
        ),
      ),
    );
  }
}

class _DetailRow extends StatelessWidget {
  final String label;
  final String value;

  const _DetailRow({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: const TextStyle(color: Colors.grey, fontSize: 14)),
          Text(value,
              style: const TextStyle(
                  fontWeight: FontWeight.w600, fontSize: 14)),
        ],
      ),
    );
  }
}

class _TransactionTile extends StatelessWidget {
  final InvestmentTransactionModel transaction;
  final NumberFormat currency;
  final DateFormat dateFormat;
  final VoidCallback onDelete;

  const _TransactionTile({
    required this.transaction,
    required this.currency,
    required this.dateFormat,
    required this.onDelete,
  });

  Color get _typeColor {
    switch (transaction.type) {
      case 'buy':
        return const Color(0xFF3B82F6);
      case 'sell':
        return const Color(0xFFEF4444);
      case 'dividend':
        return const Color(0xFF22C55E);
      default:
        return Colors.grey;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: _typeColor.withValues(alpha: 0.15),
          child: Text(
            transaction.typeLabel[0],
            style: TextStyle(
                color: _typeColor, fontWeight: FontWeight.bold),
          ),
        ),
        title: Row(
          children: [
            Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: _typeColor.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text(
                transaction.typeLabel,
                style: TextStyle(
                    color: _typeColor,
                    fontSize: 12,
                    fontWeight: FontWeight.bold),
              ),
            ),
            const SizedBox(width: 8),
            Text(
              dateFormat.format(transaction.date),
              style: const TextStyle(fontSize: 14),
            ),
          ],
        ),
        subtitle: transaction.quantity != null && transaction.price != null
            ? Text(
                '${transaction.quantity!.toStringAsFixed(4)} un. @ ${currency.format(transaction.price)}',
                style:
                    const TextStyle(fontSize: 12, color: Colors.grey),
              )
            : null,
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  currency.format(transaction.total),
                  style: const TextStyle(
                      fontWeight: FontWeight.bold, fontSize: 14),
                ),
                if (transaction.fees > 0)
                  Text(
                    'Taxa: ${currency.format(transaction.fees)}',
                    style: const TextStyle(
                        fontSize: 11, color: Colors.grey),
                  ),
              ],
            ),
            IconButton(
              icon: const Icon(Icons.delete_outline,
                  color: Colors.red, size: 20),
              onPressed: onDelete,
            ),
          ],
        ),
      ),
    );
  }
}
