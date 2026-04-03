import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/account_model.dart';
import '../repositories/account_repository.dart';

class AccountsState {
  final List<AccountModel> accounts;
  final bool isLoading;
  final String? error;

  const AccountsState({
    this.accounts = const [],
    this.isLoading = false,
    this.error,
  });

  AccountsState copyWith({
    List<AccountModel>? accounts,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      AccountsState(
        accounts: accounts ?? this.accounts,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );

  double get totalBalance =>
      accounts.fold(0.0, (sum, a) => sum + a.balance);
}

final accountRepositoryProvider = Provider<AccountRepository>((ref) {
  return AccountRepository();
});

class AccountsNotifier extends StateNotifier<AccountsState> {
  AccountsNotifier(this._repo) : super(const AccountsState());

  final AccountRepository _repo;

  Future<void> loadAccounts() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final accounts = await _repo.getAll();
      state = state.copyWith(accounts: accounts, isLoading: false);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: _extractError(e),
      );
    }
  }

  Future<void> createAccount(Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final account = await _repo.create(data);
      state = state.copyWith(
        accounts: [...state.accounts, account],
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<void> updateAccount(String id, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final updated = await _repo.update(id, data);
      state = state.copyWith(
        accounts: state.accounts
            .map((a) => a.id == id ? updated : a)
            .toList(),
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<void> deleteAccount(String id) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.delete(id);
      state = state.copyWith(
        accounts: state.accounts.where((a) => a.id != id).toList(),
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  String _extractError(Object e) {
    if (e is Exception) {
      return e.toString().replaceFirst('Exception: ', '');
    }
    return e.toString();
  }
}

final accountsProvider =
    StateNotifierProvider<AccountsNotifier, AccountsState>((ref) {
  final repo = ref.watch(accountRepositoryProvider);
  return AccountsNotifier(repo);
});
