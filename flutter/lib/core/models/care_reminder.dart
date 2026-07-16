class CareReminder {
  const CareReminder({
    required this.id,
    required this.petId,
    required this.title,
    required this.dueAt,
    this.status = 'pending',
    this.type,
  });

  final String id;
  final String petId;
  final String title;
  final DateTime dueAt;
  final String status;
  final String? type;

  bool get isDone => status == 'done';

  bool get isOverdue => !isDone && dueAt.isBefore(DateTime.now());

  factory CareReminder.fromJson(Map<String, dynamic> json) {
    final status = json['status'] as String? ?? 'pending';
    return CareReminder(
      id: json['id'] as String? ?? '',
      petId: json['petId'] as String? ?? '',
      title: json['title'] as String? ?? '',
      dueAt: DateTime.tryParse(json['dueAt'] as String? ?? '') ?? DateTime.now(),
      status: status,
      type: json['type'] as String?,
    );
  }
}
