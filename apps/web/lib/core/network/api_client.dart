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
        try {
          final token = await _storage
              .read(key: 'access_token')
              .timeout(const Duration(seconds: 3));
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
        } catch (_) {
          // Storage read failed or timed out — proceed without token.
          // API will return 401 which is handled below.
        }
        handler.next(options);
      },
      onResponse: (response, handler) {
        handler.next(response);
      },
      onError: (error, handler) {
        final statusCode = error.response?.statusCode;
        if (statusCode == 401) {
          _storage.delete(key: 'access_token');
          _storage.delete(key: 'refresh_token');
        } else if (statusCode == 402) {
          _showSnackBar('Faça upgrade do seu plano para acessar esta funcionalidade');
        } else if (statusCode != null && statusCode >= 400 && statusCode < 500) {
          final errorData = error.response?.data;
          String message = 'Erro na requisição';
          if (errorData is Map) {
            final errorObj = errorData['error'];
            if (errorObj is Map) {
              message = (errorObj['message'] as Object?)?.toString() ?? message;
            } else if (errorData['message'] != null) {
              message = errorData['message'].toString();
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
