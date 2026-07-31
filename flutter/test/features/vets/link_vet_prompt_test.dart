import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/vets/presentation/link_vet_prompt.dart';
import 'package:petsfollow_mobile/features/vets/presentation/my_vets_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.json('GET', '/api/v1/me/vets', data: <dynamic>[]);
    mock.install();
  });

  tearDown(() => mock.uninstall());

  testWidgets('skips when promptLinkVet is false', (tester) async {
    await pumpApp(
      tester,
      home: Builder(
        builder: (ctx) => Scaffold(
          body: TextButton(
            key: const Key('trigger'),
            onPressed: () => promptLinkVetIfNeeded(
              ctx,
              promptLinkVet: false,
              hasLinkedVets: true,
            ),
            child: const Text('go'),
          ),
        ),
      ),
    );
    await tester.tap(find.byKey(const Key('trigger')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('home_link_vet_dialog')), findsNothing);
  });

  testWidgets('skips dialog when hasLinkedVets is false (banner covers it)',
      (tester) async {
    await pumpApp(
      tester,
      home: Builder(
        builder: (ctx) => Scaffold(
          body: TextButton(
            key: const Key('trigger'),
            onPressed: () => promptLinkVetIfNeeded(
              ctx,
              promptLinkVet: true,
              hasLinkedVets: false,
            ),
            child: const Text('go'),
          ),
        ),
      ),
    );
    await tester.tap(find.byKey(const Key('trigger')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('home_link_vet_dialog')), findsNothing);
  });

  testWidgets('shows optional dialog then MyVets when hasLinkedVets true',
      (tester) async {
    final l10n = AppLocalizationsFr();
    await pumpApp(
      tester,
      home: Builder(
        builder: (ctx) => Scaffold(
          body: TextButton(
            key: const Key('trigger'),
            onPressed: () => promptLinkVetIfNeeded(
              ctx,
              promptLinkVet: true,
              hasLinkedVets: true,
            ),
            child: const Text('go'),
          ),
        ),
      ),
    );
    await tester.tap(find.byKey(const Key('trigger')));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('home_link_vet_dialog')), findsOneWidget);
    expect(find.text(l10n.linkVetHomeTitle), findsOneWidget);

    await tester.tap(find.byKey(const Key('home_link_vet')));
    await tester.pumpAndSettle();
    expect(find.byType(MyVetsScreen), findsOneWidget);
  });
}
