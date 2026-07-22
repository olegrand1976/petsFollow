import 'dart:convert';

import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/locale/locale_controller.dart';
import 'package:petsfollow_mobile/core/models/care_reminder.dart';
import 'package:petsfollow_mobile/core/models/notification_prefs.dart';
import 'package:petsfollow_mobile/core/models/visit.dart';
import 'package:petsfollow_mobile/core/notifications/push_navigation.dart';
import 'package:petsfollow_mobile/core/notifications/reminder_controller.dart';
import 'package:petsfollow_mobile/l10n/app_localizations.dart';
import 'package:timezone/timezone.dart' as tz;

class NotificationService {
  NotificationService._();
  static final instance = NotificationService._();

  bool _fcmRegistered = false;
  bool _handlersBound = false;
  int _localPushId = 900000;

  Future<void> init() async {
    await ReminderController.instance.init(
      onNotificationTap: (response) {
        final payload = response.payload;
        if (payload == null || payload.isEmpty) return;
        try {
          final data = jsonDecode(payload) as Map<String, dynamic>;
          PushNavigation.instance.handlePushData(data);
        } catch (_) {}
      },
    );
    await _ensureAndroidChannels();
    await _bindFcmHandlers();
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

  Future<void> _ensureAndroidChannels() async {
    final android = ReminderController.instance.plugin
        .resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>();
    if (android == null) return;
    final l10n = lookupAppLocalizations(LocaleController.instance.locale);
    await android.createNotificationChannel(
      AndroidNotificationChannel('pf_messages', l10n.notifChannelMessages, importance: Importance.high),
    );
    await android.createNotificationChannel(
      AndroidNotificationChannel('pf_visits', l10n.notifChannelVisits, importance: Importance.high),
    );
    await android.createNotificationChannel(
      AndroidNotificationChannel('pf_care', l10n.notifChannelCare, importance: Importance.defaultImportance),
    );
  }

  Future<void> _bindFcmHandlers() async {
    if (_handlersBound) return;
    _handlersBound = true;
    FirebaseMessaging.onMessage.listen(_onForegroundMessage);
    FirebaseMessaging.onMessageOpenedApp.listen(_onMessageOpened);
    final initial = await FirebaseMessaging.instance.getInitialMessage();
    if (initial != null) {
      _onMessageOpened(initial);
    }
  }

  void _onForegroundMessage(RemoteMessage message) {
    final data = Map<String, dynamic>.from(message.data);
    final type = data['type']?.toString() ?? '';
    if (type == 'message') {
      PushNavigation.instance.bumpMessageRefresh();
    }
    final title = message.notification?.title ?? _fallbackTitle(type);
    final body = message.notification?.body ?? '';
    _showLocalPush(title: title, body: body, data: data);
  }

  void _onMessageOpened(RemoteMessage message) {
    final data = Map<String, dynamic>.from(message.data);
    PushNavigation.instance.handlePushData(data);
  }

  String _fallbackTitle(String type) {
    final l10n = lookupAppLocalizations(LocaleController.instance.locale);
    return switch (type) {
      'message' => l10n.pushNewMessage,
      'visit_confirmed' => l10n.pushVisitConfirmed,
      'visit_proposed' => l10n.pushVisitProposed,
      'visit_reschedule' => l10n.pushVisitReschedule,
      _ => 'petsFollow',
    };
  }

  bool _isVisitPush(String type) =>
      type == 'visit_confirmed' || type == 'visit_proposed' || type == 'visit_reschedule';

  Future<void> _showLocalPush({
    required String title,
    required String body,
    required Map<String, dynamic> data,
  }) async {
    final type = data['type']?.toString() ?? '';
    final channelId = _isVisitPush(type) ? 'pf_visits' : 'pf_messages';
    final id = _localPushId++;
    final details = NotificationDetails(
      android: AndroidNotificationDetails(
        channelId,
        channelId,
        importance: Importance.high,
        priority: Priority.high,
      ),
      iOS: const DarwinNotificationDetails(),
    );
    await ReminderController.instance.plugin.show(
      id,
      title,
      body,
      details,
      payload: jsonEncode(data),
    );
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
      messaging.onTokenRefresh.listen((newToken) async {
        if (ApiClient.instance.token == null) return;
        try {
          await ApiClient.instance.putDeviceToken(newToken, platform);
        } catch (_) {}
      });
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
    try {
      await ReminderController.instance.init();
      await ReminderController.instance.plugin.cancel(_stableNotifId('care', reminderId));
    } catch (_) {}
  }

  Future<void> scheduleCareReminders(List<CareReminder> reminders, {String? petName}) async {
    try {
      await ReminderController.instance.init();
      final prefs = await loadPrefs();
      if (!prefs.care) return;
      final plugin = ReminderController.instance.plugin;
      for (final r in reminders) {
        final id = _stableNotifId('care', r.id);
        try {
          await plugin.cancel(id);
          if (r.isDone || r.dueAt.isBefore(DateTime.now())) continue;
          final title = petName != null ? '$petName — ${r.title}' : r.title;
          await _scheduleAt(id, r.dueAt, title, r.title, 'pf_care');
        } catch (_) {
          // Skip one reminder; keep scheduling the rest.
        }
      }
    } catch (_) {
      // Local notifications must never break care list loading.
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
