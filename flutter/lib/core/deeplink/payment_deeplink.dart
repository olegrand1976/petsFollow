import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/auth/pending_login_hint.dart';
import 'package:petsfollow_mobile/core/invite/invite_code_store.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/features/auth/presentation/confirm_email_screen.dart';
import 'package:petsfollow_mobile/features/auth/presentation/reset_password_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Result of parsing a `petsfollow://confirm-email…` URI (testable, no navigation).
enum ConfirmEmailDeepLinkKind { token, statusOk, none }

class ConfirmEmailDeepLink {
  const ConfirmEmailDeepLink._(this.kind, {this.token, this.email});

  const ConfirmEmailDeepLink.token(String token)
      : this._(ConfirmEmailDeepLinkKind.token, token: token);

  const ConfirmEmailDeepLink.statusOk({String? email})
      : this._(ConfirmEmailDeepLinkKind.statusOk, email: email);

  const ConfirmEmailDeepLink.none() : this._(ConfirmEmailDeepLinkKind.none);

  final ConfirmEmailDeepLinkKind kind;
  final String? token;
  final String? email;
}

/// Shared path normalization for `petsfollow://host/path` and `petsfollow:///path`.
String appDeepLinkPath(Uri uri) {
  var path = uri.host.isNotEmpty
      ? '${uri.host}${uri.path}'
      : uri.path.replaceFirst(RegExp(r'^/+'), '');
  return path.replaceFirst(RegExp(r'/+$'), '');
}

ConfirmEmailDeepLink parseConfirmEmailDeepLink(Uri uri) {
  if (uri.scheme != 'petsfollow') return const ConfirmEmailDeepLink.none();
  final path = appDeepLinkPath(uri);
  if (path != 'confirm-email' && !path.endsWith('confirm-email')) {
    return const ConfirmEmailDeepLink.none();
  }
  final token = uri.queryParameters['token'] ?? '';
  if (token.isNotEmpty) return ConfirmEmailDeepLink.token(token);
  if ((uri.queryParameters['status'] ?? '') == 'ok') {
    return ConfirmEmailDeepLink.statusOk(email: uri.queryParameters['email']);
  }
  return const ConfirmEmailDeepLink.none();
}

/// App deep links: `petsfollow://payment/…`, `petsfollow://reset-password?token=`,
/// `petsfollow://confirm-email?token=` / `?status=ok`, `petsfollow://invite?code=`.
///
/// Auth emails point at the Pro web URL. Custom-scheme links cover app / testing
/// and the Nuxt « Ouvrir l'app » bridge after web confirm.
class AppDeepLink {
  AppDeepLink._();
  static final instance = AppDeepLink._();

  StreamSubscription<Uri>? _sub;
  final _appLinks = AppLinks();
  VoidCallback? onPaymentSuccess;
  VoidCallback? onPaymentCancel;

  VoidCallback? _onLoginHint;

  /// Rebuild login after [PendingLoginHint] was set (confirm-email?status=ok).
  /// Flushing a pending hint when the callback is attached covers cold-start races.
  VoidCallback? get onLoginHint => _onLoginHint;
  set onLoginHint(VoidCallback? value) {
    _onLoginHint = value;
    if (value != null && PendingLoginHint.hasPending) {
      value();
    }
  }

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

    final path = appDeepLinkPath(uri);

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
    final confirm = parseConfirmEmailDeepLink(uri);
    switch (confirm.kind) {
      case ConfirmEmailDeepLinkKind.token:
        _openConfirmEmail(confirm.token!);
        return;
      case ConfirmEmailDeepLinkKind.statusOk:
        _openConfirmEmailOk(confirm.email);
        return;
      case ConfirmEmailDeepLinkKind.none:
        break;
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

  void _openConfirmEmail(String token, [int attempt = 0]) {
    final nav = PushNavigation.instance.navigatorKey.currentState;
    if (nav == null) {
      if (attempt >= 12) return;
      Future<void>.delayed(const Duration(milliseconds: 400), () {
        _openConfirmEmail(token, attempt + 1);
      });
      return;
    }
    nav.push(
      MaterialPageRoute<void>(
        builder: (_) => ConfirmEmailScreen(token: token),
      ),
    );
  }

  void _openConfirmEmailOk(String? email, [int attempt = 0]) {
    final ctx = PushNavigation.instance.navigatorKey.currentContext;
    if (ctx == null) {
      if (attempt >= 12) return;
      Future<void>.delayed(const Duration(milliseconds: 400), () {
        _openConfirmEmailOk(email, attempt + 1);
      });
      return;
    }
    final l10n = AppLocalizations.of(ctx);
    PendingLoginHint.set(
      email: email,
      infoMessage: l10n?.confirmEmailDoneSubtitle,
    );
    PushNavigation.instance.navigatorKey.currentState?.popUntil((r) => r.isFirst);
    onLoginHint?.call();
    _snack((l) => l.confirmEmailDoneTitle);
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
