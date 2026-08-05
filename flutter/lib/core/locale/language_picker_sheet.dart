import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/locale/language_label.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Opens a scrollable bottom sheet to pick among [LocaleController.supportedCodes].
/// Prefer this over [DropdownButton] — 8 locales overflow the overlay without scroll.
Future<void> showLanguagePickerSheet(BuildContext context) async {
  final l10n = AppLocalizations.of(context)!;
  final current = LocaleController.instance.languageCode;
  await showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    builder: (sheetContext) {
      final maxH = MediaQuery.sizeOf(sheetContext).height * 0.65;
      return SafeArea(
        child: ConstrainedBox(
          constraints: BoxConstraints(maxHeight: maxH),
          child: SingleChildScrollView(
            key: const Key('language_picker_list'),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 4, 16, 8),
                  child: Text(
                    l10n.language,
                    style: Theme.of(sheetContext).textTheme.titleMedium,
                  ),
                ),
                for (final code in LocaleController.supportedCodes)
                  ListTile(
                    key: Key('language_option_$code'),
                    title: Text(languageLabel(l10n, code)),
                    trailing: code == current
                        ? Icon(
                            Icons.check,
                            color: Theme.of(sheetContext).colorScheme.primary,
                          )
                        : null,
                    selected: code == current,
                    onTap: () async {
                      Navigator.of(sheetContext).pop();
                      if (code == current) return;
                      try {
                        if (ApiClient.instance.token != null) {
                          await ApiClient.instance.updateLocale(code);
                        } else {
                          await LocaleController.instance.setLocale(code);
                        }
                      } catch (_) {
                        await LocaleController.instance.setLocale(code);
                      }
                    },
                  ),
              ],
            ),
          ),
        ),
      );
    },
  );
}

/// Settings row: current language label + opens [showLanguagePickerSheet].
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
          onTap: () => showLanguagePickerSheet(context),
        );
      },
    );
  }
}
