import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/core/auth/google_login_flow.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/legal/domain/legal_document_type.dart';
import 'package:petsfollow_mobile/features/legal/presentation/legal_document_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key});

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final fullName = TextEditingController();
  final email = TextEditingController();
  final password = TextEditingController();
  final confirm = TextEditingController();
  String? error;
  String? info;
  String? success;
  bool _busy = false;
  bool consent = false;

  bool get _isIOS => defaultTargetPlatform == TargetPlatform.iOS;

  @override
  void dispose() {
    fullName.dispose();
    email.dispose();
    password.dispose();
    confirm.dispose();
    super.dispose();
  }

  Future<void> submit() async {
    final l10n = AppLocalizations.of(context)!;
    final name = fullName.text.trim();
    final mail = email.text.trim();
    final pass = password.text;
    if (name.isEmpty || mail.isEmpty || !mail.contains('@')) {
      setState(() => error = l10n.emailRequired);
      return;
    }
    if (pass.length < 8) {
      setState(() => error = l10n.passwordTooShort);
      return;
    }
    if (pass != confirm.text) {
      setState(() => error = l10n.passwordMismatch);
      return;
    }
    if (!consent) {
      setState(() => error = l10n.registerConsentRequired);
      return;
    }
    setState(() {
      error = null;
      info = null;
      success = null;
      _busy = true;
    });
    try {
      await ApiClient.instance.registerClient(
        email: mail,
        password: pass,
        fullName: name,
        locale: LocaleController.instance.locale.languageCode,
        consent: consent,
      );
      if (!mounted) return;
      setState(() => success = l10n.registerSuccess);
    } on DioException catch (e) {
      if (!mounted) return;
      final code = apiErrorCode(e);
      setState(() {
        error = code == 'email_already_exists' || code == 'conflict'
            ? l10n.registerEmailExists
            : l10n.registerFailed;
      });
    } catch (_) {
      if (mounted) setState(() => error = l10n.registerFailed);
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
      info = null;
      _busy = true;
    });
    try {
      final data = await GoogleLoginFlow.signIn();
      if (!mounted) return;
      if (data != null) {
        // L'utilisateur est connecté (ou en attente de 2FA) : le LoginScreen
        // parent finalise via _finishLogin.
        Navigator.of(context).pop(data);
        return;
      }
      setState(() => _busy = false);
    } catch (e) {
      if (!mounted) return;
      setState(() {
        error = GoogleLoginFlow.errorMessage(l10n, e);
        _busy = false;
      });
    }
  }

  void _openLegal(LegalDocumentType type) {
    Navigator.of(context).push(
      MaterialPageRoute<void>(builder: (_) => LegalDocumentScreen(type: type)),
    );
  }

  Widget _buildConsentRow(AppLocalizations l10n) {
    final linkStyle = TextStyle(
      color: AppColors.accent,
      decoration: TextDecoration.underline,
      decorationColor: AppColors.accent,
    );
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Checkbox(
          value: consent,
          onChanged: _busy ? null : (v) => setState(() => consent = v ?? false),
        ),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.only(top: 12),
            child: Wrap(
              children: [
                Text(l10n.registerConsentPrefix),
                GestureDetector(
                  onTap: () => _openLegal(LegalDocumentType.terms),
                  child: Text(l10n.legalTermsTitle, style: linkStyle),
                ),
                Text(l10n.registerConsentMiddle),
                GestureDetector(
                  onTap: () => _openLegal(LegalDocumentType.privacy),
                  child: Text(l10n.legalPrivacyTitle, style: linkStyle),
                ),
                const Text('.'),
              ],
            ),
          ),
        ),
      ],
    );
  }

  void tapApple() {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      error = null;
      info = l10n.appleComingSoon;
    });
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
          title: Text(l10n.registerTitle),
        ),
        body: SafeArea(
          bottom: false,
          child: SingleChildScrollView(
            padding: scrollPaddingWithSystemBottom(context, all: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const PetsLogo(height: 28),
                const SizedBox(height: 16),
                Text(l10n.registerSubtitle, textAlign: TextAlign.center),
                const SizedBox(height: 24),
                if (success != null) ...[
                  Text(success!, style: const TextStyle(color: AppColors.accent)),
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: () => Navigator.of(context).pop(email.text.trim()),
                    child: Text(l10n.registerBackToLogin),
                  ),
                ] else ...[
                  TextField(
                    controller: fullName,
                    textInputAction: TextInputAction.next,
                    decoration: InputDecoration(labelText: l10n.fullName),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: email,
                    keyboardType: TextInputType.emailAddress,
                    textInputAction: TextInputAction.next,
                    decoration: InputDecoration(labelText: l10n.email),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: password,
                    obscureText: true,
                    textInputAction: TextInputAction.next,
                    decoration: InputDecoration(labelText: l10n.password),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: confirm,
                    obscureText: true,
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => submit(),
                    decoration: InputDecoration(labelText: l10n.confirmNewPassword),
                  ),
                  const SizedBox(height: 8),
                  _buildConsentRow(l10n),
                  if (info != null) ...[
                    const SizedBox(height: 12),
                    Text(info!, style: const TextStyle(color: AppColors.accent)),
                  ],
                  if (error != null) ...[
                    const SizedBox(height: 12),
                    Text(error!, style: const TextStyle(color: AppColors.alert)),
                  ],
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: _busy ? null : submit,
                    child: Text(l10n.registerSubmit),
                  ),
                  if (_isIOS || GoogleAuth.isConfigured) ...[
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        const Expanded(child: Divider()),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 12),
                          child: Text(
                            l10n.loginOr,
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                        ),
                        const Expanded(child: Divider()),
                      ],
                    ),
                    const SizedBox(height: 16),
                  ],
                  if (_isIOS)
                    // Variante blanche des guidelines Apple : le fond de
                    // l'écran (loginGradient) est sombre.
                    FilledButton.icon(
                      onPressed: _busy ? null : tapApple,
                      style: FilledButton.styleFrom(
                        backgroundColor: Colors.white,
                        foregroundColor: Colors.black,
                      ),
                      icon: const Icon(Icons.apple, size: 24),
                      label: Text(l10n.loginWithApple),
                    ),
                  if (_isIOS && GoogleAuth.isConfigured) const SizedBox(height: 12),
                  if (GoogleAuth.isConfigured)
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
