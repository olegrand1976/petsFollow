import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/features/shell/presentation/main_shell_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:intl/intl.dart';

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
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
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
