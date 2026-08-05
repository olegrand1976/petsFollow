import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/config/app_env.dart';
import 'package:petsfollow_mobile/core/locale/language_picker_screen.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/appearance_settings_tile.dart';
import 'package:petsfollow_mobile/features/client_ai/presentation/explain_reports_list_screen.dart';
import 'package:petsfollow_mobile/features/client_ai/presentation/triage_chat_screen.dart';
import 'package:petsfollow_mobile/features/education/presentation/how_to_measure_screen.dart';
import 'package:petsfollow_mobile/features/invite/presentation/app_invite_qr_screen.dart';
import 'package:petsfollow_mobile/features/legal/domain/legal_document_type.dart';
import 'package:petsfollow_mobile/features/legal/presentation/legal_document_screen.dart';
import 'package:petsfollow_mobile/features/profile/presentation/profile_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/feature_modules_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/notification_preferences_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/reminder_settings_screen.dart';
import 'package:petsfollow_mobile/features/settings/presentation/switch_profile_screen.dart';
import 'package:petsfollow_mobile/features/support/presentation/support_report_screen.dart';
import 'package:petsfollow_mobile/features/vets/presentation/my_vets_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class SettingsMenuScreen extends StatelessWidget {
  const SettingsMenuScreen({super.key, required this.onLogout, this.embedded = false});

  final VoidCallback onLogout;
  final bool embedded;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final body = ListView(
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
        const LanguageSettingsTile(),
        const AppearanceSettingsTile(),
        ListTile(
          leading: const Icon(Icons.play_circle_outline),
          title: Text(l10n.howToMeasure),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const HowToMeasureScreen()),
          ),
        ),
        if (AppEnv.isClientAiEnabled) ...[
          Padding(
            key: const Key('settings_client_ai_section'),
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Row(
              children: [
                Text(
                  l10n.clientAiSectionTitle,
                  style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        color: AppColors.primary,
                        fontWeight: FontWeight.w600,
                      ),
                ),
                const SizedBox(width: 8),
                Chip(
                  label: Text(l10n.clientAiDevBadge),
                  visualDensity: VisualDensity.compact,
                  padding: EdgeInsets.zero,
                  labelPadding: const EdgeInsets.symmetric(horizontal: 6),
                ),
              ],
            ),
          ),
          ListTile(
            key: const Key('settings_client_ai_explain'),
            leading: const Icon(Icons.menu_book_outlined),
            title: Text(l10n.clientAiExplainListTitle),
            subtitle: Text(l10n.clientAiExplainListSubtitle),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const ExplainReportsListScreen()),
            ),
          ),
          ListTile(
            key: const Key('settings_client_ai_triage'),
            leading: const Icon(Icons.health_and_safety_outlined),
            title: Text(l10n.clientAiTriageTitle),
            subtitle: Text(l10n.clientAiTriageSubtitle),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const TriageChatScreen()),
            ),
          ),
        ],
        ListTile(
          leading: const Icon(Icons.local_hospital_outlined),
          title: Text(l10n.myVets),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const MyVetsScreen()),
          ),
        ),
        ListTile(
          key: const Key('settings_app_invite'),
          leading: const Icon(Icons.qr_code_2_outlined),
          title: Text(l10n.appInviteTitle),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const AppInviteQrScreen()),
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
        ListTile(
          leading: const Icon(Icons.tune),
          title: Text(l10n.notificationPreferences),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const NotificationPreferencesScreen()),
          ),
        ),
        ListTile(
          leading: const Icon(Icons.extension_outlined),
          title: Text(l10n.featureModules),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const FeatureModulesScreen()),
          ),
        ),
        ListTile(
          leading: const Icon(Icons.apps_outlined),
          title: Text(l10n.featureModulesCatalog),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const FeatureModulesScreen(catalogOnly: true)),
          ),
        ),
        ListTile(
          leading: const Icon(Icons.swap_horiz),
          title: Text(l10n.switchProfile),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const SwitchProfileScreen()),
          ),
        ),
        ListTile(
          key: const Key('settings_support'),
          leading: const Icon(Icons.support_agent_outlined),
          title: Text(l10n.supportMenu),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => openSupportReport(context, source: 'flutter_client'),
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
          key: const Key('settings_logout'),
          leading: const Icon(Icons.logout, color: AppColors.alert),
          title: Text(l10n.logout, style: const TextStyle(color: AppColors.alert)),
          onTap: () async {
            try {
              await ApiClient.instance.logout();
            } finally {
              onLogout();
            }
          },
        ),
      ],
    );

    if (embedded) {
      return Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(title: Text(l10n.navProfile)),
        body: body,
      );
    }

    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings)),
      body: body,
    );
  }

  void _openLegal(BuildContext context, LegalDocumentType type) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => LegalDocumentScreen(type: type)),
    );
  }
}
