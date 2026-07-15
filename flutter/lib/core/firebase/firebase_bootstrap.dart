import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/foundation.dart';
import 'package:petsfollow_mobile/firebase_options.dart';

/// Initialise Firebase (FCM, Analytics futurs) — **pas** d'authentification Firebase.
/// La connexion reste centralisée via l'API Go + PostgreSQL (`ApiClient.login`).
Future<void> bootstrapFirebase() async {
  if (Firebase.apps.isNotEmpty) return;
  await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);
  if (kDebugMode) {
    debugPrint('Firebase initialisé — auth via API PostgreSQL uniquement');
  }
}
