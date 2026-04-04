import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/portfolio_model.dart';
import '../models/holding_model.dart';
import '../models/investment_transaction_model.dart';
import '../models/custom_asset_model.dart';
import '../models/asset_model.dart';

class InvestmentRepository {
  final Dio _dio;

  InvestmentRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  // ---- Portfolios ----

  Future<List<PortfolioModel>> listPortfolios() async {
    final response = await _dio.get('/portfolios');
    final body = response.data as Map<String, dynamic>;
    final dataList = (body['data'] as List<dynamic>?) ?? [];
    return dataList
        .map((e) => PortfolioModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<PortfolioModel> createPortfolio(Map<String, dynamic> payload) async {
    final response = await _dio.post('/portfolios', data: payload);
    final body = response.data as Map<String, dynamic>;
    return PortfolioModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<PortfolioModel> updatePortfolio(
      String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/portfolios/$id', data: payload);
    final body = response.data as Map<String, dynamic>;
    return PortfolioModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<void> deletePortfolio(String id) async {
    await _dio.delete('/portfolios/$id');
  }

  // ---- Holdings ----

  Future<List<HoldingModel>> listHoldings(String portfolioId) async {
    final response = await _dio.get('/portfolios/$portfolioId/holdings');
    final body = response.data as Map<String, dynamic>;
    final dataList = (body['data'] as List<dynamic>?) ?? [];
    return dataList
        .map((e) => HoldingModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<HoldingModel> createHolding(
      String portfolioId, Map<String, dynamic> payload) async {
    final response =
        await _dio.post('/portfolios/$portfolioId/holdings', data: payload);
    final body = response.data as Map<String, dynamic>;
    return HoldingModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<HoldingModel> updateHolding(
      String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/holdings/$id', data: payload);
    final body = response.data as Map<String, dynamic>;
    return HoldingModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<void> deleteHolding(String id) async {
    await _dio.delete('/holdings/$id');
  }

  // ---- Investment Transactions ----

  Future<List<InvestmentTransactionModel>> listTransactions(
      String holdingId) async {
    final response = await _dio.get('/holdings/$holdingId/transactions');
    final body = response.data as Map<String, dynamic>;
    final dataList = (body['data'] as List<dynamic>?) ?? [];
    return dataList
        .map((e) =>
            InvestmentTransactionModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<InvestmentTransactionModel> createTransaction(
      String holdingId, Map<String, dynamic> payload) async {
    final response =
        await _dio.post('/holdings/$holdingId/transactions', data: payload);
    final body = response.data as Map<String, dynamic>;
    return InvestmentTransactionModel.fromJson(
        body['data'] as Map<String, dynamic>);
  }

  Future<void> deleteTransaction(String id) async {
    await _dio.delete('/investment-transactions/$id');
  }

  // ---- Assets ----

  Future<List<AssetModel>> searchAssets(String query) async {
    final response =
        await _dio.get('/assets/search', queryParameters: {'q': query});
    final body = response.data as Map<String, dynamic>;
    final dataList = (body['data'] as List<dynamic>?) ?? [];
    return dataList
        .map((e) => AssetModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  // ---- Custom Assets ----

  Future<List<CustomAssetModel>> listCustomAssets() async {
    final response = await _dio.get('/custom-assets');
    final body = response.data as Map<String, dynamic>;
    final dataList = (body['data'] as List<dynamic>?) ?? [];
    return dataList
        .map((e) => CustomAssetModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<CustomAssetModel> createCustomAsset(
      Map<String, dynamic> payload) async {
    final response = await _dio.post('/custom-assets', data: payload);
    final body = response.data as Map<String, dynamic>;
    return CustomAssetModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<CustomAssetModel> updateCustomAsset(
      String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/custom-assets/$id', data: payload);
    final body = response.data as Map<String, dynamic>;
    return CustomAssetModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<void> deleteCustomAsset(String id) async {
    await _dio.delete('/custom-assets/$id');
  }
}
