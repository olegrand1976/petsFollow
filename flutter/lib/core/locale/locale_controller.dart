import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class LocaleController extends ChangeNotifier {
  LocaleController._();
  static final instance = LocaleController._();

  static const _prefKey = 'pf_locale';
  Locale _locale = const Locale('fr');

  Locale get locale => _locale;
  String get languageCode => _locale.languageCode;

  Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    final code = prefs.getString(_prefKey);
    if (code != null && _isSupported(code)) {
      _locale = Locale(code);
      notifyListeners();
    }
  }

  Future<void> setLocale(String code) async {
    if (!_isSupported(code)) return;
    _locale = Locale(code);
    notifyListeners();
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_prefKey, code);
  }

  bool _isSupported(String code) => code == 'fr' || code == 'nl' || code == 'en';
}
