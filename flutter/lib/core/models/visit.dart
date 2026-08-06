class Visit {
  const Visit({
    required this.id,
    required this.petId,
    this.practiceId,
    this.siteId,
    this.siteName,
    this.scheduledAt,
    this.status = 'requested',
    this.notes,
    this.source,
    this.consultationSession = false,
    this.hasFinalReport = false,
    this.reportStatus = '',
    this.createdAt,
    this.proposedScheduledAt,
    this.pendingActionBy,
    this.durationMinutes,
    this.preconsultStatus,
  });

  final String id;
  final String petId;
  final String? practiceId;
  /// Practice site for this visit (multi-sites); empty on legacy payloads.
  final String? siteId;
  final String? siteName;
  final DateTime? scheduledAt;
  final String status;
  final String? notes;
  final String? source;
  final bool consultationSession;
  final bool hasFinalReport;
  /// Owner-only: `final` | `draft` | empty (never includes draft body).
  final String reportStatus;
  final DateTime? createdAt;
  final DateTime? proposedScheduledAt;
  final String? pendingActionBy;
  final int? durationMinutes;
  final String? preconsultStatus;

  bool get isUpcoming {
    if (status == 'done' || status == 'cancelled') return false;
    if (scheduledAt != null) return scheduledAt!.isAfter(DateTime.now());
    return status == 'requested' || status == 'confirmed' || status == 'reschedule_pending';
  }

  bool get awaitingClient => pendingActionBy == 'client';

  bool get preconsultPending =>
      status == 'confirmed' && preconsultStatus == 'pending';

  /// Show in Consultations section (available or pending draft).
  bool get hasConsultationSignal =>
      hasFinalReport || reportStatus == 'final' || reportStatus == 'draft';

  bool get consultationAvailable =>
      hasFinalReport || reportStatus == 'final';

  bool get consultationPending =>
      !consultationAvailable && reportStatus == 'draft';

  DateTime get displayDate =>
      proposedScheduledAt ?? scheduledAt ?? createdAt ?? DateTime.now();

  factory Visit.fromJson(Map<String, dynamic> json) {
    final siteIdRaw = (json['siteId'] as String?)?.trim();
    final siteNameRaw = (json['siteName'] as String?)?.trim();
    return Visit(
      id: json['id'] as String? ?? '',
      petId: json['petId'] as String? ?? '',
      practiceId: json['practiceId'] as String?,
      siteId: (siteIdRaw == null || siteIdRaw.isEmpty) ? null : siteIdRaw,
      siteName: (siteNameRaw == null || siteNameRaw.isEmpty) ? null : siteNameRaw,
      scheduledAt: json['scheduledAt'] != null
          ? DateTime.tryParse(json['scheduledAt'] as String)
          : null,
      status: json['status'] as String? ?? 'requested',
      notes: json['notes'] as String?,
      source: json['source'] as String?,
      consultationSession: json['consultationSession'] == true,
      hasFinalReport: json['hasFinalReport'] == true,
      reportStatus: (json['reportStatus'] as String?)?.trim() ?? '',
      createdAt: json['createdAt'] != null
          ? DateTime.tryParse(json['createdAt'] as String)
          : null,
      proposedScheduledAt: json['proposedScheduledAt'] != null
          ? DateTime.tryParse(json['proposedScheduledAt'] as String)
          : null,
      pendingActionBy: json['pendingActionBy'] as String?,
      durationMinutes: json['durationMinutes'] as int?,
      preconsultStatus: json['preconsultStatus'] as String?,
    );
  }
}
