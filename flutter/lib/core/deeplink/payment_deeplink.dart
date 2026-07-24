import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/invite/invite_code_store.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/features/auth/presentation/reset_password_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// App deep links: `petsfollow://payment/…`, `petsfollow://reset-password?token=`,
/// `petsfollow://invite?code=`.
///
/// Password-reset emails point at the Pro web URL (`/reset-password?token=`).
/// The custom-scheme reset link is kept for manual/testing and future App Links.
class AppDeepLink {
  AppDeepLink._();
  static final instance = AppDeepLink._();

  StreamSubscription<Uri>? _sub;
  final _appLinks = AppLinks();
  VoidCallback? onPaymentSuccess;
  VoidCallback? onPaymentCancel;

  String? _lastHandledKey;
  DateTime? _lastHandledAt;

  Future<void> start() async {
    try {
      final initial = await _appLinks.getInitialLink();
      if (initial != null) {
        _handle(initial);
      }
    } catch (e) {
      debugPrint('app deeplink initial: $e');
    }
    await _sub?.cancel();
    _sub = _appLinks.uriLinkStream.listen(_handle, onError: (Object e) {
      debugPrint('app deeplink stream: $e');
    });
  }

  void dispose() {
    _sub?.cancel();
    _sub = null;
  }

  void _handle(Uri uri) {
    if (uri.scheme != 'petsfollow') return;

    final key = uri.toString();
    final now = DateTime.now();
    if (_lastHandledKey == key &&
        _lastHandledAt != null &&
        now.difference(_lastHandledAt!) < const Duration(seconds: 2)) {
      return;
    }
    _lastHandledKey = key;
    _lastHandledAt = now;

    var path = uri.host.isNotEmpty
        ? '${uri.host}${uri.path}'
        : uri.path.replaceFirst(RegExp(r'^/+'), '');
    path = path.replaceFirst(RegExp(r'/+$'), '');

    if (path == 'payment/success' || path.endsWith('payment/success')) {
      onPaymentSuccess?.call();
      _snack((l10n) => l10n.paymentSuccessSnack);
      return;
    }
    if (path == 'payment/cancel' || path.endsWith('payment/cancel')) {
      onPaymentCancel?.call();
      _snack((l10n) => l10n.paymentCancelSnack);
      return;
    }
    if (path == 'reset-password' || path.endsWith('reset-password')) {
      final token = uri.queryParameters['token'] ?? '';
      if (token.isEmpty) return;
      _openResetPassword(token);
      return;
    }
    if (path == 'invite' || path.endsWith('invite')) {
      final code = uri.queryParameters['code'] ?? '';
      if (code.trim().isEmpty) return;
      unawaited(_persistAndMaybeClaim(code));
    }
  }

  Future<void> _persistAndMaybeClaim(String code) async {
    await InviteCodeStore.instance.save(code);
    final token = ApiClient.instance.token;
    if (token == null || token.isEmpty) return;
    try {
      await ApiClient.instance.claimVetInvite(code);
      await InviteCodeStore.instance.save(null);
    } catch (e) {
      debugPrint('deeplink claim invite: $e');
    }
  }

  void _openResetPassword(String token, [int attempt = 0]) {
    final nav = PushNavigation.instance.navigatorKey.currentState;
    if (nav == null) {
      if (attempt >= 12) return;
      Future<void>.delayed(const Duration(milliseconds: 400), () {
        _openResetPassword(token, attempt + 1);
      });
      return;
    }
    nav.push(
      MaterialPageRoute<void>(
        builder: (_) => ResetPasswordScreen(initialToken: token),
      ),
    );
  }

  void _snack(String Function(AppLocalizations l10n) message) {
    final ctx = PushNavigation.instance.navigatorKey.currentContext;
    if (ctx == null) return;
    final l10n = AppLocalizations.of(ctx);
    if (l10n == null) return;
    ScaffoldMessenger.maybeOf(ctx)?.showSnackBar(SnackBar(content: Text(message(l10n))));
  }
}

/// @Deprecated('Use AppDeepLink') — kept as typedef for any external refs.
typedef PaymentDeepLink = AppDeepLink;
