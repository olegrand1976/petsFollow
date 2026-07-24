import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
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
  String? success;
  bool _busy = false;

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
    setState(() {
      error = null;
      success = null;
      _busy = true;
    });
    try {
      await ApiClient.instance.registerClient(
        email: mail,
        password: pass,
        fullName: name,
        locale: LocaleController.instance.locale.languageCode,
      );
      if (!mounted) return;
      setState(() => success = l10n.registerSuccess);
    } on DioException catch (e) {
      if (!mounted) return;
      final code = _errorCode(e);
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

  String? _errorCode(DioException e) {
    final data = e.response?.data;
    if (data is Map) {
      final err = data['error'];
      if (err is Map) return err['code'] as String? ?? err['message'] as String?;
      return data['code'] as String?;
    }
    return null;
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
                  if (error != null) ...[
                    const SizedBox(height: 12),
                    Text(error!, style: const TextStyle(color: AppColors.alert)),
                  ],
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: _busy ? null : submit,
                    child: Text(l10n.registerSubmit),
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
