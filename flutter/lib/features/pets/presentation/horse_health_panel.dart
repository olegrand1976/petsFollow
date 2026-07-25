import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

const _horseCareTypes = {'farrier', 'fecal_egg'};

class HorseHealthPanel extends StatefulWidget {
  const HorseHealthPanel({super.key, required this.petId, required this.petName});

  final String petId;
  final String petName;

  @override
  State<HorseHealthPanel> createState() => _HorseHealthPanelState();
}

class _HorseHealthPanelState extends State<HorseHealthPanel> with WidgetsBindingObserver {
  List<CareReminder> reminders = [];
  List<Map<String, dynamic>> contacts = [];
  List<Map<String, dynamic>> competitions = [];
  bool loading = true;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _load();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _load();
    }
  }

  Future<void> _load() async {
    try {
      final data = await ApiClient.instance.getCareReminders(widget.petId);
      List<Map<String, dynamic>> c = [];
      List<Map<String, dynamic>> comps = [];
      try {
        c = (await ApiClient.instance.getHorseContacts(widget.petId))
            .whereType<Map>()
            .map((e) => Map<String, dynamic>.from(e))
            .toList();
      } catch (_) {}
      try {
        comps = (await ApiClient.instance.getHorseCompetitions(widget.petId))
            .whereType<Map>()
            .map((e) => Map<String, dynamic>.from(e))
            .toList();
      } catch (_) {}
      if (mounted) {
        setState(() {
          reminders = data.where((r) => !r.isDone).toList();
          contacts = c;
          competitions = comps;
          loading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() => loading = false);
      }
    }
  }

  String _careTypeLabel(AppLocalizations l10n, CareReminder reminder) {
    switch (reminder.type) {
      case 'farrier':
        return l10n.careTypeFarrier;
      case 'fecal_egg':
        return l10n.careTypeFecalEgg;
      default:
        return reminder.title;
    }
  }

  Future<void> _addContact() async {
    final l10n = AppLocalizations.of(context)!;
    final nameCtrl = TextEditingController();
    final roleCtrl = TextEditingController();
    try {
      final ok = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: Text(l10n.horseAddContact),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(controller: nameCtrl, decoration: InputDecoration(labelText: l10n.horseContactName)),
              TextField(controller: roleCtrl, decoration: InputDecoration(labelText: l10n.horseContactRole)),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(l10n.cancel)),
            FilledButton(onPressed: () => Navigator.pop(ctx, true), child: Text(l10n.save)),
          ],
        ),
      );
      if (ok != true || nameCtrl.text.trim().isEmpty) return;
      await ApiClient.instance.createHorseContact(
        widget.petId,
        fullName: nameCtrl.text.trim(),
        role: roleCtrl.text.trim(),
      );
      await _load();
    } finally {
      nameCtrl.dispose();
      roleCtrl.dispose();
    }
  }

  Future<void> _addCompetition() async {
    final l10n = AppLocalizations.of(context)!;
    final titleCtrl = TextEditingController();
    final dateCtrl = TextEditingController(text: DateFormat('yyyy-MM-dd').format(DateTime.now()));
    try {
      final ok = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: Text(l10n.horseAddCompetition),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(controller: titleCtrl, decoration: InputDecoration(labelText: l10n.horseCompetitionTitle)),
              TextField(controller: dateCtrl, decoration: InputDecoration(labelText: l10n.horseCompetitionDate)),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(l10n.cancel)),
            FilledButton(onPressed: () => Navigator.pop(ctx, true), child: Text(l10n.save)),
          ],
        ),
      );
      if (ok != true || titleCtrl.text.trim().isEmpty) return;
      await ApiClient.instance.createHorseCompetition(
        widget.petId,
        title: titleCtrl.text.trim(),
        eventDate: dateCtrl.text.trim(),
      );
      await _load();
    } finally {
      titleCtrl.dispose();
      dateCtrl.dispose();
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateFmt = DateFormat.yMMMd(Localizations.localeOf(context).toString());
    final horseReminders = reminders.where((r) => _horseCareTypes.contains(r.type)).toList();
    final otherReminders = reminders.where((r) => !_horseCareTypes.contains(r.type)).toList();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(l10n.horseHealthTitle, style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 12),
        if (loading)
          const Center(child: Padding(padding: EdgeInsets.all(16), child: CircularProgressIndicator()))
        else ...[
          if (horseReminders.isEmpty)
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Text(l10n.noCareReminders, style: TextStyle(color: AppColors.textMuted)),
              ),
            )
          else
            ...horseReminders.map(
              (r) => _HorseCareCard(
                label: _careTypeLabel(l10n, r),
                dueLabel: () {
                  final due = r.isOverdue
                      ? '${l10n.careOverdue} · ${dateFmt.format(r.dueAt)}'
                      : dateFmt.format(r.dueAt);
                  return r.hasRecurrence
                      ? '$due · ${l10n.careRecurrenceDays(r.recurrenceDays!)}'
                      : due;
                }(),
                isOverdue: r.isOverdue,
                highlighted: true,
              ),
            ),
          if (otherReminders.isNotEmpty) ...[
            const SizedBox(height: 8),
            ...otherReminders.map(
              (r) => _HorseCareCard(
                label: _careTypeLabel(l10n, r),
                dueLabel: () {
                  final due = r.isOverdue
                      ? '${l10n.careOverdue} · ${dateFmt.format(r.dueAt)}'
                      : dateFmt.format(r.dueAt);
                  return r.hasRecurrence
                      ? '$due · ${l10n.careRecurrenceDays(r.recurrenceDays!)}'
                      : due;
                }(),
                isOverdue: r.isOverdue,
                highlighted: false,
              ),
            ),
          ],
        ],
        const SizedBox(height: 16),
        Row(
          children: [
            Expanded(child: Text(l10n.horseContactsTitle, style: Theme.of(context).textTheme.titleSmall)),
            TextButton(onPressed: _addContact, child: Text(l10n.horseAddContact)),
          ],
        ),
        ...contacts.map(
          (c) => Card(
            child: ListTile(
              title: Text(c['fullName'] as String? ?? ''),
              subtitle: Text(c['role'] as String? ?? '', style: TextStyle(color: AppColors.textMuted)),
              trailing: IconButton(
                icon: const Icon(Icons.delete_outline),
                onPressed: () async {
                  await ApiClient.instance.deleteHorseContact(c['id'] as String);
                  await _load();
                },
              ),
            ),
          ),
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(child: Text(l10n.horseCompetitionsTitle, style: Theme.of(context).textTheme.titleSmall)),
            TextButton(onPressed: _addCompetition, child: Text(l10n.horseAddCompetition)),
          ],
        ),
        ...competitions.map(
          (c) => Card(
            child: ListTile(
              title: Text(c['title'] as String? ?? ''),
              subtitle: Text(
                '${c['eventDate'] ?? ''} ${c['location'] ?? ''}'.trim(),
                style: TextStyle(color: AppColors.textMuted),
              ),
              trailing: IconButton(
                icon: const Icon(Icons.delete_outline),
                onPressed: () async {
                  await ApiClient.instance.deleteHorseCompetition(c['id'] as String);
                  await _load();
                },
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _HorseCareCard extends StatelessWidget {
  const _HorseCareCard({
    required this.label,
    required this.dueLabel,
    required this.isOverdue,
    required this.highlighted,
  });

  final String label;
  final String dueLabel;
  final bool isOverdue;
  final bool highlighted;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      color: highlighted ? AppColors.gold.withValues(alpha: 0.08) : null,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: highlighted
            ? BorderSide(color: AppColors.gold.withValues(alpha: 0.35))
            : BorderSide.none,
      ),
      child: ListTile(
        leading: Icon(
          highlighted ? Icons.star_outline : Icons.medical_services_outlined,
          color: highlighted ? AppColors.gold : AppColors.primary,
        ),
        title: Text(label, style: TextStyle(fontWeight: highlighted ? FontWeight.w600 : FontWeight.normal)),
        subtitle: Text(
          dueLabel,
          style: TextStyle(color: isOverdue ? AppColors.alert : AppColors.textMuted),
        ),
      ),
    );
  }
}
