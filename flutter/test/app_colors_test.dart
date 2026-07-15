import 'package:flutter_test/flutter_test.dart';
import 'package:petsfollow_mobile/core/theme/app_colors.dart';

void main() {
  test('brand colors defined', () {
    expect(AppColors.primary, isNotNull);
    expect(AppColors.accent.value, greaterThan(0));
  });
}
