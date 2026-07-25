import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

void main() {
  test('blank comment yields null payload', () {
    expect(heartRateCommentPayload(null), isNull);
    expect(heartRateCommentPayload(''), isNull);
    expect(heartRateCommentPayload('   \n\t'), isNull);
  });

  test('trims non-empty comment', () {
    expect(heartRateCommentPayload('  agité  '), {'comment': 'agité'});
  });
}
