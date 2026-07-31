import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/app_theme.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/core/widgets/pets_logo.dart';
import 'package:petsfollow_mobile/features/legal/domain/legal_document_type.dart';
import 'package:petsfollow_mobile/features/legal/presentation/legal_document_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

/// Gate RGPD : comptes provisionnés / import sans `terms_accepted_at`.
class AcceptTermsScreen extends StatefulWidget {
  const AcceptTermsScreen({super.key, required this.onAccepted});
  final VoidCallback onAccepted;

  @override
  State<AcceptTermsScreen> createState() => _AcceptTermsScreenState();
}

class _AcceptTermsScreenState extends State<AcceptTermsScreen> {
  bool consent = false;
  String? error;
  bool _busy = false;

  Future<void> submit() async {
    final l10n = AppLocalizations.of(context)!;
    if (!consent) {
      setState(() => error = l10n.registerConsentRequired);
      return;
    }
    setState(() {
      error = null;
      _busy = true;
    });
    try {
      await ApiClient.instance.acceptTerms();
      if (mounted) widget.onAccepted();
    } catch (_) {
      if (mounted) {
        setState(() => error = l10n.acceptTermsFailed);
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void _openLegal(LegalDocumentType type) {
    Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => LegalDocumentScreen(type: type),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);
    final linkStyle = theme.textTheme.bodyMedium?.copyWith(
      color: AppColors.brandTeal,
      decoration: TextDecoration.underline,
    );
    return Container(
      decoration: BoxDecoration(gradient: AppTheme.loginGradientOf(context)),
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
                const Center(
                  child: PetsLogo(variant: PetsLogoVariant.emblem, height: 72),
                ),
                const SizedBox(height: 24),
                Text(
                  l10n.acceptTermsTitle,
                  textAlign: TextAlign.center,
                  style: theme.textTheme.headlineSmall,
                ),
                const SizedBox(height: 8),
                Text(
                  l10n.acceptTermsSubtitle,
                  textAlign: TextAlign.center,
                  style: theme.textTheme.bodyMedium,
                ),
                const SizedBox(height: 24),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Checkbox(
                      key: const Key('accept_terms_consent'),
                      value: consent,
                      onChanged: _busy
                          ? null
                          : (v) => setState(() {
                                consent = v ?? false;
                                error = null;
                              }),
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
                              child:
                                  Text(l10n.legalPrivacyTitle, style: linkStyle),
                            ),
                            const Text('.'),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
                if (error != null) ...[
                  const SizedBox(height: 8),
                  Text(error!, style: TextStyle(color: theme.colorScheme.error)),
                ],
                const SizedBox(height: 24),
                FilledButton(
                  key: const Key('accept_terms_submit'),
                  onPressed: _busy ? null : submit,
                  child: _busy
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(l10n.acceptTermsSubmit),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
