import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// app_fr.arb est le template déclaré dans l10n.yaml : c'est la source de vérité.
const _templateLocale = 'fr';

/// Lève au lieu d'utiliser `expect` : cette fonction est aussi appelée hors d'un
/// test (chargement du template au démarrage), où `expect` déclencherait
/// OutsideTestException et ferait échouer le *chargement* du fichier de test.
Map<String, Object?> _loadArb(String locale) {
  final file = File('lib/l10n/app_$locale.arb');
  if (!file.existsSync()) {
    throw StateError('lib/l10n/app_$locale.arb manquant');
  }
  return jsonDecode(file.readAsStringSync()) as Map<String, Object?>;
}

/// Clés de message, hors métadonnées (`@@locale`, `@maClé`).
Iterable<String> _messageKeys(Map<String, Object?> arb) =>
    arb.keys.where((k) => !k.startsWith('@'));

final _placeholder = RegExp(r'\{(\w+)\}');

Set<String> _placeholders(Object? value) =>
    _placeholder.allMatches('$value').map((m) => m.group(1)!).toSet();

void main() {
  final template = _loadArb(_templateLocale);
  final templateKeys = _messageKeys(template).toSet();

  // Les locales réellement exposées à l'utilisateur, pas la liste des fichiers
  // présents : un .arb orphelin ne doit pas suffire à faire passer ce test.
  final locales = LocaleController.supportedCodes;

  test('chaque locale supportée a un fichier ARB et un @@locale cohérent', () {
    for (final loc in locales) {
      final arb = _loadArb(loc);
      expect(arb['@@locale'], loc, reason: '@@locale incorrect dans app_$loc.arb');
    }
  });

  test('supportedCodes et AppLocalizations.supportedLocales concordent', () {
    // gen-l10n dérive supportedLocales des fichiers ARB présents ; si les deux
    // listes divergent, le picker propose une langue que MaterialApp ignore
    // (ou l'inverse) — le symptôme est un écran en anglais sans explication.
    final generated = AppLocalizations.supportedLocales.map((l) => l.languageCode).toSet();
    expect(generated, locales.toSet());
  });

  test('aucune clé manquante ni en trop vs le template fr', () {
    for (final loc in locales) {
      if (loc == _templateLocale) continue;
      final keys = _messageKeys(_loadArb(loc)).toSet();
      expect(
        keys.difference(templateKeys),
        isEmpty,
        reason: 'app_$loc.arb contient des clés absentes du template',
      );
      expect(
        templateKeys.difference(keys),
        isEmpty,
        reason: 'app_$loc.arb ne traduit pas toutes les clés du template',
      );
    }
  });

  test('les placeholders sont identiques au template', () {
    for (final loc in locales) {
      if (loc == _templateLocale) continue;
      final arb = _loadArb(loc);
      for (final key in templateKeys) {
        expect(
          _placeholders(arb[key]),
          _placeholders(template[key]),
          reason: 'placeholders divergents pour "$key" dans app_$loc.arb',
        );
      }
    }
  });

  test('aucune valeur vide là où le template est renseigné', () {
    for (final loc in locales) {
      final arb = _loadArb(loc);
      for (final key in templateKeys) {
        if ('${template[key]}'.trim().isEmpty) continue;
        expect(
          '${arb[key]}'.trim(),
          isNotEmpty,
          reason: '"$key" est vide dans app_$loc.arb',
        );
      }
    }
  });

  test('les libellés de langue existent pour chaque locale supportée', () {
    // languageFr, languageNl… : un oubli casse le LanguagePickerSheet.
    for (final loc in locales) {
      final key = 'language${loc[0].toUpperCase()}${loc.substring(1)}';
      expect(
        templateKeys.contains(key),
        isTrue,
        reason: 'clé $key attendue pour la locale $loc',
      );
    }
  });
}
