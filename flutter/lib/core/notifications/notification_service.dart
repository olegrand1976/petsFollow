import 'dart:convert';

import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/models/notification_prefs.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/notifications/reminder_controller.dart';
import 'package:timezone/timezone.dart' as tz;

class NotificationService {
  NotificationService._();
  static final instance = NotificationService._();

  bool _fcmRegistered = false;

  Future<void> init() async {
    await ReminderController.instance.init();
    await _registerFcmToken();
  }

  /// Call on logout so the next account can re-register its FCM token.
  void resetSession() {
    _fcmRegistered = false;
  }

  Future<void> onLogin() async {
    _fcmRegistered = false;
    await _registerFcmToken();
  }

  Future<void> _registerFcmToken() async {
    if (_fcmRegistered || ApiClient.instance.token == null) return;
    try {
      final messaging = FirebaseMessaging.instance;
      await messaging.requestPermission();
      final token = await messaging.getToken();
      if (token == null) return;
      final platform = switch (defaultTargetPlatform) {
        TargetPlatform.iOS => 'ios',
        TargetPlatform.android => 'android',
        _ => 'web',
      };
      await ApiClient.instance.putDeviceToken(token, platform);
      _fcmRegistered = true;
      if (kDebugMode) {
        debugPrint('FCM token registered ($platform)');
      }
    } catch (e) {
      if (kDebugMode) {
        debugPrint('FCM registration skipped: $e');
      }
    }
  }

  Future<NotificationPrefs> loadPrefs() async {
    if (ApiClient.instance.token == null) {
      return const NotificationPrefs(userId: '');
    }
    try {
      return await ApiClient.instance.getNotificationPrefs();
    } catch (_) {
      return const NotificationPrefs(userId: '');
    }
  }

  Future<NotificationPrefs> savePrefs(NotificationPrefs prefs) async {
    return ApiClient.instance.updateNotificationPrefs(prefs);
  }

  int _stableNotifId(String kind, String entityId) {
    final digest = utf8.encode('$kind:$entityId');
    var hash = 0;
    for (final b in digest) {
      hash = (hash * 31 + b) & 0x7fffffff;
    }
    return hash == 0 ? 1 : hash;
  }

  Future<void> cancelCareReminder(String reminderId) async {
    await ReminderController.instance.init();
    await ReminderController.instance.plugin.cancel(_stableNotifId('care', reminderId));
  }

  Future<void> scheduleCareReminders(List<CareReminder> reminders, {String? petName}) async {
    await ReminderController.instance.init();
    final prefs = await loadPrefs();
    if (!prefs.care) return;
    final plugin = ReminderController.instance.plugin;
    for (final r in reminders) {
      final id = _stableNotifId('care', r.id);
      await plugin.cancel(id);
      if (r.isDone || r.dueAt.isBefore(DateTime.now())) continue;
      final title = petName != null ? '$petName — ${r.title}' : r.title;
      await _scheduleAt(id, r.dueAt, title, r.title, 'pf_care');
    }
  }

  Future<void> scheduleVisits(List<Visit> visits, {required String visitLabel, String? petName}) async {
    await ReminderController.instance.init();
    final prefs = await loadPrefs();
    if (!prefs.visits) return;
    final plugin = ReminderController.instance.plugin;
    for (final v in visits) {
      final id = _stableNotifId('visit', v.id);
      await plugin.cancel(id);
      if (!v.isUpcoming || v.scheduledAt == null) continue;
      final title = petName != null ? '$petName — $visitLabel' : visitLabel;
      await _scheduleAt(id, v.scheduledAt!, title, v.notes ?? '', 'pf_visits');
    }
  }

  Future<void> _scheduleAt(int id, DateTime when, String title, String body, String channelId) async {
    if (when.isBefore(DateTime.now())) return;
    final scheduled = tz.TZDateTime.from(when, tz.local);
    final details = NotificationDetails(
      android: AndroidNotificationDetails(channelId, channelId, channelDescription: body),
      iOS: const DarwinNotificationDetails(),
    );
    await ReminderController.instance.plugin.zonedSchedule(
      id,
      title,
      body,
      scheduled,
      details,
      androidScheduleMode: AndroidScheduleMode.inexactAllowWhileIdle,
      uiLocalNotificationDateInterpretation: UILocalNotificationDateInterpretation.absoluteTime,
    );
  }
}
