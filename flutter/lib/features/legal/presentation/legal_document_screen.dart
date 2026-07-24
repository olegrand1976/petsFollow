import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/open_url.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/legal/domain/legal_document_type.dart';
import 'package:petsfollow_mobile/features/legal/domain/legal_urls.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class LegalDocumentScreen extends StatelessWidget {
  const LegalDocumentScreen({super.key, required this.type});

  final LegalDocumentType type;

  String _title(AppLocalizations l10n) {
    return switch (type) {
      LegalDocumentType.terms => l10n.legalTermsTitle,
      LegalDocumentType.privacy => l10n.legalPrivacyTitle,
      LegalDocumentType.legalNotice => l10n.legalNoticeTitle,
    };
  }

  String _body(AppLocalizations l10n) {
    return switch (type) {
      LegalDocumentType.terms => l10n.legalTermsBody,
      LegalDocumentType.privacy => l10n.legalPrivacyBody,
      LegalDocumentType.legalNotice => l10n.legalNoticeBody,
    };
  }

  String get _onlineUrl {
    return switch (type) {
      LegalDocumentType.terms => LegalUrls.terms,
      LegalDocumentType.privacy => LegalUrls.privacy,
      LegalDocumentType.legalNotice => LegalUrls.mentions,
    };
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(_title(l10n))),
      body: SingleChildScrollView(
        padding: scrollPaddingWithSystemBottom(context, all: 20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              _body(l10n),
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(height: 1.5),
            ),
            const SizedBox(height: 24),
            OutlinedButton.icon(
              onPressed: () => openExternalUrl(_onlineUrl),
              icon: const Icon(Icons.open_in_new),
              label: Text(l10n.legalOpenOnline),
            ),
          ],
        ),
      ),
    );
  }
}
