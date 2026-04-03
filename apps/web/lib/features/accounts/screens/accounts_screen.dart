import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/accounts_provider.dart';
import '../models/account_model.dart';

class AccountsScreen extends ConsumerStatefulWidget {
  const AccountsScreen({super.key});

  @override
  ConsumerState<AccountsScreen> createState() => _AccountsScreenState();
}

class _AccountsScreenState extends ConsumerState<AccountsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(accountsProvider.notifier).loadAccounts());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(accountsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Contas'),
        centerTitle: true,
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : state.error != null
              ? _ErrorView(
                  message: state.error!,
                  onRetry: () =>
                      ref.read(accountsProvider.notifier).loadAccounts(),
                )
              : Column(
                  children: [
                    _TotalBalanceCard(totalBalance: state.totalBalance),
                    Expanded(
                      child: state.accounts.isEmpty
                          ? const Center(
                              child: Text(
                                'Nenhuma conta cadastrada.\nAdicione sua primeira conta!',
                                textAlign: TextAlign.center,
                                style: TextStyle(color: Colors.grey),
                              ),
                            )
                          : ListView.builder(
                              padding: const EdgeInsets.all(16),
                              itemCount: state.accounts.length,
                              itemBuilder: (context, index) {
                                return _AccountCard(
                                  account: state.accounts[index],
                                );
                              },
                            ),
                    ),
                  ],
                ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/accounts/new'),
        child: const Icon(Icons.add),
      ),
    );
  }
}

class _TotalBalanceCard extends StatelessWidget {
  final double totalBalance;

  const _TotalBalanceCard({required this.totalBalance});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.all(16),
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: theme.colorScheme.primary,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Saldo Total',
            style: theme.textTheme.bodyMedium?.copyWith(
              color: theme.colorScheme.onPrimary.withValues(alpha: 0.8),
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'R\$ ${totalBalance.toStringAsFixed(2)}',
            style: theme.textTheme.headlineMedium?.copyWith(
              color: theme.colorScheme.onPrimary,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }
}

class _AccountCard extends StatelessWidget {
  final AccountModel account;

  const _AccountCard({required this.account});

  IconData get _iconData {
    switch (account.type) {
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

  Color _parseColor(BuildContext context) {
    if (account.color == null) {
      return Theme.of(context).colorScheme.primary;
    }
    try {
      return Color(
          int.parse(account.color!.replaceFirst('#', '0xFF')));
    } catch (_) {
      return Theme.of(context).colorScheme.primary;
    }
  }

  @override
  Widget build(BuildContext context) {
    final color = _parseColor(context);

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: color.withValues(alpha: 0.15),
          child: Icon(_iconData, color: color),
        ),
        title: Text(
          account.name,
          style: const TextStyle(fontWeight: FontWeight.w600),
        ),
        subtitle: Text(account.typeLabel),
        trailing: Text(
          'R\$ ${account.balance.toStringAsFixed(2)}',
          style: TextStyle(
            fontWeight: FontWeight.bold,
            color: account.balance >= 0 ? Colors.green : Colors.red,
          ),
        ),
        onTap: () => context.push('/accounts/${account.id}'),
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
          ElevatedButton(onPressed: onRetry, child: const Text('Tentar novamente')),
        ],
      ),
    );
  }
}
