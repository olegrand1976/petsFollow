import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/features/shell/presentation/main_shell_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class CareTab extends StatefulWidget {
  const CareTab({super.key});

  @override
  State<CareTab> createState() => _CareTabState();
}

class _CareTabState extends State<CareTab> {
  List<Pet> pets = [];
  Map<String, List<CareReminder>> remindersByPet = {};
  bool loading = true;

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    setState(() => loading = true);
    try {
      final data = await ApiClient.instance.getPets();
      final loadedPets = data.map((p) => Pet.fromJson(Map<String, dynamic>.from(p as Map))).toList();
      final map = <String, List<CareReminder>>{};
      for (final pet in loadedPets) {
        try {
          final reminders = await ApiClient.instance.getCareReminders(pet.id);
          map[pet.id] = reminders.where((r) => !r.isDone).toList();
          await NotificationService.instance.scheduleCareReminders(reminders, petName: pet.name);
        } catch (_) {
          map[pet.id] = [];
        }
      }
      if (mounted) {
        setState(() {
          pets = loadedPets;
          remindersByPet = map;
          loading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> markDone(CareReminder reminder) async {
    final l10n = AppLocalizations.of(context)!;
    try {
      await ApiClient.instance.markCareReminderDone(reminder.id);
      await NotificationService.instance.cancelCareReminder(reminder.id);
      await load();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('care'))),
        );
      }
    }
  }

