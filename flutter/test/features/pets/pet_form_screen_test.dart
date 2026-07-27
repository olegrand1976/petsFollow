import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_create_flow.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_form_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

/// Opens [PetFormScreen] via [openPetFormAndFollowUp] (snack on host after pop).
Future<void> pumpPetForm(WidgetTester tester) async {
  await tester.binding.setSurfaceSize(const Size(390, 844));
  addTearDown(() => tester.binding.setSurfaceSize(null));

  await pumpApp(
    tester,
    home: Scaffold(
      body: Builder(
        builder: (ctx) => Center(
          child: TextButton(
            key: const Key('open_pet_form'),
            onPressed: () => openPetFormAndFollowUp(
              ctx,
              onReload: () async {},
              hasLinkedVets: true,
            ),
            child: const Text('open'),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.byKey(const Key('open_pet_form')));
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 100));
  expect(find.byType(PetFormScreen), findsOneWidget);
}

void main() {
  late MockApi mock;
  var createPetCalled = false;
  Map<String, dynamic>? createPetBody;
  final defaultOpenUrl = openExternalUrlImpl;

  setUp(() {
    createPetCalled = false;
    createPetBody = null;
    openExternalUrlImpl = (_) async => true;
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
      final skip = createPetBody?['skipCheckout'] == true;
      return mock.ok(
        options,
        {
          'pet': {
            'id': 'pet-new-1',
            'name': 'Rex',
            'species': 'dog',
            'practiceId': 'practice-demo',
            'paymentStatus': 'pending_payment',
            'entitlement': {'status': 'pending', 'planCode': 'triennial'},
          },
          if (!skip) ...{
            'checkoutUrl': 'https://checkout.stripe.test/session/mock',
            'sessionId': 'cs_test_mock',
          },
        },
        status: 201,
      );
    });
    mock.install();
  });

  tearDown(() {
    openExternalUrlImpl = defaultOpenUrl;
    mock.uninstall();
  });

  testWidgets(
      'sticky save CTA; name required; createPet with skipCheckout; pop + snackbar',
      (tester) async {
    final l10n = AppLocalizationsFr();
    await pumpPetForm(tester);

    final saveCta = find.byKey(const Key('pet_form_save'));
    final payCta = find.byKey(const Key('pet_form_continue_payment'));
    expect(saveCta, findsOneWidget);
    expect(payCta, findsOneWidget);
    expect(find.text(l10n.petFormSave), findsOneWidget);
    expect(find.text(l10n.continueToPayment), findsOneWidget);

    final screenH = tester.getSize(find.byType(Scaffold).last).height;
    final ctaRect = tester.getRect(saveCta);
    expect(ctaRect.bottom, lessThanOrEqualTo(screenH + 0.5));
    expect(ctaRect.top, greaterThan(screenH * 0.55),
        reason: 'CTA must sit in sticky footer (lower half), not buried in scroll');

    await tester.tap(saveCta);
    await tester.pump();
    expect(find.text(l10n.petNameRequired), findsOneWidget);
    expect(createPetCalled, isFalse);

    await tester.enterText(find.byKey(const Key('pet_form_name')), 'Rex');
    await tester.pump();
    expect(find.text(l10n.petNameRequired), findsNothing);
    expect(find.byKey(const Key('pet_form_microchip')), findsOneWidget);
    expect(find.byKey(const Key('pet_form_health_book_number')), findsOneWidget);
    expect(find.byKey(const Key('pet_form_health_book_pick')), findsOneWidget);

    await tester.tap(saveCta);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));
    await tester.pumpAndSettle();

    expect(createPetCalled, isTrue);
    expect(createPetBody?['name'], 'Rex');
    expect(createPetBody?['plan'], 'triennial');
    expect(createPetBody?['billingMode'], 'subscription');
    expect(createPetBody?['skipCheckout'], isTrue);
    expect(createPetBody?['microchipNumber'], '');
    expect(createPetBody?['healthBookNumber'], '');
    expect(find.byKey(const Key('pet_form_saved')), findsOneWidget);
    expect(find.text(l10n.petSavedPendingPayment), findsOneWidget);
    expect(find.byType(PetFormScreen), findsNothing);
  });

  testWidgets('pay CTA posts createPet without skipCheckout', (tester) async {
    await pumpPetForm(tester);

    await tester.enterText(find.byKey(const Key('pet_form_name')), 'Rex');
    await tester.pump();
    await tester.tap(find.byKey(const Key('pet_form_continue_payment')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));
    await tester.pumpAndSettle();

    expect(createPetCalled, isTrue);
    expect(createPetBody?['skipCheckout'], isFalse);
    expect(find.byType(PetFormScreen), findsNothing);
  });

  testWidgets('API error keeps form open (no pop)', (tester) async {
    mock.uninstall();
    mock = MockApi();
    mock.json('GET', '/api/v1/billing/plans', data: {
      'plans': [
        {'code': 'triennial', 'label': '95 € / 3 ans', 'recommended': true},
      ],
    });
    mock.on('POST', '/api/v1/pets', (options) {
      return mock.err(
        options,
        status: 500,
        code: 'internal',
        msgKey: 'internal',
        message: 'Erreur serveur',
      );
    });
    mock.install();

    await pumpPetForm(tester);
    await tester.enterText(find.byKey(const Key('pet_form_name')), 'Rex');
    await tester.pump();
    await tester.tap(find.byKey(const Key('pet_form_save')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byType(PetFormScreen), findsOneWidget);
    expect(find.byKey(const Key('pet_form_error')), findsAtLeastNWidgets(1));
    expect(find.byKey(const Key('pet_form_saved')), findsNothing);
  });

  testWidgets('create without practice pops then optional home dialog',
      (tester) async {
    final l10n = AppLocalizationsFr();
    mock.uninstall();
    mock = MockApi();
    mock.json('GET', '/api/v1/billing/plans', data: {
      'plans': [
        {'code': 'monthly', 'label': '3,50 € / mois'},
        {'code': 'annual', 'label': '35 € / an'},
        {'code': 'triennial', 'label': '95 € / 3 ans', 'recommended': true},
      ],
    });
    mock.on('POST', '/api/v1/pets', (options) {
      return mock.ok(
        options,
        {
          'pet': {
            'id': 'pet-orphan-1',
            'name': 'Justine',
            'species': 'dog',
            'practiceId': '',
            'paymentStatus': 'pending_payment',
            'entitlement': {'status': 'pending', 'planCode': 'triennial'},
          },
        },
        status: 201,
      );
    });
    mock.json('GET', '/api/v1/me/vets', data: <dynamic>[]);
    mock.install();

    await pumpPetForm(tester);

    await tester.enterText(find.byKey(const Key('pet_form_name')), 'Justine');
    await tester.pump();
    await tester.tap(find.byKey(const Key('pet_form_save')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('pet_form_vet_link_dialog')), findsNothing);
    expect(find.byType(PetFormScreen), findsNothing);
    expect(find.byKey(const Key('pet_form_saved')), findsOneWidget);
    // hasLinkedVets: true in pumpPetForm → dialog after pop.
    expect(find.byKey(const Key('home_link_vet_dialog')), findsOneWidget);
    expect(find.text(l10n.linkVetHomeTitle), findsOneWidget);

    await tester.tap(find.byKey(const Key('home_link_vet_later')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('home_link_vet_dialog')), findsNothing);
  });

  testWidgets('monthly plan always posts billingMode subscription',
      (tester) async {
    await pumpPetForm(tester);

    await tester.enterText(find.byKey(const Key('pet_form_name')), 'Justine');
    await tester.pump();

    await tester.scrollUntilVisible(
      find.text('3,50 € / mois'),
      200,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.tap(find.text('3,50 € / mois'));
    await tester.pump();

    await tester.tap(find.byKey(const Key('pet_form_continue_payment')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));
    await tester.pumpAndSettle();

    expect(createPetCalled, isTrue);
    expect(createPetBody?['plan'], 'monthly');
    expect(createPetBody?['billingMode'], 'subscription');
    expect(find.byType(PetFormScreen), findsNothing);
  });

  testWidgets('payNow without checkoutUrl still pops (pet created)',
      (tester) async {
    final l10n = AppLocalizationsFr();
    mock.uninstall();
    mock = MockApi();
    mock.json('GET', '/api/v1/billing/plans', data: {
      'plans': [
        {'code': 'triennial', 'label': '95 € / 3 ans', 'recommended': true},
      ],
    });
    mock.on('POST', '/api/v1/pets', (options) {
      createPetCalled = true;
      return mock.ok(
        options,
        {
          'pet': {
            'id': 'pet-no-checkout',
            'name': 'Rex',
            'species': 'dog',
            'practiceId': 'practice-demo',
            'paymentStatus': 'pending_payment',
          },
        },
        status: 201,
      );
    });
    mock.install();

    await pumpPetForm(tester);
    await tester.enterText(find.byKey(const Key('pet_form_name')), 'Rex');
    await tester.pump();
    await tester.tap(find.byKey(const Key('pet_form_continue_payment')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));
    await tester.pumpAndSettle();

    expect(createPetCalled, isTrue);
    expect(find.text(l10n.petSavedPendingPayment), findsOneWidget);
    expect(find.byType(PetFormScreen), findsNothing);
  });
}
