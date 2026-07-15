import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

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
      widget.onLoggedIn();
    } catch (_) {
      if (mounted) setState(() => error = l10n.loginFailed);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 48),
              Text(l10n.appTitle, style: Theme.of(context).textTheme.headlineMedium),
              Text(l10n.appTagline),
              const SizedBox(height: 32),
              TextField(controller: email, decoration: InputDecoration(labelText: l10n.email)),
              TextField(
                controller: password,
                obscureText: true,
                decoration: InputDecoration(labelText: l10n.password),
              ),
              if (error != null) Text(error!, style: const TextStyle(color: Colors.redAccent)),
              const SizedBox(height: 16),
              FilledButton(onPressed: submit, child: Text(l10n.login)),
            ],
          ),
        ),
      ),
    );
  }
}
