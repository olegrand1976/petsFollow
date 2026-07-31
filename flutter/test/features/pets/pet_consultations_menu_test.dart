import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_detail_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/pet_timeline_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations_fr.dart';

import '../../helpers/fixtures.dart';
import '../../helpers/pump_app.dart';

void main() {
  const petId = 'pet-bella';
  const userId = 'user-1';
  late Interceptor mockInterceptor;

  setUp(() {
    ApiClient.instance.dio.interceptors.clear();
    ApiClient.instance.userId = userId;
    ApiClient.instance.token = null;
    mockInterceptor = InterceptorsWrapper(
      onRequest: (options, handler) {
        final path = options.uri.path.isNotEmpty ? options.uri.path : options.path;
        if (options.method == 'GET' && path.contains('/me/vets')) {
          handler.resolve(
            Response(requestOptions: options, statusCode: 200, data: {'data': <dynamic>[]}),
          );
          return;
        }
        if (options.method == 'GET' && path.contains('/heartrate/sessions')) {
          handler.resolve(
            Response(requestOptions: options, statusCode: 200, data: {'data': []}),
          );
          return;
        }
        if (options.method == 'GET' && path.contains('/weights')) {
          handler.resolve(
            Response(requestOptions: options, statusCode: 200, data: {'data': []}),
          );
          return;
        }
        if (options.method == 'GET' && path.contains('/pets/$petId/visits')) {
          handler.resolve(
            Response(requestOptions: options, statusCode: 200, data: {'data': <dynamic>[]}),
          );
          return;
        }
        if (options.method == 'GET' && path.contains('/pets/$petId/timeline')) {
          handler.resolve(
            Response(requestOptions: options, statusCode: 200, data: {'data': <dynamic>[]}),
          );
          return;
        }
        if (options.method == 'GET' &&
            (path.endsWith('/pets/$petId') || path.contains('/pets/$petId?'))) {
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {
                  'id': petId,
                  'name': 'Bella',
                  'ownerUserId': userId,
                  'canWriteNotes': true,
                },
              },
            ),
          );
          return;
        }
        handler.reject(
          DioException(
            requestOptions: options,
            error: 'unmocked ${options.method} $path',
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

  testWidgets('pet detail shows Consultations menu and opens timeline', (tester) async {
    final raw = Fixtures.pet(id: petId, name: 'Bella', ownerUserId: userId);
    raw['practiceId'] = 'practice-1';
    final pet = Pet.fromJson(raw);

    await pumpApp(tester, home: PetDetailScreen(pet: pet));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    final key = Key('pet_consultations_$petId');
    if (find.byKey(key).evaluate().isEmpty) {
      await tester.scrollUntilVisible(
        find.byKey(key),
        240,
        scrollable: find.byType(Scrollable).first,
      );
    }
    expect(find.byKey(key), findsOneWidget);
    expect(find.text(AppLocalizationsFr().consultationsHistory), findsOneWidget);

    await tester.ensureVisible(find.byKey(key));
    await tester.tap(find.byKey(key));
    await tester.pumpAndSettle();

    expect(find.byType(PetTimelineScreen), findsOneWidget);
    expect(find.text(AppLocalizationsFr().visitHistory), findsWidgets);
  });
}
