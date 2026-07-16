class Visit {
  const Visit({
    required this.id,
    required this.petId,
    this.scheduledAt,
    this.status = 'requested',
    this.notes,
    this.createdAt,
  });

  final String id;
  final String petId;
  final DateTime? scheduledAt;
  final String status;
  final String? notes;
  final DateTime? createdAt;

  bool get isUpcoming {
    if (status == 'done' || status == 'cancelled') return false;
    if (scheduledAt != null) return scheduledAt!.isAfter(DateTime.now());
    return status == 'requested' || status == 'confirmed';
  }

  DateTime get displayDate => scheduledAt ?? createdAt ?? DateTime.now();

  factory Visit.fromJson(Map<String, dynamic> json) {
    return Visit(
      id: json['id'] as String? ?? '',
      petId: json['petId'] as String? ?? '',
      scheduledAt: json['scheduledAt'] != null
          ? DateTime.tryParse(json['scheduledAt'] as String)
          : null,
      status: json['status'] as String? ?? 'requested',
      notes: json['notes'] as String?,
      createdAt: json['createdAt'] != null
          ? DateTime.tryParse(json['createdAt'] as String)
          : null,
    );
  }
}
