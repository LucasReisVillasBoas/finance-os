import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class WhatsAppSettingsScreen extends ConsumerStatefulWidget {
  const WhatsAppSettingsScreen({super.key});

  @override
  ConsumerState<WhatsAppSettingsScreen> createState() =>
      _WhatsAppSettingsScreenState();
}

class _WhatsAppSettingsScreenState
    extends ConsumerState<WhatsAppSettingsScreen> {
  final _phoneController = TextEditingController();
  bool _botEnabled = false;
  bool _isConnecting = false;
  bool _isConnected = false;

  // Simulated last messages
  final List<_BotMessage> _recentMessages = [
    _BotMessage(
      text: 'resumo',
      isIncoming: true,
      time: DateTime.now().subtract(const Duration(minutes: 5)),
    ),
    _BotMessage(
      text:
          '📊 Resumo de Abril\n\n💰 Saldo: R\$ 5.000,00\n📉 Gastos: R\$ 1.500,00',
      isIncoming: false,
      time: DateTime.now().subtract(const Duration(minutes: 5)),
    ),
    _BotMessage(
      text: 'gastei 50 reais em mercado',
      isIncoming: true,
      time: DateTime.now().subtract(const Duration(minutes: 2)),
    ),
    _BotMessage(
      text: 'Confirmar gasto de R\$ 50,00 em Mercado? Responda *sim* ou *não*.',
      isIncoming: false,
      time: DateTime.now().subtract(const Duration(minutes: 2)),
    ),
    _BotMessage(
      text: 'sim',
      isIncoming: true,
      time: DateTime.now().subtract(const Duration(minutes: 1)),
    ),
    _BotMessage(
      text: 'Gasto de R\$ 50,00 em Mercado registrado com sucesso! ✅',
      isIncoming: false,
      time: DateTime.now().subtract(const Duration(minutes: 1)),
    ),
  ];

  @override
  void dispose() {
    _phoneController.dispose();
    super.dispose();
  }

  void _connectWhatsApp() {
    if (_phoneController.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Informe o número de WhatsApp.')),
      );
      return;
    }

    setState(() => _isConnecting = true);

    Future.delayed(const Duration(seconds: 2), () {
      if (mounted) {
        setState(() {
          _isConnecting = false;
          _isConnected = true;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('WhatsApp conectado com sucesso!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('WhatsApp Bot'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Connection card
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: Colors.green.shade100,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: const Icon(
                          Icons.chat,
                          color: Colors.green,
                          size: 28,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Conectar WhatsApp',
                            style: theme.textTheme.titleMedium
                                ?.copyWith(fontWeight: FontWeight.bold),
                          ),
                          Text(
                            _isConnected
                                ? '✅ Conectado'
                                : 'Não conectado',
                            style: TextStyle(
                              color:
                                  _isConnected ? Colors.green : Colors.grey,
                              fontSize: 12,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: _phoneController,
                    keyboardType: TextInputType.phone,
                    decoration: const InputDecoration(
                      labelText: 'Número do WhatsApp',
                      hintText: '5511999999999',
                      border: OutlineInputBorder(),
                      prefixText: '+',
                      helperText: 'Código do país + DDD + número (sem espaços)',
                    ),
                  ),
                  const SizedBox(height: 12),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      onPressed: _isConnecting ? null : _connectWhatsApp,
                      icon: _isConnecting
                          ? const SizedBox(
                              width: 18,
                              height: 18,
                              child:
                                  CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.link),
                      label: Text(_isConnecting
                          ? 'Conectando...'
                          : _isConnected
                              ? 'Reconectar'
                              : 'Conectar'),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.green,
                        foregroundColor: Colors.white,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Bot toggle
          Card(
            child: SwitchListTile(
              title: const Text('Ativar Bot'),
              subtitle: const Text(
                'Processa mensagens recebidas automaticamente',
              ),
              value: _botEnabled,
              onChanged: _isConnected
                  ? (value) => setState(() => _botEnabled = value)
                  : null,
              secondary: const Icon(Icons.smart_toy),
            ),
          ),
          const SizedBox(height: 16),

          // Commands reference
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Comandos Disponíveis',
                    style: theme.textTheme.titleSmall
                        ?.copyWith(fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 12),
                  const _CommandRow(
                    command: 'resumo',
                    description: 'Saldo e gastos do mês atual',
                  ),
                  const _CommandRow(
                    command: 'gastei X',
                    description: 'Registrar um gasto (ex: gastei 50)',
                  ),
                  const _CommandRow(
                    command: 'gastei X em Y',
                    description:
                        'Gasto com categoria (ex: gastei 30 em mercado)',
                  ),
                  const _CommandRow(
                    command: 'recebi X',
                    description: 'Registrar uma receita (ex: recebi 5000)',
                  ),
                  const _CommandRow(
                    command: 'quanto gastei com Y?',
                    description: 'Gastos por categoria este mês',
                  ),
                  const _CommandRow(
                    command: 'carteira',
                    description: 'Resumo de contas e saldos',
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Recent messages
          if (_recentMessages.isNotEmpty) ...[
            Text(
              'Últimas Mensagens',
              style: theme.textTheme.titleSmall
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  children: _recentMessages.map((msg) {
                    return _MessageBubble(message: msg);
                  }).toList(),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _CommandRow extends StatelessWidget {
  final String command;
  final String description;

  const _CommandRow({required this.command, required this.description});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(
                color: theme.colorScheme.outline.withValues(alpha: 0.4),
              ),
            ),
            child: Text(
              command,
              style: TextStyle(
                fontFamily: 'monospace',
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: theme.colorScheme.onSurface,
              ),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              description,
              style: TextStyle(
                color: theme.colorScheme.onSurfaceVariant,
                fontSize: 13,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _BotMessage {
  final String text;
  final bool isIncoming;
  final DateTime time;

  const _BotMessage({
    required this.text,
    required this.isIncoming,
    required this.time,
  });
}

class _MessageBubble extends StatelessWidget {
  final _BotMessage message;

  const _MessageBubble({required this.message});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;
    final isUser = message.isIncoming;

    final Color bg;
    final Color fg;
    if (isUser) {
      bg = isDark
          ? const Color(0xFF1E4D2B)
          : Colors.green.shade100;
      fg = isDark ? Colors.white : Colors.black87;
    } else {
      bg = isDark
          ? theme.colorScheme.surfaceContainerHighest
          : Colors.grey.shade100;
      fg = theme.colorScheme.onSurface;
    }

    return Align(
      alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 3),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        constraints: BoxConstraints(
          maxWidth: MediaQuery.of(context).size.width * 0.7,
        ),
        decoration: BoxDecoration(
          color: bg,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Text(
          message.text,
          style: TextStyle(fontSize: 13, color: fg),
        ),
      ),
    );
  }
}
