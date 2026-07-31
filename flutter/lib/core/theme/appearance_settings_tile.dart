import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/theme/theme_controller.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Light/Dark switch for settings menus (client + Pro Light).
class AppearanceSettingsTile extends StatelessWidget {
  const AppearanceSettingsTile({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return ListenableBuilder(
      listenable: ThemeController.instance,
      builder: (context, _) {
        final dark = ThemeController.instance.isDark;
        return SwitchListTile(
          key: const Key('settings_appearance'),
          secondary: Icon(dark ? Icons.dark_mode_outlined : Icons.light_mode_outlined),
          title: Text(l10n.appearance),
          subtitle: Text(dark ? l10n.themeDark : l10n.themeLight),
          value: dark,
          onChanged: (v) => ThemeController.instance.setDark(v),
        );
      },
    );
  }
}
