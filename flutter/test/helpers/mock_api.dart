import 'package:dio/dio.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

typedef MockApiHandler = Response Function(RequestOptions options);

/// Dio interceptor that routes by `METHOD path` (path may be absolute or relative).
/// Unmatched routes fail hard so tests stay explicit.
class MockApi {
  MockApi() {
    _interceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        final key = '${options.method.toUpperCase()} ${_normalize(options.path)}';
        for (final entry in _routes.entries) {
          if (entry.key.hasMatch(key)) {
            try {
              handler.resolve(entry.value(options));
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
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked $key',
            type: DioExceptionType.badResponse,
          ),
        );
      },
    );
  }

  late final Interceptor _interceptor;
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
    ApiClient.instance.dio.interceptors
      ..clear()
      ..add(_interceptor);
  }

  void uninstall() {
    ApiClient.instance.dio.interceptors.remove(_interceptor);
    ApiClient.instance.loadToken();
  }

  Response ok(RequestOptions options, Object? data, {int status = 200}) {
    final payload = data is Map && (data as Map).containsKey('data')
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
  }) {
    return Response(
      requestOptions: options,
      statusCode: status,
      data: {
        'error': {'code': code, 'message': message, 'msgKey': code},
      },
    );
  }
}
