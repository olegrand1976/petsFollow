import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class ForceChangePasswordScreen extends StatefulWidget {
  const ForceChangePasswordScreen({super.key, required this.onChanged});
  final VoidCallback onChanged;

  @override
  State<ForceChangePasswordScreen> createState() =>
      _ForceChangePasswordScreenState();
}

class _ForceChangePasswordScreenState extends State<ForceChangePasswordScreen> {
  final password = TextEditingController();
  final confirm = TextEditingController();
  String? error;
  bool _busy = false;

  Future<void> submit() async {
    final l10n = AppLocalizations.of(context)!;
    if (password.text.length < 8) {
      setState(() => error = l10n.passwordTooShort);
      return;
    }
    if (password.text != confirm.text) {
      setState(() => error = l10n.passwordMismatch);
      return;
    }
    setState(() {
      error = null;
      _busy = true;
    });
    try {
      await ApiClient.instance.changePassword('', password.text);
      if (mounted) widget.onChanged();
    } catch (_) {
      if (mounted) setState(() => error = l10n.passwordChangeFailed);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  void dispose() {
    password.dispose();
    confirm.dispose();
    super.dispose();
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
                const Center(
                    child: PetsLogo(variant: PetsLogoVariant.emblem, height: 72)),
                const SizedBox(height: 24),
                Text(
                  l10n.forceChangePasswordTitle,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 8),
                Text(
                  l10n.forceChangePasswordSubtitle,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                const SizedBox(height: 32),
                TextField(
                  controller: password,
                  obscureText: true,
                  decoration: InputDecoration(labelText: l10n.newPassword),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: confirm,
                  obscureText: true,
                  decoration:
                      InputDecoration(labelText: l10n.confirmNewPassword),
                ),
                if (error != null) ...[
                  const SizedBox(height: 12),
                  Text(error!,
                      style: TextStyle(color: Theme.of(context).colorScheme.error)),
                ],
                const SizedBox(height: 24),
                FilledButton(
                  onPressed: _busy ? null : submit,
                  child: Text(l10n.forceChangePasswordSubmit),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
