import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class NotificationPreferencesScreen extends StatefulWidget {
  const NotificationPreferencesScreen({super.key});

  @override
  State<NotificationPreferencesScreen> createState() =>
      _NotificationPreferencesScreenState();
}

class _NotificationPreferencesScreenState
    extends State<NotificationPreferencesScreen> {
  bool _budgetAlert = true;
  bool _goalDeadline = true;
  bool _recurrenceDue = true;
  bool _weeklySummary = true;
  bool _monthlyReport = true;
  bool _loading = true;

  static const _keys = {
    'budget_alert': 'notif_budget_alert',
    'goal_deadline': 'notif_goal_deadline',
    'recurrence_due': 'notif_recurrence_due',
    'weekly_summary': 'notif_weekly_summary',
    'monthly_report': 'notif_monthly_report',
  };

  @override
  void initState() {
    super.initState();
    _loadPreferences();
  }

  Future<void> _loadPreferences() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() {
      _budgetAlert = prefs.getBool(_keys['budget_alert']!) ?? true;
      _goalDeadline = prefs.getBool(_keys['goal_deadline']!) ?? true;
      _recurrenceDue = prefs.getBool(_keys['recurrence_due']!) ?? true;
      _weeklySummary = prefs.getBool(_keys['weekly_summary']!) ?? true;
      _monthlyReport = prefs.getBool(_keys['monthly_report']!) ?? true;
      _loading = false;
    });
  }

  Future<void> _updatePref(String key, bool value) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(key, value);
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Preferências de Notificação')),
      body: ListView(
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Text(
              'Alertas',
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: Colors.grey,
              ),
            ),
          ),
          SwitchListTile(
            title: const Text('Alertas de orçamento'),
            subtitle:
                const Text('Avisa quando você está próximo do limite'),
            value: _budgetAlert,
            onChanged: (v) {
              setState(() => _budgetAlert = v);
              _updatePref(_keys['budget_alert']!, v);
            },
          ),
          SwitchListTile(
            title: const Text('Prazo de metas'),
            subtitle: const Text('Lembra quando uma meta está vencendo'),
            value: _goalDeadline,
            onChanged: (v) {
              setState(() => _goalDeadline = v);
              _updatePref(_keys['goal_deadline']!, v);
            },
          ),
          SwitchListTile(
            title: const Text('Contas recorrentes'),
            subtitle: const Text('Avisa sobre pagamentos recorrentes'),
            value: _recurrenceDue,
            onChanged: (v) {
              setState(() => _recurrenceDue = v);
              _updatePref(_keys['recurrence_due']!, v);
            },
          ),
          const Divider(),
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 8, 16, 8),
            child: Text(
              'Resumos',
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: Colors.grey,
              ),
            ),
          ),
          SwitchListTile(
            title: const Text('Resumo semanal'),
            subtitle: const Text('Receba um resumo todo domingo'),
            value: _weeklySummary,
            onChanged: (v) {
              setState(() => _weeklySummary = v);
              _updatePref(_keys['weekly_summary']!, v);
            },
          ),
          SwitchListTile(
            title: const Text('Relatório mensal'),
            subtitle: const Text('Receba um relatório no início do mês'),
            value: _monthlyReport,
            onChanged: (v) {
              setState(() => _monthlyReport = v);
              _updatePref(_keys['monthly_report']!, v);
            },
          ),
        ],
      ),
    );
  }
}
