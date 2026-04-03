import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/category_model.dart';
import '../repositories/category_repository.dart';

class CategoriesState {
  final List<CategoryModel> categories;
  final bool isLoading;
  final String? error;

  const CategoriesState({
    this.categories = const [],
    this.isLoading = false,
    this.error,
  });

  CategoriesState copyWith({
    List<CategoryModel>? categories,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      CategoriesState(
        categories: categories ?? this.categories,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );

  List<CategoryModel> byType(String type) =>
      categories.where((c) => c.type == type).toList();
}

final categoryRepositoryProvider = Provider<CategoryRepository>((ref) {
  return CategoryRepository();
});

class CategoriesNotifier extends StateNotifier<CategoriesState> {
  CategoriesNotifier(this._repo) : super(const CategoriesState());

  final CategoryRepository _repo;

  Future<void> loadCategories() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final categories = await _repo.getAll();
      state = state.copyWith(categories: categories, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<void> createCategory(Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final category = await _repo.create(data);
      state = state.copyWith(
        categories: [...state.categories, category],
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<void> updateCategory(String id, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final updated = await _repo.update(id, data);
      state = state.copyWith(
        categories: state.categories
            .map((c) => c.id == id ? updated : c)
            .toList(),
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<void> deleteCategory(String id) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.delete(id);
      state = state.copyWith(
        categories: state.categories.where((c) => c.id != id).toList(),
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

final categoriesProvider =
    StateNotifierProvider<CategoriesNotifier, CategoriesState>((ref) {
  final repo = ref.watch(categoryRepositoryProvider);
  return CategoriesNotifier(repo);
});
