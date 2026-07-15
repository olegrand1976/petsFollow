import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/notifications/reminder_controller.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/features/auth/presentation/login_screen.dart';
import 'package:petsfollow_mobile/features/home/presentation/home_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class PetsFollowApp extends StatefulWidget {
  const PetsFollowApp({super.key});

  @override
  State<PetsFollowApp> createState() => _PetsFollowAppState();
}

class _PetsFollowAppState extends State<PetsFollowApp> {
  @override
  void initState() {
    super.initState();
    LocaleController.instance.addListener(_onLocaleChanged);
    LocaleController.instance.load();
    ReminderController.instance.init();
  }

  @override
  void dispose() {
    LocaleController.instance.removeListener(_onLocaleChanged);
    super.dispose();
  }

  void _onLocaleChanged() => setState(() {});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'petsFollow',
      theme: buildAppTheme(),
      locale: LocaleController.instance.locale,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: AppLocalizations.supportedLocales,
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
  bool _ready = false;

  @override
  void initState() {
    super.initState();
    _bootstrap();
  }

  Future<void> _bootstrap() async {
    await ApiClient.instance.restoreSession();
    if (mounted) setState(() => _ready = true);
  }

  void _onAuthChanged() => setState(() {});

  @override
  Widget build(BuildContext context) {
    if (!_ready) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (ApiClient.instance.token == null) {
      return LoginScreen(onLoggedIn: _onAuthChanged);
    }
    return HomeScreen(onLogout: _onAuthChanged);
  }
}
