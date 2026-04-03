import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/transaction_model.dart';

class TransactionFilter {
  final String? startDate;
  final String? endDate;
  final String? categoryId;
  final String? accountId;
  final String? type;
  final String? search;
  final int page;
  final int pageSize;

  const TransactionFilter({
    this.startDate,
    this.endDate,
    this.categoryId,
    this.accountId,
    this.type,
    this.search,
    this.page = 1,
    this.pageSize = 20,
  });

  Map<String, dynamic> toQueryParams() {
    final params = <String, dynamic>{
      'page': page,
      'page_size': pageSize,
    };
    if (startDate != null) params['start_date'] = startDate;
    if (endDate != null) params['end_date'] = endDate;
    if (categoryId != null) params['category_id'] = categoryId;
    if (accountId != null) params['account_id'] = accountId;
    if (type != null) params['type'] = type;
    if (search != null && search!.isNotEmpty) params['search'] = search;
    return params;
  }
}

class TransactionListResult {
  final List<TransactionModel> transactions;
  final int total;
  final int page;
  final int pageSize;

  const TransactionListResult({
    required this.transactions,
    required this.total,
    required this.page,
    required this.pageSize,
  });
}

class TransactionRepository {
  final Dio _dio;

  TransactionRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<TransactionListResult> list({TransactionFilter? filter}) async {
    final f = filter ?? const TransactionFilter();
    final response = await _dio.get(
      '/transactions',
      queryParameters: f.toQueryParams(),
    );
    final body = response.data as Map<String, dynamic>;
    final meta = body['meta'] as Map<String, dynamic>;
    final dataList = body['data'] as List<dynamic>;
    final transactions = dataList
        .map((e) => TransactionModel.fromJson(e as Map<String, dynamic>))
        .toList();
    return TransactionListResult(
      transactions: transactions,
      total: meta['total'] as int? ?? 0,
      page: meta['page'] as int? ?? 1,
      pageSize: meta['page_size'] as int? ?? 20,
    );
  }

  Future<TransactionModel> getById(String id) async {
    final response = await _dio.get('/transactions/$id');
    final body = response.data as Map<String, dynamic>;
    return TransactionModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<TransactionModel> create(Map<String, dynamic> payload) async {
    final response = await _dio.post('/transactions', data: payload);
    final body = response.data as Map<String, dynamic>;
    return TransactionModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<TransactionModel> update(String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/transactions/$id', data: payload);
    final body = response.data as Map<String, dynamic>;
    return TransactionModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<void> delete(String id) async {
    await _dio.delete('/transactions/$id');
  }

  Future<Map<String, dynamic>> getSummary({
    String? startDate,
    String? endDate,
  }) async {
    final params = <String, dynamic>{};
    if (startDate != null) params['start_date'] = startDate;
    if (endDate != null) params['end_date'] = endDate;
    final response = await _dio.get(
      '/transactions/summary',
      queryParameters: params.isNotEmpty ? params : null,
    );
    final body = response.data as Map<String, dynamic>;
    return body['data'] as Map<String, dynamic>;
  }

  Future<List<TransactionModel>> createTransfer(Map<String, dynamic> payload) async {
    final response = await _dio.post('/transactions/transfer', data: payload);
    final body = response.data as Map<String, dynamic>;
    final dataList = body['data'] as List<dynamic>;
    return dataList
        .map((e) => TransactionModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
