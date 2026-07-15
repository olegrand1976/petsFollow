import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/app.dart';
import 'package:petsfollow_mobile/core/firebase/firebase_bootstrap.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await bootstrapFirebase();
  runApp(const PetsFollowApp());
}
