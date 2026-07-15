import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:timezone/data/latest.dart' as tz_data;
import 'package:timezone/timezone.dart' as tz;

class ReminderPrefs {
  const ReminderPrefs({required this.enabled, required this.hour, required this.minute});

  final bool enabled;
  final int hour;
  final int minute;
}

class ReminderController {
  ReminderController._();
  static final instance = ReminderController._();

  static const _enabledKey = 'pf_reminder_enabled';
  static const _hourKey = 'pf_reminder_hour';
  static const _minuteKey = 'pf_reminder_minute';
  static const _notificationId = 42;

  final FlutterLocalNotificationsPlugin _plugin = FlutterLocalNotificationsPlugin();
  bool _initialized = false;

  Future<void> init() async {
    if (_initialized) return;
    tz_data.initializeTimeZones();
    const android = AndroidInitializationSettings('@mipmap/ic_launcher');
    const ios = DarwinInitializationSettings();
    await _plugin.initialize(const InitializationSettings(android: android, iOS: ios));
    _initialized = true;
    final prefs = await load();
    if (prefs.enabled) {
      await _schedule(prefs.hour, prefs.minute);
    }
  }

  Future<ReminderPrefs> load() async {
    final sp = await SharedPreferences.getInstance();
    return ReminderPrefs(
      enabled: sp.getBool(_enabledKey) ?? false,
      hour: sp.getInt(_hourKey) ?? 20,
      minute: sp.getInt(_minuteKey) ?? 0,
    );
  }

  Future<void> save({required bool enabled, required int hour, required int minute}) async {
    final sp = await SharedPreferences.getInstance();
    await sp.setBool(_enabledKey, enabled);
    await sp.setInt(_hourKey, hour);
    await sp.setInt(_minuteKey, minute);
    await _plugin.cancel(_notificationId);
    if (enabled) {
      await _schedule(hour, minute);
    }
  }

  Future<void> _schedule(int hour, int minute) async {
    await init();
    final now = tz.TZDateTime.now(tz.local);
    var scheduled = tz.TZDateTime(tz.local, now.year, now.month, now.day, hour, minute);
    if (scheduled.isBefore(now)) {
      scheduled = scheduled.add(const Duration(days: 1));
    }
    const details = NotificationDetails(
      android: AndroidNotificationDetails(
        'pf_reminders',
        'Measurement reminders',
        channelDescription: 'Daily heart rate measurement reminders',
      ),
      iOS: DarwinNotificationDetails(),
    );
    await _plugin.zonedSchedule(
      _notificationId,
      'petsFollow',
      'Time for a heart rate reading for your pet',
      scheduled,
      details,
      androidScheduleMode: AndroidScheduleMode.inexactAllowWhileIdle,
      matchDateTimeComponents: DateTimeComponents.time,
    );
  }
}
