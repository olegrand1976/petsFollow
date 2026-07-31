import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

void main() {
  testWidgets('pro light consultation action keys are findable', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Scaffold(
          body: Column(
            children: [
              IconButton(
                key: const Key('pro_light_new_consultation'),
                onPressed: () {},
                icon: const Icon(Icons.medical_services_outlined),
              ),
              FloatingActionButton.extended(
                key: const Key('pro_light_new_consultation_fab'),
                onPressed: () {},
                icon: const Icon(Icons.medical_services_outlined),
                label: const Text('consult'),
              ),
              FilledButton(
                key: const Key('pro_light_consultation_cta_daf'),
                onPressed: () {},
                child: const Text('daf'),
              ),
              OutlinedButton(
                key: const Key('pro_light_consultation_cta_invoice'),
                onPressed: () {},
                child: const Text('invoice'),
              ),
              TextButton(
                key: const Key('pro_light_consultation_cta_done'),
                onPressed: () {},
                child: const Text('done'),
              ),
            ],
          ),
        ),
      ),
    );

    expect(find.byKey(const Key('pro_light_new_consultation')), findsOneWidget);
    expect(find.byKey(const Key('pro_light_new_consultation_fab')), findsOneWidget);
    expect(find.byKey(const Key('pro_light_consultation_cta_daf')), findsOneWidget);
    expect(find.byKey(const Key('pro_light_consultation_cta_invoice')), findsOneWidget);
    expect(find.byKey(const Key('pro_light_consultation_cta_done')), findsOneWidget);
  });
}
