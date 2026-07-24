import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/shell/presentation/pro_light_shell_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

void main() {
  test('proLightSpecialtyLabel maps known specialties', () {
    final l10n = AppLocalizationsFr();
    expect(proLightSpecialtyLabel(l10n, 'farrier'), l10n.proLightSpecialtyFarrier);
    expect(proLightSpecialtyLabel(l10n, 'vet_light'), l10n.proLightSpecialtyVetLight);
    expect(proLightSpecialtyLabel(l10n, 'unknown'), 'unknown');
  });

  testWidgets('pro light nav exposes four destinations', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Builder(
          builder: (context) {
            final l10n = AppLocalizations.of(context)!;
            return Scaffold(
              bottomNavigationBar: NavigationBar(
                key: const Key('pro_light_nav'),
                selectedIndex: 0,
                onDestinationSelected: (_) {},
                destinations: [
                  NavigationDestination(icon: const Icon(Icons.event), label: l10n.proLightAgenda),
                  NavigationDestination(icon: const Icon(Icons.people), label: l10n.proLightClients),
                  NavigationDestination(icon: const Icon(Icons.pets), label: l10n.proLightPets),
                  NavigationDestination(
                    icon: const Icon(Icons.settings_outlined),
                    label: l10n.proLightSettings,
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );

    expect(find.byKey(const Key('pro_light_nav')), findsOneWidget);
    expect(find.text(AppLocalizationsFr().proLightAgenda), findsWidgets);
    expect(find.text(AppLocalizationsFr().proLightClients), findsWidgets);
    expect(find.text(AppLocalizationsFr().proLightPets), findsWidgets);
    expect(find.text(AppLocalizationsFr().proLightSettings), findsWidgets);
  });
}
