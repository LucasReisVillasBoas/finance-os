import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/goal_model.dart';

class GoalRepository {
  final Dio _dio;

  GoalRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<List<GoalModel>> list() async {
    final response = await _dio.get('/goals');
    final body = response.data as Map<String, dynamic>;
    final dataList = body['data'] as List<dynamic>;
    return dataList
        .map((e) => GoalModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<GoalModel> create(Map<String, dynamic> payload) async {
    final response = await _dio.post('/goals', data: payload);
    final body = response.data as Map<String, dynamic>;
    return GoalModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<GoalModel> update(String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/goals/$id', data: payload);
    final body = response.data as Map<String, dynamic>;
    return GoalModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<void> delete(String id) async {
    await _dio.delete('/goals/$id');
  }

  Future<void> contribute(String id, Map<String, dynamic> payload) async {
    await _dio.post('/goals/$id/contribute', data: payload);
  }

  Future<List<GoalProjectionModel>> getProjections() async {
    final response = await _dio.get('/goals/projections');
    final body = response.data as Map<String, dynamic>;
    final dataList = body['data'] as List<dynamic>;
    return dataList
        .map((e) => GoalProjectionModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
