import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/dashboard_model.dart';
import '../repositories/dashboard_repository.dart';

class DashboardState {
  final DashboardOverview? overview;
  final List<MonthlyCashflowModel> cashflow;
  final bool isLoading;
  final String? error;
  final int month;
  final int year;

  const DashboardState({
    this.overview,
    this.cashflow = const [],
    this.isLoading = false,
    this.error,
    required this.month,
    required this.year,
  });

  DashboardState copyWith({
    DashboardOverview? overview,
    List<MonthlyCashflowModel>? cashflow,
    bool? isLoading,
    String? error,
    bool clearError = false,
    int? month,
    int? year,
  }) =>
      DashboardState(
        overview: overview ?? this.overview,
        cashflow: cashflow ?? this.cashflow,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
        month: month ?? this.month,
        year: year ?? this.year,
      );
}

final dashboardRepositoryProvider = Provider<DashboardRepository>((ref) {
  return DashboardRepository();
});

class DashboardNotifier extends StateNotifier<DashboardState> {
  DashboardNotifier(this._repo)
      : super(DashboardState(
          month: DateTime.now().month,
          year: DateTime.now().year,
        ));

  final DashboardRepository _repo;

  Future<void> load({int? month, int? year}) async {
    final m = month ?? state.month;
    final y = year ?? state.year;

    state = state.copyWith(isLoading: true, clearError: true, month: m, year: y);
    try {
      final results = await Future.wait([
        _repo.getOverview(month: m, year: y),
        _repo.getCashflow(),
      ]);
      state = state.copyWith(
        overview: results[0] as DashboardOverview,
        cashflow: results[1] as List<MonthlyCashflowModel>,
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

  String _extractError(Object e) {
    if (e is Exception) {
      return e.toString().replaceFirst('Exception: ', '');
    }
    return e.toString();
  }
}

final dashboardProvider =
    StateNotifierProvider<DashboardNotifier, DashboardState>((ref) {
  final repo = ref.watch(dashboardRepositoryProvider);
  return DashboardNotifier(repo);
});
