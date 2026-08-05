import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/locale/language_label.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Full-screen language picker — reliable scroll (modal sheets steal vertical drag on Android).
class LanguagePickerScreen extends StatelessWidget {
  const LanguagePickerScreen({super.key});

  Future<void> _select(BuildContext context, String code, String current) async {
    if (code == current) {
      Navigator.of(context).pop();
      return;
    }
    final l10n = AppLocalizations.of(context)!;
    if (ApiClient.instance.token == null) {
      await LocaleController.instance.setLocale(code);
      if (context.mounted) Navigator.of(context).pop();
      return;
    }
    try {
      await ApiClient.instance.updateLocale(code);
      if (context.mounted) Navigator.of(context).pop();
    } catch (e) {
      if (!context.mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(mapApiError(e, l10n))),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.language)),
      body: ListenableBuilder(
        listenable: LocaleController.instance,
        builder: (context, _) {
          final current = LocaleController.instance.languageCode;
          return ListView(
            key: const Key('language_picker_list'),
            children: [
              for (final code in LocaleController.supportedCodes)
                ListTile(
                  key: Key('language_option_$code'),
                  title: Text(languageLabel(l10n, code)),
                  trailing: code == current
                      ? Icon(
                          Icons.check,
                          color: Theme.of(context).colorScheme.primary,
                        )
                      : null,
                  selected: code == current,
                  onTap: () => _select(context, code, current),
                ),
            ],
          );
        },
      ),
    );
  }
}

/// Settings row: current language + opens [LanguagePickerScreen].
class LanguageSettingsTile extends StatelessWidget {
  const LanguageSettingsTile({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return ListenableBuilder(
      listenable: LocaleController.instance,
      builder: (context, _) {
        final code = LocaleController.instance.languageCode;
        return ListTile(
          key: const Key('settings_language'),
          leading: const Icon(Icons.language),
          title: Text(l10n.language),
          subtitle: Text(languageLabel(l10n, code)),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.of(context).push(
            MaterialPageRoute<void>(
              builder: (_) => const LanguagePickerScreen(),
            ),
          ),
        );
      },
    );
  }
}
