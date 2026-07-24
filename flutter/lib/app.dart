import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/deeplink/payment_deeplink.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/features/auth/presentation/force_change_password_screen.dart';
import 'package:petsfollow_mobile/features/auth/presentation/login_screen.dart';
import 'package:petsfollow_mobile/features/shell/presentation/commercial_field_shell_screen.dart';
import 'package:petsfollow_mobile/features/shell/presentation/main_shell_screen.dart';
import 'package:petsfollow_mobile/features/shell/presentation/pro_light_shell_screen.dart';
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
    NotificationService.instance.init();
    AppDeepLink.instance.start();
  }

  @override
  void dispose() {
    LocaleController.instance.removeListener(_onLocaleChanged);
    AppDeepLink.instance.dispose();
    super.dispose();
  }

  void _onLocaleChanged() => setState(() {});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'petsFollow',
      theme: buildAppTheme(),
      navigatorKey: PushNavigation.instance.navigatorKey,
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
  bool _mustChangePassword = false;
  int _petsRefreshTick = 0;
  int _loginHintTick = 0;

  @override
  void initState() {
    super.initState();
    ApiClient.instance.onSessionInvalidated = _onSessionInvalidated;
    ApiClient.instance.onSessionEstablished = _onAuthChanged;
    AppDeepLink.instance.onPaymentSuccess = () {
      if (mounted) setState(() => _petsRefreshTick++);
    };
    AppDeepLink.instance.onLoginHint = () {
      if (mounted) setState(() => _loginHintTick++);
    };
    _bootstrap();
  }

  @override
  void dispose() {
    if (ApiClient.instance.onSessionInvalidated == _onSessionInvalidated) {
      ApiClient.instance.onSessionInvalidated = null;
    }
    if (ApiClient.instance.onSessionEstablished == _onAuthChanged) {
      ApiClient.instance.onSessionEstablished = null;
    }
    if (AppDeepLink.instance.onLoginHint != null) {
      AppDeepLink.instance.onLoginHint = null;
    }
    super.dispose();
  }

  void _onSessionInvalidated() {
    if (!mounted) return;
    PushNavigation.instance.navigatorKey.currentState?.popUntil((r) => r.isFirst);
    setState(() {
      _mustChangePassword = false;
    });
  }

  Future<void> _bootstrap() async {
    await ApiClient.instance.restoreSession();
    if (ApiClient.instance.token != null) {
      await NotificationService.instance.init();
      // 401 on /me → interceptor clears token + notifies → login.
      _mustChangePassword = await ApiClient.instance.mustChangePassword();
    }
    if (mounted) setState(() => _ready = true);
  }

  Future<void> _onAuthChanged() async {
    var mustChange = false;
    if (ApiClient.instance.token != null) {
      mustChange = await ApiClient.instance.mustChangePassword();
    }
    if (mounted) {
      setState(() => _mustChangePassword = mustChange);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!_ready) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (ApiClient.instance.token == null) {
      return LoginScreen(
        key: ValueKey('login-$_loginHintTick'),
        onLoggedIn: _onAuthChanged,
      );
    }
    if (_mustChangePassword) {
      return ForceChangePasswordScreen(onChanged: _onAuthChanged);
    }
    if (ApiClient.instance.userRole == 'care_pro' ||
        ApiClient.instance.userRole == 'vet') {
      return ProLightShellScreen(onLogout: _onAuthChanged);
    }
    if (ApiClient.instance.userRole == 'commercial' ||
        ApiClient.instance.userRole == 'commercial_manager') {
      return CommercialFieldShellScreen(onLogout: _onAuthChanged);
    }
    if (ApiClient.instance.userRole == 'client') {
      return MainShellScreen(
        onLogout: _onAuthChanged,
        billingRefreshTick: _petsRefreshTick,
      );
    }
    return _UnsupportedRoleScreen(onLogout: _onAuthChanged);
  }
}

class _UnsupportedRoleScreen extends StatelessWidget {
  const _UnsupportedRoleScreen({required this.onLogout});
  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(l10n.unsupportedRoleApp, textAlign: TextAlign.center),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: () async {
                  await ApiClient.instance.logout();
                  onLogout();
                },
                child: Text(l10n.logout),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
