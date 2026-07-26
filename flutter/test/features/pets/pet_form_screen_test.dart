import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_form_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;
  var createPetCalled = false;
  Map<String, dynamic>? createPetBody;

  setUp(() {
    createPetCalled = false;
    createPetBody = null;
    mock = MockApi();
    mock.json('GET', '/api/v1/billing/plans', data: {
      'plans': [
        {'code': 'monthly', 'label': '3,50 € / mois'},
        {'code': 'annual', 'label': '35 € / an'},
        {'code': 'triennial', 'label': '95 € / 3 ans', 'recommended': true},
      ],
    });
    mock.on('POST', '/api/v1/pets', (options) {
      createPetCalled = true;
      createPetBody = Map<String, dynamic>.from(options.data as Map);
      return mock.ok(
        options,
        {
          'pet': {'id': 'pet-new-1', 'name': 'Rex', 'species': 'dog'},
          'checkoutUrl': 'https://checkout.stripe.test/session/mock',
          'sessionId': 'cs_test_mock',
        },
        status: 201,
      );
    });
    mock.install();
  });

  tearDown(() => mock.uninstall());

  testWidgets(
      'sticky CTA on-screen; name required feedback; posts createPet',
      (tester) async {
    final l10n = AppLocalizationsFr();
    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(tester, home: const PetFormScreen());
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    final cta = find.byKey(const Key('pet_form_continue_payment'));
    expect(cta, findsOneWidget);
    expect(find.text(l10n.continueToPayment), findsOneWidget);

    final screenH = tester.getSize(find.byType(Scaffold)).height;
    final ctaRect = tester.getRect(cta);
    expect(ctaRect.bottom, lessThanOrEqualTo(screenH + 0.5));
    expect(ctaRect.top, greaterThan(screenH * 0.55),
        reason: 'CTA must sit in sticky footer (lower half), not buried in scroll');

    // Empty name → error, no API call.
    await tester.tap(cta);
    await tester.pump();
    expect(find.text(l10n.petNameRequired), findsOneWidget);
    expect(createPetCalled, isFalse);

    await tester.enterText(find.byKey(const Key('pet_form_name')), 'Rex');
    await tester.pump();
    expect(find.text(l10n.petNameRequired), findsNothing);

    await tester.tap(cta);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(createPetCalled, isTrue);
    expect(createPetBody?['name'], 'Rex');
    expect(createPetBody?['plan'], 'triennial');
    expect(createPetBody?['billingMode'], 'subscription');
  });
}
