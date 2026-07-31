import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

void main() {
  testWidgets('pro light CR action keys are findable', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Scaffold(
          body: Column(
            children: [
              FilledButton.tonalIcon(
                key: const Key('pro_light_cr_dictation'),
                onPressed: () {},
                icon: const Icon(Icons.mic_none),
                label: const Text('dictate'),
              ),
              OutlinedButton(
                key: const Key('pro_light_cr_save'),
                onPressed: () {},
                child: const Text('save'),
              ),
              FilledButton(
                key: const Key('pro_light_cr_finalize'),
                onPressed: () {},
                child: const Text('finalize'),
              ),
              FilledButton(
                key: const Key('pro_light_cr_dictation_stop'),
                onPressed: () {},
                child: const Text('stop'),
              ),
            ],
          ),
        ),
      ),
    );
    expect(find.byKey(const Key('pro_light_cr_dictation')), findsOneWidget);
    expect(find.byKey(const Key('pro_light_cr_save')), findsOneWidget);
    expect(find.byKey(const Key('pro_light_cr_finalize')), findsOneWidget);
    expect(find.byKey(const Key('pro_light_cr_dictation_stop')), findsOneWidget);
    expect(find.byKey(const Key('pro_light_cr_improve')), findsNothing);
  });
}
