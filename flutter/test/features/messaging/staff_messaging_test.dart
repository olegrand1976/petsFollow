import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/message_thread.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    ApiClient.instance.userRole = 'vet';
    ApiClient.instance.userId = 'vet-1';
  });

  tearDown(() {
    mock.uninstall();
    ApiClient.instance.userRole = null;
    ApiClient.instance.userId = null;
  });

  test('MessageThread staff displayLabel prefers clientName', () {
    const t = MessageThread(
      id: 'thread-1',
      practiceId: 'p1',
      clientName: 'Marie Demo',
      petName: 'Rex',
    );
    expect(t.displayLabel, 'Marie Demo · Rex');
  });

  testWidgets('staff messaging shows composer without vet-link lock', (tester) async {
    mock.json('GET', '/api/v1/me', data: {
      'userId': 'vet-1',
      'id': 'vet-1',
      'role': 'vet',
    });
    mock.json('GET', '/api/v1/messaging/threads', data: [
      {
        'id': 'th-1',
        'practiceId': 'p1',
        'clientUserId': 'c1',
        'clientName': 'Client Demo',
        'petId': 'pet-1',
        'petName': 'Rex',
        'lastMessagePreview': 'Bonjour',
        'unreadCount': 0,
      },
    ]);
    mock.json('GET', RegExp(r'/api/v1/messaging/threads/.+/messages'), data: [
      {
        'id': 'm1',
        'threadId': 'th-1',
        'senderUserId': 'c1',
        'body': 'Bonjour',
        'createdAt': '2026-07-01T10:00:00Z',
      },
    ]);
    mock.json('POST', RegExp(r'/api/v1/messaging/threads/.+/read'), data: {});
    mock.install();

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

    expect(find.byKey(const Key('message_link_vet_cta')), findsNothing);
    expect(find.byKey(const Key('message_draft')), findsOneWidget);
    expect(find.byKey(const Key('message_send_btn')), findsOneWidget);
    expect(find.byKey(const Key('message_compose_fab')), findsOneWidget);
    expect(find.textContaining('Client Demo'), findsWidgets);
  });

  testWidgets('staff compose confirm posts clientUserId', (tester) async {
    Map<String, dynamic>? posted;
    mock.json('GET', '/api/v1/me', data: {'userId': 'vet-1', 'role': 'vet'});
    mock.json('GET', '/api/v1/messaging/threads', data: []);
    mock.on('POST', '/api/v1/messaging/threads', (options) {
      posted = Map<String, dynamic>.from(options.data as Map);
      return mock.ok(options, {
        'id': 'th-new',
        'practiceId': 'p1',
        'clientUserId': 'c1',
        'clientName': 'Client Demo',
        'petId': 'pet-1',
      });
    });
    mock.json('GET', RegExp(r'/api/v1/messaging/threads/.+/messages'), data: []);
    mock.install();

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

    await tester.tap(find.byKey(const Key('message_compose_empty_btn')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('message_compose_client_c1')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('message_compose_pet_pet-1')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('message_compose_confirm')));
    await tester.pumpAndSettle();

    expect(posted, isNotNull);
    expect(posted!['clientUserId'], 'c1');
    expect(posted!['petId'], 'pet-1');
    expect(posted!.containsKey('practiceId'), isFalse);
    expect(find.byKey(const Key('message_draft')), findsOneWidget);
  });
}
