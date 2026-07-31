import 'dart:collection';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';

const _window = Duration(minutes: 15);
const _maxConsole = 100;
const _maxNetwork = 50;

final _sensitiveQuery = RegExp(
  r'^(token|code|password|access_token|refresh_token|authorization)$',
  caseSensitive: false,
);

class SupportDiagnosticsBuffer {
  SupportDiagnosticsBuffer._();
  static final instance = SupportDiagnosticsBuffer._();

  final ListQueue<_ConsoleEntry> _console = ListQueue();
  final ListQueue<_NetworkEntry> _network = ListQueue();
  bool _installed = false;

  void install() {
    if (_installed) return;
    _installed = true;

    final previous = FlutterError.onError;
    FlutterError.onError = (details) {
      pushConsole(
        level: 'error',
        message: details.exceptionAsString(),
        stack: details.stack?.toString(),
      );
      previous?.call(details);
    };

    PlatformDispatcher.instance.onError = (error, stack) {
      pushConsole(
        level: 'unhandled',
        message: error.toString(),
        stack: stack.toString(),
      );
      return false;
    };
  }

  void pushConsole({
    required String level,
    required String message,
    String? stack,
  }) {
    _prune(_console);
    _console.addLast(_ConsoleEntry(
      ts: DateTime.now().toUtc().toIso8601String(),
      level: level,
      message: message.length > 2000 ? message.substring(0, 2000) : message,
      stack: stack == null
          ? null
          : (stack.length > 4000 ? stack.substring(0, 4000) : stack),
    ));
    while (_console.length > _maxConsole) {
      _console.removeFirst();
    }
  }

  void pushNetwork({
    required String method,
    required String url,
    int? status,
    int? durationMs,
    bool? ok,
  }) {
    _prune(_network);
    _network.addLast(_NetworkEntry(
      ts: DateTime.now().toUtc().toIso8601String(),
      method: method,
      url: _redactUrl(url),
      status: status,
      durationMs: durationMs,
      ok: ok,
    ));
    while (_network.length > _maxNetwork) {
      _network.removeFirst();
    }
  }

  Interceptor dioInterceptor() {
    return InterceptorsWrapper(
      onRequest: (options, handler) {
        options.extra['pf_support_started'] = DateTime.now().millisecondsSinceEpoch;
        handler.next(options);
      },
      onResponse: (response, handler) {
        final started = response.requestOptions.extra['pf_support_started'] as int?;
        final duration = started == null
            ? null
            : DateTime.now().millisecondsSinceEpoch - started;
        pushNetwork(
          method: response.requestOptions.method,
          url: response.requestOptions.uri.toString(),
          status: response.statusCode,
          durationMs: duration,
          ok: true,
        );
        handler.next(response);
      },
      onError: (error, handler) {
        final started = error.requestOptions.extra['pf_support_started'] as int?;
        final duration = started == null
            ? null
            : DateTime.now().millisecondsSinceEpoch - started;
        pushNetwork(
          method: error.requestOptions.method,
          url: error.requestOptions.uri.toString(),
          status: error.response?.statusCode,
          durationMs: duration,
          ok: false,
        );
        handler.next(error);
      },
    );
  }

  Map<String, dynamic> snapshot({String? route}) {
    _prune(_console);
    _prune(_network);
    return {
      'capturedAt': DateTime.now().toUtc().toIso8601String(),
      'windowMinutes': 15,
      'session': {
        'userId': ApiClient.instance.userId,
        'role': ApiClient.instance.userRole,
        'specialty': ApiClient.instance.userSpecialty,
      },
      'config': {
        'route': route,
        'locale': LocaleController.instance.languageCode,
        'appEnv': AppEnv.value,
        'platform': defaultTargetPlatform.name,
        'userAgent': 'petsfollow-flutter/${AppEnv.value}',
      },
      'consoleErrors': _console.map((e) => e.toJson()).toList(),
      'networkEntries': _network.map((e) => e.toJson()).toList(),
    };
  }

  void _prune<T extends _Timestamped>(ListQueue<T> q) {
    final cutoff = DateTime.now().toUtc().subtract(_window);
    while (q.isNotEmpty) {
      final ts = DateTime.tryParse(q.first.ts)?.toUtc();
      if (ts == null || !ts.isBefore(cutoff)) break;
      q.removeFirst();
    }
  }

  static String _redactUrl(String raw) {
    try {
      final u = Uri.parse(raw);
      final q = Map<String, String>.from(u.queryParameters);
      for (final key in q.keys.toList()) {
        if (_sensitiveQuery.hasMatch(key)) q[key] = '[redacted]';
      }
      return u.replace(queryParameters: q.isEmpty ? null : q).toString();
    } catch (_) {
      return raw.length > 500 ? raw.substring(0, 500) : raw;
    }
  }
}

abstract class _Timestamped {
  String get ts;
}

class _ConsoleEntry implements _Timestamped {
  _ConsoleEntry({
    required this.ts,
    required this.level,
    required this.message,
    this.stack,
  });
  @override
  final String ts;
  final String level;
  final String message;
  final String? stack;

  Map<String, dynamic> toJson() => {
        'ts': ts,
        'level': level,
        'message': message,
        if (stack != null) 'stack': stack,
      };
}

class _NetworkEntry implements _Timestamped {
  _NetworkEntry({
    required this.ts,
    required this.method,
    required this.url,
    this.status,
    this.durationMs,
    this.ok,
  });
  @override
  final String ts;
  final String method;
  final String url;
  final int? status;
  final int? durationMs;
  final bool? ok;

  Map<String, dynamic> toJson() => {
        'ts': ts,
        'method': method,
        'url': url,
        if (status != null) 'status': status,
        if (durationMs != null) 'durationMs': durationMs,
        if (ok != null) 'ok': ok,
      };
}
