import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/transaction_model.dart';
import '../repositories/transaction_repository.dart';

class TransactionsState {
  final List<TransactionModel> transactions;
  final bool isLoading;
  final String? error;
  final int total;
  final int page;
  final TransactionFilter filter;

  const TransactionsState({
    this.transactions = const [],
    this.isLoading = false,
    this.error,
    this.total = 0,
    this.page = 1,
    this.filter = const TransactionFilter(),
  });

  TransactionsState copyWith({
    List<TransactionModel>? transactions,
    bool? isLoading,
    String? error,
    bool clearError = false,
    int? total,
    int? page,
    TransactionFilter? filter,
  }) =>
      TransactionsState(
        transactions: transactions ?? this.transactions,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
        total: total ?? this.total,
        page: page ?? this.page,
        filter: filter ?? this.filter,
      );
}

final transactionRepositoryProvider = Provider<TransactionRepository>((ref) {
  return TransactionRepository();
});

class TransactionsNotifier extends StateNotifier<TransactionsState> {
  TransactionsNotifier(this._repo) : super(const TransactionsState());

  final TransactionRepository _repo;

  Future<void> loadTransactions({TransactionFilter? filter, bool reset = false}) async {
    final activeFilter = filter ?? state.filter;
    state = state.copyWith(
      isLoading: true,
      clearError: true,
      filter: activeFilter,
      page: reset ? 1 : state.page,
    );
    try {
      final result = await _repo.list(
        filter: reset
            ? TransactionFilter(
                startDate: activeFilter.startDate,
                endDate: activeFilter.endDate,
                categoryId: activeFilter.categoryId,
                accountId: activeFilter.accountId,
                type: activeFilter.type,
                search: activeFilter.search,
                page: 1,
                pageSize: activeFilter.pageSize,
              )
            : activeFilter,
      );
      state = state.copyWith(
        transactions: result.transactions,
        total: result.total,
        page: result.page,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: _extractError(e),
      );
    }
  }

  Future<void> applyFilter(TransactionFilter filter) async {
    await loadTransactions(filter: filter, reset: true);
  }

  Future<void> clearFilter() async {
    await loadTransactions(filter: const TransactionFilter(), reset: true);
  }

  Future<TransactionModel?> createTransaction(Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final tx = await _repo.create(data);
      state = state.copyWith(
        transactions: [tx, ...state.transactions],
        total: state.total + 1,
        isLoading: false,
      );
      return tx;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return null;
    }
  }

  Future<bool> updateTransaction(String id, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final updated = await _repo.update(id, data);
      state = state.copyWith(
        transactions: state.transactions
            .map((t) => t.id == id ? updated : t)
            .toList(),
        isLoading: false,
      );
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<bool> deleteTransaction(String id) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.delete(id);
      state = state.copyWith(
        transactions: state.transactions.where((t) => t.id != id).toList(),
        total: state.total > 0 ? state.total - 1 : 0,
        isLoading: false,
      );
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<List<TransactionModel>?> createTransfer(Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final txs = await _repo.createTransfer(data);
      // Reload to reflect balance changes
      await loadTransactions(reset: true);
      return txs;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return null;
    }
  }

  String _extractError(Object e) {
    if (e is Exception) {
      return e.toString().replaceFirst('Exception: ', '');
    }
    return e.toString();
  }
}

final transactionsProvider =
    StateNotifierProvider<TransactionsNotifier, TransactionsState>((ref) {
  final repo = ref.watch(transactionRepositoryProvider);
  return TransactionsNotifier(repo);
});
