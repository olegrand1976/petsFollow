import 'package:flutter/services.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Flux Google partagé entre les écrans de login et d'inscription :
/// Google Sign-In → POST /auth/google (create-if-absent côté API).
abstract final class GoogleLoginFlow {
  /// Retourne la réponse de login (peut contenir `requires2FA`/`mfaToken`),
  /// ou `null` si l'utilisateur a annulé le sélecteur Google.
  ///
  /// [consent] est requis par l'API pour créer un compte client inconnu (RGPD).
  static Future<Map<String, dynamic>?> signIn({bool consent = false}) async {
    final idToken = await GoogleAuth.signInForIdToken();
    if (idToken == null) return null;
    return ApiClient.instance.loginWithGoogle(idToken, consent: consent);
  }

  static String errorMessage(AppLocalizations l10n, Object e) {
    switch (apiErrorCode(e)) {
      case 'not_configured':
        return l10n.googleNotConfigured;
      case 'google_client_only':
      case 'google_pro_only':
        return l10n.googleWrongAudience;
      case 'email_not_verified':
        return l10n.emailNotVerified;
      case 'consent_required':
        return l10n.registerConsentRequired;
      default:
        break;
    }
    if (e is PlatformException) {
      final code = e.code.toLowerCase();
      final details = '${e.message ?? ''} ${e.details ?? ''}'.toLowerCase();
      // Play Services: DEVELOPER_ERROR / status 10 = SHA / OAuth Android mal configuré.
      if (code.contains('sign_in_failed') ||
          details.contains('developer_error') ||
          details.contains('status{statuscode=10') ||
          RegExp(r'\b10\b').hasMatch(details)) {
        return l10n.googleNotConfigured;
      }
      if (code == 'network_error' || details.contains('network')) {
        return l10n.googleLoginFailed;
      }
    }
    if (e is StateError && e.message.contains('idToken missing')) {
      return l10n.googleNotConfigured;
    }
    return l10n.googleLoginFailed;
  }
}
