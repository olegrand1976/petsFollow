import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/invite/presentation/app_invite_qr_screen.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;

  setUp(() {
    mock = MockApi();
    mock.install();
  });

  tearDown(() {
    mock.uninstall();
  });

  testWidgets('commercial invite shows dual cabinet/client copy CTAs', (tester) async {
    mock.json('GET', r'/api/v1/me/app-invite', data: {
      'code': 'ABCD2345',
      'role': 'commercial',
      'inviteUrl': 'https://pro.example/invite/ABCD2345',
      'vetRegisterUrl': 'https://pro.example/register?invite=ABCD2345',
      'displayName': 'Camille Demo',
      'qrCodeDataUrl': null,
    });

    await pumpApp(tester, home: const AppInviteQrScreen());
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('app_invite_code')), findsOneWidget);
    expect(find.byKey(const Key('app_invite_copy_vet')), findsOneWidget);
    expect(find.byKey(const Key('app_invite_copy_client')), findsOneWidget);
  });

  testWidgets('client invite shows single copy CTA without cabinet link', (tester) async {
    mock.json('GET', r'/api/v1/me/app-invite', data: {
      'code': 'CLNT9XYZ',
      'role': 'client',
      'inviteUrl': 'https://pro.example/invite/CLNT9XYZ',
      'displayName': 'Marie Demo',
      'qrCodeDataUrl': null,
    });

    await pumpApp(tester, home: const AppInviteQrScreen());
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('app_invite_code')), findsOneWidget);
    expect(find.byKey(const Key('app_invite_copy')), findsOneWidget);
    expect(find.byKey(const Key('app_invite_copy_vet')), findsNothing);
    expect(find.byKey(const Key('app_invite_copy_client')), findsNothing);
  });
}
