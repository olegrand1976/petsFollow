import 'package:dio/dio.dart';
import 'package:flutter/services.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Levée quand l'API exige le consentement CGU pour créer un compte client Google.
/// Contient le [idToken] déjà obtenu pour éviter un second passage Google.
class GoogleConsentRequired implements Exception {
  GoogleConsentRequired(this.idToken);
  final String idToken;
}

/// Flux Google partagé entre les écrans de login et d'inscription :
/// Google Sign-In → POST /auth/google (create-if-absent côté API).
abstract final class GoogleLoginFlow {
  /// Retourne la réponse de login (peut contenir `requires2FA`/`mfaToken`),
  /// ou `null` si l'utilisateur a annulé le sélecteur Google.
  ///
  /// [consent] est requis par l'API pour créer un compte client inconnu (RGPD).
  /// Sans consent et email inconnu → [GoogleConsentRequired].
  static Future<Map<String, dynamic>?> signIn({
    bool consent = false,
    String? commercialUserId,
  }) async {
    final idToken = await GoogleAuth.signInForIdToken();
    if (idToken == null) return null;
    return complete(idToken, consent: consent, commercialUserId: commercialUserId);
  }

  /// Finalise le login API avec un [idToken] déjà obtenu.
  static Future<Map<String, dynamic>> complete(
    String idToken, {
    bool consent = false,
    String? commercialUserId,
  }) async {
    try {
      return await ApiClient.instance.loginWithGoogle(
        idToken,
        consent: consent,
        commercialUserId: commercialUserId,
      );
    } catch (e) {
      if (!consent && isConsentRequired(e)) {
        throw GoogleConsentRequired(idToken);
      }
      rethrow;
    }
  }

  /// Détecte `consent_required` (code API) ou message localisé de repli.
  static bool isConsentRequired(Object e) {
    final code = apiErrorCode(e);
    if (code == 'consent_required') return true;
    if (e is! DioException) return false;
    final data = e.response?.data;
    if (data is! Map) return false;
    final err = data['error'];
    if (err is! Map) return false;
    final message = err['message']?.toString().toLowerCase() ?? '';
    return message.contains('confidentialité') ||
        message.contains('privacy policy') ||
        message.contains('privacybeleid') ||
        message.contains('privaatsuspoliitikaga') ||
        message.contains('política de privacidad') ||
        message.contains('politica de privacidad');
  }

  static String errorMessage(AppLocalizations l10n, Object e) {
    if (e is GoogleConsentRequired || isConsentRequired(e)) {
      return l10n.registerConsentRequired;
    }
    switch (apiErrorCode(e)) {
      case 'not_configured':
        return l10n.googleNotConfigured;
      case 'google_client_only':
      case 'google_pro_only':
        return l10n.googleWrongAudience;
      case 'email_not_verified':
        return l10n.emailNotVerified;
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
