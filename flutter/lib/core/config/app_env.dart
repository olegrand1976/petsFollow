import 'package:flutter/foundation.dart';

/// Build-time environment — set via `--dart-define=FLAVOR=` and/or `APP_ENV=`.
///
/// Scripts always pass **both** with the same value (`staging` | `prod`) alongside
/// `--flavor`. Prefer `FLAVOR` when set; otherwise `APP_ENV`; default `staging`.
class AppEnv {
  AppEnv._();

  static const String _appEnv = String.fromEnvironment('APP_ENV');
  static const String _flavor = String.fromEnvironment('FLAVOR');

  /// Resolved canal (`staging` | `prod`).
  static String get value {
    if (_flavor.isNotEmpty) return _flavor;
    if (_appEnv.isNotEmpty) return _appEnv;
    return 'staging';
  }

  static bool get isStaging => value == 'staging';
  static bool get isProd => value == 'prod';

  static String get appTitle => isStaging ? 'petsFollow Staging' : 'petsFollow';

  /// Call once at process start (before Firebase init).
  /// Throws if `FLAVOR` and `APP_ENV` disagree, or if an unknown value is used.
  static void validate() {
    if (_flavor.isNotEmpty &&
        _appEnv.isNotEmpty &&
        _flavor != _appEnv) {
      throw StateError(
        'FLAVOR="$_flavor" != APP_ENV="$_appEnv" — pass the same value for both.',
      );
    }
    final v = value;
    if (v != 'staging' && v != 'prod') {
      throw StateError('APP_ENV/FLAVOR must be staging|prod (got: "$v")');
    }
    // Bare `flutter build --flavor prod` without dart-defines would default to
    // staging while packaging the store applicationId — refuse in release.
    if (kReleaseMode &&
        defaultTargetPlatform == TargetPlatform.android &&
        _flavor.isEmpty &&
        _appEnv.isEmpty) {
      throw StateError(
        'Release Android requires --dart-define=FLAVOR=staging|prod '
        '(and matching --flavor). Use make flutter-dev / firebase-android-dist / '
        'play-android-bundle.',
      );
    }
  }
}
