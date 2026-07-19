import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/models/notification_prefs.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class NotificationPreferencesScreen extends StatefulWidget {
  const NotificationPreferencesScreen({super.key});

  @override
  State<NotificationPreferencesScreen> createState() =>
      _NotificationPreferencesScreenState();
}

class _NotificationPreferencesScreenState
    extends State<NotificationPreferencesScreen> {
  bool loading = true;
  bool hr = true;
  bool care = true;
  bool visits = true;
  bool messages = true;
  bool discovery = true;
  bool billing = true;
  bool saving = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final prefs = await NotificationService.instance.loadPrefs();
    if (mounted) {
      setState(() {
        hr = prefs.hr;
        care = prefs.care;
        visits = prefs.visits;
        messages = prefs.messages;
        discovery = prefs.discovery;
        billing = prefs.billing;
        loading = false;
      });
    }
  }

  Future<void> _save() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => saving = true);
    try {
      await NotificationService.instance.savePrefs(
        NotificationPrefs(
          userId: '',
          hr: hr,
          care: care,
          visits: visits,
          messages: messages,
          discovery: discovery,
          billing: billing,
        ),
      );
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(l10n.notificationPrefsSaved)));
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('prefs'))),
        );
      }
    }
    if (mounted) setState(() => saving = false);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    if (loading) {
      return Scaffold(
        appBar: AppBar(title: Text(l10n.notificationPreferences)),
        body: const Center(child: CircularProgressIndicator()),
      );
    }
    return Scaffold(
      appBar: AppBar(title: Text(l10n.notificationPreferences)),
      body: SafeArea(
        top: false,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(l10n.notificationPrefsHint,
                style: TextStyle(color: AppColors.textMuted, height: 1.4)),
            const SizedBox(height: 8),
            SwitchListTile(
                title: Text(l10n.notificationPrefHr),
                value: hr,
                onChanged: (v) => setState(() => hr = v)),
            SwitchListTile(
                title: Text(l10n.notificationPrefCare),
                value: care,
                onChanged: (v) => setState(() => care = v)),
            SwitchListTile(
                title: Text(l10n.notificationPrefVisits),
                value: visits,
                onChanged: (v) => setState(() => visits = v)),
            SwitchListTile(
                title: Text(l10n.notificationPrefMessages),
                value: messages,
                onChanged: (v) => setState(() => messages = v)),
            SwitchListTile(
                title: Text(l10n.notificationPrefDiscovery),
                value: discovery,
                onChanged: (v) => setState(() => discovery = v)),
            SwitchListTile(
                title: Text(l10n.notificationPrefBilling),
                value: billing,
                onChanged: (v) => setState(() => billing = v)),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: saving ? null : _save,
              child: saving
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2))
                  : Text(l10n.save),
            ),
          ],
        ),
      ),
    );
  }
}
