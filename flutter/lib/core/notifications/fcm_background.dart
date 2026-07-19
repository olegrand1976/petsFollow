import 'package:firebase_messaging/firebase_messaging.dart';

/// Top-level background handler required by firebase_messaging.
@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  // Data-only / system tray display is handled by the OS when a notification
  // payload is present. No Dart UI work here.
}
