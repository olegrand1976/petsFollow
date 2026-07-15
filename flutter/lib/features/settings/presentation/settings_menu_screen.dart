import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/features/education/presentation/how_to_measure_screen.dart';
import 'package:petsfollow_mobile/features/legal/domain/legal_document_type.dart';
import 'package:petsfollow_mobile/features/legal/presentation/legal_document_screen.dart';
import 'package:petsfollow_mobile/features/profile/presentation/profile_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/reminder_settings_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class SettingsMenuScreen extends StatelessWidget {
  const SettingsMenuScreen({super.key, required this.onLogout});

  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings)),
      body: ListView(
        children: [
          ListTile(
            leading: const Icon(Icons.person_outline),
            title: Text(l10n.myData),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const ProfileScreen()),
            ),
          ),
          ListTile(
            leading: const Icon(Icons.play_circle_outline),
            title: Text(l10n.howToMeasure),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const HowToMeasureScreen()),
            ),
          ),
          ListTile(
            leading: const Icon(Icons.notifications_outlined),
            title: Text(l10n.reminders),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const ReminderSettingsScreen()),
            ),
          ),
          const Divider(),
          ListTile(
            leading: const Icon(Icons.shield_outlined),
            title: Text(l10n.legalPrivacyTitle),
            onTap: () => _openLegal(context, LegalDocumentType.privacy),
          ),
          ListTile(
            leading: const Icon(Icons.description_outlined),
            title: Text(l10n.legalTermsTitle),
            onTap: () => _openLegal(context, LegalDocumentType.terms),
          ),
          ListTile(
            leading: const Icon(Icons.info_outline),
            title: Text(l10n.legalNoticeTitle),
            onTap: () => _openLegal(context, LegalDocumentType.legalNotice),
          ),
          const Divider(),
          ListTile(
            leading: const Icon(Icons.logout, color: Colors.redAccent),
            title: Text(l10n.logout, style: const TextStyle(color: Colors.redAccent)),
            onTap: () async {
              await ApiClient.instance.logout();
              onLogout();
            },
          ),
        ],
      ),
    );
  }

  void _openLegal(BuildContext context, LegalDocumentType type) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => LegalDocumentScreen(type: type)),
    );
  }
}
