import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/accounts_provider.dart';
import '../models/account_model.dart';

class AccountFormScreen extends ConsumerStatefulWidget {
  /// If [accountId] is provided, the screen is in edit mode.
  final String? accountId;

  const AccountFormScreen({super.key, this.accountId});

  @override
  ConsumerState<AccountFormScreen> createState() => _AccountFormScreenState();
}

class _AccountFormScreenState extends ConsumerState<AccountFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _institutionController = TextEditingController();
  final _balanceController = TextEditingController(text: '0.00');

  String _selectedType = 'checking';
  String? _selectedColor;
  AccountModel? _existingAccount;
  bool _initialized = false;

  static const _accountTypes = [
    ('checking', 'Conta Corrente'),
    ('savings', 'Poupança'),
    ('credit_card', 'Cartão de Crédito'),
    ('investment', 'Investimento'),
    ('wallet', 'Carteira'),
    ('other', 'Outro'),
  ];

  static const _predefinedColors = [
    '#2196F3', // Blue
    '#4CAF50', // Green
    '#F44336', // Red
    '#FF9800', // Orange
    '#9C27B0', // Purple
    '#00BCD4', // Cyan
    '#795548', // Brown
    '#607D8B', // Blue Grey
  ];

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_initialized && widget.accountId != null) {
      _initialized = true;
      final state = ref.read(accountsProvider);
      try {
        _existingAccount = state.accounts
            .firstWhere((a) => a.id == widget.accountId);
        _nameController.text = _existingAccount!.name;
        _institutionController.text = _existingAccount!.institution ?? '';
        _balanceController.text =
            _existingAccount!.balance.toStringAsFixed(2);
        _selectedType = _existingAccount!.type;
        _selectedColor = _existingAccount!.color;
      } catch (_) {
        // Account not in state — load will happen when user navigates here
      }
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _institutionController.dispose();
    _balanceController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    final payload = {
      'name': _nameController.text.trim(),
      'type': _selectedType,
      if (_institutionController.text.trim().isNotEmpty)
        'institution': _institutionController.text.trim(),
      'balance': double.tryParse(_balanceController.text) ?? 0.0,
      if (_selectedColor != null) 'color': _selectedColor,
    };

    final notifier = ref.read(accountsProvider.notifier);

    if (widget.accountId != null) {
      await notifier.updateAccount(widget.accountId!, payload);
    } else {
      await notifier.createAccount(payload);
    }

    if (mounted) {
      final state = ref.read(accountsProvider);
      if (state.error == null) {
        context.pop();
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(state.error!)),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final isEditing = widget.accountId != null;
    final state = ref.watch(accountsProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(isEditing ? 'Editar Conta' : 'Nova Conta'),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(
                  labelText: 'Nome da conta *',
                  border: OutlineInputBorder(),
                ),
                validator: (v) {
                  if (v == null || v.trim().length < 2) {
                    return 'Nome deve ter pelo menos 2 caracteres';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                initialValue: _selectedType,
                decoration: const InputDecoration(
                  labelText: 'Tipo de conta *',
                  border: OutlineInputBorder(),
                ),
                items: _accountTypes
                    .map((t) => DropdownMenuItem(
                          value: t.$1,
                          child: Text(t.$2),
                        ))
                    .toList(),
                onChanged: (v) => setState(() => _selectedType = v!),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _institutionController,
                decoration: const InputDecoration(
                  labelText: 'Instituição',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _balanceController,
                decoration: InputDecoration(
                  labelText: isEditing ? 'Saldo' : 'Saldo inicial',
                  border: const OutlineInputBorder(),
                  prefixText: 'R\$ ',
                ),
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                validator: (v) {
                  if (v == null || double.tryParse(v) == null) {
                    return 'Informe um valor válido';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 24),
              Text(
                'Cor',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: _predefinedColors.map((c) {
                  final color =
                      Color(int.parse(c.replaceFirst('#', '0xFF')));
                  final isSelected = _selectedColor == c;
                  return GestureDetector(
                    onTap: () => setState(() => _selectedColor = c),
                    child: Container(
                      width: 40,
                      height: 40,
                      decoration: BoxDecoration(
                        color: color,
                        shape: BoxShape.circle,
                        border: isSelected
                            ? Border.all(
                                color: Colors.black, width: 3)
                            : null,
                      ),
                      child: isSelected
                          ? const Icon(Icons.check,
                              color: Colors.white, size: 20)
                          : null,
                    ),
                  );
                }).toList(),
              ),
              const SizedBox(height: 32),
              ElevatedButton(
                onPressed: state.isLoading ? null : _submit,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: state.isLoading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Text(isEditing ? 'Salvar' : 'Criar Conta'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
