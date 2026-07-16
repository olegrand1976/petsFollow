import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_chart.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:intl/intl.dart';

class PetTimelineScreen extends StatefulWidget {
  const PetTimelineScreen({super.key, required this.petId, this.petName});

  final String petId;
  final String? petName;

  @override
  State<PetTimelineScreen> createState() => _PetTimelineScreenState();
}

class _PetTimelineScreenState extends State<PetTimelineScreen> {
  List<dynamic> items = [];
  List<Visit> visits = [];
  List<({DateTime date, int bpm, bool isAlert})> chartPoints = [];
  bool loading = true;

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    final l10n = AppLocalizations.of(context)!;
    try {
      final data = await ApiClient.instance.getTimeline(widget.petId);
      final sessions = await ApiClient.instance.getHeartRateSessions(widget.petId);
      final visitData = await ApiClient.instance.getVisits(widget.petId);
      await NotificationService.instance.scheduleVisits(
        visitData,
        visitLabel: l10n.upcomingVisit,
        petName: widget.petName,
      );
      if (mounted) {
        setState(() {
          items = data;
          visits = visitData;
          chartPoints = sessions
              .where((s) => s['bpm'] != null)
              .map((s) => (
                    date: DateTime.parse(s['startedAt'] as String),
                    bpm: s['bpm'] as int,
                    isAlert: s['isAlert'] as bool? ?? false,
                  ))
              .toList()
              .reversed
              .toList();
          loading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => loading = false);
    }
  }

  String _visitStatusLabel(AppLocalizations l10n, String status) {
    switch (status) {
      case 'requested':
        return l10n.visitStatusRequested;
      case 'confirmed':
        return l10n.visitStatusConfirmed;
      case 'done':
        return l10n.visitStatusDone;
      case 'cancelled':
        return l10n.visitStatusCancelled;
      default:
        return status;
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateFmt = DateFormat.yMMMd(Localizations.localeOf(context).toString());

    return Scaffold(
      appBar: AppBar(title: Text(l10n.history)),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: load,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  if (chartPoints.isNotEmpty) ...[
                    Text(l10n.latestValues, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 8),
                    HeartRateChart(points: chartPoints, height: 200),
                    const SizedBox(height: 24),
                  ],
                  if (visits.isNotEmpty) ...[
                    Text(l10n.visitHistory, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 8),
                    ...visits.map(
                      (v) => Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          leading: Icon(
                            v.isUpcoming ? Icons.event : Icons.event_available,
                            color: v.isUpcoming ? AppColors.primary : AppColors.textMuted,
                          ),
                          title: Text(_visitStatusLabel(l10n, v.status)),
                          subtitle: Text(
                            [
                              dateFmt.format(v.displayDate),
                              if (v.notes != null && v.notes!.isNotEmpty) v.notes,
                            ].join(' · '),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],
                  ...items.map((item) {
                    final m = item as Map<String, dynamic>;
                    return ListTile(
                      title: Text(m['title'] as String? ?? ''),
                      subtitle: Text(m['body'] as String? ?? ''),
                    );
                  }),
                ],
              ),
            ),
    );
  }
}
