import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Handles Stripe return deep links (`petsfollow://payment/success|cancel`).
class PaymentDeepLink {
  PaymentDeepLink._();
  static final instance = PaymentDeepLink._();

  StreamSubscription<Uri>? _sub;
  final _appLinks = AppLinks();
  VoidCallback? onPaymentSuccess;
  VoidCallback? onPaymentCancel;

  Future<void> start() async {
    try {
      final initial = await _appLinks.getInitialLink();
      if (initial != null) {
        _handle(initial);
      }
    } catch (e) {
      debugPrint('payment deeplink initial: $e');
    }
    await _sub?.cancel();
    _sub = _appLinks.uriLinkStream.listen(_handle, onError: (Object e) {
      debugPrint('payment deeplink stream: $e');
    });
  }

  void dispose() {
    _sub?.cancel();
    _sub = null;
  }

  void _handle(Uri uri) {
    if (uri.scheme != 'petsfollow') return;
    final path = uri.host.isNotEmpty
        ? '${uri.host}${uri.path}'
        : uri.path.replaceFirst(RegExp(r'^/+'), '');
    if (path == 'payment/success' || path.endsWith('payment/success')) {
      onPaymentSuccess?.call();
      _snack((l10n) => l10n.paymentSuccessSnack);
      return;
    }
    if (path == 'payment/cancel' || path.endsWith('payment/cancel')) {
      onPaymentCancel?.call();
      _snack((l10n) => l10n.paymentCancelSnack);
    }
  }

  void _snack(String Function(AppLocalizations l10n) message) {
    final ctx = PushNavigation.instance.navigatorKey.currentContext;
    if (ctx == null) return;
    final l10n = AppLocalizations.of(ctx);
    if (l10n == null) return;
    ScaffoldMessenger.maybeOf(ctx)?.showSnackBar(SnackBar(content: Text(message(l10n))));
  }
}
