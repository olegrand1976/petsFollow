import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

void main() {
  testWidgets('audio consent accept disabled until checkbox checked', (tester) async {
    var checked = false;
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Builder(
          builder: (context) {
            final l10n = AppLocalizations.of(context)!;
            return Scaffold(
              body: StatefulBuilder(
                builder: (context, setLocal) {
                  return AlertDialog(
                    title: Text(l10n.proLightAudioConsentTitle),
                    content: CheckboxListTile(
                      key: const Key('pro_light_audio_consent_checkbox'),
                      value: checked,
                      onChanged: (v) => setLocal(() => checked = v ?? false),
                      title: Text(l10n.proLightAudioConsentClientCheck),
                    ),
                    actions: [
                      FilledButton(
                        key: const Key('pro_light_audio_consent_accept'),
                        onPressed: checked ? () {} : null,
                        child: Text(l10n.proLightAudioConsentAccept),
                      ),
                    ],
                  );
                },
              ),
            );
          },
        ),
      ),
    );

    final accept = tester.widget<FilledButton>(
      find.byKey(const Key('pro_light_audio_consent_accept')),
    );
    expect(accept.onPressed, isNull);

    await tester.tap(find.byKey(const Key('pro_light_audio_consent_checkbox')));
    await tester.pump();

    final acceptAfter = tester.widget<FilledButton>(
      find.byKey(const Key('pro_light_audio_consent_accept')),
    );
    expect(acceptAfter.onPressed, isNotNull);
  });
}
