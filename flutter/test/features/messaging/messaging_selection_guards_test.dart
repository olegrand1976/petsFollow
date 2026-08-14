import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';

import '../../helpers/fixtures.dart';
import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    ApiClient.instance.userId = 'user-1';
    ApiClient.instance.userRole = 'client';
  });

  tearDown(() {
    mock.uninstall();
    ApiClient.instance.userId = null;
    ApiClient.instance.userRole = null;
  });

  void stubClientBase({
    required List<Map<String, dynamic>> pets,
    List<Map<String, dynamic>> threads = const [],
  }) {
    mock.json('GET', '/api/v1/me', data: {
      'userId': 'user-1',
      'id': 'user-1',
      'role': 'client',
    });
    mock.json('GET', '/api/v1/me/vets', data: [
      {
        'practiceId': 'practice-1',
        'practiceName': 'VetPlus',
        'vetFullName': 'Dr Demo',
        'status': 'active',
      },
    ]);
    mock.json('GET', '/api/v1/pets', data: pets);
    mock.json('GET', '/api/v1/messaging/threads', data: threads);
    mock.json('GET', RegExp(r'/api/v1/messaging/threads/.+/messages'), data: []);
    mock.json('POST', RegExp(r'/api/v1/messaging/threads/.+/read'), data: {});
  }

  testWidgets('after ensure, message_draft is enabled', (tester) async {
    mock.json('GET', '/api/v1/me', data: {'userId': 'vet-1', 'role': 'vet'});
    mock.json('GET', '/api/v1/messaging/threads', data: []);
    mock.on('POST', '/api/v1/messaging/threads', (options) {
      return mock.ok(options, {
        'id': 'th-new',
        'practiceId': 'p1',
        'clientUserId': 'c1',
        'clientName': 'Client Demo',
        'petId': 'pet-1',
        'petName': 'Rex',
      });
    });
    mock.json('GET', RegExp(r'/api/v1/messaging/threads/.+/messages'), data: []);
    mock.install();
    ApiClient.instance.userRole = 'vet';
    ApiClient.instance.userId = 'vet-1';

    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(
      tester,
      home: MessagingScreen(
        staffMode: true,
        staffClients: const [
          {'userId': 'c1', 'fullName': 'Client Demo', 'email': 'c@test'},
        ],
        staffPets: const [
          {'id': 'pet-1', 'name': 'Rex', 'ownerUserId': 'c1'},
        ],
      ),
    );
    await tester.pumpAndSettle();

    final draft = tester.widget<TextField>(find.byKey(const Key('message_draft')));
    expect(draft.enabled, isTrue);
    expect(tester.widget<IconButton>(find.byKey(const Key('message_send_btn'))).onPressed,
        isNotNull);
  });

  testWidgets('changing pet during ensure ignores stale thread', (tester) async {
    final firstEnsureGate = Completer<void>();
    final ensures = <Map<String, dynamic>>[];

    stubClientBase(pets: [
      {...Fixtures.pet(id: 'pet-1', name: 'Rex'), 'practiceId': 'practice-1'},
      {...Fixtures.pet(id: 'pet-2', name: 'Mimi'), 'practiceId': 'practice-1'},
    ]);
    mock.on('POST', '/api/v1/messaging/threads', (options) async {
      final body = Map<String, dynamic>.from(options.data as Map);
      ensures.add(body);
      if (ensures.length == 1) {
        await firstEnsureGate.future;
        return mock.ok(options, {
          'id': 'th-stale',
          'practiceId': 'practice-1',
          'clientUserId': 'user-1',
          'petId': 'pet-1',
          'petName': 'Rex',
        });
      }
      return mock.ok(options, {
        'id': 'th-current',
        'practiceId': 'practice-1',
        'clientUserId': 'user-1',
        'petId': body['petId'],
        'petName': 'Mimi',
      });
    });
    mock.install();

    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(tester, home: const MessagingScreen());
    // Avoid pumpAndSettle while ensure is gated — progress indicator never settles.
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(ensures, isNotEmpty);
    expect(ensures.first['petId'], 'pet-1');
    expect(find.byKey(const Key('message_select_pet')), findsOneWidget);

    await tester.tap(find.byKey(const Key('message_select_pet')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.tap(find.text('Mimi').last);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(ensures.length, greaterThanOrEqualTo(2));
    expect(ensures.last['petId'], 'pet-2');

    firstEnsureGate.complete();
    await tester.pumpAndSettle();

    final draft =
        tester.widget<TextField>(find.byKey(const Key('message_draft')));
    expect(draft.enabled, isTrue);
    expect(find.text('Mimi'), findsWidgets);
  });

  testWidgets('changing selection clears draft text', (tester) async {
    stubClientBase(
      pets: [
        {...Fixtures.pet(id: 'pet-1', name: 'Rex'), 'practiceId': 'practice-1'},
        {...Fixtures.pet(id: 'pet-2', name: 'Mimi'), 'practiceId': 'practice-1'},
      ],
      threads: [
        {
          'id': 'thread-1',
          'practiceId': 'practice-1',
          'clientUserId': 'user-1',
          'petId': 'pet-1',
          'petName': 'Rex',
          'unreadCount': 0,
        },
        {
          'id': 'thread-2',
          'practiceId': 'practice-1',
          'clientUserId': 'user-1',
          'petId': 'pet-2',
          'petName': 'Mimi',
          'unreadCount': 0,
        },
      ],
    );
    mock.install();

    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(tester, home: const MessagingScreen());
    await tester.pumpAndSettle();

    await tester.enterText(find.byKey(const Key('message_draft')), 'brouillon Rex');
    await tester.pump();
    expect(find.text('brouillon Rex'), findsOneWidget);

    await tester.tap(find.byKey(const Key('message_select_pet')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Mimi').last);
    await tester.pumpAndSettle();

    expect(find.text('brouillon Rex'), findsNothing);
    final draft = tester.widget<TextField>(find.byKey(const Key('message_draft')));
    expect(draft.controller?.text ?? '', isEmpty);
  });

  testWidgets('unread counts appear on vet/pet selects', (tester) async {
    stubClientBase(
      pets: [
        {...Fixtures.pet(id: 'pet-1', name: 'Rex'), 'practiceId': 'practice-1'},
        {...Fixtures.pet(id: 'pet-2', name: 'Mimi'), 'practiceId': 'practice-1'},
      ],
      threads: [
        {
          'id': 'thread-1',
          'practiceId': 'practice-1',
          'clientUserId': 'user-1',
          'petId': 'pet-1',
          'petName': 'Rex',
          'practiceName': 'VetPlus',
          'unreadCount': 0,
        },
        {
          'id': 'thread-2',
          'practiceId': 'practice-1',
          'clientUserId': 'user-1',
          'petId': 'pet-2',
          'petName': 'Mimi',
          'practiceName': 'VetPlus',
          'unreadCount': 2,
        },
      ],
    );
    mock.install();

    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(tester, home: const MessagingScreen());
    await tester.pumpAndSettle();

    // Closed vet select shows practice-level unread (other pet's thread).
    expect(find.textContaining('(2)'), findsWidgets);

    await tester.tap(find.byKey(const Key('message_select_pet')));
    await tester.pumpAndSettle();
    expect(find.text('Mimi (2)'), findsOneWidget);
  });

  testWidgets('staff multi-client shows choose-client hint when none selected',
      (tester) async {
    mock.json('GET', '/api/v1/me', data: {'userId': 'vet-1', 'role': 'vet'});
    mock.json('GET', '/api/v1/messaging/threads', data: []);
    mock.install();
    ApiClient.instance.userRole = 'vet';
    ApiClient.instance.userId = 'vet-1';

    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(
      tester,
      home: MessagingScreen(
        staffMode: true,
        staffClients: const [
          {'userId': 'c1', 'fullName': 'Client A', 'email': 'a@test'},
          {'userId': 'c2', 'fullName': 'Client B', 'email': 'b@test'},
        ],
        staffPets: const [
          {'id': 'pet-1', 'name': 'Rex', 'ownerUserId': 'c1'},
          {'id': 'pet-2', 'name': 'Mimi', 'ownerUserId': 'c2'},
        ],
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Aucune conversation'), findsNothing);
    expect(find.text('Client'), findsWidgets);
  });
}
