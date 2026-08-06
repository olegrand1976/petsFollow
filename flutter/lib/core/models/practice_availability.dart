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

class PracticeBookableSite {
  const PracticeBookableSite({
    required this.siteId,
    required this.siteName,
    required this.enabled,
    this.phone = '',
  });

  final String siteId;
  final String siteName;
  final bool enabled;
  final String phone;

  factory PracticeBookableSite.fromJson(Map<String, dynamic> json) {
    return PracticeBookableSite(
      siteId: (json['siteId'] as String?)?.trim() ?? '',
      siteName: (json['siteName'] as String?)?.trim() ?? '',
      enabled: json['enabled'] == true,
      phone: (json['phone'] as String?)?.trim() ?? '',
    );
  }
}

class PracticeAvailability {
  const PracticeAvailability({
    required this.enabled,
    this.slots = const [],
    this.practicePhone = '',
    this.practiceName = '',
    this.siteId = '',
    this.siteName = '',
    this.sites = const [],
  });

  final bool enabled;
  final List<PracticeAvailabilitySlot> slots;
  final String practicePhone;
  final String practiceName;
  final String siteId;
  final String siteName;
  final List<PracticeBookableSite> sites;

  factory PracticeAvailability.fromJson(Map<String, dynamic> json) {
    final raw = json['slots'] as List<dynamic>? ?? const [];
    final slots = <PracticeAvailabilitySlot>[];
    for (final e in raw) {
      if (e is! Map) continue;
      final slot = PracticeAvailabilitySlot.fromJson(Map<String, dynamic>.from(e));
      if (slot.start.millisecondsSinceEpoch == 0) continue;
      slots.add(slot);
    }
    final rawSites = json['sites'] as List<dynamic>? ?? const [];
    final sites = <PracticeBookableSite>[];
    for (final e in rawSites) {
      if (e is! Map) continue;
      final s = PracticeBookableSite.fromJson(Map<String, dynamic>.from(e));
      if (s.siteId.isEmpty) continue;
      sites.add(s);
    }
    return PracticeAvailability(
      enabled: json['enabled'] == true,
      slots: slots,
      practicePhone: (json['practicePhone'] as String?)?.trim() ?? '',
      practiceName: (json['practiceName'] as String?)?.trim() ?? '',
      siteId: (json['siteId'] as String?)?.trim() ?? '',
      siteName: (json['siteName'] as String?)?.trim() ?? '',
      sites: sites,
    );
  }
}
