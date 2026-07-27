import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/features/heartrate/supports_heart_rate.dart';

void main() {
  test('supports dog/cat/horse only', () {
    expect(supportsHeartRateControl('dog'), isTrue);
    expect(supportsHeartRateControl('cat'), isTrue);
    expect(supportsHeartRateControl('horse'), isTrue);
    expect(supportsHeartRateControl('other'), isFalse);
    expect(supportsHeartRateControl(null), isFalse);
    expect(supportsHeartRateControl(''), isFalse);
  });
}
