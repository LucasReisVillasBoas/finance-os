import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/dashboard_model.dart';

class DashboardRepository {
  final Dio _dio;

  DashboardRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<DashboardOverview> getOverview({int? month, int? year}) async {
    final params = <String, dynamic>{};
    if (month != null) params['month'] = month;
    if (year != null) params['year'] = year;

    final response = await _dio.get(
      '/dashboard/overview',
      queryParameters: params.isNotEmpty ? params : null,
    );
    final body = response.data as Map<String, dynamic>;
    return DashboardOverview.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<List<MonthlyCashflowModel>> getCashflow() async {
    final response = await _dio.get('/dashboard/cashflow');
    final body = response.data as Map<String, dynamic>;
    final dataList = ((body['data'] as List<dynamic>?) ?? []);
    return dataList
        .map((e) => MonthlyCashflowModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<PatrimonySnapshotModel>> getPatrimonyHistory() async {
    final response = await _dio.get('/dashboard/patrimony');
    final body = response.data as Map<String, dynamic>;
    final dataList = (body['data'] as List<dynamic>?) ?? [];
    return dataList
        .map((e) => PatrimonySnapshotModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
