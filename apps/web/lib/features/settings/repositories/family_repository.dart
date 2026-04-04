import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../models/family_model.dart';

class FamilyRepository {
  final Dio _dio;

  FamilyRepository({Dio? dioClient}) : _dio = dioClient ?? dio;

  Future<FamilyGroup?> getGroup() async {
    try {
      final response = await _dio.get('/family');
      return FamilyGroup.fromJson(response.data['data'] as Map<String, dynamic>);
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<FamilyGroup> createGroup(String name) async {
    final response = await _dio.post('/family', data: {'name': name});
    return FamilyGroup.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<String> getInviteCode() async {
    final response = await _dio.post('/family/invite');
    return response.data['data']['invite_code'] as String;
  }

  Future<FamilyGroup> joinGroup(String inviteCode) async {
    final response =
        await _dio.post('/family/join', data: {'invite_code': inviteCode});
    return FamilyGroup.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<void> removeMember(String memberId) async {
    await _dio.delete('/family/members/$memberId');
  }

  Future<List<FamilyMember>> getMembers() async {
    final response = await _dio.get('/family');
    final data = response.data['data'];
    if (data is! Map) return [];
    final members = data['members'];
    if (members is! List) return [];
    return members
        .map((e) => FamilyMember.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
