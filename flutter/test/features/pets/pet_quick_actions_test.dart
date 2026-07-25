import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_quick_actions.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

void main() {
  const petId = 'pet-1';
  late Interceptor mockInterceptor;
  Map<String, dynamic>? lastWeightBody;

  setUp(() {
    lastWeightBody = null;
    ApiClient.instance.dio.interceptors.clear();
    mockInterceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        if (options.method == 'POST' &&
            options.path.contains('/pets/$petId/weights')) {
          lastWeightBody = Map<String, dynamic>.from(
            options.data as Map<String, dynamic>,
          );
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 201,
              data: {
                'data': {
                  'id': 'w-1',
                  'petId': petId,
                  'weightKg': lastWeightBody!['weightKg'],
                  'recordedAt': DateTime.now().toUtc().toIso8601String(),
                },
              },
            ),
          );
          return;
        }
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked ${options.method} ${options.path}',
          ),
        );
      },
    );
    ApiClient.instance.dio.interceptors.add(mockInterceptor);
  });

  tearDown(() {
    ApiClient.instance.dio.interceptors.remove(mockInterceptor);
  });

  testWidgets('PetQuickActions exposes heart and weight keys per pet', (
    tester,
  ) async {
    var heartTapped = false;
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Scaffold(
          body: PetQuickActions(
            petId: petId,
            onHeartRate: () => heartTapped = true,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final l10n = AppLocalizationsFr();
    expect(find.byKey(Key('pet_action_heartrate_$petId')), findsOneWidget);
    expect(find.byKey(Key('pet_action_weight_$petId')), findsOneWidget);
    expect(find.text(l10n.heartRateShort), findsOneWidget);
    expect(find.text(l10n.weightShort), findsOneWidget);

    await tester.tap(find.byKey(Key('pet_action_heartrate_$petId')));
    await tester.pump();
    expect(heartTapped, isTrue);
  });

  testWidgets('weight sheet rejects below 0.01 and saves valid kg', (
    tester,
  ) async {
    var saved = false;
    final l10n = AppLocalizationsFr();
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('fr'),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Scaffold(
          body: PetQuickActions(
            petId: petId,
            onHeartRate: () {},
            onWeightRecorded: () => saved = true,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(Key('pet_action_weight_$petId')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('weight_sheet')), findsOneWidget);

    await tester.enterText(find.byKey(const Key('weight_kg_field')), '0.005');
    await tester.tap(find.byKey(const Key('weight_save_btn')));
    await tester.pump();
    expect(find.byKey(const Key('weight_error')), findsOneWidget);
    expect(find.text(l10n.weightInvalid), findsOneWidget);
    expect(lastWeightBody, isNull);

    await tester.enterText(find.byKey(const Key('weight_kg_field')), '12,5');
    await tester.enterText(
      find.byKey(const Key('weight_comment_field')),
      'après balade',
    );
    await tester.tap(find.byKey(const Key('weight_save_btn')));
    await tester.pump(); // start async save
    await tester.pump(); // resolve Dio + Navigator.pop
    await tester.pump(const Duration(seconds: 1)); // sheet dismiss animation

    expect(lastWeightBody?['weightKg'], 12.5);
    expect(lastWeightBody?['comment'], 'après balade');
    expect(saved, isTrue);
    expect(find.text(l10n.weightSentToVet), findsOneWidget);
  });
}
