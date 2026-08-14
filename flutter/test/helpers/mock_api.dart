import 'dart:async';

import 'package:dio/dio.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

typedef MockApiHandler = FutureOr<Response> Function(RequestOptions options);

/// Dio interceptor that routes by `METHOD path` (path may be absolute or relative).
/// Unmatched routes fail hard so tests stay explicit.
/// Responses with status ≥ 400 are rejected as [DioException] (real API behavior).
class MockApi {
  Interceptor? _interceptor;
  final Map<RegExp, MockApiHandler> _routes = {};

  static String _normalize(String path) {
    if (path.startsWith('http')) {
      final uri = Uri.parse(path);
      return uri.path;
    }
    return path;
  }

  void on(String method, Pattern pathPattern, MockApiHandler handler) {
    final path = pathPattern is RegExp
        ? pathPattern.pattern
        : RegExp.escape(pathPattern.toString());
    _routes[RegExp('^${method.toUpperCase()} $path\$')] = handler;
  }

  void json(
    String method,
    Pattern pathPattern, {
    int status = 200,
    Object? data,
  }) {
    on(method, pathPattern, (options) {
      return Response(
        requestOptions: options,
        statusCode: status,
        data: data is Map && data.containsKey('data')
            ? data
            : {'data': data},
      );
    });
  }

  void install() {
    _interceptor = InterceptorsWrapper(
      onRequest: (options, handler) async {
        final method = options.method.toUpperCase();
        final candidates = <String>{
          '$method ${_normalize(options.path)}',
          '$method ${_normalize(options.uri.path)}',
        };
        for (final entry in _routes.entries) {
          if (candidates.any(entry.key.hasMatch)) {
            try {
              final res = await entry.value(options);
              final status = res.statusCode ?? 200;
              // handler.resolve bypasses Dio validateStatus — reject 4xx/5xx
              // so callers get DioException like a real API.
              if (status >= 400) {
                handler.reject(
                  DioException(
                    requestOptions: options,
                    response: res,
                    type: DioExceptionType.badResponse,
                    message: 'Mock API $status',
                  ),
                );
              } else {
                handler.resolve(res);
              }
            } catch (e, st) {
              handler.reject(
                DioException(
                  requestOptions: options,
                  error: e,
                  stackTrace: st,
                ),
              );
            }
            return;
          }
        }
        final key = candidates.first;
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked $key',
            type: DioExceptionType.badResponse,
          ),
        );
      },
    );
    ApiClient.instance.dio.interceptors
      ..clear()
      ..add(_interceptor!);
  }

  void uninstall() {
    final interceptor = _interceptor;
    if (interceptor != null) {
      ApiClient.instance.dio.interceptors.remove(interceptor);
      _interceptor = null;
    }
    ApiClient.instance.loadToken();
  }

  Response ok(RequestOptions options, Object? data, {int status = 200}) {
    final payload = data is Map && data.containsKey('data')
        ? data
        : {'data': data};
    return Response(
      requestOptions: options,
      statusCode: status,
      data: payload,
    );
  }

  Response err(
    RequestOptions options, {
    int status = 400,
    String code = 'bad_request',
    String message = 'error',
    String? msgKey,
  }) {
    return Response(
      requestOptions: options,
      statusCode: status,
      data: {
        'error': {
          'code': code,
          'message': message,
          'msgKey': msgKey ?? code,
        },
      },
    );
  }
}
