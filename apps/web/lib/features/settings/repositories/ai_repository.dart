import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';

class AIRepository {
  final Dio _dio;

  AIRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<String> chat(String message) async {
    final response = await _dio.post('/ai/chat', data: {'message': message});
    return response.data['data']['response'] as String;
  }

  Future<Map<String, dynamic>> getSpendingForecast() async {
    final response = await _dio.get('/ai/spending-forecast');
    return response.data['data'] as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getPortfolioAnalysis() async {
    final response = await _dio.get('/ai/portfolio-analysis');
    return response.data['data'] as Map<String, dynamic>;
  }
}
