import 'package:petsfollow_mobile/core/api/api_client.dart';
import 'package:petsfollow_mobile/core/models/discovery_card.dart';
import 'package:petsfollow_mobile/core/models/discovery_progress.dart';
import 'package:shared_preferences/shared_preferences.dart';

class DiscoveryController {
  DiscoveryController._();
  static final instance = DiscoveryController._();

  static const _startedAtKeyPrefix = 'pf_discovery_started_at';
  static const _completedKeyPrefix = 'pf_discovery_completed';

  String? _userId;

  void bindUser(String? userId) {
    _userId = (userId != null && userId.isNotEmpty) ? userId : null;
  }

  Future<void> clearLocal() async {
    final sp = await SharedPreferences.getInstance();
    final uid = _userId;
    if (uid != null) {
      await sp.remove('${_startedAtKeyPrefix}_$uid');
      await sp.remove('${_completedKeyPrefix}_$uid');
    }
    // legacy unscoped keys
    await sp.remove('pf_discovery_started_at');
    await sp.remove('pf_discovery_completed');
    _userId = null;
  }

  String get _startedAtKey =>
      _userId != null ? '${_startedAtKeyPrefix}_$_userId' : 'pf_discovery_started_at';

  String get _completedKey =>
      _userId != null ? '${_completedKeyPrefix}_$_userId' : 'pf_discovery_completed';

  Future<DiscoveryProgress> load() async {
    if (ApiClient.instance.token != null) {
      try {
        final me = await ApiClient.instance.getMe();
        bindUser(me['userId'] as String? ?? me['id'] as String?);
      } catch (_) {}
    }
    final local = await _loadLocal();
    if (ApiClient.instance.token == null) return local;
    try {
      final remote = await ApiClient.instance.getDiscovery();
      bindUser(remote.userId.isNotEmpty ? remote.userId : _userId);
      await _saveLocal(remote);
      return remote;
    } catch (_) {
      return local;
    }
  }

  Future<DiscoveryProgress> completeCard(String cardKey) async {
    final local = await _loadLocal();
    final completed = {...local.completedCards};
    completed.add(cardKey);
    var progress = DiscoveryProgress(
      userId: local.userId.isNotEmpty ? local.userId : (_userId ?? ''),
      startedAt: local.startedAt,
      completedCards: completed.toList(),
      streakDays: local.streakDays,
    );
    await _saveLocal(progress);
    if (ApiClient.instance.token != null) {
      try {
        progress = await ApiClient.instance.completeDiscoveryCard(cardKey);
        await _saveLocal(progress);
      } catch (_) {}
    }
    return progress;
  }

  DiscoveryCard? missionCardForToday(List<DiscoveryCard> cards, DiscoveryProgress progress) {
    for (final dayIndex in DiscoveryCard.journeyDays) {
      if (!progress.isCardUnlocked(dayIndex)) continue;
      if (progress.isCardCompleted(dayIndex)) continue;
      for (final c in cards) {
        if (c.dayIndex == dayIndex) return c;
      }
    }
    return null;
  }

  List<DiscoveryCard> cardsWithProgress(List<DiscoveryCard> cards, DiscoveryProgress progress) {
    return cards
        .map(
          (c) => DiscoveryCard(
            dayIndex: c.dayIndex,
            title: c.title,
            body: c.body,
            completed: progress.isCardCompleted(c.dayIndex),
            locked: !progress.isCardUnlocked(c.dayIndex),
          ),
        )
        .toList();
  }

  Future<DiscoveryProgress> _loadLocal() async {
    final sp = await SharedPreferences.getInstance();
    final startedRaw = sp.getString(_startedAtKey);
    final startedAt = startedRaw != null ? DateTime.tryParse(startedRaw) ?? DateTime.now() : DateTime.now();
    if (startedRaw == null) {
      await sp.setString(_startedAtKey, startedAt.toIso8601String());
    }
    final completedRaw = sp.getStringList(_completedKey) ?? [];
    return DiscoveryProgress(
      userId: _userId ?? '',
      startedAt: startedAt,
      completedCards: completedRaw,
    );
  }

  Future<void> _saveLocal(DiscoveryProgress progress) async {
    if (progress.userId.isNotEmpty) {
      bindUser(progress.userId);
    }
    final sp = await SharedPreferences.getInstance();
    await sp.setString(_startedAtKey, progress.startedAt.toIso8601String());
    await sp.setStringList(_completedKey, progress.completedCards);
  }
}
