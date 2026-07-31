import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/messaging/presentation/messaging_screen.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.json('GET', '/api/v1/me', data: {
      'userId': 'user-1',
      'email': 'client.demo@petsfollow.test',
      'role': 'client',
    });
    mock.json('GET', '/api/v1/messaging/threads', data: [
      {
        'id': 'thread-1',
        'practiceId': 'practice-1',
        'clientUserId': 'user-1',
        'vetUserId': 'vet-1',
        'petId': 'pet-1',
        'lastMessagePreview': 'Bonjour',
        'unreadCount': 0,
      },
    ]);
    mock.json('GET', RegExp(r'/api/v1/messaging/threads/.+/messages'), data: []);
    mock.json('POST', RegExp(r'/api/v1/messaging/threads/.+/read'), data: {});
    mock.json('GET', '/api/v1/me/vets', data: [
      {
        'practiceId': 'practice-1',
        'practiceName': 'VetPlus',
        'vetFullName': 'Dr Demo',
        'status': 'active',
      },
    ]);
    mock.install();
  });

  tearDown(() => mock.uninstall());

  testWidgets('attach sheet opens camera/gallery after photo choice', (tester) async {
    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(tester, home: const MessagingScreen());
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('message_attach_btn')), findsOneWidget);
    await tester.tap(find.byKey(const Key('message_attach_btn')));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('message_attach_photo')), findsOneWidget);
    expect(find.byKey(const Key('message_attach_video')), findsOneWidget);

    await tester.tap(find.byKey(const Key('message_attach_photo')));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('message_attach_camera')), findsOneWidget);
    expect(find.byKey(const Key('message_attach_gallery')), findsOneWidget);
    expect(find.text('Prendre une photo'), findsOneWidget);
  });

  testWidgets('video attach sheet shows takeVideo label', (tester) async {
    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await pumpApp(tester, home: const MessagingScreen());
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('message_attach_btn')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('message_attach_video')));
    await tester.pumpAndSettle();

    expect(find.text('Filmer une vidéo'), findsOneWidget);
    expect(find.byKey(const Key('message_attach_gallery')), findsOneWidget);
  });
}
