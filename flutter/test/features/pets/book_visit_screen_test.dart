import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/practice_availability.dart';
import 'package:petsfollow_mobile/core/models/vet_link.dart';
import 'package:petsfollow_mobile/features/pets/presentation/book_visit_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

void main() {
  testWidgets('shows vet picker when several vets', (tester) async {
    final vets = [
      const VetLink(
        practiceId: 'p1',
        vetEmail: 'a@test.com',
        vetFullName: 'Dr Alpha',
        practiceName: 'Cabinet A',
      ),
      const VetLink(
        practiceId: 'p2',
        vetEmail: 'b@test.com',
        vetFullName: 'Dr Beta',
        practiceName: 'Cabinet B',
      ),
    ];

    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: BookVisitScreen(
          petId: 'pet-1',
          petName: 'Rex',
          initialVets: vets,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text(AppLocalizationsFr().calendarSelectVet), findsOneWidget);
    expect(find.text('Dr Alpha'), findsOneWidget);
    expect(find.text('Dr Beta'), findsOneWidget);
    expect(find.byKey(const Key('book_visit_vet_p1')), findsOneWidget);
  });

  testWidgets('filters out vets outside practiceIdFilter', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: BookVisitScreen(
          petId: 'pet-1',
          petName: 'Rex',
          practiceIdFilter: 'p1',
          initialVets: const [
            VetLink(
              practiceId: 'p1',
              vetEmail: 'a@test.com',
              vetFullName: 'Dr Alpha',
              practiceName: 'Cabinet A',
            ),
            VetLink(
              practiceId: 'p2',
              vetEmail: 'b@test.com',
              vetFullName: 'Dr Beta',
              practiceName: 'Cabinet B',
            ),
          ],
          availabilityOverride: const PracticeAvailability(
            enabled: false,
            practicePhone: '01 23 45 67 89',
            practiceName: 'Cabinet A',
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Dr Beta'), findsNothing);
    expect(find.byKey(const Key('book_visit_call_practice')), findsOneWidget);
  });

  testWidgets('single vet with booking disabled shows call CTA', (tester) async {
    final l10n = AppLocalizationsFr();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: BookVisitScreen(
          petId: 'pet-1',
          petName: 'Rex',
          practiceIdFilter: 'p1',
          initialVets: const [
            VetLink(
              practiceId: 'p1',
              vetEmail: 'a@test.com',
              vetFullName: 'Dr Alpha',
              practiceName: 'Cabinet A',
            ),
          ],
          availabilityOverride: const PracticeAvailability(
            enabled: false,
            practicePhone: '01 23 45 67 89',
            practiceName: 'Cabinet A',
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text(l10n.calendarBookingDisabled), findsOneWidget);
    expect(find.byKey(const Key('book_visit_call_practice')), findsOneWidget);
    expect(find.textContaining('01 23 45 67 89'), findsOneWidget);
  });

  testWidgets('enabled slots are listed', (tester) async {
    final l10n = AppLocalizationsFr();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: BookVisitScreen(
          petId: 'pet-1',
          petName: 'Rex',
          practiceId: 'p1',
          rescheduleVisitId: 'visit-1',
          availabilityOverride: PracticeAvailability(
            enabled: true,
            practicePhone: '01 23 45 67 89',
            practiceName: 'Cabinet A',
            slots: [
              PracticeAvailabilitySlot(start: DateTime.parse('2099-06-01T10:00:00.000Z')),
            ],
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text(l10n.calendarPickSlot), findsOneWidget);
  });
}
