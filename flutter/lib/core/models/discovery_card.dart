class DiscoveryCard {
  const DiscoveryCard({
    required this.dayIndex,
    required this.title,
    required this.body,
    this.completed = false,
    this.locked = false,
  });

  /// Stage index in the journey (API keys: day0 / day2 / day4 / day6).
  final int dayIndex;
  final String title;
  final String body;
  final bool completed;
  final bool locked;

  String get cardKey => 'day$dayIndex';

  static const journeyDays = [0, 2, 4, 6];
}
