class DiscoveryProgress {
  const DiscoveryProgress({
    required this.userId,
    required this.startedAt,
    this.completedCards = const [],
    this.streakDays = 0,
  });

  final String userId;
  final DateTime startedAt;
  final List<String> completedCards;
  final int streakDays;

  /// Stage indices (API keys remain `day0`/`day2`/`day4`/`day6` for compat).
  static const journeyDays = [0, 2, 4, 6];

  factory DiscoveryProgress.fromJson(Map<String, dynamic> json) {
    final raw = json['completedCards'] ?? json['completed_cards'];
    List<String> cards = const [];
    if (raw is List) {
      cards = raw.map((e) => e.toString().trim()).where((e) => e.isNotEmpty).toList();
    } else if (raw is String && raw.trim().isNotEmpty) {
      // Defensive: some gateways double-encode JSONB as a string.
      final trimmed = raw.trim();
      if (trimmed.startsWith('[')) {
        try {
          final decoded = trimmed; // parsed below via simple split if needed
          final inner = decoded.replaceAll(RegExp(r'[\[\]"]'), '');
          cards = inner.split(',').map((e) => e.trim()).where((e) => e.isNotEmpty).toList();
        } catch (_) {
          cards = const [];
        }
      }
    }
    return DiscoveryProgress(
      userId: json['userId'] as String? ?? json['user_id'] as String? ?? '',
      startedAt: DateTime.tryParse(json['startedAt'] as String? ?? json['started_at'] as String? ?? '') ??
          DateTime.now(),
      completedCards: cards,
      streakDays: json['streakDays'] as int? ?? json['streak_days'] as int? ?? 0,
    );
  }

  static String cardKeyForDay(int dayIndex) => 'day$dayIndex';

  /// Next stage unlocks as soon as the previous stage is completed (no calendar wait).
  bool isCardUnlocked(int dayIndex) {
    final i = journeyDays.indexOf(dayIndex);
    if (i <= 0) return true;
    return isCardCompleted(journeyDays[i - 1]);
  }

  bool isCardCompleted(int dayIndex) => completedCards.contains(cardKeyForDay(dayIndex));

  /// True when every journey stage has been marked complete.
  bool get isJourneyComplete => journeyDays.every(isCardCompleted);
}
