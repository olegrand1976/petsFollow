import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/api/api_errors.dart';
import 'package:petsfollow_mobile/core/models/pet.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/notifications/notification_service.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';
import 'package:petsfollow_mobile/core/theme/pets_palette.dart';
import 'package:petsfollow_mobile/core/ui/load_error_view.dart';
import 'package:petsfollow_mobile/core/ui/safe_bottom.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_chart.dart';
import 'package:petsfollow_mobile/features/pets/presentation/book_visit_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/consultation_view_screen.dart';
import 'package:petsfollow_mobile/features/pets/presentation/preconsult_screen.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';

class PetTimelineScreen extends StatefulWidget {
  const PetTimelineScreen({
    super.key,
    required this.petId,
    this.petName,
    this.canWriteNotes,
  });

  final String petId;
  final String? petName;
  /// When null, resolved from [ApiClient.getPet] on load.
  final bool? canWriteNotes;

  @override
  State<PetTimelineScreen> createState() => _PetTimelineScreenState();
}

class _PetTimelineScreenState extends State<PetTimelineScreen> {
  List<Map<String, dynamic>> items = [];
  List<Visit> visits = [];
  List<({DateTime date, int bpm, bool isAlert})> chartPoints = [];
  bool loading = true;
  String? loadError;
  bool _hasLoadedOnce = false;
  bool _canWriteNotes = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) load();
    });
  }

  Future<void> load() async {
    final l10n = AppLocalizations.of(context)!;
    final keepStale = _hasLoadedOnce;
    if (!keepStale && mounted) {
      setState(() {
        loading = true;
        loadError = null;
      });
    }
    try {
      var canWrite = widget.canWriteNotes ?? false;
      var resolvedAcl = widget.canWriteNotes != null;
      if (!resolvedAcl) {
        try {
          final raw = await ApiClient.instance.getPet(widget.petId);
          canWrite = Pet.fromJson(Map<String, dynamic>.from(raw)).canWriteNotes;
          resolvedAcl = true;
        } catch (_) {
          try {
            final list = await ApiClient.instance.getPets();
            for (final row in list) {
              final map = Map<String, dynamic>.from(row as Map);
              if (map['id']?.toString() == widget.petId) {
                canWrite = Pet.fromJson(map).canWriteNotes;
                resolvedAcl = true;
                break;
              }
            }
          } catch (_) {}
        }
      }
      if (!resolvedAcl && _hasLoadedOnce) {
        canWrite = _canWriteNotes;
      }
      final data = await ApiClient.instance.getTimeline(widget.petId);
      final sessions = await ApiClient.instance.getHeartRateSessions(widget.petId);
      final visitData = await ApiClient.instance.getVisits(widget.petId);
      if (mounted) {
        setState(() {
          if (resolvedAcl || _hasLoadedOnce) {
            _canWriteNotes = canWrite;
          }
          items = data
              .map((e) => Map<String, dynamic>.from(e as Map))
              .toList();
          visits = visitData;
          chartPoints = sessions
              .where((s) => s['bpm'] != null)
              .map((s) => (
                    date: DateTime.parse(s['startedAt'] as String),
                    bpm: (s['bpm'] as num).toInt(),
                    isAlert: s['isAlert'] as bool? ?? false,
                  ))
              .toList()
              .reversed
              .toList();
          loading = false;
          loadError = null;
          _hasLoadedOnce = true;
        });
      }
      // Best-effort: never block / fail the timeline UI.
      // Only schedule/cancel when ACL was positively resolved.
      if (resolvedAcl) {
        if (canWrite) {
          await NotificationService.instance.scheduleVisits(
            visitData,
            visitLabel: l10n.upcomingVisit,
            petName: widget.petName,
          );
        } else {
          await NotificationService.instance.cancelVisitReminders(
            visitData.map((v) => v.id),
          );
        }
      }
    } catch (e) {
      if (!mounted) return;
      final msg = mapApiError(e, l10n);
      if (keepStale) {
        setState(() => loading = false);
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
      } else {
        setState(() {
          loading = false;
          loadError = msg;
        });
      }
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
      case 'weight':
        return Icons.monitor_weight_outlined;
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

  Color _colorForType(String type, Color textMuted) {
    switch (type) {
      case 'heartrate':
        return AppColors.alert;
      case 'weight':
        return AppColors.brandTeal;
      case 'message':
        return AppColors.primary;
      case 'care':
        return AppColors.gold;
      case 'visit':
        return AppColors.primary;
      default:
        return textMuted;
    }
  }

  String _typeLabel(AppLocalizations l10n, String type) {
    switch (type) {
      case 'heartrate':
        return l10n.timelineTypeHeartrate;
      case 'weight':
        return l10n.timelineTypeWeight;
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

  Future<void> _proposeReschedule(Visit visit) async {
    final l10n = AppLocalizations.of(context)!;
    final practiceId = visit.practiceId;
    if (practiceId == null || practiceId.isEmpty) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorGeneric('visit'))),
        );
      }
      return;
    }
    final ok = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => BookVisitScreen(
          petId: widget.petId,
          petName: widget.petName ?? '',
          practiceId: practiceId,
          rescheduleVisitId: visit.id,
        ),
      ),
    );
    if (ok == true && mounted) await load();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final p = PetsPalette.of(context);
    final dateFmt = DateFormat.yMMMd(Localizations.localeOf(context).toString());
    final upcoming = visits.where((v) => v.isUpcoming).toList();
    final pastWithReport = visits.where((v) => v.hasFinalReport).toList();

    return Scaffold(
      appBar: AppBar(title: Text(l10n.visitHistory)),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : loadError != null
              ? LoadErrorView(message: loadError!, onRetry: load)
              : RefreshIndicator(
              onRefresh: load,
              child: ListView(
                padding: scrollPaddingWithSystemBottom(context, all: 16),
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
                              if (_canWriteNotes)
                                Wrap(
                                  alignment: WrapAlignment.end,
                                  children: [
                                    if (v.preconsultPending && _canWriteNotes)
                                      TextButton(
                                        onPressed: () async {
                                          await Navigator.push<void>(
                                            context,
                                            MaterialPageRoute(
                                              builder: (_) => PreconsultScreen(
                                                visitId: v.id,
                                                petName: widget.petName,
                                              ),
                                            ),
                                          );
                                          if (mounted) await load();
                                        },
                                        child: Text(l10n.preconsultFillCta),
                                      ),
                                    if (v.status == 'requested' && !v.awaitingClient)
                                      TextButton(
                                        onPressed: () => _cancelVisit(v),
                                        child: Text(l10n.visitCancelAction),
                                      ),
                                    if ((v.status == 'requested' || v.status == 'confirmed') &&
                                        !v.awaitingClient)
                                      TextButton(
                                        onPressed: () => _proposeReschedule(v),
                                        child: Text(l10n.visitProposeReschedule),
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
                  if (pastWithReport.isNotEmpty) ...[
                    Text(l10n.consultationsHistory, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 8),
                    ...pastWithReport.map(
                      (v) => Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          key: Key('visit_consultation_tile_${v.id}'),
                          leading: CircleAvatar(
                            backgroundColor: AppColors.primary.withValues(alpha: 0.15),
                            child: Icon(Icons.description_outlined, color: AppColors.primary, size: 20),
                          ),
                          title: Text(l10n.consultationTitle),
                          subtitle: Text(dateFmt.format(v.displayDate)),
                          trailing: const Icon(Icons.chevron_right),
                          onTap: () {
                            Navigator.push<void>(
                              context,
                              MaterialPageRoute(
                                builder: (_) => ConsultationViewScreen(
                                  visitId: v.id,
                                  petName: widget.petName,
                                ),
                              ),
                            );
                          },
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
                      child: Text(l10n.timelineEmpty, style: TextStyle(color: p.textMuted)),
                    )
                  else
                    ...items.map((m) {
                      final type = m['type'] as String? ?? 'event';
                      final createdAt = DateTime.tryParse(m['createdAt'] as String? ?? '');
                      return Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          leading: CircleAvatar(
                            backgroundColor: _colorForType(type, p.textMuted).withValues(alpha: 0.15),
                            child: Icon(_iconForType(type), color: _colorForType(type, p.textMuted), size: 20),
                          ),
                          title: Text(_typeLabel(l10n, type)),
                          subtitle: Text(
                            [
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