  Future<void> postpone(CareReminder reminder) async {
    final l10n = AppLocalizations.of(context)!;
    final days = await showModalBottomSheet<int>(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(title: Text(l10n.carePostponeDays(7)), onTap: () => Navigator.pop(ctx, 7)),
            ListTile(title: Text(l10n.carePostponeDays(14)), onTap: () => Navigator.pop(ctx, 14)),
            ListTile(title: Text(l10n.carePostponeDays(30)), onTap: () => Navigator.pop(ctx, 30)),
          ],
        ),
      ),
    );
    if (days == null) return;
    try {
      await ApiClient.instance.postponeCareReminder(reminder.id, days);
      await load();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('care'))),
        );
      }
    }
  }

  Future<void> _showCreateSheet() async {
    if (pets.isEmpty || !mounted) return;
    final l10n = AppLocalizations.of(context)!;
    var selectedPetId = pets.first.id;
    var selectedType = 'vaccination';
    var dueDays = 30;
    final titleCtrl = TextEditingController();

    final created = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setModal) {
            final hasHorse = pets.any((p) => p.species == 'horse');
            final types = <MapEntry<String, String>>[
              MapEntry('vaccination', l10n.careTypeVaccination),
              MapEntry('deworming', l10n.careTypeDeworming),
              MapEntry('vet_check', l10n.careTypeVetCheck),
              MapEntry('dental', l10n.careTypeDental),
              if (hasHorse) ...[
                MapEntry('farrier', l10n.careTypeFarrier),
                MapEntry('fecal_egg', l10n.careTypeFecalEgg),
              ],
              MapEntry('custom', l10n.careTypeCustom),
            ];
            return Padding(
              padding: EdgeInsets.only(
                left: 16,
                right: 16,
                top: 16,
                bottom: MediaQuery.of(ctx).viewInsets.bottom + 16,
              ),
              child: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(l10n.careAddReminder, style: Theme.of(ctx).textTheme.titleMedium),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      value: selectedPetId,
                      decoration: InputDecoration(labelText: l10n.careSelectPet),
                      items: pets
                          .map((p) => DropdownMenuItem(value: p.id, child: Text(p.name)))
                          .toList(),
                      onChanged: (v) {
                        if (v != null) setModal(() => selectedPetId = v);
                      },
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      value: selectedType,
                      items: types
                          .map((e) => DropdownMenuItem(value: e.key, child: Text(e.value)))
                          .toList(),
                      onChanged: (v) {
                        if (v != null) setModal(() => selectedType = v);
                      },
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: titleCtrl,
                      decoration: InputDecoration(hintText: l10n.careTypeCustom),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<int>(
                      value: dueDays,
                      decoration: InputDecoration(labelText: l10n.careDueInDays(dueDays)),
                      items: const [7, 14, 30, 90]
                          .map((d) => DropdownMenuItem(value: d, child: Text(l10n.careDueInDays(d))))
                          .toList(),
                      onChanged: (v) {
                        if (v != null) setModal(() => dueDays = v);
                      },
                    ),
                    const SizedBox(height: 16),
                    FilledButton(
                      onPressed: () => Navigator.pop(ctx, true),
                      child: Text(l10n.careAddReminder),
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );

    final title = titleCtrl.text;
    titleCtrl.dispose();
    if (created != true) return;

    try {
      await ApiClient.instance.createCareReminder(
        selectedPetId,
        title: title.isEmpty ? null : title,
        type: selectedType,
        dueDays: dueDays,
      );
      await load();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('care'))),
        );
      }
    }
  }

  String _careTitle(AppLocalizations l10n, CareReminder r) {
    switch (r.type) {
      case 'vaccination':
        return l10n.careTypeVaccination;
      case 'deworming':
        return l10n.careTypeDeworming;
      case 'vet_check':
        return l10n.careTypeVetCheck;
      case 'dental':
        return l10n.careTypeDental;
      case 'farrier':
        return l10n.careTypeFarrier;
      case 'fecal_egg':
        return l10n.careTypeFecalEgg;
      case 'custom':
        return r.title.isNotEmpty ? r.title : l10n.careTypeCustom;
      default:
        if (r.title.isNotEmpty) return r.title;
        return r.type ?? l10n.careTypeCustom;
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateFmt = DateFormat.yMMMd(Localizations.localeOf(context).toString());

    return PetsTabScaffold(
      title: Text(l10n.careTitle),
      floatingActionButton: pets.isEmpty
          ? null
          : FloatingActionButton(
              onPressed: _showCreateSheet,
              tooltip: l10n.careAddReminder,
              child: const Icon(Icons.add),
            ),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: load,
              child: pets.isEmpty
                  ? ListView(
                      children: [
                        SizedBox(
                          height: MediaQuery.of(context).size.height * 0.5,
                          child: Center(child: Text(l10n.emptyPetsTitle, style: TextStyle(color: AppColors.textMuted))),
                        ),
                      ],
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 88),
                      itemCount: pets.length,
                      itemBuilder: (_, i) {
                        final pet = pets[i];
                        final reminders = remindersByPet[pet.id] ?? [];
                        return Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Padding(
                              padding: const EdgeInsets.only(top: 8, bottom: 8),
                              child: Text(pet.name, style: Theme.of(context).textTheme.titleMedium),
                            ),
                            if (reminders.isEmpty)
                              Card(
                                child: Padding(
                                  padding: const EdgeInsets.all(16),
                                  child: Text(l10n.noCareReminders, style: TextStyle(color: AppColors.textMuted)),
                                ),
                              )
                            else
                              ...reminders.map(
                                (r) => Card(
                                  margin: const EdgeInsets.only(bottom: 8),
                                  child: ListTile(
                                    title: Text(_careTitle(l10n, r)),
                                    subtitle: Text(
                                      r.isOverdue
                                          ? '${l10n.careOverdue} · ${dateFmt.format(r.dueAt)}'
                                          : dateFmt.format(r.dueAt),
                                      style: TextStyle(
                                        color: r.isOverdue ? AppColors.alert : AppColors.textMuted,
                                      ),
                                    ),
                                    trailing: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        IconButton(
                                          icon: const Icon(Icons.check_circle_outline),
                                          tooltip: l10n.careDone,
                                          onPressed: () => markDone(r),
                                        ),
                                        IconButton(
                                          icon: const Icon(Icons.schedule),
                                          tooltip: l10n.carePostpone,
                                          onPressed: () => postpone(r),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ),
                            const SizedBox(height: 8),
                          ],
                        );
                      },
                    ),
            ),
    );
  }
}
