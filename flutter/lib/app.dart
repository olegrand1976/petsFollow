import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/features/auth/presentation/login_screen.dart';
import 'package:petsfollow_mobile/features/home/presentation/home_screen.dart';

class PetsFollowApp extends StatelessWidget {
  const PetsFollowApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'petsFollow',
      theme: buildAppTheme(),
      home: const AuthGate(),
    );
  }
}

class AuthGate extends StatefulWidget {
  const AuthGate({super.key});

  @override
  State<AuthGate> createState() => _AuthGateState();
}

class _AuthGateState extends State<AuthGate> {
  @override
  void initState() {
    super.initState();
    ApiClient.instance.loadToken();
  }

  @override
  Widget build(BuildContext context) {
    if (ApiClient.instance.token == null) {
      return LoginScreen(onLoggedIn: () => setState(() {}));
    }
    return const HomeScreen();
  }
}
