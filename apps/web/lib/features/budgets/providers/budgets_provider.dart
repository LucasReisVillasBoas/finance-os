import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/budget_model.dart';
import '../repositories/budget_repository.dart';

class BudgetsState {
  final List<BudgetModel> budgets;
  final List<BudgetProgressModel> progress;
  final bool isLoading;
  final String? error;
  final int month;
  final int year;

  BudgetsState({
    this.budgets = const [],
    this.progress = const [],
    this.isLoading = false,
    this.error,
    required this.month,
    required this.year,
  });

  BudgetsState copyWith({
    List<BudgetModel>? budgets,
    List<BudgetProgressModel>? progress,
    bool? isLoading,
    String? error,
    bool clearError = false,
    int? month,
    int? year,
  }) =>
      BudgetsState(
        budgets: budgets ?? this.budgets,
        progress: progress ?? this.progress,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
        month: month ?? this.month,
        year: year ?? this.year,
      );
}

final budgetRepositoryProvider = Provider<BudgetRepository>((ref) {
  return BudgetRepository();
});

class BudgetsNotifier extends StateNotifier<BudgetsState> {
  BudgetsNotifier(this._repo)
      : super(BudgetsState(
          month: DateTime.now().month,
          year: DateTime.now().year,
        ));

  final BudgetRepository _repo;

  Future<void> load({int? month, int? year}) async {
    final m = month ?? state.month;
    final y = year ?? state.year;

    state = state.copyWith(isLoading: true, clearError: true, month: m, year: y);
    try {
      final results = await Future.wait([
        _repo.list(month: m, year: y),
        _repo.getProgress(month: m, year: y),
      ]);
      state = state.copyWith(
        budgets: results[0] as List<BudgetModel>,
        progress: results[1] as List<BudgetProgressModel>,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  void changeMonth(int delta) {
    var month = state.month + delta;
    var year = state.year;
    if (month > 12) {
      month = 1;
      year += 1;
    } else if (month < 1) {
      month = 12;
      year -= 1;
    }
    load(month: month, year: year);
  }

  Future<BudgetModel?> create(Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final budget = await _repo.create(data);
      await load();
      return budget;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return null;
    }
  }

  Future<bool> update(String id, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.update(id, data);
      await load();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<bool> delete(String id) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.delete(id);
      state = state.copyWith(
        budgets: state.budgets.where((b) => b.id != id).toList(),
        progress: state.progress.where((p) => p.budgetId != id).toList(),
        isLoading: false,
      );
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  String _extractError(Object e) {
    if (e is Exception) {
      return e.toString().replaceFirst('Exception: ', '');
    }
    return e.toString();
  }
}

final budgetsProvider =
    StateNotifierProvider<BudgetsNotifier, BudgetsState>((ref) {
  final repo = ref.watch(budgetRepositoryProvider);
  return BudgetsNotifier(repo);
});
