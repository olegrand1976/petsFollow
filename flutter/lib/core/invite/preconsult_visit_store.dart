import 'package:shared_preferences/shared_preferences.dart';

/// Persists a pending pre-consult visit id across cold start until opened after login.
class PreconsultVisitStore {
  PreconsultVisitStore._();
  static final instance = PreconsultVisitStore._();

  static const _key = 'pf_preconsult_visit_id';

  Future<void> save(String? visitId) async {
    final normalized = (visitId ?? '').trim();
    final prefs = await SharedPreferences.getInstance();
    if (normalized.isEmpty) {
      await prefs.remove(_key);
      return;
    }
    await prefs.setString(_key, normalized);
  }

  Future<String?> peek() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_key)?.trim() ?? '';
    return raw.isEmpty ? null : raw;
  }

  Future<String?> take() async {
    final id = await peek();
    if (id == null) return null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_key);
    return id;
  }
}
