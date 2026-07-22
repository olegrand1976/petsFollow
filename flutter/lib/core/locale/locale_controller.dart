import 'dart:ui' show PlatformDispatcher;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class LocaleController extends ChangeNotifier {
  LocaleController._();
  static final instance = LocaleController._();

  static const _prefKey = 'pf_locale';
  static const supportedCodes = ['fr', 'nl', 'en', 'es'];
  Locale _locale = const Locale('fr');

  Locale get locale => _locale;
  String get languageCode => _locale.languageCode;

  /// Loads saved locale, or silently adopts the OS language when supported.
  Future<void> load() async {
    final prefs = await SharedPreferences.getInstance();
    final saved = prefs.getString(_prefKey);
    if (saved != null && _isSupported(saved)) {
      _locale = Locale(saved);
      notifyListeners();
      return;
    }
    final device = PlatformDispatcher.instance.locale.languageCode;
    final resolved = _isSupported(device) ? device : 'fr';
    _locale = Locale(resolved);
    await prefs.setString(_prefKey, resolved);
    notifyListeners();
  }

  Future<void> setLocale(String code) async {
    if (!_isSupported(code)) return;
    _locale = Locale(code);
    notifyListeners();
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_prefKey, code);
  }

  bool _isSupported(String code) => supportedCodes.contains(code);
}
