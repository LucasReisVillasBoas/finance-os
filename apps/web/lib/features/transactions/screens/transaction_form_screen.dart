import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/transactions_provider.dart';
import '../../accounts/providers/accounts_provider.dart';
import '../../settings/providers/categories_provider.dart';
import '../../settings/models/category_model.dart';

class TransactionFormScreen extends ConsumerStatefulWidget {
  final String? transactionId;

  const TransactionFormScreen({super.key, this.transactionId});

  bool get isEditing => transactionId != null;

  @override
  ConsumerState<TransactionFormScreen> createState() =>
      _TransactionFormScreenState();
}

class _TransactionFormScreenState
    extends ConsumerState<TransactionFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _descriptionController = TextEditingController();
  final _notesController = TextEditingController();
  final _amountController = TextEditingController();
  final _tagController = TextEditingController();

  String _type = 'expense';
  DateTime _date = DateTime.now();
  String? _selectedAccountId;
  String? _selectedCategoryId;
  List<String> _tags = [];
  bool _isSaving = false;

  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(accountsProvider.notifier).loadAccounts();
      ref.read(categoriesProvider.notifier).loadCategories();
    });
  }

  @override
  void dispose() {
    _descriptionController.dispose();
    _notesController.dispose();
    _amountController.dispose();
    _tagController.dispose();
    super.dispose();
  }

  Future<void> _selectDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _date,
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );
    if (picked != null && mounted) {
      setState(() => _date = picked);
    }
  }

  void _addTag(String tag) {
    final trimmed = tag.trim();
    if (trimmed.isNotEmpty && !_tags.contains(trimmed)) {
      setState(() {
        _tags = [..._tags, trimmed];
        _tagController.clear();
      });
    }
  }

  void _removeTag(String tag) {
    setState(() => _tags = _tags.where((t) => t != tag).toList());
  }

  double? _parseAmount(String value) {
    final clean = value.replaceAll('.', '').replaceAll(',', '.');
    return double.tryParse(clean);
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedAccountId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Selecione uma conta')),
      );
      return;
    }

    final amount = _parseAmount(_amountController.text);
    if (amount == null || amount <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Valor inválido')),
      );
      return;
    }

    setState(() => _isSaving = true);

    final payload = {
      'account_id': _selectedAccountId,
      if (_selectedCategoryId != null) 'category_id': _selectedCategoryId,
      'type': _type,
      'amount': amount,
      if (_descriptionController.text.isNotEmpty)
        'description': _descriptionController.text,
      if (_notesController.text.isNotEmpty) 'notes': _notesController.text,
      'date': _date.toUtc().toIso8601String(),
      'tags': _tags,
    };

    try {
      if (widget.isEditing) {
        final success = await ref
            .read(transactionsProvider.notifier)
            .updateTransaction(widget.transactionId!, payload);
        if (mounted) {
          if (success) {
            if (context.canPop()) {
              context.pop();
            } else {
              context.go('/transactions');
            }
          } else {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Erro ao atualizar transação')),
            );
          }
        }
      } else {
        final tx = await ref
            .read(transactionsProvider.notifier)
            .createTransaction(payload);
        if (mounted) {
          if (tx != null) {
            if (context.canPop()) {
              context.pop();
            } else {
              context.go('/transactions');
            }
          } else {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Erro ao criar transação')),
            );
          }
        }
      }
    } finally {
      if (mounted) setState(() => _isSaving = false);
    }
  }

  List<CategoryModel> _filteredCategories(List<CategoryModel> all) {
    return all.where((c) => c.type == _type || c.type == 'transfer').toList();
  }

  @override
  Widget build(BuildContext context) {
    final accountsState = ref.watch(accountsProvider);
    final categoriesState = ref.watch(categoriesProvider);
    final dateFormat = DateFormat('dd/MM/yyyy', 'pt_BR');

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.isEditing ? 'Editar Transação' : 'Nova Transação'),
        centerTitle: true,
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // Type selector
            SegmentedButton<String>(
              segments: const [
                ButtonSegment(
                  value: 'expense',
                  label: Text('Despesa'),
                  icon: Icon(Icons.arrow_upward),
                ),
                ButtonSegment(
                  value: 'income',
                  label: Text('Receita'),
                  icon: Icon(Icons.arrow_downward),
                ),
              ],
              selected: {_type},
              onSelectionChanged: (Set<String> selected) {
                setState(() {
                  _type = selected.first;
                  _selectedCategoryId = null;
                });
              },
            ),
            const SizedBox(height: 16),

            // Amount field
            TextFormField(
              controller: _amountController,
              decoration: const InputDecoration(
                labelText: 'Valor *',
                prefixText: 'R\$ ',
                border: OutlineInputBorder(),
              ),
              keyboardType:
                  const TextInputType.numberWithOptions(decimal: true),
              inputFormatters: [
                FilteringTextInputFormatter.allow(RegExp(r'[\d,.]')),
              ],
              validator: (v) {
                if (v == null || v.isEmpty) return 'Informe o valor';
                final parsed = _parseAmount(v);
                if (parsed == null || parsed <= 0) return 'Valor inválido';
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
              maxLength: 255,
            ),
            const SizedBox(height: 8),

            // Notes
            TextFormField(
              controller: _notesController,
              decoration: const InputDecoration(
                labelText: 'Observações',
                border: OutlineInputBorder(),
              ),
              maxLines: 3,
              maxLength: 1000,
            ),
            const SizedBox(height: 8),

            // Date picker
            InkWell(
              onTap: _selectDate,
              child: InputDecorator(
                decoration: const InputDecoration(
                  labelText: 'Data *',
                  border: OutlineInputBorder(),
                  suffixIcon: Icon(Icons.calendar_today),
                ),
                child: Text(dateFormat.format(_date)),
              ),
            ),
            const SizedBox(height: 16),

            // Account dropdown
            DropdownButtonFormField<String>(
              initialValue: _selectedAccountId,
              decoration: const InputDecoration(
                labelText: 'Conta *',
                border: OutlineInputBorder(),
              ),
              items: accountsState.accounts
                  .map((a) => DropdownMenuItem(
                        value: a.id,
                        child: Text(a.name),
                      ))
                  .toList(),
              onChanged: (v) => setState(() => _selectedAccountId = v),
              hint: accountsState.isLoading
                  ? const Text('Carregando...')
                  : const Text('Selecione a conta'),
            ),
            const SizedBox(height: 16),

            // Category dropdown
            DropdownButtonFormField<String>(
              initialValue: _selectedCategoryId,
              decoration: const InputDecoration(
                labelText: 'Categoria',
                border: OutlineInputBorder(),
              ),
              items: [
                const DropdownMenuItem<String>(
                  value: null,
                  child: Text('Sem categoria'),
                ),
                ..._filteredCategories(categoriesState.categories)
                    .map((c) => DropdownMenuItem(
                          value: c.id,
                          child: Text(c.name),
                        )),
              ],
              onChanged: (v) => setState(() => _selectedCategoryId = v),
              hint: categoriesState.isLoading
                  ? const Text('Carregando...')
                  : const Text('Selecione a categoria'),
            ),
            const SizedBox(height: 16),

            // Tags
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    controller: _tagController,
                    decoration: const InputDecoration(
                      labelText: 'Tags',
                      border: OutlineInputBorder(),
                      hintText: 'Adicionar tag...',
                    ),
                    onFieldSubmitted: _addTag,
                  ),
                ),
                const SizedBox(width: 8),
                IconButton(
                  onPressed: () => _addTag(_tagController.text),
                  icon: const Icon(Icons.add_circle_outline),
                  tooltip: 'Adicionar tag',
                ),
              ],
            ),
            if (_tags.isNotEmpty)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Wrap(
                  spacing: 8,
                  children: _tags
                      .map((tag) => Chip(
                            label: Text(tag),
                            onDeleted: () => _removeTag(tag),
                          ))
                      .toList(),
                ),
              ),
            const SizedBox(height: 32),

            // Save button
            ElevatedButton(
              onPressed: _isSaving ? null : _submit,
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: _isSaving
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(
                      widget.isEditing ? 'Salvar Alterações' : 'Salvar',
                      style: const TextStyle(fontSize: 16),
                    ),
            ),
          ],
        ),
      ),
    );
  }
}
