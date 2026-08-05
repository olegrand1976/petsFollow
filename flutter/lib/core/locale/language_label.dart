import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Native language name for a supported locale code.
String languageLabel(AppLocalizations l10n, String code) {
  switch (code) {
    case 'fr':
      return l10n.languageFr;
    case 'nl':
      return l10n.languageNl;
    case 'en':
      return l10n.languageEn;
    case 'es':
      return l10n.languageEs;
    case 'et':
      return l10n.languageEt;
    case 'it':
      return l10n.languageIt;
    case 'uk':
      return l10n.languageUk;
    case 'ru':
      return l10n.languageRu;
    default:
      return code;
  }
}
