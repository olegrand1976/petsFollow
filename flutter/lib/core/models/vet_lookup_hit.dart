class VetLookupHit {
  const VetLookupHit({
    required this.vetUserId,
    required this.practiceId,
    required this.practiceName,
    required this.vetFullName,
    required this.vetEmail,
  });

  final String vetUserId;
  final String practiceId;
  final String practiceName;
  final String vetFullName;
  final String vetEmail;

  factory VetLookupHit.fromJson(Map<String, dynamic> json) {
    return VetLookupHit(
      vetUserId: (json['vetUserId'] as String?) ?? '',
      practiceId: (json['practiceId'] as String?) ?? '',
      practiceName: (json['practiceName'] as String?) ?? '',
      vetFullName: (json['vetFullName'] as String?) ?? '',
      vetEmail: (json['vetEmail'] as String?) ?? '',
    );
  }
}
