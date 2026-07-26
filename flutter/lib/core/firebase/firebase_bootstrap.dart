import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/foundation.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/firebase_options.dart';

/// Initialise Firebase (FCM, Analytics futurs) — **pas** d'authentification Firebase.
/// La connexion reste centralisée via l'API Go + PostgreSQL (`ApiClient.login`).
Future<void> bootstrapFirebase() async {
  if (Firebase.apps.isNotEmpty) return;
  await Firebase.initializeApp(options: DefaultFirebaseOptions.forEnv(AppEnv.value));
  if (kDebugMode) {
    debugPrint(
      'Firebase initialisé (${AppEnv.value}) — auth via API PostgreSQL uniquement',
    );
  }
}
