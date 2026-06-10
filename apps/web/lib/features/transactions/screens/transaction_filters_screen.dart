import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/transactions_provider.dart';
import '../repositories/transaction_repository.dart';
import '../../accounts/providers/accounts_provider.dart';
import '../../settings/providers/categories_provider.dart';

class TransactionFiltersScreen extends ConsumerStatefulWidget {
  const TransactionFiltersScreen({super.key});

  @override
  ConsumerState<TransactionFiltersScreen> createState() =>
      _TransactionFiltersScreenState();
}

class _TransactionFiltersScreenState
    extends ConsumerState<TransactionFiltersScreen> {
  DateTime? _startDate;
  DateTime? _endDate;
  String? _selectedType;
  List<String> _selectedCategoryIds = [];
  List<String> _selectedAccountIds = [];

  final _dateFormat = DateFormat('dd/MM/yyyy', 'pt_BR');

  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(accountsProvider.notifier).loadAccounts();
      ref.read(categoriesProvider.notifier).loadCategories();
    });

    // Pre-populate from current filter
    final current = ref.read(transactionsProvider).filter;
    if (current.startDate != null) {
      _startDate = DateTime.tryParse(current.startDate!);
    }
    if (current.endDate != null) {
      _endDate = DateTime.tryParse(current.endDate!);
    }
    _selectedType = current.type;
    if (current.categoryId != null) {
      _selectedCategoryIds = [current.categoryId!];
    }
    if (current.accountId != null) {
      _selectedAccountIds = [current.accountId!];
    }
  }

  Future<void> _selectDateRange() async {
    final picked = await showDateRangePicker(
      context: context,
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
      initialDateRange: _startDate != null && _endDate != null
          ? DateTimeRange(start: _startDate!, end: _endDate!)
          : null,
    );
    if (picked != null && mounted) {
      setState(() {
        _startDate = picked.start;
        _endDate = picked.end;
      });
    }
  }

  void _applyFilters() {
    final filter = TransactionFilter(
      startDate: _startDate != null
          ? DateFormat('yyyy-MM-dd', 'pt_BR').format(_startDate!)
          : null,
      endDate: _endDate != null
          ? DateFormat('yyyy-MM-dd', 'pt_BR').format(_endDate!)
          : null,
      type: _selectedType,
      categoryId: _selectedCategoryIds.isNotEmpty
          ? _selectedCategoryIds.first
          : null,
      accountId:
          _selectedAccountIds.isNotEmpty ? _selectedAccountIds.first : null,
    );
    ref.read(transactionsProvider.notifier).applyFilter(filter);
    context.pop();
  }

  void _clearFilters() {
    setState(() {
      _startDate = null;
      _endDate = null;
      _selectedType = null;
      _selectedCategoryIds = [];
      _selectedAccountIds = [];
    });
    ref.read(transactionsProvider.notifier).clearFilter();
    context.pop();
  }

  @override
  Widget build(BuildContext context) {
    final accountsState = ref.watch(accountsProvider);
    final categoriesState = ref.watch(categoriesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Filtros'),
        centerTitle: true,
        actions: [
          TextButton(
            onPressed: _clearFilters,
            child: const Text('Limpar'),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Date range
          const Text('Período',
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
          const SizedBox(height: 8),
          InkWell(
            onTap: _selectDateRange,
            child: InputDecorator(
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                suffixIcon: Icon(Icons.date_range),
              ),
              child: Text(
                _startDate != null && _endDate != null
                    ? '${_dateFormat.format(_startDate!)} — ${_dateFormat.format(_endDate!)}'
                    : 'Selecionar período...',
                style: _startDate == null
                    ? TextStyle(color: Colors.grey[500])
                    : null,
              ),
            ),
          ),
          if (_startDate != null || _endDate != null)
            Align(
              alignment: Alignment.centerRight,
              child: TextButton(
                onPressed: () =>
                    setState(() => _startDate = _endDate = null),
                child: const Text('Remover período'),
              ),
            ),
          const SizedBox(height: 20),

          // Type filter
          const Text('Tipo',
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
          const SizedBox(height: 8),
          DropdownButtonFormField<String>(
            initialValue: _selectedType,
            decoration: const InputDecoration(
              border: OutlineInputBorder(),
            ),
            hint: const Text('Todos os tipos'),
            items: const [
              DropdownMenuItem(value: null, child: Text('Todos os tipos')),
              DropdownMenuItem(value: 'income', child: Text('Receita')),
              DropdownMenuItem(value: 'expense', child: Text('Despesa')),
              DropdownMenuItem(
                  value: 'transfer', child: Text('Transferência')),
            ],
            onChanged: (v) => setState(() => _selectedType = v),
          ),
          const SizedBox(height: 20),

          // Accounts filter
          const Text('Contas',
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
          const SizedBox(height: 8),
          if (accountsState.isLoading)
            const Center(child: CircularProgressIndicator())
          else
            ...accountsState.accounts.map((a) => CheckboxListTile(
                  title: Text(a.name),
                  value: _selectedAccountIds.contains(a.id),
                  onChanged: (checked) {
                    setState(() {
                      if (checked == true) {
                        _selectedAccountIds = [a.id];
                      } else {
                        _selectedAccountIds =
                            _selectedAccountIds.where((id) => id != a.id).toList();
                      }
                    });
                  },
                  controlAffinity: ListTileControlAffinity.leading,
                  dense: true,
                )),
          const SizedBox(height: 20),

          // Categories filter
          const Text('Categorias',
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
          const SizedBox(height: 8),
          if (categoriesState.isLoading)
            const Center(child: CircularProgressIndicator())
          else
            ...categoriesState.categories.map((c) => CheckboxListTile(
                  title: Text(c.name),
                  value: _selectedCategoryIds.contains(c.id),
                  onChanged: (checked) {
                    setState(() {
                      if (checked == true) {
                        _selectedCategoryIds =
                            [..._selectedCategoryIds, c.id];
                      } else {
                        _selectedCategoryIds =
                            _selectedCategoryIds.where((id) => id != c.id).toList();
                      }
                    });
                  },
                  controlAffinity: ListTileControlAffinity.leading,
                  dense: true,
                )),
          const SizedBox(height: 32),
        ],
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: ElevatedButton(
            onPressed: _applyFilters,
            style: ElevatedButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 16),
            ),
            child: const Text('Aplicar Filtros', style: TextStyle(fontSize: 16)),
          ),
        ),
      ),
    );
  }
}
