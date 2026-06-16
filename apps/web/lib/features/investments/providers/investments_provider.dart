import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/portfolio_model.dart';
import '../models/holding_model.dart';
import '../models/investment_transaction_model.dart';
import '../models/custom_asset_model.dart';
import '../models/asset_model.dart';
import '../models/currency_quote_model.dart';
import '../repositories/investment_repository.dart';

class InvestmentsState {
  final List<PortfolioModel> portfolios;
  final List<HoldingModel> holdings;
  final List<InvestmentTransactionModel> transactions;
  final List<CustomAssetModel> customAssets;
  final List<AssetModel> searchResults;
  final List<CurrencyQuoteModel> currencyQuotes;
  final bool isLoading;
  final String? error;
  final String? selectedPortfolioId;

  const InvestmentsState({
    this.portfolios = const [],
    this.holdings = const [],
    this.transactions = const [],
    this.customAssets = const [],
    this.searchResults = const [],
    this.currencyQuotes = const [],
    this.isLoading = false,
    this.error,
    this.selectedPortfolioId,
  });

  InvestmentsState copyWith({
    List<PortfolioModel>? portfolios,
    List<HoldingModel>? holdings,
    List<InvestmentTransactionModel>? transactions,
    List<CustomAssetModel>? customAssets,
    List<AssetModel>? searchResults,
    List<CurrencyQuoteModel>? currencyQuotes,
    bool? isLoading,
    String? error,
    bool clearError = false,
    String? selectedPortfolioId,
  }) =>
      InvestmentsState(
        portfolios: portfolios ?? this.portfolios,
        holdings: holdings ?? this.holdings,
        transactions: transactions ?? this.transactions,
        customAssets: customAssets ?? this.customAssets,
        searchResults: searchResults ?? this.searchResults,
        currencyQuotes: currencyQuotes ?? this.currencyQuotes,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
        selectedPortfolioId: selectedPortfolioId ?? this.selectedPortfolioId,
      );

  double get totalInvested =>
      holdings.fold(0, (sum, h) => sum + h.totalInvested);

  double get totalCurrentValue =>
      holdings.fold(0, (sum, h) => sum + h.currentValue);

  double get totalUnrealizedPnl =>
      holdings.fold(0, (sum, h) => sum + h.unrealizedPnl);

  double get totalRealizedPnl =>
      holdings.fold(0, (sum, h) => sum + h.realizedPnl);
}

final investmentRepositoryProvider = Provider<InvestmentRepository>((ref) {
  return InvestmentRepository();
});

class InvestmentsNotifier extends StateNotifier<InvestmentsState> {
  InvestmentsNotifier(this._repo) : super(const InvestmentsState());

  final InvestmentRepository _repo;

  Future<void> loadPortfolios() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final portfolios = await _repo.listPortfolios();
      state = state.copyWith(portfolios: portfolios, isLoading: false);
      if (portfolios.isNotEmpty) {
        final defaultPortfolio = portfolios.firstWhere(
          (p) => p.isDefault,
          orElse: () => portfolios.first,
        );
        await loadHoldings(defaultPortfolio.id);
      }
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<void> loadHoldings(String portfolioId) async {
    state = state.copyWith(isLoading: true, clearError: true,
        selectedPortfolioId: portfolioId);
    try {
      final holdings = await _repo.listHoldings(portfolioId);
      state = state.copyWith(holdings: holdings, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<PortfolioModel?> createPortfolio(Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final portfolio = await _repo.createPortfolio(data);
      await loadPortfolios();
      return portfolio;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return null;
    }
  }

  Future<bool> updatePortfolio(String id, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.updatePortfolio(id, data);
      await loadPortfolios();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<bool> deletePortfolio(String id) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.deletePortfolio(id);
      await loadPortfolios();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<HoldingModel?> createHolding(
      String portfolioId, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final holding = await _repo.createHolding(portfolioId, data);
      if (state.selectedPortfolioId != null) {
        await loadHoldings(state.selectedPortfolioId!);
      }
      return holding;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return null;
    }
  }

  Future<bool> deleteHolding(String id) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.deleteHolding(id);
      state = state.copyWith(
        holdings: state.holdings.where((h) => h.id != id).toList(),
        isLoading: false,
      );
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<void> loadTransactions(String holdingId) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final transactions = await _repo.listTransactions(holdingId);
      state = state.copyWith(transactions: transactions, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<InvestmentTransactionModel?> createTransaction(
      String holdingId, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final tx = await _repo.createTransaction(holdingId, data);
      await loadTransactions(holdingId);
      if (state.selectedPortfolioId != null) {
        await loadHoldings(state.selectedPortfolioId!);
      }
      return tx;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return null;
    }
  }

  Future<bool> deleteTransaction(String id, String holdingId) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.deleteTransaction(id);
      await loadTransactions(holdingId);
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<void> searchAssets(String query) async {
    if (query.isEmpty) {
      state = state.copyWith(searchResults: []);
      return;
    }
    try {
      final results = await _repo.searchAssets(query);
      state = state.copyWith(searchResults: results);
    } catch (e) {
      state = state.copyWith(searchResults: []);
    }
  }

  /// Loads USD/EUR currency quotes. Failures are silent so they never block
  /// the rest of the portfolio screen.
  Future<void> loadCurrencyQuotes() async {
    try {
      final quotes = await _repo.getCurrencyQuotes();
      state = state.copyWith(currencyQuotes: quotes);
    } catch (_) {
      // Keep any previously loaded quotes; quotes are non-critical.
    }
  }

  Future<void> loadCustomAssets() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final customAssets = await _repo.listCustomAssets();
      state = state.copyWith(customAssets: customAssets, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
    }
  }

  Future<CustomAssetModel?> createCustomAsset(
      Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final asset = await _repo.createCustomAsset(data);
      await loadCustomAssets();
      return asset;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return null;
    }
  }

  Future<bool> updateCustomAsset(String id, Map<String, dynamic> data) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.updateCustomAsset(id, data);
      await loadCustomAssets();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _extractError(e));
      return false;
    }
  }

  Future<bool> deleteCustomAsset(String id) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.deleteCustomAsset(id);
      state = state.copyWith(
        customAssets:
            state.customAssets.where((a) => a.id != id).toList(),
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

final investmentsProvider =
    StateNotifierProvider<InvestmentsNotifier, InvestmentsState>((ref) {
  final repo = ref.watch(investmentRepositoryProvider);
  return InvestmentsNotifier(repo);
});
