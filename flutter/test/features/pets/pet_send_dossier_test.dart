import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_detail_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/fixtures.dart';
import '../../helpers/pump_app.dart';

void main() {
  const petId = 'pet-bella';
  const userId = 'user-1';
  late Interceptor mockInterceptor;
  String? postedEmail;

  setUp(() {
    postedEmail = null;
    ApiClient.instance.dio.interceptors.clear();
    ApiClient.instance.userId = userId;
    ApiClient.instance.token = null;
    mockInterceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        if (options.method == 'GET' && options.path.contains('/me/vets')) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {'data': <dynamic>[]},
            ),
          );
          return;
        }
        if (options.method == 'GET' && options.path.contains('/heartrate/sessions')) {
          handler.resolve(
            Response(requestOptions: options, statusCode: 200, data: {'data': []}),
          );
          return;
        }
        if (options.method == 'GET' && options.path.contains('/weights')) {
          handler.resolve(
            Response(requestOptions: options, statusCode: 200, data: {'data': []}),
          );
          return;
        }
        if (options.method == 'POST' &&
            options.path.contains('/pets/$petId/dossier-shares')) {
          postedEmail = (options.data as Map?)?['email'] as String?;
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 201,
              data: {
                'data': {
                  'ok': true,
                  'expiresAt': '2026-07-28T12:00:00Z',
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
    ApiClient.instance.userId = null;
  });

  testWidgets('send dossier posts email to API', (tester) async {
    final raw = Fixtures.pet(id: petId, name: 'Bella', ownerUserId: userId);
    raw['practiceId'] = 'practice-1';
    final pet = Pet.fromJson(raw);
    expect(pet.isOwner, isTrue);
    expect(pet.isActive, isTrue);
    expect(pet.needsVetLink, isFalse);

    await pumpApp(tester, home: PetDetailScreen(pet: pet));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    final l10n = AppLocalizationsFr();
    final key = Key('pet_send_dossier_$petId');
    await tester.ensureVisible(find.byKey(key));
    await tester.pump();

    expect(find.text(l10n.sendDossierToPro), findsOneWidget);
    await tester.tap(find.byKey(key));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('pet_send_dossier_dialog')), findsOneWidget);
    await tester.enterText(find.byKey(const Key('pet_send_dossier_email')), 'pro@clinic.test');
    await tester.tap(find.byKey(const Key('pet_send_dossier_confirm')));
    await tester.pumpAndSettle();

    expect(postedEmail, 'pro@clinic.test');
    expect(find.text(l10n.sendDossierSuccess), findsOneWidget);
  });
}
