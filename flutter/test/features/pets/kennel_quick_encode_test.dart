import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/pets/presentation/kennel_quick_encode_screen.dart';
import 'package:petsfollow_mobile/features/vets/presentation/my_vets_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;
  var batchCalled = false;
  List<dynamic>? batchPets;

  setUp(() {
    batchCalled = false;
    batchPets = null;
    mock = MockApi();
    mock.on('POST', '/api/v1/pets/batch', (options) {
      batchCalled = true;
      final body = Map<String, dynamic>.from(options.data as Map);
      batchPets = body['pets'] as List<dynamic>?;
      final pets = (batchPets ?? []).map((raw) {
        final p = Map<String, dynamic>.from(raw as Map);
        return {
          'id': 'pet-${p['name']}',
          'name': p['name'],
          'species': p['species'],
          'practiceId': 'practice-demo',
          'paymentStatus': 'pending_payment',
          'entitlement': {'status': 'pending', 'planCode': 'triennial'},
        };
      }).toList();
      return mock.ok(
        options,
        {'pets': pets, 'count': pets.length},
        status: 201,
      );
    });
    mock.install();
  });

  tearDown(() => mock.uninstall());

  testWidgets('kennel submit posts batch with steer plan and pops',
      (tester) async {
    final l10n = AppLocalizationsFr();
    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(
      tester,
      home: Scaffold(
        body: Builder(
          builder: (ctx) => TextButton(
            key: const Key('open_kennel'),
            onPressed: () {
              Navigator.of(ctx).push(
                MaterialPageRoute<void>(
                  builder: (_) => const KennelQuickEncodeScreen(),
                ),
              );
            },
            child: const Text('open'),
          ),
        ),
      ),
    );
    await tester.tap(find.byKey(const Key('open_kennel')));
    await tester.pumpAndSettle();

    expect(find.byType(KennelQuickEncodeScreen), findsOneWidget);
    expect(find.text(l10n.kennelQuickEncodeTitle), findsOneWidget);
    expect(find.text(l10n.petBirthDate), findsOneWidget);

    await tester.enterText(find.byType(TextField).first, 'Rex');
    await tester.pump();

    await tester.tap(find.byKey(const Key('kennel_submit')));
    await tester.pump();
    await tester.pumpAndSettle();

    expect(batchCalled, isTrue);
    expect(batchPets, isNotNull);
    expect(batchPets!.length, 1);
    final row = Map<String, dynamic>.from(batchPets!.first as Map);
    expect(row['name'], 'Rex');
    expect(row['species'], 'dog');
    expect(row['plan'], 'triennial');
    expect(row['billingMode'], 'subscription');
    expect(find.byType(KennelQuickEncodeScreen), findsNothing);
  });

  testWidgets('kennel empty names does not call API', (tester) async {
    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(tester, home: const KennelQuickEncodeScreen());
    await tester.pump();

    await tester.tap(find.byKey(const Key('kennel_submit')));
    await tester.pump();

    expect(batchCalled, isFalse);
    expect(find.byType(KennelQuickEncodeScreen), findsOneWidget);
  });

  testWidgets('kennel without practice shows link-vet dialog then pops',
      (tester) async {
    final l10n = AppLocalizationsFr();
    mock.uninstall();
    mock = MockApi();
    mock.on('POST', '/api/v1/pets/batch', (options) {
      return mock.ok(
        options,
        {
          'pets': [
            {
              'id': 'pet-orphan',
              'name': 'Rex',
              'species': 'dog',
              'practiceId': '',
              'paymentStatus': 'pending_payment',
            },
          ],
          'count': 1,
        },
        status: 201,
      );
    });
    mock.json('GET', '/api/v1/me/vets', data: <dynamic>[]);
    mock.install();

    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(tester, home: const KennelQuickEncodeScreen());
    await tester.pump();
    await tester.enterText(find.byType(TextField).first, 'Rex');
    await tester.pump();
    await tester.tap(find.byKey(const Key('kennel_submit')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byKey(const Key('kennel_vet_link_dialog')), findsOneWidget);
    expect(find.text(l10n.linkVetAfterSaveTitle), findsOneWidget);
    expect(find.byKey(const Key('kennel_link_vet')), findsOneWidget);

    await tester.tap(find.byKey(const Key('kennel_link_vet')));
    await tester.pumpAndSettle();
    expect(find.byType(MyVetsScreen), findsOneWidget);
    expect(find.byType(KennelQuickEncodeScreen), findsNothing);
  });
}
