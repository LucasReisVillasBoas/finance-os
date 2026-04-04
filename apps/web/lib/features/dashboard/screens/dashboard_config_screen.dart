import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

// Provider for dashboard widget visibility config
final dashboardConfigProvider =
    StateNotifierProvider<DashboardConfigNotifier, Map<String, bool>>((ref) {
  return DashboardConfigNotifier();
});

class DashboardConfigNotifier extends StateNotifier<Map<String, bool>> {
  DashboardConfigNotifier()
      : super({
          'balance': true,
          'chart': true,
          'categories': true,
          'budgets': true,
          'wallet': true,
        }) {
    _loadConfig();
  }

  Future<void> _loadConfig() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final updated = <String, bool>{};
      for (final key in state.keys) {
        updated[key] = prefs.getBool('dashboard_widget_$key') ?? state[key]!;
      }
      state = updated;
    } catch (_) {
      // Use defaults on error
    }
  }

  Future<void> toggle(String key) async {
    final current = state[key] ?? true;
    state = {...state, key: !current};
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool('dashboard_widget_$key', !current);
    } catch (_) {
      // Silently ignore persistence errors
    }
  }
}

class DashboardConfigScreen extends ConsumerWidget {
  const DashboardConfigScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final config = ref.watch(dashboardConfigProvider);

    const widgets = [
      _WidgetConfig(
        key: 'balance',
        label: 'Saldo Líquido',
        description: 'Card principal com saldo, receitas e despesas do mês',
        icon: Icons.account_balance_wallet,
      ),
      _WidgetConfig(
        key: 'chart',
        label: 'Gráfico de Fluxo',
        description: 'Barras de receitas vs despesas dos últimos 6 meses',
        icon: Icons.bar_chart,
      ),
      _WidgetConfig(
        key: 'categories',
        label: 'Top Categorias',
        description: 'Maiores gastos por categoria no mês',
        icon: Icons.pie_chart,
      ),
      _WidgetConfig(
        key: 'budgets',
        label: 'Orçamentos em Alerta',
        description: 'Orçamentos próximos ou acima do limite',
        icon: Icons.warning_amber,
      ),
      _WidgetConfig(
        key: 'wallet',
        label: 'Carteira / Contas',
        description: 'Lista horizontal das suas contas',
        icon: Icons.credit_card,
      ),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('Configurar Dashboard'),
        centerTitle: true,
      ),
      body: Column(
        children: [
          Container(
            margin: const EdgeInsets.all(16),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.primaryContainer,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.info_outline,
                  color: Theme.of(context).colorScheme.onPrimaryContainer,
                  size: 18,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Ative ou desative os widgets que aparecem na tela inicial.',
                    style: TextStyle(
                      fontSize: 13,
                      color: Theme.of(context).colorScheme.onPrimaryContainer,
                    ),
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              itemCount: widgets.length,
              separatorBuilder: (context, index) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final w = widgets[index];
                final isEnabled = config[w.key] ?? true;

                return CheckboxListTile(
                  value: isEnabled,
                  onChanged: (_) =>
                      ref.read(dashboardConfigProvider.notifier).toggle(w.key),
                  secondary: Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: isEnabled
                          ? Theme.of(context)
                              .colorScheme
                              .primary
                              .withValues(alpha: 0.15)
                          : Colors.grey.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Icon(
                      w.icon,
                      color: isEnabled
                          ? Theme.of(context).colorScheme.primary
                          : Colors.grey,
                      size: 20,
                    ),
                  ),
                  title: Text(
                    w.label,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: isEnabled ? null : Colors.grey,
                    ),
                  ),
                  subtitle: Text(
                    w.description,
                    style: TextStyle(
                      fontSize: 12,
                      color: isEnabled ? Colors.grey : Colors.grey.shade400,
                    ),
                  ),
                  controlAffinity: ListTileControlAffinity.trailing,
                  activeColor: Theme.of(context).colorScheme.primary,
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _WidgetConfig {
  const _WidgetConfig({
    required this.key,
    required this.label,
    required this.description,
    required this.icon,
  });

  final String key;
  final String label;
  final String description;
  final IconData icon;
}
