import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../shared/providers/auth_provider.dart';
import '../models/dashboard_model.dart';
import '../repositories/dashboard_repository.dart';

class DashboardState {
  final DashboardOverview? overview;
  final List<MonthlyCashflowModel> cashflow;
  final List<PatrimonySnapshotModel> patrimonyHistory;
  final bool isLoading;
  final String? error;
  final int month;
  final int year;

  const DashboardState({
    this.overview,
    this.cashflow = const [],
    this.patrimonyHistory = const [],
    this.isLoading = false,
    this.error,
    required this.month,
    required this.year,
  });

  DashboardState copyWith({
    DashboardOverview? overview,
    List<MonthlyCashflowModel>? cashflow,
    List<PatrimonySnapshotModel>? patrimonyHistory,
    bool? isLoading,
    String? error,
    bool clearError = false,
    int? month,
    int? year,
  }) =>
      DashboardState(
        overview: overview ?? this.overview,
        cashflow: cashflow ?? this.cashflow,
        patrimonyHistory: patrimonyHistory ?? this.patrimonyHistory,
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
  DashboardNotifier(this._repo, this._ref)
      : super(DashboardState(
          month: DateTime.now().month,
          year: DateTime.now().year,
        ));

  final DashboardRepository _repo;
  final Ref _ref;

  Future<void> load({int? month, int? year}) async {
    final m = month ?? state.month;
    final y = year ?? state.year;

    state = state.copyWith(isLoading: true, clearError: true, month: m, year: y);
    try {
      final results = await Future.wait([
        _repo.getOverview(month: m, year: y),
        _repo.getCashflow(),
        _repo.getPatrimonyHistory(),
      ]);
      state = state.copyWith(
        overview: results[0] as DashboardOverview,
        cashflow: results[1] as List<MonthlyCashflowModel>,
        patrimonyHistory: results[2] as List<PatrimonySnapshotModel>,
        isLoading: false,
      );
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        await _ref.read(authProvider.notifier).logout();
        return;
      }
      state = state.copyWith(isLoading: false, error: _extractError(e));
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
    if (e is DioException) {
      final data = e.response?.data;
      if (data is Map) {
        final err = data['error'];
        if (err is Map && err['message'] != null) {
          return err['message'].toString();
        }
        if (data['message'] != null) return data['message'].toString();
      }
      if (e.type == DioExceptionType.connectionError ||
          e.type == DioExceptionType.connectionTimeout) {
        return 'Não foi possível conectar ao servidor.';
      }
      final status = e.response?.statusCode;
      return status != null ? 'Erro $status do servidor.' : 'Erro de rede.';
    }
    if (e is Exception) {
      return e.toString().replaceFirst('Exception: ', '');
    }
    return e.toString();
  }
}

final dashboardProvider =
    StateNotifierProvider<DashboardNotifier, DashboardState>((ref) {
  final repo = ref.watch(dashboardRepositoryProvider);
  return DashboardNotifier(repo, ref);
});
