import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key, required this.onLoggedIn});
  final VoidCallback onLoggedIn;

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  // Prefill demo credentials only in debug — never ship them in Play release builds.
  final email = TextEditingController(
    text: kDebugMode ? 'client.demo@petsfollow.test' : '',
  );
  final password = TextEditingController(
    text: kDebugMode ? 'ClientDemo123!' : '',
  );
  String? error;
  bool _busy = false;

  Future<void> submit() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      error = null;
      _busy = true;
    });
    try {
      await ApiClient.instance.login(email.text, password.text);
      await NotificationService.instance.init();
      widget.onLoggedIn();
    } catch (_) {
      if (mounted) setState(() => error = l10n.loginFailed);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> submitGoogle() async {
    final l10n = AppLocalizations.of(context)!;
    if (!GoogleAuth.isConfigured) {
      setState(() => error = l10n.googleNotConfigured);
      return;
    }
    setState(() {
      error = null;
      _busy = true;
    });
    try {
      final idToken = await GoogleAuth.signInForIdToken();
      if (idToken == null) {
        if (mounted) setState(() => _busy = false);
        return;
      }
      await ApiClient.instance.loginWithGoogle(idToken);
      await NotificationService.instance.init();
      widget.onLoggedIn();
    } on DioException catch (e) {
      if (!mounted) return;
      setState(() => error = _googleErrorMessage(l10n, e));
    } catch (_) {
      if (mounted) setState(() => error = l10n.googleLoginFailed);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  String _googleErrorMessage(AppLocalizations l10n, DioException e) {
    final data = e.response?.data;
    String? code;
    if (data is Map) {
      final err = data['error'];
      if (err is Map) {
        code = err['code'] as String?;
      } else {
        code = data['code'] as String?;
      }
    }
    switch (code) {
      case 'not_configured':
        return l10n.googleNotConfigured;
      case 'google_client_not_found':
        return l10n.googleClientNotFound;
      case 'google_client_only':
      case 'google_pro_only':
        return l10n.googleWrongAudience;
      default:
        return l10n.googleLoginFailed;
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
          bottom: false,
          child: SingleChildScrollView(
            padding: scrollPaddingWithSystemBottom(context, all: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SizedBox(height: 48),
                const Center(child: PetsLogo(variant: PetsLogoVariant.emblem, height: 72)),
                const SizedBox(height: 24),
                const Center(child: PetsLogo(height: 36)),
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
                  Text(error!, style: const TextStyle(color: AppColors.alert)),
                ],
                const SizedBox(height: 24),
                FilledButton(
                  onPressed: _busy ? null : submit,
                  child: Text(l10n.login),
                ),
                if (GoogleAuth.isConfigured) ...[
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      const Expanded(child: Divider()),
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 12),
                        child: Text(l10n.loginOr, style: Theme.of(context).textTheme.bodySmall),
                      ),
                      const Expanded(child: Divider()),
                    ],
                  ),
                  const SizedBox(height: 16),
                  OutlinedButton.icon(
                    onPressed: _busy ? null : submitGoogle,
                    icon: const Icon(Icons.g_mobiledata, size: 28),
                    label: Text(l10n.loginWithGoogle),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
