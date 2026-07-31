import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/models/manager_overview.dart';

void main() {
  test('ManagerOverview.fromJson parses team + self', () {
    final ov = ManagerOverview.fromJson({
      'teamProspectsTotal': 10,
      'teamProspectsConverted': 3,
      'conversionRateBps': 1250,
      'teamMonthEarnedCents': 3500,
      'team': [
        {
          'userId': 'u1',
          'fullName': 'Alex',
          'email': 'a@test',
          'prospectsConverted': 2,
          'monthEarnedCents': 2000,
          'staleInPipeline': 1,
        },
      ],
      'self': {
        'assignedVets': 4,
        'prospectsTotal': 5,
        'prospectsConverted': 1,
        'monthEarnedCents': 500,
      },
    });
    expect(ov.teamProspectsTotal, 10);
    expect(ov.conversionRateLabel, '12.5 %');
    expect(ov.team, hasLength(1));
    expect(ov.team.first.fullName, 'Alex');
    expect(ov.self.assignedVets, 4);
    expect(formatEuroCents(3500), '35 €');
    expect(formatEuroCents(3550), '35.50 €');
  });
}
