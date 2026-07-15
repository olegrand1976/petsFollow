import 'package:flutter/material.dart';
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
    try {
      await ApiClient.instance.login(email.text, password.text);
      widget.onLoggedIn();
    } catch (_) {
      setState(() => error = 'Connexion impossible');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 48),
              Text('petsFollow', style: Theme.of(context).textTheme.headlineMedium),
              const Text('Suivi santé de votre animal'),
              const SizedBox(height: 32),
              TextField(controller: email, decoration: const InputDecoration(labelText: 'Email')),
              TextField(controller: password, obscureText: true, decoration: const InputDecoration(labelText: 'Mot de passe')),
              if (error != null) Text(error!, style: const TextStyle(color: Colors.redAccent)),
              const SizedBox(height: 16),
              FilledButton(onPressed: submit, child: const Text('Se connecter')),
            ],
          ),
        ),
      ),
    );
  }
}
