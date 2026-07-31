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

  Future<void> openDialog(WidgetTester tester, {Size? screen}) async {
    if (screen != null) {
      tester.view.physicalSize = screen;
      tester.view.devicePixelRatio = 1.0;
      // Le champ e-mail a autofocus : le clavier est ouvert dès l'affichage et
      // ampute la hauteur disponible pour le contenu du dialogue.
      tester.view.viewInsets = const FakeViewPadding(bottom: 280);
      addTearDown(tester.view.reset);
    }
    final raw = Fixtures.pet(id: petId, name: 'Bella', ownerUserId: userId);
    raw['practiceId'] = 'practice-1';
    final pet = Pet.fromJson(raw);
    expect(pet.isOwner, isTrue);
    expect(pet.isActive, isTrue);
    expect(pet.needsVetLink, isFalse);

    await pumpApp(tester, home: PetDetailScreen(pet: pet));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    final key = Key('pet_send_dossier_$petId');
    if (find.byKey(key).evaluate().isEmpty) {
      await tester.scrollUntilVisible(find.byKey(key), 240,
          scrollable: find.byType(Scrollable).first);
    }
    await tester.ensureVisible(find.byKey(key));
    await tester.pump();
    await tester.tap(find.byKey(key));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('pet_send_dossier_dialog')), findsOneWidget);
  }

  testWidgets('send dossier posts email to API once consent is given', (tester) async {
    final l10n = AppLocalizationsFr();
    await openDialog(tester);

    expect(find.text(l10n.sendDossierPhiWarning), findsOneWidget);
    await tester.enterText(find.byKey(const Key('pet_send_dossier_email')), 'pro@clinic.test');
    await tester.tap(find.byKey(const Key('pet_send_dossier_consent')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('pet_send_dossier_confirm')));
    await tester.pumpAndSettle();

    expect(postedEmail, 'pro@clinic.test');
    expect(find.text(l10n.sendDossierSuccess), findsOneWidget);
  });

  // Données de santé vers un tiers : rien ne part tant que le consentement
  // explicite n'est pas coché.
  testWidgets('send dossier stays blocked without consent', (tester) async {
    await openDialog(tester);

    await tester.enterText(find.byKey(const Key('pet_send_dossier_email')), 'pro@clinic.test');
    await tester.pumpAndSettle();

    final confirm = tester.widget<FilledButton>(
      find.byKey(const Key('pet_send_dossier_confirm')),
    );
    expect(confirm.onPressed, isNull);

    await tester.tap(find.byKey(const Key('pet_send_dossier_confirm')));
    await tester.pumpAndSettle();

    expect(postedEmail, isNull);
    expect(find.byKey(const Key('pet_send_dossier_dialog')), findsOneWidget);
  });

  // Deux débordements sur la largeur Android la plus courante : le dialogue
  // allongé par l'avertissement PHI passait sous le clavier (d'où
  // `scrollable: true`), et l'en-tête « Mes vétérinaires » de l'écran écrasait
  // son titre à zéro pour loger le bouton (d'où le `Flexible`).
  testWidgets('pet detail and send dossier dialog fit a 360 dp screen', (tester) async {
    await openDialog(tester, screen: const Size(360, 560));

    expect(tester.takeException(), isNull);
    expect(find.byKey(const Key('pet_send_dossier_consent')), findsOneWidget);
  });
}
