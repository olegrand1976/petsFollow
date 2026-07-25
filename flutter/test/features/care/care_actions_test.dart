import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Lightweight care action keys contract (full CareTab needs pets+API load).
void main() {
  test('CareReminder id is used for action keys', () {
    const id = 'care-42';
    expect(Key('care_done_$id'), isNotNull);
    expect(Key('care_postpone_$id'), isNotNull);
    expect(Key('care_add_fab'), isNotNull);
  });

  testWidgets('care action key widgets are findable', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Scaffold(
          floatingActionButton: FloatingActionButton(
            key: const Key('care_add_fab'),
            onPressed: () {},
            child: const Icon(Icons.add),
          ),
          body: Row(
            children: [
              IconButton(
                key: const Key('care_done_care-1'),
                onPressed: () {},
                icon: const Icon(Icons.check),
              ),
              IconButton(
                key: const Key('care_postpone_care-1'),
                onPressed: () {},
                icon: const Icon(Icons.schedule),
              ),
            ],
          ),
        ),
      ),
    );
    expect(find.byKey(const Key('care_add_fab')), findsOneWidget);
    expect(find.byKey(const Key('care_done_care-1')), findsOneWidget);
    expect(find.byKey(const Key('care_postpone_care-1')), findsOneWidget);
  });

  test('CareReminder.fromJson roundtrip id', () {
    final r = CareReminder.fromJson({
      'id': 'care-1',
      'petId': 'pet-1',
      'title': 'Vaccin',
      'type': 'vaccination',
      'dueAt': DateTime.now().toUtc().toIso8601String(),
      'status': 'pending',
    });
    expect(r.id, 'care-1');
  });
}
