import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/notifications/reminder_controller.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class ReminderSettingsScreen extends StatefulWidget {
  const ReminderSettingsScreen({super.key});

  @override
  State<ReminderSettingsScreen> createState() => _ReminderSettingsScreenState();
}

class _ReminderSettingsScreenState extends State<ReminderSettingsScreen> {
  bool enabled = false;
  TimeOfDay time = const TimeOfDay(hour: 20, minute: 0);
  bool loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final prefs = await ReminderController.instance.load();
    setState(() {
      enabled = prefs.enabled;
      time = TimeOfDay(hour: prefs.hour, minute: prefs.minute);
      loading = false;
    });
  }

  Future<void> _pickTime() async {
    final picked = await showTimePicker(context: context, initialTime: time);
    if (picked != null) setState(() => time = picked);
  }

  Future<void> _save() async {
    await ReminderController.instance.save(
      enabled: enabled,
      hour: time.hour,
      minute: time.minute,
    );
    if (mounted) {
      final l10n = AppLocalizations.of(context)!;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(l10n.remindersSaved)));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    if (loading) {
      return Scaffold(appBar: AppBar(title: Text(l10n.reminders)), body: const Center(child: CircularProgressIndicator()));
    }
    return Scaffold(
      appBar: AppBar(title: Text(l10n.reminders)),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(l10n.remindersHint, style: const TextStyle(height: 1.4)),
            const SizedBox(height: 20),
            SwitchListTile(
              title: Text(l10n.remindersEnabled),
              value: enabled,
              onChanged: (v) => setState(() => enabled = v),
            ),
            ListTile(
              title: Text(l10n.remindersTime),
              subtitle: Text(time.format(context)),
              trailing: const Icon(Icons.schedule),
              onTap: enabled ? _pickTime : null,
            ),
            const SizedBox(height: 24),
            FilledButton(onPressed: _save, child: Text(l10n.save)),
          ],
        ),
      ),
    );
  }
}
