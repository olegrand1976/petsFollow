import 'package:in_app_review/in_app_review.dart';
import 'package:shared_preferences/shared_preferences.dart';

class InAppReviewHelper {
  InAppReviewHelper._();

  static const _lastAskKey = 'pf_review_last_ask';
  static const _cooldownDays = 90;

  static Future<bool> shouldShowDialog() async {
    final sp = await SharedPreferences.getInstance();
    final lastRaw = sp.getString(_lastAskKey);
    if (lastRaw == null) return true;
    final last = DateTime.tryParse(lastRaw);
    if (last == null) return true;
    return DateTime.now().difference(last).inDays >= _cooldownDays;
  }

  static Future<void> recordAsked() async {
    final sp = await SharedPreferences.getInstance();
    await sp.setString(_lastAskKey, DateTime.now().toIso8601String());
  }

  static Future<void> openStoreReview() async {
    final review = InAppReview.instance;
    if (await review.isAvailable()) {
      await review.requestReview();
    }
  }
}
