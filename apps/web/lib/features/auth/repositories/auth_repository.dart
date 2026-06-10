import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../models/user_model.dart';
import '../../../core/network/api_client.dart';

class AuthResponse {
  final String accessToken;
  final String refreshToken;
  final UserModel user;

  const AuthResponse({
    required this.accessToken,
    required this.refreshToken,
    required this.user,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    final data = json['data'] as Map<String, dynamic>;
    return AuthResponse(
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
      user: UserModel.fromJson(data['user'] as Map<String, dynamic>),
    );
  }
}

class AuthRepository {
  final Dio _dio;
  final FlutterSecureStorage _storage;

  AuthRepository({
    Dio? dioClient,
    FlutterSecureStorage? storage,
  })  : _dio = dioClient ?? dio,
        _storage = storage ?? const FlutterSecureStorage();

  Future<AuthResponse> register({
    required String name,
    required String email,
    required String password,
  }) async {
    final response = await _dio.post('/auth/register', data: {
      'name': name,
      'email': email,
      'password': password,
    });
    final authResp = AuthResponse.fromJson(response.data as Map<String, dynamic>);
    await _saveTokens(authResp.accessToken, authResp.refreshToken);
    return authResp;
  }

  Future<void> forgotPassword({required String email}) async {
    await _dio.post('/auth/forgot-password', data: {'email': email});
  }

  Future<AuthResponse> login({
    required String email,
    required String password,
  }) async {
    final response = await _dio.post('/auth/login', data: {
      'email': email,
      'password': password,
    });
    final authResp = AuthResponse.fromJson(response.data as Map<String, dynamic>);
    await _saveTokens(authResp.accessToken, authResp.refreshToken);
    return authResp;
  }

  Future<AuthResponse> refresh() async {
    final refreshToken = await _storage.read(key: 'refresh_token');
    if (refreshToken == null) {
      throw Exception('No refresh token stored');
    }
    final response = await _dio.post('/auth/refresh', data: {
      'refresh_token': refreshToken,
    });
    final authResp = AuthResponse.fromJson(response.data as Map<String, dynamic>);
    await _saveTokens(authResp.accessToken, authResp.refreshToken);
    return authResp;
  }

  Future<void> logout() async {
    final refreshToken = await _storage.read(key: 'refresh_token');
    if (refreshToken != null) {
      try {
        await _dio.post('/auth/logout', data: {'refresh_token': refreshToken});
      } catch (_) {
        // Best effort — still clear local tokens
      }
    }
    await _clearTokens();
  }

  Future<String?> getAccessToken() => _storage.read(key: 'access_token');

  Future<void> _saveTokens(String accessToken, String refreshToken) async {
    await Future.wait([
      _storage.write(key: 'access_token', value: accessToken),
      _storage.write(key: 'refresh_token', value: refreshToken),
    ]);
  }

  Future<void> _clearTokens() async {
    await Future.wait([
      _storage.delete(key: 'access_token'),
      _storage.delete(key: 'refresh_token'),
    ]);
  }
}
