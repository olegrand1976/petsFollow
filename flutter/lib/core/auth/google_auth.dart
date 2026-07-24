import 'package:google_sign_in/google_sign_in.dart';

/// Google Sign-In for pets clients (API Go validates the idToken).
///
/// [serverClientId] must match API `GOOGLE_OAUTH_CLIENT_ID` (Web OAuth client).
class GoogleAuth {
  GoogleAuth._();

  static const serverClientId = String.fromEnvironment(
    'GOOGLE_SERVER_CLIENT_ID',
    defaultValue: '',
  );

  static final GoogleSignIn _signIn = GoogleSignIn(
    scopes: const ['email', 'profile'],
    serverClientId: serverClientId.isEmpty ? null : serverClientId,
  );

  static bool get isConfigured => serverClientId.isNotEmpty;

  /// Returns a Google ID token, or null if the user cancelled.
  static Future<String?> signInForIdToken() async {
    if (!isConfigured) {
      throw StateError('GOOGLE_SERVER_CLIENT_ID not configured');
    }
    final account = await _signIn.signIn();
    if (account == null) return null;
    final auth = await account.authentication;
    final idToken = auth.idToken;
    if (idToken == null || idToken.isEmpty) {
      throw StateError('Google idToken missing — check Web client ID / SHA-1');
    }
    return idToken;
  }

  static Future<void> signOut() async {
    try {
      await _signIn.signOut();
    } catch (_) {}
  }
}
