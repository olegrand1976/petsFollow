import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_edit_screen.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;
  Map<String, dynamic>? updateBody;

  setUp(() {
    updateBody = null;
    mock = MockApi();
    mock.on('PUT', '/api/v1/pets/pet-1', (options) {
      updateBody = Map<String, dynamic>.from(options.data as Map);
      return mock.ok(options, {'status': 'updated'});
    });
    mock.install();
  });

  tearDown(() => mock.uninstall());

  testWidgets('edit sends microchip + health book number', (tester) async {
    const pet = Pet(
      id: 'pet-1',
      name: 'Rex',
      species: 'dog',
      breed: 'Labrador',
      ownerUserId: 'user-1',
      microchipNumber: '250269601234567',
      healthBookNumber: 'HB-42',
    );

    await pumpApp(tester, home: const PetEditScreen(pet: pet));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('pet_edit_microchip')), findsOneWidget);
    expect(find.byKey(const Key('pet_edit_health_book_number')), findsOneWidget);
    expect(find.byKey(const Key('pet_edit_health_book_pick')), findsOneWidget);

    await tester.enterText(find.byKey(const Key('pet_edit_microchip')), '999');
    await tester.enterText(
        find.byKey(const Key('pet_edit_health_book_number')), 'HB-99');
    await tester.tap(find.byKey(const Key('pet_edit_save')));
    await tester.pump();
    await tester.pumpAndSettle();

    expect(updateBody?['microchipNumber'], '999');
    expect(updateBody?['healthBookNumber'], 'HB-99');
    expect(updateBody?['name'], 'Rex');
    expect(updateBody?['domicileLocation'], '');
    expect(find.byType(PetEditScreen), findsNothing);
  });

  testWidgets('horse edit sends domicile location', (tester) async {
    const pet = Pet(
      id: 'pet-1',
      name: 'Spirit',
      species: 'horse',
      breed: 'Selle Français',
      ownerUserId: 'user-1',
      domicileLocation: 'Old stable',
      foodChainStatus: 'excluded_from_food_chain',
    );

    await pumpApp(tester, home: const PetEditScreen(pet: pet));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('pet_edit_domicile')), findsOneWidget);
    await tester.enterText(
        find.byKey(const Key('pet_edit_domicile')), 'Écurie Demo — Bruxelles');
    await tester.ensureVisible(find.byKey(const Key('pet_edit_save')));
    await tester.tap(find.byKey(const Key('pet_edit_save')));
    await tester.pump();
    await tester.pumpAndSettle();

    expect(updateBody, isNotNull);
    expect(updateBody?['domicileLocation'], 'Écurie Demo — Bruxelles');
    expect(updateBody?['species'], 'horse');
    expect(find.byType(PetEditScreen), findsNothing);
  });

  testWidgets('leaving horse clears domicile', (tester) async {
    const pet = Pet(
      id: 'pet-1',
      name: 'Spirit',
      species: 'horse',
      breed: 'Selle Français',
      ownerUserId: 'user-1',
      domicileLocation: 'Écurie Demo',
    );

    await pumpApp(tester, home: const PetEditScreen(pet: pet));
    await tester.pumpAndSettle();

    await tester.tap(find.byType(DropdownButtonFormField<String>));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Chien').last);
    await tester.pumpAndSettle();

    await tester.ensureVisible(find.byKey(const Key('pet_edit_save')));
    await tester.tap(find.byKey(const Key('pet_edit_save')));
    await tester.pump();
    await tester.pumpAndSettle();

    expect(updateBody?['species'], 'dog');
    expect(updateBody?['domicileLocation'], '');
  });
}
