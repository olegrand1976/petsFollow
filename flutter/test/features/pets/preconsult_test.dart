import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/preconsult.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';

void main() {
  test('PreconsultAnswers.toJson omits empty comment', () {
    const a = PreconsultAnswers(
      chiefComplaint: ' Vomiting ',
      duration: 'few_days',
      behavior: 'lethargic',
      appetite: 'decreased',
      thirst: 'normal',
      elimination: 'normal',
      urgency: 'high',
    );
    final json = a.toJson();
    expect(json['chiefComplaint'], 'Vomiting');
    expect(json.containsKey('comment'), isFalse);
    expect(json['urgency'], 'high');
  });

  test('Visit.preconsultPending requires confirmed + pending', () {
    final pending = Visit.fromJson({
      'id': 'v1',
      'petId': 'p1',
      'status': 'confirmed',
      'preconsultStatus': 'pending',
    });
    expect(pending.preconsultPending, isTrue);
    final submitted = Visit.fromJson({
      'id': 'v1',
      'petId': 'p1',
      'status': 'confirmed',
      'preconsultStatus': 'submitted',
    });
    expect(submitted.preconsultPending, isFalse);
  });

  test('PreconsultIntake.fromJson parses answers', () {
    final in_ = PreconsultIntake.fromJson({
      'id': 'i1',
      'visitId': 'v1',
      'status': 'submitted',
      'petName': 'Bella',
      'answers': {
        'chiefComplaint': 'Cough',
        'duration': 'week',
        'behavior': 'normal',
        'appetite': 'normal',
        'thirst': 'increased',
        'elimination': 'unknown',
        'urgency': 'low',
        'comment': 'night only',
      },
    });
    expect(in_.isSubmitted, isTrue);
    expect(in_.petName, 'Bella');
    expect(in_.answers.thirst, 'increased');
    expect(in_.answers.comment, 'night only');
  });
}
