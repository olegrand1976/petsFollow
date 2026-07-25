class PreconsultAnswers {
  const PreconsultAnswers({
    this.chiefComplaint = '',
    this.duration = 'unknown',
    this.behavior = 'unknown',
    this.appetite = 'unknown',
    this.thirst = 'unknown',
    this.elimination = 'unknown',
    this.urgency = 'medium',
    this.comment = '',
  });

  final String chiefComplaint;
  final String duration;
  final String behavior;
  final String appetite;
  final String thirst;
  final String elimination;
  final String urgency;
  final String comment;

  Map<String, dynamic> toJson() => {
        'chiefComplaint': chiefComplaint.trim(),
        'duration': duration,
        'behavior': behavior,
        'appetite': appetite,
        'thirst': thirst,
        'elimination': elimination,
        'urgency': urgency,
        if (comment.trim().isNotEmpty) 'comment': comment.trim(),
      };

  factory PreconsultAnswers.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const PreconsultAnswers();
    return PreconsultAnswers(
      chiefComplaint: json['chiefComplaint'] as String? ?? '',
      duration: json['duration'] as String? ?? 'unknown',
      behavior: json['behavior'] as String? ?? 'unknown',
      appetite: json['appetite'] as String? ?? 'unknown',
      thirst: json['thirst'] as String? ?? 'unknown',
      elimination: json['elimination'] as String? ?? 'unknown',
      urgency: json['urgency'] as String? ?? 'medium',
      comment: json['comment'] as String? ?? '',
    );
  }
}

class PreconsultIntake {
  const PreconsultIntake({
    required this.id,
    required this.visitId,
    required this.status,
    required this.answers,
    this.petId,
    this.petName,
    this.submittedAt,
  });

  final String id;
  final String visitId;
  final String status;
  final PreconsultAnswers answers;
  final String? petId;
  final String? petName;
  final DateTime? submittedAt;

  bool get isPending => status == 'pending';
  bool get isSubmitted => status == 'submitted';

  factory PreconsultIntake.fromJson(Map<String, dynamic> json) {
    return PreconsultIntake(
      id: json['id'] as String? ?? '',
      visitId: json['visitId'] as String? ?? '',
      status: json['status'] as String? ?? 'pending',
      answers: PreconsultAnswers.fromJson(
        json['answers'] is Map
            ? Map<String, dynamic>.from(json['answers'] as Map)
            : null,
      ),
      petId: json['petId'] as String?,
      petName: json['petName'] as String?,
      submittedAt: json['submittedAt'] != null
          ? DateTime.tryParse(json['submittedAt'] as String)
          : null,
    );
  }
}
