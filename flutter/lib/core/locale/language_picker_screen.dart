import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
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
    // Always apply locally first so uk/ru work even if PATCH fails (DB lag, etc.).
    await LocaleController.instance.setLocale(code);
    if (ApiClient.instance.token != null) {
      try {
        await ApiClient.instance.patchPreferredLocale(code);
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(mapApiError(e, l10n))),
          );
        }
      }
    }
    if (context.mounted) Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final bottom = MediaQuery.paddingOf(context).bottom;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.language)),
      body: ListenableBuilder(
        listenable: LocaleController.instance,
        builder: (context, _) {
          final current = LocaleController.instance.languageCode;
          return ListView(
            key: const Key('language_picker_list'),
            padding: EdgeInsets.only(bottom: bottom + 24),
            children: [
              for (final code in LocaleController.supportedCodes)
                ListTile(
                  key: Key('language_option_$code'),
                  title: Text(languageLabel(l10n, code)),
                  subtitle: Text(code.toUpperCase()),
                  trailing: code == current
                      ? Icon(Icons.check, color: Theme.of(context).colorScheme.primary)
                      : null,
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
          subtitle: Text('${languageLabel(l10n, code)} (${code.toUpperCase()})'),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.of(context, rootNavigator: true).push(
            MaterialPageRoute<void>(
              builder: (_) => const LanguagePickerScreen(),
            ),
          ),
        );
      },
    );
  }
}

/// Build/version row so testers can confirm the installed App Distribution APK.
class AppBuildInfoTile extends StatelessWidget {
  const AppBuildInfoTile({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final version = AppEnv.buildVersion;
    final label = version.isNotEmpty
        ? version
        : '${AppEnv.value} (dev)';
    return ListTile(
      key: const Key('settings_build_info'),
      leading: const Icon(Icons.info_outline),
      title: Text(l10n.appBuildInfo),
      subtitle: Text(label),
    );
  }
}
