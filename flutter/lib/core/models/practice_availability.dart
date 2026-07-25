class PracticeAvailabilitySlot {
  const PracticeAvailabilitySlot({required this.start, this.end});

  final DateTime start;
  final DateTime? end;

  factory PracticeAvailabilitySlot.fromJson(Map<String, dynamic> json) {
    return PracticeAvailabilitySlot(
      start: DateTime.tryParse(json['start']?.toString() ?? '') ?? DateTime.fromMillisecondsSinceEpoch(0),
      end: json['end'] != null ? DateTime.tryParse(json['end'].toString()) : null,
    );
  }
}

class PracticeAvailability {
  const PracticeAvailability({
    required this.enabled,
    this.slots = const [],
    this.practicePhone = '',
    this.practiceName = '',
  });

  final bool enabled;
  final List<PracticeAvailabilitySlot> slots;
  final String practicePhone;
  final String practiceName;

  factory PracticeAvailability.fromJson(Map<String, dynamic> json) {
    final raw = json['slots'] as List<dynamic>? ?? const [];
    final slots = <PracticeAvailabilitySlot>[];
    for (final e in raw) {
      if (e is! Map) continue;
      final slot = PracticeAvailabilitySlot.fromJson(Map<String, dynamic>.from(e));
      if (slot.start.millisecondsSinceEpoch == 0) continue;
      slots.add(slot);
    }
    return PracticeAvailability(
      enabled: json['enabled'] == true,
      slots: slots,
      practicePhone: (json['practicePhone'] as String?)?.trim() ?? '',
      practiceName: (json['practiceName'] as String?)?.trim() ?? '',
    );
  }
}
