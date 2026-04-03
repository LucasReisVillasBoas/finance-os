import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/accounts_provider.dart';
import '../models/account_model.dart';

class AccountDetailScreen extends ConsumerWidget {
  final String accountId;

  const AccountDetailScreen({super.key, required this.accountId});

  IconData _iconForType(String type) {
    switch (type) {
      case 'checking':
        return Icons.account_balance;
      case 'savings':
        return Icons.savings;
      case 'credit_card':
        return Icons.credit_card;
      case 'investment':
        return Icons.trending_up;
      case 'wallet':
        return Icons.account_balance_wallet;
      default:
        return Icons.attach_money;
    }
  }

  Color _parseColor(String? hexColor, BuildContext context) {
    if (hexColor == null) return Theme.of(context).colorScheme.primary;
    try {
      return Color(int.parse(hexColor.replaceFirst('#', '0xFF')));
    } catch (_) {
      return Theme.of(context).colorScheme.primary;
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(accountsProvider);

    AccountModel? account;
    try {
      account = state.accounts.firstWhere((a) => a.id == accountId);
    } catch (_) {
      account = null;
    }

    if (state.isLoading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    if (account == null) {
      return Scaffold(
        appBar: AppBar(title: const Text('Conta')),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Text('Conta não encontrada'),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => context.pop(),
                child: const Text('Voltar'),
              ),
            ],
          ),
        ),
      );
    }

    final color = _parseColor(account.color, context);
    final icon = _iconForType(account.type);

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 200,
            pinned: true,
            backgroundColor: color,
            flexibleSpace: FlexibleSpaceBar(
              title: Text(
                account.name,
                style: const TextStyle(color: Colors.white),
              ),
              background: Container(
                color: color,
                child: Center(
                  child: Icon(icon, size: 80, color: Colors.white54),
                ),
              ),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.edit, color: Colors.white),
                onPressed: () =>
                    context.push('/accounts/${account!.id}/edit'),
              ),
            ],
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  _BalanceCard(account: account, color: color),
                  const SizedBox(height: 16),
                  _InfoSection(account: account),
                  const SizedBox(height: 16),
                  _DeactivateButton(account: account),
                  const SizedBox(height: 24),
                  _TransactionsPlaceholder(),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _BalanceCard extends StatelessWidget {
  final AccountModel account;
  final Color color;

  const _BalanceCard({required this.account, required this.color});

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            Text(
              'Saldo atual',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey,
                  ),
            ),
            const SizedBox(height: 8),
            Text(
              'R\$ ${account.balance.toStringAsFixed(2)}',
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                    color: account.balance >= 0 ? Colors.green : Colors.red,
                  ),
            ),
            if (account.creditLimit != null) ...[
              const SizedBox(height: 8),
              Text(
                'Limite: R\$ ${account.creditLimit!.toStringAsFixed(2)}',
                style: Theme.of(context)
                    .textTheme
                    .bodySmall
                    ?.copyWith(color: Colors.grey),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _InfoSection extends StatelessWidget {
  final AccountModel account;

  const _InfoSection({required this.account});

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Informações',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const Divider(),
            _InfoRow(label: 'Tipo', value: account.typeLabel),
            if (account.institution != null)
              _InfoRow(label: 'Instituição', value: account.institution!),
            _InfoRow(
              label: 'Status',
              value: account.isActive ? 'Ativa' : 'Inativa',
            ),
          ],
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  final String label;
  final String value;

  const _InfoRow({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: Colors.grey)),
          Text(value, style: const TextStyle(fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}

class _DeactivateButton extends ConsumerWidget {
  final AccountModel account;

  const _DeactivateButton({required this.account});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return OutlinedButton.icon(
      style: OutlinedButton.styleFrom(
        foregroundColor: Colors.red,
        side: const BorderSide(color: Colors.red),
        padding: const EdgeInsets.symmetric(vertical: 14),
      ),
      onPressed: () async {
        final confirmed = await showDialog<bool>(
          context: context,
          builder: (ctx) => AlertDialog(
            title: const Text('Desativar conta'),
            content: const Text(
              'Deseja desativar esta conta? Ela não aparecerá mais na lista de contas ativas.',
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx, false),
                child: const Text('Cancelar'),
              ),
              TextButton(
                onPressed: () => Navigator.pop(ctx, true),
                child: const Text(
                  'Desativar',
                  style: TextStyle(color: Colors.red),
                ),
              ),
            ],
          ),
        );
        if (confirmed == true && context.mounted) {
          await ref
              .read(accountsProvider.notifier)
              .deleteAccount(account.id);
          if (context.mounted) context.pop();
        }
      },
      icon: const Icon(Icons.block),
      label: const Text('Desativar conta'),
    );
  }
}

class _TransactionsPlaceholder extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Últimas transações',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const Divider(),
            const SizedBox(height: 32),
            Center(
              child: Text(
                'Transações disponíveis em breve',
                style: TextStyle(color: Colors.grey[500]),
              ),
            ),
            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }
}
