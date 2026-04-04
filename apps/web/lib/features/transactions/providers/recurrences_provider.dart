import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/recurrence_model.dart';
import '../repositories/recurrence_repository.dart';

class RecurrencesState {
  final List<RecurrenceModel> items;
  final bool isLoading;
  final String? error;

  const RecurrencesState({
    this.items = const [],
    this.isLoading = false,
    this.error,
  });

  RecurrencesState copyWith({
    List<RecurrenceModel>? items,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      RecurrencesState(
        items: items ?? this.items,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );
}

final recurrenceRepositoryProvider = Provider<RecurrenceRepository>((ref) {
  return RecurrenceRepository();
});

class RecurrencesNotifier extends StateNotifier<RecurrencesState> {
  RecurrencesNotifier(this._repo) : super(const RecurrencesState());

  final RecurrenceRepository _repo;

  Future<void> load() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final items = await _repo.list();
      state = state.copyWith(items: items, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<RecurrenceModel?> create(Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final rec = await _repo.create(data);
      state = state.copyWith(
        items: [rec, ...state.items],
        isLoading: false,
      );
      return rec;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return null;
    }
  }

  Future<bool> update(String id, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final updated = await _repo.update(id, data);
      state = state.copyWith(
        items: state.items.map((r) => r.id == id ? updated : r).toList(),
        isLoading: false,
      );
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
        items: state.items.where((r) => r.id != id).toList(),
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

final recurrencesProvider =
    StateNotifierProvider<RecurrencesNotifier, RecurrencesState>((ref) {
  final repo = ref.watch(recurrenceRepositoryProvider);
  return RecurrencesNotifier(repo);
});
