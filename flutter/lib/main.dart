import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/app.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/core/firebase/firebase_bootstrap.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/notifications/fcm_background.dart';
import 'package:petsfollow_mobile/core/theme/theme_controller.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  AppEnv.validate();
  await bootstrapFirebase();
  // Prefs before first frame — avoid locale/theme flash.
  await Future.wait([
    LocaleController.instance.load(),
    ThemeController.instance.load(),
  ]);
  FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);
  runApp(const PetsFollowApp());
}
