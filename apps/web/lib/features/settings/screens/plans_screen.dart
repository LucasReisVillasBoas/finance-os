import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../auth/providers/auth_provider.dart';

class PlansScreen extends ConsumerWidget {
  const PlansScreen({super.key});

  static const _plans = [
    _PlanInfo(
      name: 'Free',
      price: 'Grátis',
      id: 'free',
      color: Colors.grey,
      features: [
        'Até 3 contas',
        'Transações ilimitadas',
        'Orçamentos básicos',
        'Dashboard',
      ],
    ),
    _PlanInfo(
      name: 'Pro',
      price: 'R\$ 29/mês',
      id: 'pro',
      color: Colors.blue,
      features: [
        'Tudo do Free',
        'Contas ilimitadas',
        'Importação OFX/CSV',
        'IA: Previsão de gastos',
        'IA: Análise de portfólio',
        'Grupo familiar',
        'Metas financeiras',
      ],
    ),
    _PlanInfo(
      name: 'Premium',
      price: 'R\$ 59/mês',
      id: 'premium',
      color: Colors.amber,
      features: [
        'Tudo do Pro',
        'Bot WhatsApp',
        'Múltiplos portfólios',
        'Relatórios avançados',
        'Suporte prioritário',
        'API access',
      ],
    ),
  ];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final currentPlan = authState.user?.plan ?? 'free';

    return Scaffold(
      appBar: AppBar(title: const Text('Planos')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const Text(
            'Escolha o plano ideal para você',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          ..._plans.map(
            (plan) => _PlanCard(
              plan: plan,
              isCurrent: plan.id == currentPlan,
            ),
          ),
        ],
      ),
    );
  }
}

class _PlanInfo {
  final String name;
  final String price;
  final String id;
  final Color color;
  final List<String> features;

  const _PlanInfo({
    required this.name,
    required this.price,
    required this.id,
    required this.color,
    required this.features,
  });
}

class _PlanCard extends StatelessWidget {
  final _PlanInfo plan;
  final bool isCurrent;

  const _PlanCard({required this.plan, required this.isCurrent});

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      elevation: isCurrent ? 4 : 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: isCurrent ? plan.color : Colors.transparent,
          width: 2,
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      plan.name,
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: plan.color,
                      ),
                    ),
                    Text(
                      plan.price,
                      style: const TextStyle(fontSize: 16),
                    ),
                  ],
                ),
                if (isCurrent)
                  Chip(
                    label: const Text('Plano atual'),
                    backgroundColor: plan.color.withAlpha(30),
                    labelStyle: TextStyle(color: plan.color),
                  ),
              ],
            ),
            const SizedBox(height: 16),
            ...plan.features.map(
              (f) => Padding(
                padding: const EdgeInsets.symmetric(vertical: 3),
                child: Row(
                  children: [
                    Icon(Icons.check_circle, size: 16, color: plan.color),
                    const SizedBox(width: 8),
                    Text(f),
                  ],
                ),
              ),
            ),
            if (!isCurrent) ...[
              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: plan.color,
                    foregroundColor: Colors.white,
                  ),
                  onPressed: () {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: Text('Upgrade para ${plan.name} em breve!'),
                      ),
                    );
                  },
                  child: Text('Upgrade para ${plan.name}'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
