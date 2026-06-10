import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/recurrence_model.dart';

class RecurrenceRepository {
  final Dio _dio;

  RecurrenceRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<List<RecurrenceModel>> list() async {
    final response = await _dio.get('/recurrences');
    final body = response.data as Map<String, dynamic>;
    final dataList = ((body['data'] as List<dynamic>?) ?? []);
    return dataList
        .map((e) => RecurrenceModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<RecurrenceModel> create(Map<String, dynamic> payload) async {
    final response = await _dio.post('/recurrences', data: payload);
    final body = response.data as Map<String, dynamic>;
    return RecurrenceModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<RecurrenceModel> update(String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/recurrences/$id', data: payload);
    final body = response.data as Map<String, dynamic>;
    return RecurrenceModel.fromJson(body['data'] as Map<String, dynamic>);
  }

  Future<void> delete(String id) async {
    await _dio.delete('/recurrences/$id');
  }
}
