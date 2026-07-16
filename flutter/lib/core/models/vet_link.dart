class VetLink {
  const VetLink({
    required this.practiceId,
    required this.vetEmail,
    required this.vetFullName,
    required this.practiceName,
    this.vetUserId,
    this.linkedAt,
    this.isPrimary = false,
  });

  final String practiceId;
  final String vetEmail;
  final String vetFullName;
  final String practiceName;
  final String? vetUserId;
  final String? linkedAt;
  final bool isPrimary;

  factory VetLink.fromJson(Map<String, dynamic> json, {String? primaryPracticeId}) {
    final practiceId = json['practiceId'] as String? ?? '';
    return VetLink(
      practiceId: practiceId,
      vetEmail: json['vetEmail'] as String? ?? '',
      vetFullName: json['vetFullName'] as String? ?? json['vetName'] as String? ?? '',
      practiceName: json['practiceName'] as String? ?? '',
      vetUserId: json['vetUserId'] as String?,
      linkedAt: json['linkedAt'] as String?,
      isPrimary: primaryPracticeId != null && practiceId == primaryPracticeId,
    );
  }
}
