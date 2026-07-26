import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/vets/presentation/my_vets_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/mock_api.dart';
import '../../helpers/pump_app.dart';

void main() {
  late MockApi mock;
  var inviteCalled = false;
  var suggestCalled = false;
  Map<String, dynamic>? inviteBody;
  Map<String, dynamic>? suggestBody;

  setUp(() {
    inviteCalled = false;
    suggestCalled = false;
    inviteBody = null;
    suggestBody = null;
    mock = MockApi();
    mock.json('GET', '/api/v1/me/vets', data: []);
    mock.on('GET', RegExp(r'/api/v1/me/vets/lookup'), (options) {
      final q = options.uri.queryParameters['q'] ?? '';
      final hits = q.toLowerCase().contains('vetplus')
          ? [
              {
                'vetUserId': 'vet-1',
                'practiceId': 'pr-1',
                'practiceName': 'Cabinet VetPlus Demo',
                'vetFullName': 'Dr Demo',
                'vetEmail': 'vet.demo@petsfollow.test',
              }
            ]
          : <Map<String, dynamic>>[];
      return mock.ok(options, hits);
    });
    mock.on('POST', '/api/v1/me/vets/invite', (options) {
      inviteCalled = true;
      inviteBody = Map<String, dynamic>.from(options.data as Map);
      return mock.ok(options, {
        'found': true,
        'status': 'pending',
        'practiceName': 'Cabinet VetPlus Demo',
      });
    });
    mock.on('POST', '/api/v1/me/vets/suggest', (options) {
      suggestCalled = true;
      suggestBody = Map<String, dynamic>.from(options.data as Map);
      return mock.ok(
        options,
        {'status': 'suggested', 'found': false, 'leadId': 'lead-1'},
        status: 201,
      );
    });
    mock.install();
  });

  tearDown(() => mock.uninstall());

  testWidgets('lookup hit invites by vetUserId; suggest posts email+phone', (tester) async {
    final l10n = AppLocalizationsFr();
    await pumpApp(tester, home: const MyVetsScreen());
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    expect(find.byKey(const Key('add_vet_search')), findsOneWidget);

    await tester.enterText(find.byKey(const Key('add_vet_search')), 'VetPlus');
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pump();

    final hit = find.byKey(const Key('add_vet_hit_vet-1'));
    expect(hit, findsOneWidget);
    await tester.tap(hit);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(inviteCalled, isTrue);
    expect(inviteBody?['vetUserId'], 'vet-1');
    expect(find.textContaining('Cabinet VetPlus Demo'), findsWidgets);

    await tester.tap(find.byKey(const Key('add_vet_not_listed')));
    await tester.pump();
    expect(find.text(l10n.addVetSuggestTitle), findsOneWidget);

    await tester.enterText(find.byKey(const Key('add_vet_suggest_email')), 'new.vet@clinic.test');
    await tester.enterText(find.byKey(const Key('add_vet_suggest_phone')), '+32470000000');
    final submit = find.byKey(const Key('add_vet_suggest_submit'));
    await tester.ensureVisible(submit);
    await tester.pump();
    await tester.tap(submit);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 200));

    expect(suggestCalled, isTrue);
    expect(suggestBody?['email'], 'new.vet@clinic.test');
    expect(suggestBody?['phone'], '+32470000000');
    // Form closes after successful suggest.
    expect(find.byKey(const Key('add_vet_suggest_submit')), findsNothing);
  });
}
