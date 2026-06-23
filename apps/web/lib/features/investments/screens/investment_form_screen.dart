import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/investments_provider.dart';
import '../models/asset_model.dart';

class InvestmentFormScreen extends ConsumerStatefulWidget {
  final String? holdingId;
  const InvestmentFormScreen({super.key, this.holdingId});

  @override
  ConsumerState<InvestmentFormScreen> createState() =>
      _InvestmentFormScreenState();
}

class _InvestmentFormScreenState extends ConsumerState<InvestmentFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _searchController = TextEditingController();
  final _qtyController = TextEditingController();
  final _priceController = TextEditingController();
  final _feesController = TextEditingController(text: '0');
  final _notesController = TextEditingController();

  Timer? _debounce;

  String _transactionType = 'buy';
  DateTime _selectedDate = DateTime.now();
  String? _selectedAssetId;
  String _holdingType = 'stock';
  bool _isSubmitting = false;
  bool _showSuggestions = false;

  final _transactionTypes = [
    {'value': 'buy', 'label': 'Compra'},
    {'value': 'sell', 'label': 'Venda'},
    {'value': 'dividend', 'label': 'Dividendo'},
    {'value': 'split', 'label': 'Desdobramento'},
    {'value': 'bonus', 'label': 'Bonificação'},
  ];

  final _holdingTypes = [
    {'value': 'stock', 'label': 'Ação'},
    {'value': 'fii', 'label': 'FII'},
    {'value': 'etf', 'label': 'ETF'},
    {'value': 'crypto', 'label': 'Cripto'},
    {'value': 'fixed_income', 'label': 'Renda Fixa'},
    {'value': 'fund', 'label': 'Fundo'},
    {'value': 'other', 'label': 'Outro'},
  ];

  @override
  void dispose() {
    _debounce?.cancel();
    _searchController.dispose();
    _qtyController.dispose();
    _priceController.dispose();
    _feesController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _selectDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _selectedDate,
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );
    if (picked != null) {
      setState(() => _selectedDate = picked);
    }
  }

  void _onSearchChanged(String query) {
    _debounce?.cancel();
    if (query.length < 2) {
      setState(() {
        _showSuggestions = false;
      });
      return;
    }
    _debounce = Timer(const Duration(milliseconds: 400), () {
      ref.read(investmentsProvider.notifier).searchAssets(query);
      setState(() => _showSuggestions = true);
    });
  }

  void _selectAsset(AssetModel asset) {
    setState(() {
      _selectedAssetId = asset.id;
      _searchController.text = asset.displayName;
      _showSuggestions = false;
    });
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    final state = ref.read(investmentsProvider);
    if (state.portfolios.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
            content: Text('Crie um portfólio antes de adicionar operações.')),
      );
      return;
    }

    final portfolioId = state.selectedPortfolioId;
    if (portfolioId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Nenhum portfólio selecionado.')),
      );
      return;
    }

    final qty = double.tryParse(_qtyController.text.replaceAll(',', '.'));
    final price = double.tryParse(_priceController.text.replaceAll(',', '.'));
    final fees =
        double.tryParse(_feesController.text.replaceAll(',', '.')) ?? 0.0;
    final total = (qty ?? 0) * (price ?? 0) + fees;

    setState(() => _isSubmitting = true);

    // If we don't have an existing holdingId, create a holding first.
    String? holdingId = widget.holdingId;
    if (holdingId == null) {
      final assetName = _searchController.text.trim();
      final holding = await ref
          .read(investmentsProvider.notifier)
          .createHolding(portfolioId, {
        'name': assetName.isNotEmpty ? assetName : 'Ativo',
        'type': _holdingType,
        if (_selectedAssetId != null) 'asset_id': _selectedAssetId,
      });
      if (holding == null) {
        setState(() => _isSubmitting = false);
        final err = ref.read(investmentsProvider).error;
        if (mounted && err != null) {
          ScaffoldMessenger.of(context)
              .showSnackBar(SnackBar(content: Text(err)));
        }
        return;
      }
      holdingId = holding.id;
    }

    final payload = <String, dynamic>{
      'type': _transactionType,
      'quantity': qty,
      'price': price,
      'fees': fees,
      'total': total,
      'date': _selectedDate.toUtc().toIso8601String(),
      if (_selectedAssetId != null) 'asset_id': _selectedAssetId,
      if (_notesController.text.isNotEmpty) 'notes': _notesController.text,
    };
    payload.removeWhere((_, v) => v == null);

    final tx = await ref
        .read(investmentsProvider.notifier)
        .createTransaction(holdingId, payload);
    setState(() => _isSubmitting = false);

    if (tx != null && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Operação registrada com sucesso!')),
      );
      context.pop();
    } else {
      final err = ref.read(investmentsProvider).error;
      if (mounted && err != null) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(err)));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(investmentsProvider);
    final dateFormat = DateFormat('dd/MM/yyyy', 'pt_BR');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Nova Operação'),
      ),
      body: Form(
        key: _formKey,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Asset search
              const Text('Ativo',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              TextFormField(
                controller: _searchController,
                decoration: const InputDecoration(
                  hintText: 'Buscar por ticker ou nome (ex: PETR4)',
                  prefixIcon: Icon(Icons.search),
                  border: OutlineInputBorder(),
                ),
                onChanged: _onSearchChanged,
                validator: (_) => null, // asset is optional (custom holding)
              ),
              if (_showSuggestions && state.searchResults.isNotEmpty)
                Container(
                  decoration: BoxDecoration(
                    border: Border.all(color: Colors.grey.shade300),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  constraints: const BoxConstraints(maxHeight: 200),
                  child: ListView.builder(
                    shrinkWrap: true,
                    itemCount: state.searchResults.length,
                    itemBuilder: (context, index) {
                      final asset = state.searchResults[index];
                      return ListTile(
                        title: Text(asset.displayName),
                        subtitle: Text('${asset.type} · ${asset.exchange ?? ''}'),
                        dense: true,
                        onTap: () => _selectAsset(asset),
                      );
                    },
                  ),
                ),
              const SizedBox(height: 16),

              // Asset type (only shown when creating a new holding)
              if (widget.holdingId == null) ...[
                const Text('Tipo de ativo',
                    style: TextStyle(fontWeight: FontWeight.w600)),
                const SizedBox(height: 4),
                DropdownButtonFormField<String>(
                  initialValue: _holdingType,
                  decoration: const InputDecoration(
                    border: OutlineInputBorder(),
                  ),
                  items: _holdingTypes
                      .map((t) => DropdownMenuItem<String>(
                            value: t['value'],
                            child: Text(t['label']!),
                          ))
                      .toList(),
                  onChanged: (v) {
                    if (v != null) setState(() => _holdingType = v);
                  },
                ),
                const SizedBox(height: 16),
              ],

              // Transaction type
              const Text('Tipo de operação',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              DropdownButtonFormField<String>(
                initialValue: _transactionType,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                ),
                items: _transactionTypes
                    .map((t) => DropdownMenuItem<String>(
                          value: t['value'],
                          child: Text(t['label']!),
                        ))
                    .toList(),
                onChanged: (v) {
                  if (v != null) setState(() => _transactionType = v);
                },
                validator: (v) =>
                    v == null ? 'Selecione o tipo de operação' : null,
              ),
              const SizedBox(height: 16),

              // Quantity and price
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Quantidade',
                            style: TextStyle(fontWeight: FontWeight.w600)),
                        const SizedBox(height: 4),
                        TextFormField(
                          controller: _qtyController,
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                            hintText: '0',
                          ),
                          keyboardType: const TextInputType.numberWithOptions(
                              decimal: true),
                          validator: (v) {
                            if (_transactionType == 'dividend') return null;
                            if (v == null || v.isEmpty) {
                              return 'Informe a quantidade';
                            }
                            if (double.tryParse(v.replaceAll(',', '.')) ==
                                null) {
                              return 'Valor inválido';
                            }
                            return null;
                          },
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Preço (R\$)',
                            style: TextStyle(fontWeight: FontWeight.w600)),
                        const SizedBox(height: 4),
                        TextFormField(
                          controller: _priceController,
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                            hintText: '0,00',
                          ),
                          keyboardType: const TextInputType.numberWithOptions(
                              decimal: true),
                          validator: (v) {
                            if (_transactionType == 'dividend') return null;
                            if (v == null || v.isEmpty) {
                              return 'Informe o preço';
                            }
                            if (double.tryParse(v.replaceAll(',', '.')) ==
                                null) {
                              return 'Valor inválido';
                            }
                            return null;
                          },
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),

              // Fees
              const Text('Taxas (R\$)',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              TextFormField(
                controller: _feesController,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  hintText: '0,00',
                ),
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                validator: (v) {
                  if (v == null || v.isEmpty) return null;
                  if (double.tryParse(v.replaceAll(',', '.')) == null) {
                    return 'Valor inválido';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),

              // Date
              const Text('Data',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              InkWell(
                onTap: _selectDate,
                child: Container(
                  padding: const EdgeInsets.symmetric(
                      horizontal: 12, vertical: 14),
                  decoration: BoxDecoration(
                    border: Border.all(color: Colors.grey.shade400),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.calendar_today,
                          size: 18, color: Colors.grey),
                      const SizedBox(width: 8),
                      Text(dateFormat.format(_selectedDate)),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),

              // Notes
              const Text('Observações (opcional)',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              TextFormField(
                controller: _notesController,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  hintText: 'Notas sobre a operação...',
                ),
                maxLines: 2,
              ),
              const SizedBox(height: 24),

              ElevatedButton(
                onPressed: _isSubmitting ? null : _submit,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: _isSubmitting
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Registrar Operação',
                        style: TextStyle(fontSize: 16)),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
