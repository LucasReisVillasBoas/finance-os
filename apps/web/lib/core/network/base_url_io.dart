import 'dart:io';

String getDefaultBaseUrl() {
  if (Platform.isAndroid) return 'http://10.0.2.2:8000';
  return 'http://localhost:8000';
}
