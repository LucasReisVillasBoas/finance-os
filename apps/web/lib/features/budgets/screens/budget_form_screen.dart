import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/budgets_provider.dart';
import '../models/budget_model.dart';
import '../../settings/providers/categories_provider.dart';

class BudgetFormScreen extends ConsumerStatefulWidget {
  final String? budgetId;

  const BudgetFormScreen({super.key, this.budgetId});

  @override
  ConsumerState<BudgetFormScreen> createState() => _BudgetFormScreenState();
}

class _BudgetFormScreenState extends ConsumerState<BudgetFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _amountController = TextEditingController();

  String? _selectedCategoryId;
  String _selectedPeriod = 'monthly';
  double _thresholdPct = 80.0;
  bool _initialized = false;

  static const _periods = [
    ('weekly', 'Semanal'),
    ('monthly', 'Mensal'),
    ('yearly', 'Anual'),
  ];

  bool get _isEditing => widget.budgetId != null;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_initialized) {
      _initialized = true;
      Future.microtask(() {
        ref.read(categoriesProvider.notifier).loadCategories();
        if (_isEditing) {
          _loadExisting();
        }
      });
    }
  }

  void _loadExisting() {
    final budgetsState = ref.read(budgetsProvider);
    final existing =
        budgetsState.budgets.where((b) => b.id == widget.budgetId);
    if (existing.isNotEmpty) {
      _populateFromModel(existing.first);
    }
  }

  void _populateFromModel(BudgetModel budget) {
    _amountController.text = budget.amount.toStringAsFixed(2);
    setState(() {
      _selectedCategoryId = budget.categoryId;
      _selectedPeriod = budget.period;
      _thresholdPct = budget.thresholdPct;
    });
  }

  @override
  void dispose() {
    _amountController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    final amount = double.tryParse(
          _amountController.text.replaceAll(',', '.'),
        ) ??
        0.0;

    final budgetsState = ref.read(budgetsProvider);
    final payload = <String, dynamic>{
      if (_selectedCategoryId != null) 'category_id': _selectedCategoryId,
      'amount': amount,
      'period': _selectedPeriod,
      'month': budgetsState.month,
      'year': budgetsState.year,
      'threshold_pct': _thresholdPct,
    };

    bool success;
    if (_isEditing) {
      success = await ref
          .read(budgetsProvider.notifier)
          .update(widget.budgetId!, payload);
    } else {
      final budget =
          await ref.read(budgetsProvider.notifier).create(payload);
      success = budget != null;
    }

    if (success && mounted) {
      context.pop();
    } else if (mounted) {
      final error = ref.read(budgetsProvider).error;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error ?? 'Erro ao salvar')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final categoriesState = ref.watch(categoriesProvider);
    final budgetsState = ref.watch(budgetsProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditing ? 'Editar Orçamento' : 'Novo Orçamento'),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Category
              DropdownButtonFormField<String>(
                initialValue: _selectedCategoryId,
                decoration: const InputDecoration(
                  labelText: 'Categoria',
                  border: OutlineInputBorder(),
                ),
                items: [
                  const DropdownMenuItem(
                    value: null,
                    child: Text('Geral - sem categoria'),
                  ),
                  ...categoriesState.categories.map((c) => DropdownMenuItem(
                        value: c.id,
                        child: Text(c.name),
                      )),
                ],
                onChanged: (v) => setState(() => _selectedCategoryId = v),
              ),
              const SizedBox(height: 16),
              // Amount
              TextFormField(
                controller: _amountController,
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                decoration: const InputDecoration(
                  labelText: 'Valor planejado',
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
              // Period
              DropdownButtonFormField<String>(
                initialValue: _selectedPeriod,
                decoration: const InputDecoration(
                  labelText: 'Período',
                  border: OutlineInputBorder(),
                ),
                items: _periods
                    .map((p) => DropdownMenuItem(
                          value: p.$1,
                          child: Text(p.$2),
                        ))
                    .toList(),
                onChanged: (v) =>
                    setState(() => _selectedPeriod = v ?? 'monthly'),
              ),
              const SizedBox(height: 24),
              // Threshold slider
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Alerta de limite',
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                      Text(
                        '${_thresholdPct.toStringAsFixed(0)}%',
                        style: const TextStyle(fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'Notifica quando o gasto atingir ${_thresholdPct.toStringAsFixed(0)}% do orçamento',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey[600],
                    ),
                  ),
                  Slider(
                    value: _thresholdPct,
                    min: 50,
                    max: 100,
                    divisions: 10,
                    label: '${_thresholdPct.toStringAsFixed(0)}%',
                    onChanged: (v) => setState(() => _thresholdPct = v),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              ElevatedButton(
                onPressed: budgetsState.isLoading ? null : _submit,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: budgetsState.isLoading
                    ? const CircularProgressIndicator()
                    : Text(_isEditing ? 'Salvar alterações' : 'Criar orçamento'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
