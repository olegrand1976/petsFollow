import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Flux Google partagé entre les écrans de login et d'inscription :
/// Google Sign-In → POST /auth/google (create-if-absent côté API).
abstract final class GoogleLoginFlow {
  /// Retourne la réponse de login (peut contenir `requires2FA`/`mfaToken`),
  /// ou `null` si l'utilisateur a annulé le sélecteur Google.
  static Future<Map<String, dynamic>?> signIn() async {
    final idToken = await GoogleAuth.signInForIdToken();
    if (idToken == null) return null;
    return ApiClient.instance.loginWithGoogle(idToken);
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
      default:
        return l10n.googleLoginFailed;
    }
  }
}
