import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

class GoalContributeDialog extends StatefulWidget {
  final String goalId;
  final String goalName;
  final Future<void> Function(double amount, DateTime date, String? notes) onConfirm;

  const GoalContributeDialog({
    super.key,
    required this.goalId,
    required this.goalName,
    required this.onConfirm,
  });

  @override
  State<GoalContributeDialog> createState() => _GoalContributeDialogState();
}

class _GoalContributeDialogState extends State<GoalContributeDialog> {
  final _formKey = GlobalKey<FormState>();
  final _amountController = TextEditingController();
  final _notesController = TextEditingController();
  DateTime _selectedDate = DateTime.now();
  bool _isLoading = false;

  @override
  void dispose() {
    _amountController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
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

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    final amount = double.tryParse(
      _amountController.text.replaceAll(',', '.'),
    );
    if (amount == null || amount <= 0) return;

    setState(() => _isLoading = true);

    try {
      await widget.onConfirm(
        amount,
        _selectedDate,
        _notesController.text.isNotEmpty ? _notesController.text : null,
      );
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('Aportar em "${widget.goalName}"'),
      content: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextFormField(
              controller: _amountController,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              autofocus: true,
              decoration: const InputDecoration(
                labelText: 'Valor do Aporte (R\$) *',
                hintText: '0,00',
                prefixText: 'R\$ ',
                border: OutlineInputBorder(),
              ),
              validator: (v) {
                if (v == null || v.isEmpty) return 'Informe o valor';
                final amount = double.tryParse(v.replaceAll(',', '.'));
                if (amount == null || amount <= 0) return 'Valor inválido';
                return null;
              },
            ),
            const SizedBox(height: 12),
            ListTile(
              contentPadding: EdgeInsets.zero,
              title: Text(
                'Data: ${DateFormat('dd/MM/yyyy').format(_selectedDate)}',
              ),
              leading: const Icon(Icons.calendar_today),
              onTap: _pickDate,
            ),
            const SizedBox(height: 4),
            TextFormField(
              controller: _notesController,
              decoration: const InputDecoration(
                labelText: 'Observações (opcional)',
                border: OutlineInputBorder(),
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isLoading ? null : () => Navigator.of(context).pop(),
          child: const Text('Cancelar'),
        ),
        ElevatedButton(
          onPressed: _isLoading ? null : _submit,
          child: _isLoading
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Text('Confirmar'),
        ),
      ],
    );
  }
}
