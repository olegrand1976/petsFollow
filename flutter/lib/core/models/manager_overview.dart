/// Team member row from `GET /commercial-manager/overview` / `…/team`.
class ManagerTeamMember {
  const ManagerTeamMember({
    required this.userId,
    required this.fullName,
    required this.email,
    this.assignedVets = 0,
    this.prospectsConverted = 0,
    this.contacts30d = 0,
    this.appointmentsUpcoming = 0,
    this.appointmentsDone = 0,
    this.staleInPipeline = 0,
    this.monthEarnedCents = 0,
    this.lifetimeEarnedCents = 0,
  });

  final String userId;
  final String fullName;
  final String email;
  final int assignedVets;
  final int prospectsConverted;
  final int contacts30d;
  final int appointmentsUpcoming;
  final int appointmentsDone;
  final int staleInPipeline;
  final int monthEarnedCents;
  final int lifetimeEarnedCents;

  factory ManagerTeamMember.fromJson(Map<String, dynamic> json) {
    int i(String k) => (json[k] as num?)?.toInt() ?? 0;
    return ManagerTeamMember(
      userId: json['userId'] as String? ?? '',
      fullName: json['fullName'] as String? ?? '',
      email: json['email'] as String? ?? '',
      assignedVets: i('assignedVets'),
      prospectsConverted: i('prospectsConverted'),
      contacts30d: i('contacts30d'),
      appointmentsUpcoming: i('appointmentsUpcoming'),
      appointmentsDone: i('appointmentsDone'),
      staleInPipeline: i('staleInPipeline'),
      monthEarnedCents: i('monthEarnedCents'),
      lifetimeEarnedCents: i('lifetimeEarnedCents'),
    );
  }
}

/// Personal commercial KPIs (manager `self` or member overview).
class CommercialSelfStats {
  const CommercialSelfStats({
    this.assignedVets = 0,
    this.prospectsTotal = 0,
    this.prospectsConverted = 0,
    this.monthEarnedCents = 0,
    this.lifetimeEarnedCents = 0,
    this.appointmentsUpcoming = 0,
    this.staleInPipeline = 0,
  });

  final int assignedVets;
  final int prospectsTotal;
  final int prospectsConverted;
  final int monthEarnedCents;
  final int lifetimeEarnedCents;
  final int appointmentsUpcoming;
  final int staleInPipeline;

  factory CommercialSelfStats.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const CommercialSelfStats();
    int i(String k) => (json[k] as num?)?.toInt() ?? 0;
    return CommercialSelfStats(
      assignedVets: i('assignedVets'),
      prospectsTotal: i('prospectsTotal'),
      prospectsConverted: i('prospectsConverted'),
      monthEarnedCents: i('monthEarnedCents'),
      lifetimeEarnedCents: i('lifetimeEarnedCents'),
      appointmentsUpcoming: i('appointmentsUpcoming'),
      staleInPipeline: i('staleInPipeline'),
    );
  }
}

/// `GET /api/v1/commercial-manager/overview`.
class ManagerOverview {
  const ManagerOverview({
    this.team = const [],
    this.teamProspectsTotal = 0,
    this.teamProspectsContacted = 0,
    this.teamProspectsConverted = 0,
    this.teamAppointmentsUpcoming = 0,
    this.teamStaleInPipeline = 0,
    this.teamMonthEarnedCents = 0,
    this.directoryTotal = 0,
    this.conversionRateBps = 0,
    this.self = const CommercialSelfStats(),
  });

  final List<ManagerTeamMember> team;
  final int teamProspectsTotal;
  final int teamProspectsContacted;
  final int teamProspectsConverted;
  final int teamAppointmentsUpcoming;
  final int teamStaleInPipeline;
  final int teamMonthEarnedCents;
  final int directoryTotal;
  final int conversionRateBps;
  final CommercialSelfStats self;

  /// Conversion rate as percent string, e.g. `12.5 %`, or `—` if zero.
  String get conversionRateLabel {
    if (conversionRateBps <= 0) return '—';
    return '${(conversionRateBps / 100).toStringAsFixed(1)} %';
  }

  factory ManagerOverview.fromJson(Map<String, dynamic> json) {
    int i(String k) => (json[k] as num?)?.toInt() ?? 0;
    final rawTeam = json['team'];
    final team = rawTeam is List
        ? rawTeam
            .whereType<Map>()
            .map((e) => ManagerTeamMember.fromJson(Map<String, dynamic>.from(e)))
            .toList()
        : <ManagerTeamMember>[];
    final selfRaw = json['self'];
    return ManagerOverview(
      team: team,
      teamProspectsTotal: i('teamProspectsTotal'),
      teamProspectsContacted: i('teamProspectsContacted'),
      teamProspectsConverted: i('teamProspectsConverted'),
      teamAppointmentsUpcoming: i('teamAppointmentsUpcoming'),
      teamStaleInPipeline: i('teamStaleInPipeline'),
      teamMonthEarnedCents: i('teamMonthEarnedCents'),
      directoryTotal: i('directoryTotal'),
      conversionRateBps: i('conversionRateBps'),
      self: CommercialSelfStats.fromJson(
        selfRaw is Map ? Map<String, dynamic>.from(selfRaw) : null,
      ),
    );
  }
}

String formatEuroCents(int cents) {
  final euros = cents / 100.0;
  return '${euros.toStringAsFixed(euros.truncateToDouble() == euros ? 0 : 2)} €';
}
