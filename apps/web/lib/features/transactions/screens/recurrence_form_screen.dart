import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/recurrences_provider.dart';
import '../models/recurrence_model.dart';
import '../../accounts/providers/accounts_provider.dart';
import '../../settings/providers/categories_provider.dart';

class RecurrenceFormScreen extends ConsumerStatefulWidget {
  final String? recurrenceId;

  const RecurrenceFormScreen({super.key, this.recurrenceId});

  @override
  ConsumerState<RecurrenceFormScreen> createState() =>
      _RecurrenceFormScreenState();
}

class _RecurrenceFormScreenState extends ConsumerState<RecurrenceFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _amountController = TextEditingController();
  final _descriptionController = TextEditingController();

  String _selectedType = 'expense';
  String _selectedFrequency = 'monthly';
  String? _selectedAccountId;
  String? _selectedCategoryId;
  DateTime _startDate = DateTime.now();
  DateTime? _endDate;
  bool _autoLaunch = false;
  bool _initialized = false;

  static const _frequencies = [
    ('daily', 'Diário'),
    ('weekly', 'Semanal'),
    ('biweekly', 'Quinzenal'),
    ('monthly', 'Mensal'),
    ('yearly', 'Anual'),
  ];

  bool get _isEditing => widget.recurrenceId != null;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_initialized) {
      _initialized = true;
      Future.microtask(() {
        ref.read(accountsProvider.notifier).loadAccounts();
        ref.read(categoriesProvider.notifier).loadCategories();
        if (_isEditing) {
          _loadExisting();
        }
      });
    }
  }

  void _loadExisting() {
    final recurrencesState = ref.read(recurrencesProvider);
    final existing = recurrencesState.items.where(
      (r) => r.id == widget.recurrenceId,
    );
    if (existing.isNotEmpty) {
      _populateFromModel(existing.first);
    }
  }

  void _populateFromModel(RecurrenceModel rec) {
    _amountController.text = rec.amount.toStringAsFixed(2);
    _descriptionController.text = rec.description ?? '';
    setState(() {
      _selectedType = rec.type;
      _selectedFrequency = rec.frequency;
      _selectedAccountId = rec.accountId;
      _selectedCategoryId = rec.categoryId;
      _startDate = rec.startDate;
      _endDate = rec.endDate;
      _autoLaunch = rec.autoLaunch;
    });
  }

  @override
  void dispose() {
    _amountController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  Future<void> _pickDate({required bool isStart}) async {
    final picked = await showDatePicker(
      context: context,
      initialDate: isStart ? _startDate : (_endDate ?? DateTime.now()),
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );
    if (picked != null) {
      setState(() {
        if (isStart) {
          _startDate = picked;
        } else {
          _endDate = picked;
        }
      });
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedAccountId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Selecione uma conta')),
      );
      return;
    }

    final amount = double.tryParse(
          _amountController.text.replaceAll(',', '.'),
        ) ??
        0.0;

    final payload = <String, dynamic>{
      'account_id': _selectedAccountId,
      if (_selectedCategoryId != null) 'category_id': _selectedCategoryId,
      'type': _selectedType,
      'amount': amount,
      if (_descriptionController.text.isNotEmpty)
        'description': _descriptionController.text,
      'frequency': _selectedFrequency,
      'start_date': _startDate.toIso8601String(),
      if (_endDate != null) 'end_date': _endDate!.toIso8601String(),
      'auto_launch': _autoLaunch,
    };

    bool success;
    if (_isEditing) {
      success = await ref
          .read(recurrencesProvider.notifier)
          .update(widget.recurrenceId!, payload);
    } else {
      final rec =
          await ref.read(recurrencesProvider.notifier).create(payload);
      success = rec != null;
    }

    if (success && mounted) {
      context.pop();
    } else if (mounted) {
      final error = ref.read(recurrencesProvider).error;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error ?? 'Erro ao salvar')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final accountsState = ref.watch(accountsProvider);
    final categoriesState = ref.watch(categoriesProvider);
    final recurrencesState = ref.watch(recurrencesProvider);
    final dateFormatter = DateFormat('dd/MM/yyyy');

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditing ? 'Editar Recorrência' : 'Nova Recorrência'),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Type selector
              const Text('Tipo', style: TextStyle(fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              SegmentedButton<String>(
                segments: const [
                  ButtonSegment(value: 'expense', label: Text('Despesa')),
                  ButtonSegment(value: 'income', label: Text('Receita')),
                ],
                selected: {_selectedType},
                onSelectionChanged: (values) =>
                    setState(() => _selectedType = values.first),
              ),
              const SizedBox(height: 16),
              // Amount
              TextFormField(
                controller: _amountController,
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                decoration: const InputDecoration(
                  labelText: 'Valor',
                  prefixText: 'R\$ ',
                  border: OutlineInputBorder(),
                ),
                validator: (v) {
                  if (v == null || v.isEmpty) return 'Informe o valor';
                  final parsed = double.tryParse(v.replaceAll(',', '.'));
                  if (parsed == null || parsed <= 0) {
                    return 'Valor deve ser maior que zero';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              // Description
              TextFormField(
                controller: _descriptionController,
                decoration: const InputDecoration(
                  labelText: 'Descrição',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              // Frequency
              DropdownButtonFormField<String>(
                initialValue: _selectedFrequency,
                decoration: const InputDecoration(
                  labelText: 'Frequência',
                  border: OutlineInputBorder(),
                ),
                items: _frequencies
                    .map((f) => DropdownMenuItem(
                          value: f.$1,
                          child: Text(f.$2),
                        ))
                    .toList(),
                onChanged: (v) =>
                    setState(() => _selectedFrequency = v ?? 'monthly'),
              ),
              const SizedBox(height: 16),
              // Account
              DropdownButtonFormField<String>(
                initialValue: _selectedAccountId,
                decoration: const InputDecoration(
                  labelText: 'Conta',
                  border: OutlineInputBorder(),
                ),
                items: accountsState.accounts
                    .map((a) => DropdownMenuItem(
                          value: a.id,
                          child: Text(a.name),
                        ))
                    .toList(),
                onChanged: (v) => setState(() => _selectedAccountId = v),
                validator: (v) =>
                    v == null ? 'Selecione uma conta' : null,
              ),
              const SizedBox(height: 16),
              // Category (optional)
              DropdownButtonFormField<String>(
                initialValue: _selectedCategoryId,
                decoration: const InputDecoration(
                  labelText: 'Categoria (opcional)',
                  border: OutlineInputBorder(),
                ),
                items: [
                  const DropdownMenuItem(
                    value: null,
                    child: Text('Sem categoria'),
                  ),
                  ...categoriesState.categories.map((c) => DropdownMenuItem(
                        value: c.id,
                        child: Text(c.name),
                      )),
                ],
                onChanged: (v) => setState(() => _selectedCategoryId = v),
              ),
              const SizedBox(height: 16),
              // Start date
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Data de início'),
                subtitle: Text(dateFormatter.format(_startDate)),
                trailing: const Icon(Icons.calendar_today),
                onTap: () => _pickDate(isStart: true),
              ),
              // End date
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Data de término (opcional)'),
                subtitle: Text(
                  _endDate != null
                      ? dateFormatter.format(_endDate!)
                      : 'Sem data de término',
                ),
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (_endDate != null)
                      IconButton(
                        icon: const Icon(Icons.clear),
                        onPressed: () => setState(() => _endDate = null),
                      ),
                    const Icon(Icons.calendar_today),
                  ],
                ),
                onTap: () => _pickDate(isStart: false),
              ),
              const SizedBox(height: 8),
              // Auto launch
              SwitchListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Lançar automaticamente'),
                subtitle: const Text(
                  'Cria a transação automaticamente na data de vencimento',
                ),
                value: _autoLaunch,
                onChanged: (v) => setState(() => _autoLaunch = v),
              ),
              const SizedBox(height: 24),
              ElevatedButton(
                onPressed: recurrencesState.isLoading ? null : _submit,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: recurrencesState.isLoading
                    ? const CircularProgressIndicator()
                    : Text(_isEditing ? 'Salvar alterações' : 'Criar recorrência'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
