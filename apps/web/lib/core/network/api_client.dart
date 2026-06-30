import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart' show defaultTargetPlatform, kIsWeb, TargetPlatform;
import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'base_url.dart';

const _envUrl = String.fromEnvironment('API_BASE_URL');

// Resolves the API base URL at runtime:
//   1. If API_BASE_URL was provided at build time (--dart-define), use it.
//   2. Otherwise fall back to a platform-aware default so Android emulators
//      reach the host machine via 10.0.2.2 instead of the loopback address.
String get _baseUrl => _envUrl.isNotEmpty ? _envUrl : getDefaultBaseUrl();

const _storage = FlutterSecureStorage();

/// Global navigator key for showing SnackBars outside widget tree.
final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

/// Invoked when the interceptor gives up on a 401 (refresh failed or unavailable).
/// The auth provider wires this up so it can clear its own state and let the
/// router redirect to /login.
void Function()? onUnauthorized;

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

// Shared in-flight refresh — prevents N concurrent 401s from triggering N
// refreshes (which would rotate the refresh_token N times and break it).
Future<String?>? _refreshInFlight;

Future<String?> _refreshAccessToken(Dio originalDio) {
  return _refreshInFlight ??= _doRefresh(originalDio).whenComplete(() {
    _refreshInFlight = null;
  });
}

Future<String?> _doRefresh(Dio originalDio) async {
  final refreshToken = await _storage.read(key: 'refresh_token');
  if (refreshToken == null) return null;
  // Bare Dio so the refresh call doesn't recurse through our interceptor.
  final bare = Dio(BaseOptions(
    baseUrl: originalDio.options.baseUrl,
    connectTimeout: originalDio.options.connectTimeout,
    receiveTimeout: originalDio.options.receiveTimeout,
    headers: {'Content-Type': 'application/json'},
  ));
  try {
    final resp = await bare.post('/auth/refresh', data: {'refresh_token': refreshToken});
    final data = (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
    final newAccess = data['access_token'] as String;
    final newRefresh = data['refresh_token'] as String;
    await Future.wait([
      _storage.write(key: 'access_token', value: newAccess),
      _storage.write(key: 'refresh_token', value: newRefresh),
    ]);
    return newAccess;
  } catch (_) {
    return null;
  }
}

Future<void> _clearTokens() => Future.wait([
      _storage.delete(key: 'access_token'),
      _storage.delete(key: 'refresh_token'),
    ]);

bool _isAuthEndpoint(String path) =>
    path.contains('/auth/login') ||
    path.contains('/auth/refresh') ||
    path.contains('/auth/register');

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
        }
        handler.next(options);
      },
      onResponse: (response, handler) {
        handler.next(response);
      },
      onError: (error, handler) async {
        final statusCode = error.response?.statusCode;
        final requestOptions = error.requestOptions;
        final isAuthEndpoint = _isAuthEndpoint(requestOptions.path);
        final alreadyRetried = requestOptions.extra['_retriedAfterRefresh'] == true;

        if (statusCode == 401 && !isAuthEndpoint && !alreadyRetried) {
          final newToken = await _refreshAccessToken(dio);
          if (newToken != null) {
            requestOptions.headers['Authorization'] = 'Bearer $newToken';
            requestOptions.extra['_retriedAfterRefresh'] = true;
            try {
              final retryResp = await dio.fetch(requestOptions);
              return handler.resolve(retryResp);
            } on DioException catch (e) {
              return handler.next(e);
            }
          }
          await _clearTokens();
          onUnauthorized?.call();
          return handler.next(error);
        }

        if (statusCode == 401) {
          await _clearTokens();
          if (!isAuthEndpoint) onUnauthorized?.call();
          return handler.next(error);
        }

        if (statusCode == 402) {
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
