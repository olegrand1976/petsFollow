import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter_timezone/flutter_timezone.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:timezone/data/latest.dart' as tz_data;
import 'package:timezone/timezone.dart' as tz;

class ReminderPrefs {
  const ReminderPrefs({
    required this.enabled,
    required this.hour,
    required this.minute,
    this.notificationTitle,
    this.notificationBody,
  });

  final bool enabled;
  final int hour;
  final int minute;
  final String? notificationTitle;
  final String? notificationBody;
}

class ReminderController {
  ReminderController._();
  static final instance = ReminderController._();

  static const _enabledKey = 'pf_reminder_enabled';
  static const _hourKey = 'pf_reminder_hour';
  static const _minuteKey = 'pf_reminder_minute';
  static const _titleKey = 'pf_reminder_title';
  static const _bodyKey = 'pf_reminder_body';
  static const _notificationId = 42;

  final FlutterLocalNotificationsPlugin _plugin = FlutterLocalNotificationsPlugin();
  bool _initialized = false;

  FlutterLocalNotificationsPlugin get plugin => _plugin;

  Future<void> init({DidReceiveNotificationResponseCallback? onNotificationTap}) async {
    if (_initialized) return;
    tz_data.initializeTimeZones();
    try {
      final name = await FlutterTimezone.getLocalTimezone();
      tz.setLocalLocation(tz.getLocation(name));
    } catch (_) {
      try {
        tz.setLocalLocation(tz.getLocation('Europe/Brussels'));
      } catch (_) {
        // keep tz.UTC fallback
      }
    }
    const android = AndroidInitializationSettings('@mipmap/ic_launcher');
    const ios = DarwinInitializationSettings();
    await _plugin.initialize(
      const InitializationSettings(android: android, iOS: ios),
      onDidReceiveNotificationResponse: onNotificationTap,
    );
    _initialized = true;
    final prefs = await load();
    if (prefs.enabled && prefs.notificationTitle != null && prefs.notificationBody != null) {
      await _schedule(
        prefs.hour,
        prefs.minute,
        title: prefs.notificationTitle!,
        body: prefs.notificationBody!,
      );
    }
  }

  Future<ReminderPrefs> load() async {
    final sp = await SharedPreferences.getInstance();
    return ReminderPrefs(
      enabled: sp.getBool(_enabledKey) ?? false,
      hour: sp.getInt(_hourKey) ?? 20,
      minute: sp.getInt(_minuteKey) ?? 0,
      notificationTitle: sp.getString(_titleKey),
      notificationBody: sp.getString(_bodyKey),
    );
  }

  Future<void> save({
    required bool enabled,
    required int hour,
    required int minute,
    required String notificationTitle,
    required String notificationBody,
  }) async {
    final sp = await SharedPreferences.getInstance();
    await sp.setBool(_enabledKey, enabled);
    await sp.setInt(_hourKey, hour);
    await sp.setInt(_minuteKey, minute);
    await sp.setString(_titleKey, notificationTitle);
    await sp.setString(_bodyKey, notificationBody);
    await _plugin.cancel(_notificationId);
    if (enabled) {
      await _schedule(
        hour,
        minute,
        title: notificationTitle,
        body: notificationBody,
      );
    }
  }

  Future<void> _schedule(
    int hour,
    int minute, {
    required String title,
    required String body,
  }) async {
    await init();
    final now = tz.TZDateTime.now(tz.local);
    var scheduled = tz.TZDateTime(tz.local, now.year, now.month, now.day, hour, minute);
    if (scheduled.isBefore(now)) {
      scheduled = scheduled.add(const Duration(days: 1));
    }
    final details = NotificationDetails(
      android: AndroidNotificationDetails(
        'pf_reminders',
        title,
        channelDescription: body,
      ),
      iOS: DarwinNotificationDetails(),
    );
    await _plugin.zonedSchedule(
      _notificationId,
      title,
      body,
      scheduled,
      details,
      androidScheduleMode: AndroidScheduleMode.inexactAllowWhileIdle,
      uiLocalNotificationDateInterpretation: UILocalNotificationDateInterpretation.absoluteTime,
      matchDateTimeComponents: DateTimeComponents.time,
    );
  }
}
