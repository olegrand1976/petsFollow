import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/auth/google_auth.dart';
import 'package:petsfollow_mobile/core/auth/google_login_flow.dart';
import 'package:petsfollow_mobile/core/auth/pending_login_hint.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/auth/presentation/forgot_password_screen.dart';
import 'package:petsfollow_mobile/features/auth/presentation/register_screen.dart';
import 'package:petsfollow_mobile/features/legal/domain/legal_document_type.dart';
import 'package:petsfollow_mobile/features/legal/presentation/legal_document_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({
    super.key,
    required this.onLoggedIn,
    this.initialEmail,
    this.infoMessage,
  });

  final VoidCallback onLoggedIn;
  final String? initialEmail;
  final String? infoMessage;

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  // Prefill demo credentials only in debug — never ship them in Play release builds.
  late final TextEditingController email;
  final password = TextEditingController(
    text: kDebugMode ? 'ClientDemo123!' : '',
  );
  final totpCode = TextEditingController();
  String? error;
  String? info;
  bool _busy = false;
  String? _mfaToken;
  bool consent = false;

  bool get _awaiting2FA => _mfaToken != null;

  @override
  void initState() {
    super.initState();
    final hint = PendingLoginHint.take();
    final seedEmail = widget.initialEmail?.trim().isNotEmpty == true
        ? widget.initialEmail!.trim()
        : hint.email;
    email = TextEditingController(
      text: seedEmail?.isNotEmpty == true
          ? seedEmail!
          : (kDebugMode ? 'client.demo@petsfollow.test' : ''),
    );
    info = widget.infoMessage ?? hint.infoMessage;
  }

  @override
  void dispose() {
    email.dispose();
    password.dispose();
    totpCode.dispose();
    super.dispose();
  }

  Future<void> _finishLogin(Map<String, dynamic> data) async {
    if (data['requires2FA'] == true &&
        (data['mfaToken'] as String?)?.isNotEmpty == true) {
      setState(() {
        _mfaToken = data['mfaToken'] as String;
        totpCode.clear();
        error = null;
      });
      return;
    }
    await NotificationService.instance.init();
    widget.onLoggedIn();
  }

  Future<void> submit() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() {
      error = null;
      info = null;
      _busy = true;
    });
    try {
      final data = await ApiClient.instance.login(email.text, password.text);
      if (!mounted) return;
      await _finishLogin(data);
    } catch (e) {
      if (!mounted) return;
      final code = apiErrorCode(e);
      setState(() {
        error = code == 'email_not_verified' ? l10n.emailNotVerified : l10n.loginFailed;
      });
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> submit2FA() async {
    final l10n = AppLocalizations.of(context)!;
    final token = _mfaToken;
    if (token == null) return;
    setState(() {
      error = null;
      _busy = true;
    });
    try {
      final data = await ApiClient.instance.verify2FA(token, totpCode.text.trim());
      if (!mounted) return;
      await _finishLogin(data);
    } catch (_) {
      if (mounted) setState(() => error = l10n.twoFaInvalid);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void reset2FA() {
    setState(() {
      _mfaToken = null;
      totpCode.clear();
      error = null;
    });
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
      // Compte existant : consent ignoré côté API.
      // Compte inconnu : consent=true requis (checkbox ci-dessous).
      final data = await GoogleLoginFlow.signIn(consent: consent);
      if (!mounted) return;
      if (data != null) await _finishLogin(data);
    } on GoogleConsentRequired catch (e) {
      if (!mounted) return;
      // Checkbox non cochée + email Google inconnu → forcer l'acceptation.
      final accepted = await _promptGoogleConsent(l10n);
      if (!mounted) return;
      if (accepted != true) {
        setState(() => error = l10n.registerConsentRequired);
        return;
      }
      setState(() => consent = true);
      try {
        final data = await GoogleLoginFlow.complete(e.idToken, consent: true);
        if (!mounted) return;
        await _finishLogin(data);
      } catch (retryErr) {
        if (!mounted) return;
        setState(() => error = GoogleLoginFlow.errorMessage(l10n, retryErr));
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => error = GoogleLoginFlow.errorMessage(l10n, e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  /// Compte Google inconnu : demander l'acceptation CGU/privacy (RGPD) avant create-if-absent.
  Future<bool?> _promptGoogleConsent(AppLocalizations l10n) {
    final linkStyle = TextStyle(
      color: AppColors.accent,
      decoration: TextDecoration.underline,
      decorationColor: AppColors.accent,
    );
    return showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.registerTitle),
        content: Wrap(
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
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: Text(l10n.proLightAudioConsentAccept),
          ),
        ],
      ),
    );
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

  Future<void> _openRegister() async {
    final result = await Navigator.of(context).push<Object>(
      MaterialPageRoute(builder: (_) => const RegisterScreen()),
    );
    if (!mounted || result == null) return;
    if (result is Map<String, dynamic>) {
      // Inscription via Google : session déjà ouverte (ou 2FA à finaliser).
      await _finishLogin(result);
      return;
    }
    if (result is! String || result.isEmpty) return;
    setState(() {
      email.text = result;
      info = AppLocalizations.of(context)!.registerSuccess;
      error = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Container(
      decoration: BoxDecoration(gradient: AppTheme.loginGradientOf(context)),
      child: Scaffold(
        backgroundColor: Colors.transparent,
        body: SafeArea(
          bottom: false,
          child: SingleChildScrollView(
            padding: scrollPaddingWithSystemBottom(context, all: 24),
            child: _awaiting2FA ? _build2FA(l10n) : _buildCredentials(l10n),
          ),
        ),
      ),
    );
  }

  Widget _buildCredentials(AppLocalizations l10n) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 48),
        const Center(child: PetsLogo(variant: PetsLogoVariant.emblem, height: 72)),
        const SizedBox(height: 24),
        const Center(
          child: PetsLogo(
            variant: PetsLogoVariant.wordmark,
            height: 36,
            excludeSemantics: true,
          ),
        ),
        const SizedBox(height: 8),
        Text(
          l10n.appTagline,
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.bodyLarge,
        ),
        const SizedBox(height: 40),
        TextField(
          key: const Key('login_email'),
          controller: email,
          decoration: InputDecoration(labelText: l10n.email),
        ),
        const SizedBox(height: 12),
        TextField(
          key: const Key('login_password'),
          controller: password,
          obscureText: true,
          decoration: InputDecoration(labelText: l10n.password),
        ),
        Align(
          alignment: Alignment.centerRight,
          child: TextButton(
            onPressed: _busy
                ? null
                : () {
                    Navigator.of(context).push(
                      MaterialPageRoute<void>(
                        builder: (_) => const ForgotPasswordScreen(),
                      ),
                    );
                  },
            child: Text(l10n.forgotPassword),
          ),
        ),
        Align(
          alignment: Alignment.center,
          child: TextButton(
            onPressed: _busy ? null : _openRegister,
            child: Text(l10n.registerCta),
          ),
        ),
        if (info != null) ...[
          const SizedBox(height: 4),
          Text(info!, style: const TextStyle(color: AppColors.accent)),
        ],
        if (error != null) ...[
          const SizedBox(height: 4),
          Text(error!, key: const Key('login_error'), style: const TextStyle(color: AppColors.alert)),
        ],
        const SizedBox(height: 16),
        FilledButton(
          key: const Key('login_submit'),
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
          _buildConsentRow(l10n),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            onPressed: _busy ? null : submitGoogle,
            icon: const Icon(Icons.g_mobiledata, size: 28),
            label: Text(l10n.loginWithGoogle),
          ),
        ],
      ],
    );
  }

  Widget _build2FA(AppLocalizations l10n) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 48),
        const Center(child: PetsLogo(variant: PetsLogoVariant.emblem, height: 72)),
        const SizedBox(height: 24),
        Text(
          l10n.twoFaTitle,
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: 8),
        Text(
          l10n.twoFaSubtitle,
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.bodyMedium,
        ),
        const SizedBox(height: 32),
        TextField(
          controller: totpCode,
          keyboardType: TextInputType.number,
          maxLength: 6,
          autofillHints: const [AutofillHints.oneTimeCode],
          decoration: InputDecoration(labelText: l10n.twoFaCode, counterText: ''),
          onSubmitted: (_) => _busy ? null : submit2FA(),
        ),
        if (error != null) ...[
          const SizedBox(height: 12),
          Text(error!, style: const TextStyle(color: AppColors.alert)),
        ],
        const SizedBox(height: 24),
        FilledButton(
          onPressed: _busy ? null : submit2FA,
          child: Text(l10n.twoFaSubmit),
        ),
        TextButton(
          onPressed: _busy ? null : reset2FA,
          child: Text(l10n.twoFaBack),
        ),
      ],
    );
  }
}
