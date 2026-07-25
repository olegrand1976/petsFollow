import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Opened via `petsfollow://reset-password?token=` (testing / future App Links).
/// Production reset emails use the Pro web URL (`/reset-password?token=`).
class ResetPasswordScreen extends StatefulWidget {
  const ResetPasswordScreen({super.key, required this.initialToken});

  final String initialToken;

  @override
  State<ResetPasswordScreen> createState() => _ResetPasswordScreenState();
}

class _ResetPasswordScreenState extends State<ResetPasswordScreen> {
  final password = TextEditingController();
  final confirm = TextEditingController();
  String? error;
  bool _busy = false;
  bool _done = false;

  bool get _hasToken => widget.initialToken.trim().isNotEmpty;

  @override
  void dispose() {
    password.dispose();
    confirm.dispose();
    super.dispose();
  }

  Future<void> submit() async {
    final l10n = AppLocalizations.of(context)!;
    if (!_hasToken) {
      setState(() => error = l10n.resetPasswordInvalidLink);
      return;
    }
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
      await ApiClient.instance.resetPassword(widget.initialToken.trim(), password.text);
      if (mounted) setState(() => _done = true);
    } catch (_) {
      if (mounted) setState(() => error = l10n.resetPasswordFailed);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Container(
      decoration: const BoxDecoration(gradient: AppTheme.loginGradient),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
        ),
        body: SafeArea(
          bottom: false,
          child: SingleChildScrollView(
            padding: scrollPaddingWithSystemBottom(context, all: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Center(
                    child: PetsLogo(variant: PetsLogoVariant.emblem, height: 64)),
                const SizedBox(height: 24),
                Text(
                  _done ? l10n.resetPasswordDoneTitle : l10n.resetPasswordTitle,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 8),
                Text(
                  !_hasToken
                      ? l10n.resetPasswordInvalidLink
                      : _done
                          ? l10n.resetPasswordDoneSubtitle
                          : l10n.resetPasswordSubtitle,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                if (_hasToken && !_done) ...[
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
                    decoration: InputDecoration(labelText: l10n.confirmNewPassword),
                  ),
                  if (error != null) ...[
                    const SizedBox(height: 12),
                    Text(error!, style: const TextStyle(color: AppColors.alert)),
                  ],
                  const SizedBox(height: 24),
                  FilledButton(
                    onPressed: _busy ? null : submit,
                    child: Text(l10n.resetPasswordSubmit),
                  ),
                ],
                TextButton(
                  onPressed: () => Navigator.of(context).popUntil((r) => r.isFirst),
                  child: Text(l10n.resetPasswordBackToLogin),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
