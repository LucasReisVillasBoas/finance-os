import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/family_model.dart';
import '../repositories/family_repository.dart';

final familyRepositoryProvider = Provider<FamilyRepository>(
  (ref) => FamilyRepository(),
);

class FamilyState {
  final FamilyGroup? group;
  final List<FamilyMember> members;
  final bool isLoading;
  final String? error;

  const FamilyState({
    this.group,
    this.members = const [],
    this.isLoading = false,
    this.error,
  });

  FamilyState copyWith({
    FamilyGroup? group,
    List<FamilyMember>? members,
    bool? isLoading,
    String? error,
  }) {
    return FamilyState(
      group: group ?? this.group,
      members: members ?? this.members,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class FamilyNotifier extends StateNotifier<FamilyState> {
  final FamilyRepository _repo;

  FamilyNotifier(this._repo) : super(const FamilyState(isLoading: true)) {
    load();
  }

  Future<void> load() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final group = await _repo.getGroup();
      state = state.copyWith(group: group, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> createGroup(String name) async {
    state = state.copyWith(isLoading: true);
    try {
      final group = await _repo.createGroup(name);
      state = state.copyWith(group: group, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      rethrow;
    }
  }

  Future<String> getInviteCode() async {
    return _repo.getInviteCode();
  }

  Future<void> joinGroup(String inviteCode) async {
    state = state.copyWith(isLoading: true);
    try {
      final group = await _repo.joinGroup(inviteCode);
      state = state.copyWith(group: group, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      rethrow;
    }
  }

  Future<void> removeMember(String memberId) async {
    try {
      await _repo.removeMember(memberId);
      await load();
    } catch (e) {
      state = state.copyWith(error: e.toString());
      rethrow;
    }
  }
}

final familyProvider =
    StateNotifierProvider<FamilyNotifier, FamilyState>(
  (ref) => FamilyNotifier(ref.watch(familyRepositoryProvider)),
);
