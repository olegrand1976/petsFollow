import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key, required this.onLoggedIn});
  final VoidCallback onLoggedIn;

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final email = TextEditingController(text: 'client.demo@petsfollow.test');
  final password = TextEditingController(text: 'ClientDemo123!');
  String? error;

  Future<void> submit() async {
    final l10n = AppLocalizations.of(context)!;
    try {
      await ApiClient.instance.login(email.text, password.text);
      await NotificationService.instance.init();
      widget.onLoggedIn();
    } catch (_) {
      if (mounted) setState(() => error = l10n.loginFailed);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Container(
      decoration: const BoxDecoration(gradient: AppTheme.loginGradient),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        body: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SizedBox(height: 48),
                const Center(child: PetsLogo(variant: PetsLogoVariant.emblem, height: 72)),
                const SizedBox(height: 24),
                Center(child: PetsLogo(height: 36)),
                const SizedBox(height: 8),
                Text(
                  l10n.appTagline,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.bodyLarge,
                ),
                const SizedBox(height: 40),
                TextField(controller: email, decoration: InputDecoration(labelText: l10n.email)),
                const SizedBox(height: 12),
                TextField(
                  controller: password,
                  obscureText: true,
                  decoration: InputDecoration(labelText: l10n.password),
                ),
                if (error != null) ...[
                  const SizedBox(height: 12),
                  Text(error!, style: const TextStyle(color: Colors.redAccent)),
                ],
                const SizedBox(height: 24),
                FilledButton(onPressed: submit, child: Text(l10n.login)),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
