import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Opened via `petsfollow://confirm-email?token=`.
/// Production emails use the Pro web URL; this screen covers app / testing links.
class ConfirmEmailScreen extends StatefulWidget {
  const ConfirmEmailScreen({super.key, required this.token});

  final String token;

  @override
  State<ConfirmEmailScreen> createState() => _ConfirmEmailScreenState();
}

class _ConfirmEmailScreenState extends State<ConfirmEmailScreen> {
  bool _loading = true;
  bool _done = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _confirm());
  }

  Future<void> _confirm() async {
    final l10n = AppLocalizations.of(context)!;
    final token = widget.token.trim();
    if (token.isEmpty) {
      setState(() {
        _loading = false;
        _error = l10n.confirmEmailInvalidLink;
      });
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ApiClient.instance.confirmEmail(token);
      // Session + onSessionEstablished are handled inside confirmEmail → _completeLogin.
      if (!mounted) return;
      setState(() {
        _loading = false;
        _done = true;
      });
      // Let AuthGate rebuild to the shell, then drop this route.
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        Navigator.of(context).popUntil((r) => r.isFirst);
      });
    } on DioException catch (e) {
      if (!mounted) return;
      final code = _errorCode(e);
      setState(() {
        _loading = false;
        _error = code == 'not_found' || code == 'invalid_confirm_link'
            ? l10n.confirmEmailInvalidLink
            : l10n.confirmEmailFailed;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = l10n.confirmEmailFailed;
      });
    }
  }

  String? _errorCode(DioException e) {
    final data = e.response?.data;
    if (data is Map) {
      final err = data['error'];
      if (err is Map) {
        return (err['code'] ?? err['msgKey'] ?? err['message'])?.toString();
      }
      return data['code']?.toString();
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
        ),
        body: SafeArea(
          bottom: false,
          child: SingleChildScrollView(
            padding: scrollPaddingWithSystemBottom(context, all: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Center(
                  child: PetsLogo(variant: PetsLogoVariant.emblem, height: 64),
                ),
                const SizedBox(height: 24),
                Text(
                  _done
                      ? l10n.confirmEmailDoneTitle
                      : _error != null
                          ? l10n.confirmEmailFailedTitle
                          : l10n.confirmEmailTitle,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 12),
                if (_loading) ...[
                  Text(
                    l10n.confirmEmailLoading,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 24),
                  const Center(child: CircularProgressIndicator()),
                ] else if (_done) ...[
                  Text(
                    l10n.confirmEmailDoneSubtitle,
                    textAlign: TextAlign.center,
                  ),
                ] else ...[
                  Text(
                    _error ?? l10n.confirmEmailFailed,
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: AppColors.alert),
                  ),
                  const SizedBox(height: 24),
                  FilledButton(
                    onPressed: _confirm,
                    child: Text(l10n.retryAction),
                  ),
                  TextButton(
                    onPressed: () => Navigator.of(context).popUntil((r) => r.isFirst),
                    child: Text(l10n.confirmEmailBackToLogin),
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
