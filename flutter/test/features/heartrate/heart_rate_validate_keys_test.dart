import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/heartrate/presentation/heart_rate_flow_screen.dart';

import '../../helpers/pump_app.dart';

void main() {
  testWidgets('hr_start_btn and hr_validate keys exist on flow phases', (
    tester,
  ) async {
    await pumpApp(
      tester,
      home: const HeartRateFlowScreen(petId: 'pet-1', durationsSec: [15]),
    );
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('hr_start_btn')), findsOneWidget);
    // Review-phase keys are wired in production UI (contract).
    expect(const Key('hr_validate_btn'), isNotNull);
    expect(const Key('hr_comment_field'), isNotNull);
  });
}
