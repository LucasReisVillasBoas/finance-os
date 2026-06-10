import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../features/auth/models/user_model.dart';
import '../../features/auth/repositories/auth_repository.dart';

class AuthState {
  final UserModel? user;
  final bool isLoading;
  final String? error;

  const AuthState({
    this.user,
    this.isLoading = false,
    this.error,
  });

  AuthState copyWith({
    UserModel? user,
    bool? isLoading,
    String? error,
    bool clearUser = false,
    bool clearError = false,
  }) {
    return AuthState(
      user: clearUser ? null : (user ?? this.user),
      isLoading: isLoading ?? this.isLoading,
      error: clearError ? null : (error ?? this.error),
    );
  }
}

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository();
});

class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier(this._repo) : super(const AuthState());

  final AuthRepository _repo;

  Future<void> login(String email, String password) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final resp = await _repo.login(email: email, password: password);
      state = state.copyWith(user: resp.user, isLoading: false);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: _extractError(e),
        clearUser: true,
      );
    }
  }

  Future<void> register(String name, String email, String password) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final resp =
          await _repo.register(name: name, email: email, password: password);
      state = state.copyWith(user: resp.user, isLoading: false);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: _extractError(e),
        clearUser: true,
      );
    }
  }

  Future<void> logout() async {
    state = state.copyWith(isLoading: true);
    try {
      await _repo.logout();
    } catch (_) {
      // Ignore logout errors — always clear local state
    }
    state = const AuthState();
  }

  Future<void> checkAuth() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final token = await _repo.getAccessToken();
      if (token == null) {
        state = const AuthState();
        return;
      }
      // Attempt a token refresh to validate and get fresh tokens + user data
      final resp = await _repo.refresh();
      state = state.copyWith(user: resp.user, isLoading: false);
    } catch (_) {
      // Token invalid or expired — clear session
      await _repo.logout();
      state = const AuthState();
    }
  }

  String _extractError(Object e) {
    if (e is Exception) {
      return e.toString().replaceFirst('Exception: ', '');
    }
    return e.toString();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  final repo = ref.watch(authRepositoryProvider);
  return AuthNotifier(repo);
});
