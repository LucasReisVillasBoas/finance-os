import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/category_model.dart';

class CategoryRepository {
  final Dio _dio;

  CategoryRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<List<CategoryModel>> getAll() async {
    final response = await _dio.get('/categories');
    final data = response.data as Map<String, dynamic>;
    final list = data['data'] as List<dynamic>;
    return list
        .map((e) => CategoryModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<CategoryModel> create(Map<String, dynamic> payload) async {
    final response = await _dio.post('/categories', data: payload);
    final data = response.data as Map<String, dynamic>;
    return CategoryModel.fromJson(data['data'] as Map<String, dynamic>);
  }

  Future<CategoryModel> update(String id, Map<String, dynamic> payload) async {
    final response = await _dio.put('/categories/$id', data: payload);
    final data = response.data as Map<String, dynamic>;
    return CategoryModel.fromJson(data['data'] as Map<String, dynamic>);
  }

  Future<void> delete(String id) async {
    await _dio.delete('/categories/$id');
  }
}
