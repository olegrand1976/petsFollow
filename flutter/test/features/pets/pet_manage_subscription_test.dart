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
  var portalCalled = false;

  setUp(() {
    portalCalled = false;
    ApiClient.instance.dio.interceptors.clear();
    ApiClient.instance.userId = userId;
    ApiClient.instance.token = null; // FeatureModulesController skips fetch
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
        if (options.method == 'POST' &&
            options.path.contains('/pets/$petId/billing/portal')) {
          portalCalled = true;
          handler.resolve(
            Response(
              requestOptions: options,
              statusCode: 200,
              data: {
                'data': {
                  'url':
                      'http://localhost:8291/api/v1/billing/dev/mock-portal?customer=cus_mock',
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

  testWidgets('manage subscription POSTs billing portal', (tester) async {
    final pet = Pet.fromJson(
      Fixtures.pet(id: petId, name: 'Bella', ownerUserId: userId),
    );
    expect(pet.isOwner, isTrue);
    expect(pet.isActive && pet.entitlement!.isSubscription, isTrue);

    await pumpApp(tester, home: PetDetailScreen(pet: pet));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    final l10n = AppLocalizationsFr();
    final key = Key('pet_manage_subscription_$petId');
    await tester.scrollUntilVisible(
      find.byKey(key),
      300,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pump();

    expect(find.text(l10n.manageSubscription), findsOneWidget);
    await tester.tap(find.byKey(key));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(portalCalled, isTrue);
  });
}
