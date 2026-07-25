import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/deeplink/payment_deeplink.dart';

void main() {
  test('appDeepLinkPath normalizes payment success', () {
    expect(
      appDeepLinkPath(Uri.parse('petsfollow://payment/success')),
      'payment/success',
    );
  });

  test('appDeepLinkPath normalizes invite', () {
    expect(
      appDeepLinkPath(Uri.parse('petsfollow://invite?code=abc')),
      'invite',
    );
  });
}
