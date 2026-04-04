import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

const String _baseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8000',
);

const _storage = FlutterSecureStorage();

/// Global navigator key for showing SnackBars outside widget tree.
final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

void _showSnackBar(String message, {bool isError = true}) {
  final context = navigatorKey.currentContext;
  if (context != null) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: isError ? Colors.red : null,
      ),
    );
  }
}

Dio createDio() {
  final dio = Dio(
    BaseOptions(
      baseUrl: '$_baseUrl/api/v1',
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
      headers: {'Content-Type': 'application/json'},
    ),
  );

  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _storage.read(key: 'access_token');
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onResponse: (response, handler) {
        handler.next(response);
      },
      onError: (error, handler) {
        final statusCode = error.response?.statusCode;
        if (statusCode == 401) {
          // Clear tokens and redirect to login
          _storage.delete(key: 'access_token');
          _storage.delete(key: 'refresh_token');
          navigatorKey.currentState?.pushNamedAndRemoveUntil(
            '/login',
            (route) => false,
          );
        } else if (statusCode == 402) {
          _showSnackBar('Faça upgrade do seu plano para acessar esta funcionalidade');
        } else if (statusCode != null && statusCode >= 400 && statusCode < 500) {
          final errorData = error.response?.data;
          String message = 'Erro na requisição';
          if (errorData is Map) {
            final errorObj = errorData['error'];
            if (errorObj is Map) {
              message = errorObj['message'] as String? ?? message;
            } else if (errorData['message'] is String) {
              message = errorData['message'] as String;
            }
          }
          _showSnackBar(message);
        } else if (statusCode != null && statusCode >= 500) {
          _showSnackBar('Erro interno. Tente novamente.');
        }
        handler.next(error);
      },
    ),
  );

  return dio;
}

final dio = createDio();
