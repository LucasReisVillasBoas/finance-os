import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/goal_model.dart';
import '../repositories/goal_repository.dart';

class GoalsState {
  final List<GoalModel> goals;
  final List<GoalProjectionModel> projections;
  final bool isLoading;
  final String? error;

  const GoalsState({
    this.goals = const [],
    this.projections = const [],
    this.isLoading = false,
    this.error,
  });

  GoalsState copyWith({
    List<GoalModel>? goals,
    List<GoalProjectionModel>? projections,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      GoalsState(
        goals: goals ?? this.goals,
        projections: projections ?? this.projections,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );
}

final goalRepositoryProvider = Provider<GoalRepository>((ref) {
  return GoalRepository();
});

class GoalsNotifier extends StateNotifier<GoalsState> {
  GoalsNotifier(this._repo) : super(const GoalsState());

  final GoalRepository _repo;

  Future<void> load() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final results = await Future.wait([
        _repo.list(),
        _repo.getProjections(),
      ]);
      state = state.copyWith(
        goals: results[0] as List<GoalModel>,
        projections: results[1] as List<GoalProjectionModel>,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<GoalModel?> create(Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final goal = await _repo.create(data);
      await load();
      return goal;
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
        goals: state.goals.where((g) => g.id != id).toList(),
        isLoading: false,
      );
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<bool> contribute(String goalId, double amount, DateTime date, {String? notes}) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final payload = <String, dynamic>{
        'amount': amount,
        'date': date.toIso8601String(),
        if (notes != null && notes.isNotEmpty) 'notes': notes,
      };
      await _repo.contribute(goalId, payload);
      await load();
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

final goalsProvider = StateNotifierProvider<GoalsNotifier, GoalsState>((ref) {
  final repo = ref.watch(goalRepositoryProvider);
  return GoalsNotifier(repo);
});
