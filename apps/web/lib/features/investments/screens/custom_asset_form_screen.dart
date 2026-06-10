import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/investments_provider.dart';
import '../models/custom_asset_model.dart';

class CustomAssetFormScreen extends ConsumerStatefulWidget {
  final String? customAssetId;
  const CustomAssetFormScreen({super.key, this.customAssetId});

  @override
  ConsumerState<CustomAssetFormScreen> createState() =>
      _CustomAssetFormScreenState();
}

class _CustomAssetFormScreenState
    extends ConsumerState<CustomAssetFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _currentValueController = TextEditingController();
  final _purchaseValueController = TextEditingController();
  final _monthlyIncomeController = TextEditingController(text: '0');
  final _descriptionController = TextEditingController();

  String _selectedType = 'imovel';
  DateTime? _purchaseDate;
  bool _isSubmitting = false;
  bool _isEditing = false;

  final _assetTypes = [
    {'value': 'imovel', 'label': 'Imóvel'},
    {'value': 'veiculo', 'label': 'Veículo'},
    {'value': 'renda_fixa', 'label': 'Renda Fixa'},
    {'value': 'previdencia', 'label': 'Previdência'},
    {'value': 'consorcio', 'label': 'Consórcio'},
    {'value': 'outro', 'label': 'Outro'},
  ];

  @override
  void initState() {
    super.initState();
    if (widget.customAssetId != null) {
      _isEditing = true;
      _loadExistingData();
    }
  }

  void _loadExistingData() {
    final state = ref.read(investmentsProvider);
    final asset = state.customAssets.cast<CustomAssetModel?>().firstWhere(
          (a) => a?.id == widget.customAssetId,
          orElse: () => null,
        );
    if (asset != null) {
      _nameController.text = asset.name;
      _selectedType = asset.type;
      _currentValueController.text = asset.currentValue.toStringAsFixed(2);
      if (asset.purchaseValue != null) {
        _purchaseValueController.text =
            asset.purchaseValue!.toStringAsFixed(2);
      }
      _purchaseDate = asset.purchaseDate;
      _monthlyIncomeController.text = asset.monthlyIncome.toStringAsFixed(2);
      if (asset.description != null) {
        _descriptionController.text = asset.description!;
      }
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _currentValueController.dispose();
    _purchaseValueController.dispose();
    _monthlyIncomeController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  Future<void> _selectDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _purchaseDate ?? DateTime.now(),
      firstDate: DateTime(1950),
      lastDate: DateTime.now(),
    );
    if (picked != null) {
      setState(() => _purchaseDate = picked);
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    final currentValue = double.tryParse(
        _currentValueController.text.replaceAll(',', '.'));
    final purchaseValue = _purchaseValueController.text.isNotEmpty
        ? double.tryParse(
            _purchaseValueController.text.replaceAll(',', '.'))
        : null;
    final monthlyIncome = double.tryParse(
            _monthlyIncomeController.text.replaceAll(',', '.')) ??
        0.0;

    final payload = <String, dynamic>{
      'name': _nameController.text.trim(),
      'type': _selectedType,
      'current_value': currentValue,
      'purchase_value': purchaseValue,
      if (_purchaseDate != null)
        'purchase_date': _purchaseDate!.toUtc().toIso8601String(),
      'monthly_income': monthlyIncome,
      if (_descriptionController.text.isNotEmpty)
        'description': _descriptionController.text.trim(),
    };

    setState(() => _isSubmitting = true);

    bool success;
    if (_isEditing && widget.customAssetId != null) {
      success = await ref
          .read(investmentsProvider.notifier)
          .updateCustomAsset(widget.customAssetId!, payload);
    } else {
      final created = await ref
          .read(investmentsProvider.notifier)
          .createCustomAsset(payload);
      success = created != null;
    }

    setState(() => _isSubmitting = false);

    if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(_isEditing
              ? 'Ativo atualizado com sucesso!'
              : 'Ativo criado com sucesso!'),
        ),
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
    final dateFormat = DateFormat('dd/MM/yyyy', 'pt_BR');

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditing ? 'Editar Ativo' : 'Novo Ativo Personalizado'),
      ),
      body: Form(
        key: _formKey,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Name
              const Text('Nome',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(
                  hintText: 'Ex: Apartamento Centro, Tesouro Direto...',
                  border: OutlineInputBorder(),
                ),
                validator: (v) {
                  if (v == null || v.trim().isEmpty) {
                    return 'Informe o nome do ativo';
                  }
                  if (v.trim().length < 2) {
                    return 'Nome deve ter pelo menos 2 caracteres';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),

              // Type
              const Text('Tipo',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              DropdownButtonFormField<String>(
                initialValue: _selectedType,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                ),
                items: _assetTypes
                    .map((t) => DropdownMenuItem<String>(
                          value: t['value'],
                          child: Text(t['label']!),
                        ))
                    .toList(),
                onChanged: (v) {
                  if (v != null) setState(() => _selectedType = v);
                },
              ),
              const SizedBox(height: 16),

              // Current value
              const Text('Valor Atual (R\$)',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              TextFormField(
                controller: _currentValueController,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  hintText: '0,00',
                  prefixText: 'R\$ ',
                ),
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                validator: (v) {
                  if (v == null || v.isEmpty) {
                    return 'Informe o valor atual';
                  }
                  final parsed = double.tryParse(v.replaceAll(',', '.'));
                  if (parsed == null) return 'Valor inválido';
                  if (parsed < 0) return 'Valor não pode ser negativo';
                  return null;
                },
              ),
              const SizedBox(height: 16),

              // Purchase value
              const Text('Valor de Compra (opcional)',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              TextFormField(
                controller: _purchaseValueController,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  hintText: '0,00',
                  prefixText: 'R\$ ',
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

              // Purchase date
              const Text('Data de Compra (opcional)',
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
                      Text(
                        _purchaseDate != null
                            ? dateFormat.format(_purchaseDate!)
                            : 'Selecione a data',
                        style: TextStyle(
                          color: _purchaseDate != null
                              ? Colors.black87
                              : Colors.grey,
                        ),
                      ),
                      const Spacer(),
                      if (_purchaseDate != null)
                        GestureDetector(
                          onTap: () => setState(() => _purchaseDate = null),
                          child: const Icon(Icons.clear,
                              size: 18, color: Colors.grey),
                        ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),

              // Monthly income
              const Text('Renda Mensal (R\$)',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              TextFormField(
                controller: _monthlyIncomeController,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  hintText: '0,00',
                  prefixText: 'R\$ ',
                  helperText: 'Aluguel, juros, rendimentos mensais...',
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

              // Description
              const Text('Descrição (opcional)',
                  style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              TextFormField(
                controller: _descriptionController,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  hintText: 'Detalhes sobre o ativo...',
                ),
                maxLines: 3,
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
                    : Text(
                        _isEditing ? 'Atualizar Ativo' : 'Criar Ativo',
                        style: const TextStyle(fontSize: 16),
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
