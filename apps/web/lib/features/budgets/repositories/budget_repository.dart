import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/budget_model.dart';

class BudgetRepository {
  final Dio _dio;

  BudgetRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<List<BudgetModel>> list({int? month, int? year}) async {
    final params = <String, dynamic>{};
    if (month != null) params['month'] = month;
    if (year != null) params['year'] = year;

    final response = await _dio.get(
      '/budgets',
      queryParameters: params.isNotEmpty ? params : null,
    );
    final body = response.data as Map<String, dynamic>;
    final dataList = body['data'] as List<dynamic>;
    return dataList
        .map((e) => BudgetModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<BudgetModel> create(Map<String, dynamic> payload) async {
    final response = await _dio.post('/budgets', data: payload);
    final body = response.data as Map<String, dynamic>;
    return BudgetModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<BudgetModel> update(String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/budgets/$id', data: payload);
    final body = response.data as Map<String, dynamic>;
    return BudgetModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<void> delete(String id) async {
    await _dio.delete('/budgets/$id');
  }

  Future<List<BudgetProgressModel>> getProgress({int? month, int? year}) async {
    final params = <String, dynamic>{};
    if (month != null) params['month'] = month;
    if (year != null) params['year'] = year;

    final response = await _dio.get(
      '/budgets/progress',
      queryParameters: params.isNotEmpty ? params : null,
    );
    final body = response.data as Map<String, dynamic>;
    final dataList = body['data'] as List<dynamic>;
    return dataList
        .map((e) => BudgetProgressModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
