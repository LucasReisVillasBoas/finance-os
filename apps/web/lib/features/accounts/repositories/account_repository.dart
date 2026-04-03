import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/account_model.dart';

class AccountRepository {
  final Dio _dio;

  AccountRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<List<AccountModel>> getAll() async {
    final response = await _dio.get('/accounts');
    final data = response.data as Map<String, dynamic>;
    final list = data['data'] as List<dynamic>;
    return list
        .map((e) => AccountModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<AccountModel> getById(String id) async {
    final response = await _dio.get('/accounts/$id');
    final data = response.data as Map<String, dynamic>;
    return AccountModel.fromJson(data['data'] as Map<String, dynamic>);
  }

  Future<AccountModel> create(Map<String, dynamic> payload) async {
    final response = await _dio.post('/accounts', data: payload);
    final data = response.data as Map<String, dynamic>;
    return AccountModel.fromJson(data['data'] as Map<String, dynamic>);
  }

  Future<AccountModel> update(String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/accounts/$id', data: payload);
    final data = response.data as Map<String, dynamic>;
    return AccountModel.fromJson(data['data'] as Map<String, dynamic>);
  }

  Future<void> delete(String id) async {
    await _dio.delete('/accounts/$id');
  }

  Future<Map<String, dynamic>> getSummary() async {
    final response = await _dio.get('/accounts/summary');
    final data = response.data as Map<String, dynamic>;
    return data['data'] as Map<String, dynamic>;
  }
}
