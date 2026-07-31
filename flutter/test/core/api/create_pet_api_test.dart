import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';

import '../../helpers/mock_api.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.install();
  });

  tearDown(() => mock.uninstall());

  test('createPet unwraps data envelope via _asMap', () async {
    mock.on('POST', '/api/v1/pets', (options) {
      return mock.ok(
        options,
        {
          'pet': {
            'id': 'pet-1',
            'name': 'Rex',
            'species': 'dog',
            'entitlement': {'status': 'pending', 'planCode': 'triennial'},
          },
        },
        status: 201,
      );
    });

    final res = await ApiClient.instance.createPet({
      'name': 'Rex',
      'species': 'dog',
      'plan': 'triennial',
      'billingMode': 'subscription',
      'skipCheckout': true,
    });

    expect(res['pet'], isA<Map>());
    final pet = Map<String, dynamic>.from(res['pet'] as Map);
    expect(pet['id'], 'pet-1');
  });

  test('createPet tolerates nested maps from jsonDecode (dynamic keys)',
      () async {
    // Simulate Dio JSON where nested maps are Map<String, dynamic> after decode.
    final decoded = jsonDecode(jsonEncode({
      'data': {
        'pet': {
          'id': 'pet-dyn',
          'name': 'Mimi',
          'species': 'cat',
          'paymentStatus': 'pending_payment',
          'entitlement': {'status': 'pending', 'planCode': 'annual'},
        },
      },
    })) as Map<String, dynamic>;

    mock.on('POST', '/api/v1/pets', (options) {
      return mock.ok(options, decoded, status: 201);
    });

    final res = await ApiClient.instance.createPet({'name': 'Mimi'});
    final rawPet = res['pet'];
    expect(rawPet, isA<Map>());
    final pet = Map<String, dynamic>.from(rawPet as Map);
    expect(pet['id'], 'pet-dyn');

    final model = Pet.fromJson(pet);
    expect(model.id, 'pet-dyn');
    expect(model.entitlement?.status, 'pending');
    expect(model.entitlement?.planCode, 'annual');
  });

  test('Pet.fromJson accepts entitlement Map without typed cast crash', () {
    final json = Map<String, dynamic>.from(
      jsonDecode(jsonEncode({
        'id': 'p1',
        'name': 'Bella',
        'species': 'dog',
        'breed': 'Mix',
        'paymentStatus': 'pending_payment',
        'entitlement': {
          'status': 'pending',
          'planCode': 'triennial',
          'billingMode': 'subscription',
        },
      })) as Map,
    );
    final pet = Pet.fromJson(json);
    expect(pet.id, 'p1');
    expect(pet.entitlement?.allowsAccess, isFalse);
  });
}
