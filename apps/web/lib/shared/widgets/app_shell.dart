import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class AppShell extends StatefulWidget {
  final Widget child;
  final String location;

  const AppShell({super.key, required this.child, required this.location});

  @override
  State<AppShell> createState() => _AppShellState();
}

class _AppShellState extends State<AppShell> {
  int _currentIndex = 0;

  // Returns -1 for routes without a dedicated tab (investments, goals, notifications, etc.)
  // -1 is a sentinel — the tab bar should not change highlight for these routes.
  static int _indexFromLocation(String location) {
    if (location.startsWith('/home')) return 0;
    if (location.startsWith('/transactions')) return 1;
    if (location.startsWith('/accounts')) return 2;
    if (location.startsWith('/budgets')) return 3;
    if (location.startsWith('/settings')) return 4;
    return -1;
  }

  @override
  void initState() {
    super.initState();
    final idx = _indexFromLocation(widget.location);
    _currentIndex = idx == -1 ? 0 : idx;
  }

  @override
  void didUpdateWidget(covariant AppShell oldWidget) {
    super.didUpdateWidget(oldWidget);
    final newIndex = _indexFromLocation(widget.location);

    // Only update the active tab when the new route has a dedicated tab.
    // Routes like /investments, /goals, /notifications keep the last known tab.
    if (newIndex != -1 && newIndex != _currentIndex) {
      setState(() {
        _currentIndex = newIndex;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: widget.child,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex,
        onTap: (index) {
          switch (index) {
            case 0:
              context.go('/home');
            case 1:
              context.go('/transactions');
            case 2:
              context.go('/accounts');
            case 3:
              context.go('/budgets');
            case 4:
              context.go('/settings');
          }
        },
        type: BottomNavigationBarType.fixed,
        selectedItemColor: Theme.of(context).colorScheme.primary,
        unselectedItemColor: Colors.grey,
        items: const [
          BottomNavigationBarItem(icon: Icon(Icons.home), label: 'Home'),
          BottomNavigationBarItem(
              icon: Icon(Icons.swap_horiz), label: 'Transações'),
          BottomNavigationBarItem(
              icon: Icon(Icons.account_balance_wallet), label: 'Contas'),
          BottomNavigationBarItem(
              icon: Icon(Icons.pie_chart), label: 'Orçamentos'),
          BottomNavigationBarItem(
              icon: Icon(Icons.settings), label: 'Config'),
        ],
      ),
    );
  }
}
