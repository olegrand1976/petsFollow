class CareReminder {
  const CareReminder({
    required this.id,
    required this.petId,
    required this.title,
    required this.dueAt,
    this.status = 'pending',
    this.type,
    this.recurrenceDays,
  });

  final String id;
  final String petId;
  final String title;
  final DateTime dueAt;
  final String status;
  final String? type;
  final int? recurrenceDays;

  bool get isDone => status == 'done';

  bool get isOverdue => !isDone && dueAt.isBefore(DateTime.now());

  bool get hasRecurrence => recurrenceDays != null && recurrenceDays! > 0;

  /// Échéance = date de référence (+ récurrence si définie).
  static DateTime computeDueAt(DateTime referenceDate, int? recurrenceDays) {
    final day = DateTime(referenceDate.year, referenceDate.month, referenceDate.day);
    if (recurrenceDays == null || recurrenceDays <= 0) return day;
    return day.add(Duration(days: recurrenceDays));
  }

  factory CareReminder.fromJson(Map<String, dynamic> json) {
    final status = json['status'] as String? ?? 'pending';
    final rawRecurrence = json['recurrenceDays'];
    int? recurrenceDays;
    if (rawRecurrence is int) {
      recurrenceDays = rawRecurrence;
    } else if (rawRecurrence is num) {
      recurrenceDays = rawRecurrence.toInt();
    }
    return CareReminder(
      id: json['id'] as String? ?? '',
      petId: json['petId'] as String? ?? '',
      title: json['title'] as String? ?? '',
      dueAt: DateTime.tryParse(json['dueAt'] as String? ?? '') ?? DateTime.now(),
      status: status,
      type: json['type'] as String?,
      recurrenceDays: recurrenceDays,
    );
  }
}
