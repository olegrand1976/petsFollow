class Visit {
  const Visit({
    required this.id,
    required this.petId,
    this.practiceId,
    this.scheduledAt,
    this.status = 'requested',
    this.notes,
    this.createdAt,
    this.proposedScheduledAt,
    this.pendingActionBy,
    this.durationMinutes,
  });

  final String id;
  final String petId;
  final String? practiceId;
  final DateTime? scheduledAt;
  final String status;
  final String? notes;
  final DateTime? createdAt;
  final DateTime? proposedScheduledAt;
  final String? pendingActionBy;
  final int? durationMinutes;

  bool get isUpcoming {
    if (status == 'done' || status == 'cancelled') return false;
    if (scheduledAt != null) return scheduledAt!.isAfter(DateTime.now());
    return status == 'requested' || status == 'confirmed' || status == 'reschedule_pending';
  }

  bool get awaitingClient => pendingActionBy == 'client';

  DateTime get displayDate =>
      proposedScheduledAt ?? scheduledAt ?? createdAt ?? DateTime.now();

  factory Visit.fromJson(Map<String, dynamic> json) {
    return Visit(
      id: json['id'] as String? ?? '',
      petId: json['petId'] as String? ?? '',
      practiceId: json['practiceId'] as String?,
      scheduledAt: json['scheduledAt'] != null
          ? DateTime.tryParse(json['scheduledAt'] as String)
          : null,
      status: json['status'] as String? ?? 'requested',
      notes: json['notes'] as String?,
      createdAt: json['createdAt'] != null
          ? DateTime.tryParse(json['createdAt'] as String)
          : null,
      proposedScheduledAt: json['proposedScheduledAt'] != null
          ? DateTime.tryParse(json['proposedScheduledAt'] as String)
          : null,
      pendingActionBy: json['pendingActionBy'] as String?,
      durationMinutes: json['durationMinutes'] as int?,
    );
  }
}
