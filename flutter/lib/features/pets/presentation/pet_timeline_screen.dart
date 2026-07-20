import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_chart.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class PetTimelineScreen extends StatefulWidget {
  const PetTimelineScreen({super.key, required this.petId, this.petName});

  final String petId;
  final String? petName;

  @override
  State<PetTimelineScreen> createState() => _PetTimelineScreenState();
}

class _PetTimelineScreenState extends State<PetTimelineScreen> {
  List<Map<String, dynamic>> items = [];
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
          items = data
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
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
      case 'reschedule_pending':
        return l10n.visitStatusReschedulePending;
      default:
        return status;
    }
  }

  IconData _iconForType(String type) {
    switch (type) {
      case 'heartrate':
        return Icons.favorite_outline;
      case 'message':
        return Icons.chat_bubble_outline;
      case 'care':
        return Icons.medical_services_outlined;
      case 'visit':
        return Icons.event_available;
      case 'event':
        return Icons.flag_outlined;
      default:
        return Icons.circle_outlined;
    }
  }

  Color _colorForType(String type) {
    switch (type) {
      case 'heartrate':
        return AppColors.alert;
      case 'message':
        return AppColors.primary;
      case 'care':
        return AppColors.gold;
      case 'visit':
        return AppColors.primary;
      default:
        return AppColors.textMuted;
    }
  }

  String _typeLabel(AppLocalizations l10n, String type) {
    switch (type) {
      case 'heartrate':
        return l10n.timelineTypeHeartrate;
      case 'message':
        return l10n.timelineTypeMessage;
      case 'care':
        return l10n.timelineTypeCare;
      case 'visit':
        return l10n.timelineTypeVisit;
      case 'event':
        return l10n.timelineTypeEvent;
      default:
        return type;
    }
  }

  Future<void> _cancelVisit(Visit visit) async {
    final l10n = AppLocalizations.of(context)!;
    try {
      await ApiClient.instance.updateVisit(visit.id, 'cancelled');
      await load();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('visit'))),
        );
      }
    }
  }

  Future<void> _visitAction(Visit visit, String action) async {
    final l10n = AppLocalizations.of(context)!;
    try {
      await ApiClient.instance.visitAction(visit.id, action: action);
      await load();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('visit'))),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateFmt = DateFormat.yMMMd(Localizations.localeOf(context).toString());
    final upcoming = visits.where((v) => v.isUpcoming).toList();

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
                  if (upcoming.isNotEmpty) ...[
                    Text(l10n.upcomingVisits, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 8),
                    ...upcoming.map(
                      (v) => Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: Padding(
                          padding: const EdgeInsets.only(bottom: 8),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              ListTile(
                                leading: Icon(
                                  Icons.event,
                                  color: AppColors.primary,
                                ),
                                title: Text(_visitStatusLabel(l10n, v.status)),
                                subtitle: Text(
                                  [
                                    dateFmt.format(v.displayDate),
                                    if (v.notes != null && v.notes!.isNotEmpty) v.notes,
                                  ].join(' · '),
                                ),
                              ),
                              Wrap(
                                alignment: WrapAlignment.end,
                                children: [
                                  if (v.status == 'requested' && !v.awaitingClient)
                                    TextButton(
                                      onPressed: () => _cancelVisit(v),
                                      child: Text(l10n.visitCancelAction),
                                    ),
                                  if (v.awaitingClient && v.status == 'requested')
                                    TextButton(
                                      onPressed: () => _visitAction(v, 'confirm'),
                                      child: Text(l10n.visitConfirm),
                                    ),
                                  if (v.awaitingClient && v.status == 'reschedule_pending') ...[
                                    TextButton(
                                      onPressed: () => _visitAction(v, 'accept_reschedule'),
                                      child: Text(l10n.visitAcceptReschedule),
                                    ),
                                    TextButton(
                                      onPressed: () => _visitAction(v, 'reject_reschedule'),
                                      child: Text(l10n.visitRejectReschedule),
                                    ),
                                  ],
                                ],
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],
                  Text(l10n.history, style: Theme.of(context).textTheme.titleMedium),
                  const SizedBox(height: 8),
                  if (items.isEmpty)
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 24),
                      child: Text(l10n.timelineEmpty, style: TextStyle(color: AppColors.textMuted)),
                    )
                  else
                    ...items.map((m) {
                      final type = m['type'] as String? ?? 'event';
                      final createdAt = DateTime.tryParse(m['createdAt'] as String? ?? '');
                      return Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          leading: CircleAvatar(
                            backgroundColor: _colorForType(type).withValues(alpha: 0.15),
                            child: Icon(_iconForType(type), color: _colorForType(type), size: 20),
                          ),
                          title: Text(m['title'] as String? ?? _typeLabel(l10n, type)),
                          subtitle: Text(
                            [
                              _typeLabel(l10n, type),
                              if (createdAt != null) dateFmt.format(createdAt.toLocal()),
                              if ((m['body'] as String?)?.isNotEmpty == true) m['body'] as String,
                            ].join(' · '),
                          ),
                        ),
                      );
                    }),
                ],
              ),
            ),
    );
  }
}
