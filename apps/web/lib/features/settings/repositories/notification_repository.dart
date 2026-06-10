import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/notification_model.dart';

class NotificationRepository {
  final Dio _dio;

  NotificationRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<List<NotificationModel>> getAll() async {
    final response = await _dio.get('/notifications');
    final data = (response.data['data'] as List<dynamic>?) ?? [];
    return data
        .map((e) => NotificationModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> markAsRead(String id) async {
    await _dio.put('/notifications/$id/read');
  }

  Future<void> markAllAsRead() async {
    await _dio.put('/notifications/read-all');
  }

  Future<void> deleteAll() async {
    await _dio.delete('/notifications');
  }
}
