import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../providers/goals_provider.dart';
import '../models/goal_model.dart';

class GoalFormScreen extends ConsumerStatefulWidget {
  final String? goalId;

  const GoalFormScreen({super.key, this.goalId});

  @override
  ConsumerState<GoalFormScreen> createState() => _GoalFormScreenState();
}

class _GoalFormScreenState extends ConsumerState<GoalFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _targetAmountController = TextEditingController();
  final _monthlyContributionController = TextEditingController();

  DateTime? _targetDate;
  String? _selectedIcon;
  Color? _selectedColor;

  bool _isLoading = false;

  GoalModel? _existingGoal;

  static const _icons = ['🎯', '🏠', '✈️', '🚗', '📱', '💰', '🎓', '💍', '🏖️', '💪'];

  static const _colors = [
    Colors.blue,
    Colors.green,
    Colors.orange,
    Colors.purple,
    Colors.red,
    Colors.teal,
    Colors.pink,
    Colors.amber,
  ];

  @override
  void initState() {
    super.initState();
    if (widget.goalId != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _loadExistingGoal();
      });
    }
  }

  void _loadExistingGoal() {
    final state = ref.read(goalsProvider);
    final goal = state.goals.where((g) => g.id == widget.goalId).firstOrNull;
    if (goal != null) {
      setState(() {
        _existingGoal = goal;
        _nameController.text = goal.name;
        _targetAmountController.text = goal.targetAmount.toStringAsFixed(2);
        if (goal.monthlyContribution != null) {
          _monthlyContributionController.text =
              goal.monthlyContribution!.toStringAsFixed(2);
        }
        _targetDate = goal.targetDate;
        _selectedIcon = goal.icon;
      });
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _targetAmountController.dispose();
    _monthlyContributionController.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _targetDate ?? DateTime.now().add(const Duration(days: 365)),
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365 * 30)),
    );
    if (picked != null) {
      setState(() => _targetDate = picked);
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    final targetAmount = double.tryParse(
      _targetAmountController.text.replaceAll(',', '.'),
    );
    final monthly = _monthlyContributionController.text.isNotEmpty
        ? double.tryParse(
            _monthlyContributionController.text.replaceAll(',', '.'),
          )
        : null;

    final data = <String, dynamic>{
      'name': _nameController.text.trim(),
      'target_amount': targetAmount,
      if (_targetDate != null)
        'target_date': _targetDate!.toUtc().toIso8601String(),
      'monthly_contribution': monthly,
      'icon': _selectedIcon,
      if (_selectedColor != null)
        'color':
            '#${_selectedColor!.toARGB32().toRadixString(16).padLeft(8, '0').substring(2)}',
    };

    bool success;
    if (_existingGoal != null) {
      success = await ref
          .read(goalsProvider.notifier)
          .update(_existingGoal!.id, data);
    } else {
      final result = await ref.read(goalsProvider.notifier).create(data);
      success = result != null;
    }

    setState(() => _isLoading = false);

    if (success && mounted) {
      context.pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    final isEditing = _existingGoal != null;

    return Scaffold(
      appBar: AppBar(
        title: Text(isEditing ? 'Editar Meta' : 'Nova Meta'),
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // Icon picker
            const Text('Ícone (opcional)', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              children: _icons.map((icon) {
                final selected = _selectedIcon == icon;
                return GestureDetector(
                  onTap: () => setState(() => _selectedIcon = selected ? null : icon),
                  child: Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      border: Border.all(
                        color: selected ? Colors.blue : Colors.grey.shade300,
                        width: selected ? 2 : 1,
                      ),
                      borderRadius: BorderRadius.circular(8),
                      color: selected ? Colors.blue.shade50 : null,
                    ),
                    child: Text(icon, style: const TextStyle(fontSize: 24)),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 16),

            // Color picker
            const Text('Cor', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              children: _colors.map((color) {
                final selected = _selectedColor == color;
                return GestureDetector(
                  onTap: () => setState(() => _selectedColor = selected ? null : color),
                  child: Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(
                      color: color,
                      shape: BoxShape.circle,
                      border: selected
                          ? Border.all(color: Colors.black, width: 3)
                          : null,
                    ),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 16),

            TextFormField(
              controller: _nameController,
              decoration: const InputDecoration(
                labelText: 'Nome da Meta *',
                hintText: 'Ex.: Viagem para Europa',
                border: OutlineInputBorder(),
              ),
              validator: (v) {
                if (v == null || v.trim().length < 2) {
                  return 'Nome deve ter ao menos 2 caracteres';
                }
                return null;
              },
            ),
            const SizedBox(height: 16),

            TextFormField(
              controller: _targetAmountController,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              decoration: const InputDecoration(
                labelText: 'Valor da Meta (R\$) *',
                hintText: '0,00',
                border: OutlineInputBorder(),
                prefixText: 'R\$ ',
              ),
              validator: (v) {
                if (v == null || v.isEmpty) return 'Informe o valor da meta';
                final amount = double.tryParse(v.replaceAll(',', '.'));
                if (amount == null || amount <= 0) return 'Valor inválido';
                return null;
              },
            ),
            const SizedBox(height: 16),

            TextFormField(
              controller: _monthlyContributionController,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              decoration: const InputDecoration(
                labelText: 'Aporte Mensal (R\$) — opcional',
                hintText: '0,00',
                border: OutlineInputBorder(),
                prefixText: 'R\$ ',
              ),
            ),
            const SizedBox(height: 16),

            // Target date
            ListTile(
              contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
              horizontalTitleGap: 12,
              title: Text(
                _targetDate != null
                    ? 'Data alvo: ${DateFormat('dd/MM/yyyy', 'pt_BR').format(_targetDate!)}'
                    : 'Data alvo (opcional)',
              ),
              leading: const Icon(Icons.calendar_today),
              trailing: _targetDate != null
                  ? IconButton(
                      icon: const Icon(Icons.clear),
                      onPressed: () => setState(() => _targetDate = null),
                    )
                  : null,
              onTap: _pickDate,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
                side: BorderSide(color: Colors.grey.shade300),
              ),
            ),
            const SizedBox(height: 24),

            SizedBox(
              height: 48,
              child: ElevatedButton(
                onPressed: _isLoading ? null : _submit,
                child: _isLoading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Text(isEditing ? 'Salvar' : 'Criar Meta'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
