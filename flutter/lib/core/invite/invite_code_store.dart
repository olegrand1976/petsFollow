import 'package:shared_preferences/shared_preferences.dart';

/// Persists a vet app-invite code across install / cold start until claimed.
class InviteCodeStore {
  InviteCodeStore._();
  static final instance = InviteCodeStore._();

  static const _key = 'pf_vet_invite_code';

  Future<void> save(String? code) async {
    final normalized = (code ?? '').trim().toUpperCase();
    final prefs = await SharedPreferences.getInstance();
    if (normalized.isEmpty) {
      await prefs.remove(_key);
      return;
    }
    await prefs.setString(_key, normalized);
  }

  Future<String?> peek() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_key)?.trim().toUpperCase() ?? '';
    return raw.isEmpty ? null : raw;
  }

  Future<String?> take() async {
    final code = await peek();
    if (code == null) return null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_key);
    return code;
  }
}
