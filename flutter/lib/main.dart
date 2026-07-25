import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/app.dart';
import 'package:petsfollow_mobile/core/firebase/firebase_bootstrap.dart';
import 'package:petsfollow_mobile/core/notifications/fcm_background.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await bootstrapFirebase();
  FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);
  runApp(const PetsFollowApp());
}
