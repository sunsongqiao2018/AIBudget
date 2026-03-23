import 'dart:convert';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:http/http.dart' as http;

/// Base URL for edge-api.
/// On web (browser), use localhost. On Android emulator, 10.0.2.2 maps to host localhost.
final String kBaseUrl = kIsWeb ? 'http://localhost:8080' : 'http://10.0.2.2:8080';

class ApiClient {
  ApiClient({String? baseUrl, http.Client? httpClient})
      : _baseUrl = baseUrl ?? kBaseUrl,
        _client = httpClient ?? http.Client();

  final String _baseUrl;
  final http.Client _client;

  /// Performs a POST request to [path] with a JSON [body].
  ///
  /// Throws [ApiException] on non-2xx responses or network errors.
  Future<Map<String, dynamic>> post(
    String path,
    Map<String, dynamic> body,
  ) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await _client.post(
      uri,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: jsonEncode(body),
    );

    final decoded = jsonDecode(response.body) as Map<String, dynamic>;

    if (response.statusCode >= 200 && response.statusCode < 300) {
      return decoded;
    }

    final message = decoded['error'] as String? ?? 'Unknown error';
    throw ApiException(statusCode: response.statusCode, message: message);
  }

  void dispose() => _client.close();
}

class ApiException implements Exception {
  const ApiException({required this.statusCode, required this.message});

  final int statusCode;
  final String message;

  @override
  String toString() => 'ApiException($statusCode): $message';
}
